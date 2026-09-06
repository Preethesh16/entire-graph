package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/entireio/entire-graph/internal/coordinate"
)

type connectAgentFlags struct {
	server            string
	teamID            string
	inviteCode        string
	name              string
	role              string
	missionID         string
	blocker           string
	repo              string
	interval          time.Duration
	once              bool
	allowInsecureHTTP bool
}

type connectorSession struct {
	SessionID  string `json:"session_id"`
	Agent      string `json:"agent"`
	Model      string `json:"model,omitempty"`
	Branch     string `json:"branch,omitempty"`
	LastActive string `json:"last_active,omitempty"`
}

type connectorCheckpoint struct {
	ID        string `json:"id"`
	Date      string `json:"date,omitempty"`
	SessionID string `json:"session_id"`
}

func runConnectAgent(ctx context.Context, opts Options, args []string) error {
	flags, err := parseConnectAgentFlags(args)
	if err != nil {
		return err
	}
	if err := validateConnectorServer(flags.server, flags.allowInsecureHTTP); err != nil {
		return err
	}
	flags.server = strings.TrimRight(flags.server, "/")
	repo, err := resolveRepo(ctx, opts.Env, flags.repo)
	if err != nil {
		return err
	}
	session, err := detectConnectorSession(ctx)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	var joined coordinate.JoinCredentials
	if err := connectorPOST(ctx, client, flags.server+"/api/v1/teams/"+url.PathEscape(flags.teamID)+"/join", "", coordinate.JoinTeamRequest{InviteCode: flags.inviteCode, Name: flags.name, Role: flags.role}, &joined); err != nil {
		return fmt.Errorf("join team: %w", err)
	}
	var connected coordinate.AgentCredentials
	if err := connectorPOST(ctx, client, flags.server+"/api/v1/agents/connect", joined.ConnectorToken, coordinate.ConnectAgentRequest{
		TeamID: joined.TeamID, MemberID: joined.MemberID, SessionID: session.SessionID,
		AgentType: session.Agent, Provider: "Entire", Model: session.Model,
	}, &connected); err != nil {
		return fmt.Errorf("connect agent: %w", err)
	}
	fmt.Fprintf(opts.Stdout, "Connected %s session %s as agent %s in team %s\n", session.Agent, session.SessionID, connected.AgentID, joined.TeamID)
	send := func() error {
		fresh, err := detectConnectorSession(ctx)
		if err != nil {
			return err
		}
		files, err := connectorChangedFiles(ctx, repo)
		if err != nil {
			return err
		}
		checkpoint, checkpointErr := latestConnectorCheckpoint(ctx, fresh.SessionID)
		if checkpointErr != nil {
			fmt.Fprintf(opts.Stderr, "Checkpoint metadata unavailable; heartbeat continues without it: %v\n", checkpointErr)
		}
		heartbeat := coordinate.HeartbeatRequest{
			MissionID: flags.missionID, Branch: fresh.Branch, ChangedFiles: files,
			CheckpointID: checkpoint.ID, Blocker: flags.blocker, LastActivity: fresh.LastActive,
		}
		var agent coordinate.ConnectedAgent
		if err := connectorPOST(ctx, client, flags.server+"/api/v1/agents/"+url.PathEscape(connected.AgentID)+"/heartbeat", connected.AgentToken, heartbeat, &agent); err != nil {
			return fmt.Errorf("send heartbeat: %w", err)
		}
		fmt.Fprintf(opts.Stdout, "Heartbeat sent at %s (branch=%s changed_files=%d checkpoint=%s)\n", agent.LastHeartbeat, agent.Branch, len(agent.ChangedFiles), agent.CheckpointID)
		return nil
	}
	if err := send(); err != nil {
		return err
	}
	if flags.once {
		return nil
	}
	ticker := time.NewTicker(flags.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := send(); err != nil {
				return err
			}
		}
	}
}

func parseConnectAgentFlags(args []string) (connectAgentFlags, error) {
	flags := connectAgentFlags{interval: 15 * time.Second}
	for index := 0; index < len(args); index++ {
		name := args[index]
		var value string
		var err error
		switch name {
		case "--server", "--team", "--invite", "--name", "--role", "--mission", "--blocker", "--repo", "--interval":
			value, err = connectAgentFlagValue(args, &index, name)
			if err != nil {
				return flags, err
			}
			switch name {
			case "--server":
				flags.server = value
			case "--team":
				flags.teamID = value
			case "--invite":
				flags.inviteCode = value
			case "--name":
				flags.name = value
			case "--role":
				flags.role = value
			case "--mission":
				flags.missionID = value
			case "--blocker":
				flags.blocker = value
			case "--repo":
				flags.repo = value
			case "--interval":
				flags.interval, err = time.ParseDuration(value)
				if err != nil {
					return flags, fmt.Errorf("invalid --interval: %w", err)
				}
			}
		case "--once":
			flags.once = true
		case "--allow-insecure-http":
			flags.allowInsecureHTTP = true
		default:
			return flags, fmt.Errorf("unknown connect-agent argument %q", name)
		}
	}
	for name, value := range map[string]string{"--server": flags.server, "--team": flags.teamID, "--invite": flags.inviteCode, "--name": flags.name} {
		if strings.TrimSpace(value) == "" {
			return flags, fmt.Errorf("connect-agent requires %s", name)
		}
	}
	if flags.interval < 5*time.Second {
		return flags, errors.New("connect-agent --interval must be at least 5s")
	}
	return flags, nil
}

func connectAgentFlagValue(args []string, index *int, name string) (string, error) {
	*index = *index + 1
	if *index >= len(args) {
		return "", fmt.Errorf("connect-agent %s requires a value", name)
	}
	return args[*index], nil
}

func validateConnectorServer(raw string, allowInsecure bool) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return errors.New("connect-agent --server must be an absolute URL")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return errors.New("connect-agent --server must not include a path")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme != "http" {
		return errors.New("connect-agent --server must use https or http")
	}
	host := parsed.Hostname()
	ip := net.ParseIP(host)
	if host == "localhost" || ip != nil && ip.IsLoopback() {
		return nil
	}
	if !allowInsecure {
		return errors.New("plain HTTP to a non-loopback server requires --allow-insecure-http (LAN use only)")
	}
	return nil
}

func detectConnectorSession(ctx context.Context) (connectorSession, error) {
	command := exec.CommandContext(ctx, "entire", "session", "current", "--json")
	output, err := command.Output()
	if err != nil {
		return connectorSession{}, fmt.Errorf("detect Entire session: %w", err)
	}
	var session connectorSession
	if err := json.Unmarshal(output, &session); err != nil {
		return connectorSession{}, fmt.Errorf("decode Entire session: %w", err)
	}
	if session.SessionID == "" || session.Agent == "" {
		return connectorSession{}, errors.New("Entire session current did not return a session_id and agent")
	}
	return session, nil
}

func latestConnectorCheckpoint(ctx context.Context, sessionID string) (connectorCheckpoint, error) {
	command := exec.CommandContext(ctx, "entire", "checkpoint", "list", "--pending", "--json")
	output, err := command.Output()
	if err != nil {
		return connectorCheckpoint{}, fmt.Errorf("list Entire checkpoints: %w", err)
	}
	var checkpoints []connectorCheckpoint
	if err := json.Unmarshal(output, &checkpoints); err != nil {
		return connectorCheckpoint{}, fmt.Errorf("decode Entire checkpoints: %w", err)
	}
	var matches []connectorCheckpoint
	for _, checkpoint := range checkpoints {
		if checkpoint.SessionID == sessionID {
			matches = append(matches, checkpoint)
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Date > matches[j].Date })
	if len(matches) == 0 {
		return connectorCheckpoint{}, nil
	}
	return matches[0], nil
}

func connectorChangedFiles(ctx context.Context, repo string) ([]string, error) {
	command := exec.CommandContext(ctx, "git", "-C", repo, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read changed files: %w", err)
	}
	records := bytes.Split(output, []byte{0})
	files := map[string]bool{}
	for index := 0; index < len(records); index++ {
		record := records[index]
		if len(record) < 4 {
			continue
		}
		path := string(record[3:])
		files[filepathSlash(path)] = true
		if (record[0] == 'R' || record[0] == 'C' || record[1] == 'R' || record[1] == 'C') && index+1 < len(records) {
			index++
		}
	}
	result := make([]string, 0, len(files))
	for path := range files {
		result = append(result, path)
	}
	sort.Strings(result)
	if len(result) > 128 {
		result = result[:128]
	}
	return result, nil
}

func filepathSlash(path string) string { return strings.ReplaceAll(path, "\\", "/") }

func connectorPOST(ctx context.Context, client *http.Client, endpoint, token string, input, output any) error {
	content, err := json.Marshal(input)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(content))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("server returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
