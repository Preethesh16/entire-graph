package coordinate

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	NetworkSchemaVersion = "spidey-sense/network-v1alpha1"
	AgentOnlineWindow    = 45 * time.Second
	maxChangedFiles      = 128
	maxNetworkFileBytes  = 4 << 20
)

var (
	ErrTeamNotFound  = errors.New("team not found")
	ErrAgentNotFound = errors.New("agent not found")
	ErrUnauthorized  = errors.New("unauthorized")
)

type CreateTeamRequest struct {
	Name       string `json:"name"`
	Objective  string `json:"objective,omitempty"`
	LeaderName string `json:"leader_name"`
	LeaderRole string `json:"leader_role,omitempty"`
}

type TeamCredentials struct {
	TeamID         string `json:"team_id"`
	InviteCode     string `json:"invite_code"`
	AdminToken     string `json:"admin_token"`
	MemberID       string `json:"member_id"`
	ConnectorToken string `json:"connector_token"`
	ViewerToken    string `json:"viewer_token"`
}

type JoinTeamRequest struct {
	InviteCode string `json:"invite_code"`
	Name       string `json:"name"`
	Role       string `json:"role,omitempty"`
}

type JoinCredentials struct {
	TeamID         string `json:"team_id"`
	MemberID       string `json:"member_id"`
	ConnectorToken string `json:"connector_token"`
	ViewerToken    string `json:"viewer_token"`
}

type ConnectAgentRequest struct {
	TeamID    string `json:"team_id"`
	MemberID  string `json:"member_id"`
	SessionID string `json:"session_id"`
	AgentType string `json:"agent"`
	Provider  string `json:"provider"`
	Model     string `json:"model,omitempty"`
}

type AgentCredentials struct {
	AgentID    string `json:"agent_id"`
	AgentToken string `json:"agent_token"`
}

type HeartbeatRequest struct {
	MissionID    string   `json:"mission_id,omitempty"`
	Branch       string   `json:"branch,omitempty"`
	ChangedFiles []string `json:"changed_files,omitempty"`
	CheckpointID string   `json:"checkpoint_id,omitempty"`
	Blocker      string   `json:"blocker,omitempty"`
	LastActivity string   `json:"last_activity,omitempty"`
}

type ConnectedAgent struct {
	ID            string   `json:"id"`
	TeamID        string   `json:"team_id"`
	MemberID      string   `json:"member_id"`
	Name          string   `json:"name"`
	Role          string   `json:"role,omitempty"`
	SessionID     string   `json:"session_id"`
	AgentType     string   `json:"agent"`
	Provider      string   `json:"provider"`
	Model         string   `json:"model,omitempty"`
	Online        bool     `json:"online"`
	MissionID     string   `json:"mission_id,omitempty"`
	Branch        string   `json:"branch,omitempty"`
	ChangedFiles  []string `json:"changed_files,omitempty"`
	CheckpointID  string   `json:"checkpoint_id,omitempty"`
	Blocker       string   `json:"blocker,omitempty"`
	LastActivity  string   `json:"last_activity,omitempty"`
	LastHeartbeat string   `json:"last_heartbeat"`
}

type TeamEvent struct {
	ID      uint64          `json:"id"`
	Type    string          `json:"type"`
	TeamID  string          `json:"team_id"`
	AgentID string          `json:"agent_id,omitempty"`
	At      string          `json:"at"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type networkState struct {
	SchemaVersion string                 `json:"schema_version"`
	Teams         map[string]*teamRecord `json:"teams"`
}

type teamRecord struct {
	ID         string                    `json:"id"`
	Name       string                    `json:"name"`
	Objective  string                    `json:"objective,omitempty"`
	InviteHash string                    `json:"invite_hash"`
	AdminHash  string                    `json:"admin_hash"`
	Members    map[string]*networkMember `json:"members"`
	Agents     map[string]*agentRecord   `json:"agents"`
	CreatedAt  string                    `json:"created_at"`
}

type networkMember struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Role          string `json:"role,omitempty"`
	ConnectorHash string `json:"connector_hash"`
	ViewerHash    string `json:"viewer_hash"`
}

type agentRecord struct {
	ID            string   `json:"id"`
	MemberID      string   `json:"member_id"`
	SessionID     string   `json:"session_id"`
	AgentType     string   `json:"agent"`
	Provider      string   `json:"provider"`
	Model         string   `json:"model,omitempty"`
	TokenHash     string   `json:"token_hash"`
	MissionID     string   `json:"mission_id,omitempty"`
	Branch        string   `json:"branch,omitempty"`
	ChangedFiles  []string `json:"changed_files,omitempty"`
	CheckpointID  string   `json:"checkpoint_id,omitempty"`
	Blocker       string   `json:"blocker,omitempty"`
	LastActivity  string   `json:"last_activity,omitempty"`
	LastHeartbeat string   `json:"last_heartbeat"`
}

type NetworkStore struct {
	mu          sync.RWMutex
	path        string
	state       networkState
	now         func() time.Time
	eventID     uint64
	subscribers map[string]map[chan TeamEvent]bool
}

func OpenNetworkStore(path string) (*NetworkStore, error) {
	store := &NetworkStore{
		path: path, now: time.Now,
		state:       networkState{SchemaVersion: NetworkSchemaVersion, Teams: map[string]*teamRecord{}},
		subscribers: map[string]map[chan TeamEvent]bool{},
	}
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read network state: %w", err)
	}
	if len(content) > maxNetworkFileBytes {
		return nil, fmt.Errorf("network state exceeds %d bytes", maxNetworkFileBytes)
	}
	if err := json.Unmarshal(content, &store.state); err != nil {
		return nil, fmt.Errorf("decode network state: %w", err)
	}
	if store.state.SchemaVersion != NetworkSchemaVersion {
		return nil, fmt.Errorf("unsupported network schema %q", store.state.SchemaVersion)
	}
	if store.state.Teams == nil {
		store.state.Teams = map[string]*teamRecord{}
	}
	return store, nil
}

func (store *NetworkStore) CreateTeam(request CreateTeamRequest) (TeamCredentials, error) {
	if err := validateText("team name", request.Name, 1, 100); err != nil {
		return TeamCredentials{}, err
	}
	if err := validateText("leader name", request.LeaderName, 1, 100); err != nil {
		return TeamCredentials{}, err
	}
	if err := validateText("objective", request.Objective, 0, 500); err != nil {
		return TeamCredentials{}, err
	}
	if err := validateText("leader role", request.LeaderRole, 0, 100); err != nil {
		return TeamCredentials{}, err
	}
	teamID, err := randomValue("team_", 12)
	if err != nil {
		return TeamCredentials{}, err
	}
	memberID, err := randomValue("member_", 12)
	if err != nil {
		return TeamCredentials{}, err
	}
	invite, err := randomValue("invite_", 12)
	if err != nil {
		return TeamCredentials{}, err
	}
	admin, err := randomValue("admin_", 32)
	if err != nil {
		return TeamCredentials{}, err
	}
	connector, err := randomValue("connector_", 32)
	if err != nil {
		return TeamCredentials{}, err
	}
	viewer, err := randomValue("viewer_", 32)
	if err != nil {
		return TeamCredentials{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.state.Teams[teamID] = &teamRecord{
		ID: teamID, Name: strings.TrimSpace(request.Name), Objective: strings.TrimSpace(request.Objective),
		InviteHash: tokenHash(invite), AdminHash: tokenHash(admin), CreatedAt: store.now().UTC().Format(time.RFC3339Nano),
		Members: map[string]*networkMember{memberID: {ID: memberID, Name: strings.TrimSpace(request.LeaderName), Role: strings.TrimSpace(request.LeaderRole), ConnectorHash: tokenHash(connector), ViewerHash: tokenHash(viewer)}},
		Agents:  map[string]*agentRecord{},
	}
	if err := store.persistLocked(); err != nil {
		delete(store.state.Teams, teamID)
		return TeamCredentials{}, err
	}
	return TeamCredentials{TeamID: teamID, InviteCode: invite, AdminToken: admin, MemberID: memberID, ConnectorToken: connector, ViewerToken: viewer}, nil
}

func (store *NetworkStore) JoinTeam(teamID string, request JoinTeamRequest) (JoinCredentials, error) {
	if err := validateText("name", request.Name, 1, 100); err != nil {
		return JoinCredentials{}, err
	}
	if err := validateText("role", request.Role, 0, 100); err != nil {
		return JoinCredentials{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	team, ok := store.state.Teams[teamID]
	if !ok {
		return JoinCredentials{}, ErrTeamNotFound
	}
	if !tokenMatches(request.InviteCode, team.InviteHash) {
		return JoinCredentials{}, ErrUnauthorized
	}
	memberID, err := randomValue("member_", 12)
	if err != nil {
		return JoinCredentials{}, err
	}
	connector, err := randomValue("connector_", 32)
	if err != nil {
		return JoinCredentials{}, err
	}
	viewer, err := randomValue("viewer_", 32)
	if err != nil {
		return JoinCredentials{}, err
	}
	team.Members[memberID] = &networkMember{ID: memberID, Name: strings.TrimSpace(request.Name), Role: strings.TrimSpace(request.Role), ConnectorHash: tokenHash(connector), ViewerHash: tokenHash(viewer)}
	if err := store.persistLocked(); err != nil {
		delete(team.Members, memberID)
		return JoinCredentials{}, err
	}
	return JoinCredentials{TeamID: teamID, MemberID: memberID, ConnectorToken: connector, ViewerToken: viewer}, nil
}

func (store *NetworkStore) ConnectAgent(token string, request ConnectAgentRequest) (AgentCredentials, error) {
	for name, value := range map[string]string{"team_id": request.TeamID, "member_id": request.MemberID, "session_id": request.SessionID, "agent": request.AgentType, "provider": request.Provider} {
		if err := validateText(name, value, 1, 200); err != nil {
			return AgentCredentials{}, err
		}
	}
	if err := validateText("model", request.Model, 0, 200); err != nil {
		return AgentCredentials{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	team, ok := store.state.Teams[request.TeamID]
	if !ok {
		return AgentCredentials{}, ErrTeamNotFound
	}
	member, ok := team.Members[request.MemberID]
	if !ok || !tokenMatches(token, member.ConnectorHash) {
		return AgentCredentials{}, ErrUnauthorized
	}
	agentID, err := randomValue("agent_", 12)
	if err != nil {
		return AgentCredentials{}, err
	}
	agentToken, err := randomValue("agent_token_", 32)
	if err != nil {
		return AgentCredentials{}, err
	}
	now := store.now().UTC().Format(time.RFC3339Nano)
	team.Agents[agentID] = &agentRecord{ID: agentID, MemberID: member.ID, SessionID: strings.TrimSpace(request.SessionID), AgentType: strings.TrimSpace(request.AgentType), Provider: strings.TrimSpace(request.Provider), Model: strings.TrimSpace(request.Model), TokenHash: tokenHash(agentToken), LastHeartbeat: now, LastActivity: now}
	if err := store.persistLocked(); err != nil {
		delete(team.Agents, agentID)
		return AgentCredentials{}, err
	}
	store.publishLocked(request.TeamID, "agent.connected", agentID, store.publicAgentLocked(team, team.Agents[agentID]))
	return AgentCredentials{AgentID: agentID, AgentToken: agentToken}, nil
}

func (store *NetworkStore) Heartbeat(agentID, token string, request HeartbeatRequest) (ConnectedAgent, error) {
	if err := validateHeartbeat(request); err != nil {
		return ConnectedAgent{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	team, record := store.findAgentLocked(agentID)
	if record == nil {
		return ConnectedAgent{}, ErrAgentNotFound
	}
	if !tokenMatches(token, record.TokenHash) {
		return ConnectedAgent{}, ErrUnauthorized
	}
	previous := *record
	previous.ChangedFiles = append([]string(nil), record.ChangedFiles...)
	record.MissionID = strings.TrimSpace(request.MissionID)
	record.Branch = strings.TrimSpace(request.Branch)
	record.ChangedFiles = append([]string(nil), request.ChangedFiles...)
	record.CheckpointID = strings.TrimSpace(request.CheckpointID)
	record.Blocker = strings.TrimSpace(request.Blocker)
	if strings.TrimSpace(request.LastActivity) != "" {
		record.LastActivity = request.LastActivity
	}
	record.LastHeartbeat = store.now().UTC().Format(time.RFC3339Nano)
	if err := store.persistLocked(); err != nil {
		*record = previous
		return ConnectedAgent{}, err
	}
	agent := store.publicAgentLocked(team, record)
	store.publishLocked(team.ID, "agent.heartbeat", agentID, agent)
	return agent, nil
}

func (store *NetworkStore) Agents(teamID, token string) ([]ConnectedAgent, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	team, ok := store.state.Teams[teamID]
	if !ok {
		return nil, ErrTeamNotFound
	}
	if !teamCanRead(team, token) {
		return nil, ErrUnauthorized
	}
	agents := make([]ConnectedAgent, 0, len(team.Agents))
	for _, record := range team.Agents {
		agents = append(agents, store.publicAgentLocked(team, record))
	}
	sort.Slice(agents, func(i, j int) bool { return agents[i].ID < agents[j].ID })
	return agents, nil
}

func (store *NetworkStore) Subscribe(teamID, token string) (<-chan TeamEvent, func(), error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	team, ok := store.state.Teams[teamID]
	if !ok {
		return nil, nil, ErrTeamNotFound
	}
	if !teamCanRead(team, token) {
		return nil, nil, ErrUnauthorized
	}
	channel := make(chan TeamEvent, 16)
	if store.subscribers[teamID] == nil {
		store.subscribers[teamID] = map[chan TeamEvent]bool{}
	}
	store.subscribers[teamID][channel] = true
	cancel := func() {
		store.mu.Lock()
		defer store.mu.Unlock()
		if store.subscribers[teamID][channel] {
			delete(store.subscribers[teamID], channel)
			close(channel)
		}
	}
	return channel, cancel, nil
}

func teamCanRead(team *teamRecord, token string) bool {
	if tokenMatches(token, team.AdminHash) {
		return true
	}
	for _, member := range team.Members {
		if tokenMatches(token, member.ViewerHash) {
			return true
		}
	}
	return false
}

func (store *NetworkStore) findAgentLocked(agentID string) (*teamRecord, *agentRecord) {
	for _, team := range store.state.Teams {
		if agent := team.Agents[agentID]; agent != nil {
			return team, agent
		}
	}
	return nil, nil
}

func (store *NetworkStore) publicAgentLocked(team *teamRecord, record *agentRecord) ConnectedAgent {
	member := team.Members[record.MemberID]
	lastHeartbeat, _ := time.Parse(time.RFC3339Nano, record.LastHeartbeat)
	return ConnectedAgent{ID: record.ID, TeamID: team.ID, MemberID: record.MemberID, Name: member.Name, Role: member.Role, SessionID: record.SessionID, AgentType: record.AgentType, Provider: record.Provider, Model: record.Model, Online: store.now().Sub(lastHeartbeat) <= AgentOnlineWindow, MissionID: record.MissionID, Branch: record.Branch, ChangedFiles: append([]string(nil), record.ChangedFiles...), CheckpointID: record.CheckpointID, Blocker: record.Blocker, LastActivity: record.LastActivity, LastHeartbeat: record.LastHeartbeat}
}

func (store *NetworkStore) publishLocked(teamID, eventType, agentID string, value any) {
	store.eventID++
	data, _ := json.Marshal(value)
	event := TeamEvent{ID: store.eventID, Type: eventType, TeamID: teamID, AgentID: agentID, At: store.now().UTC().Format(time.RFC3339Nano), Data: data}
	for channel := range store.subscribers[teamID] {
		select {
		case channel <- event:
		default:
		}
	}
}

func (store *NetworkStore) persistLocked() error {
	content, err := json.MarshalIndent(store.state, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	if len(content) > maxNetworkFileBytes {
		return fmt.Errorf("network state exceeds %d bytes", maxNetworkFileBytes)
	}
	directory := filepath.Dir(store.path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create network state directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".spidey-network-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, store.path); err != nil {
		return fmt.Errorf("replace network state: %w", err)
	}
	return nil
}

func validateHeartbeat(request HeartbeatRequest) error {
	for name, value := range map[string]string{"mission_id": request.MissionID, "branch": request.Branch, "checkpoint_id": request.CheckpointID, "blocker": request.Blocker, "last_activity": request.LastActivity} {
		limit := 200
		if name == "blocker" {
			limit = 500
		}
		if err := validateText(name, value, 0, limit); err != nil {
			return err
		}
	}
	if len(request.ChangedFiles) > maxChangedFiles {
		return fmt.Errorf("changed_files exceeds %d entries", maxChangedFiles)
	}
	for _, path := range request.ChangedFiles {
		if err := validateText("changed file", path, 1, 500); err != nil {
			return err
		}
		clean := filepath.ToSlash(filepath.Clean(path))
		if filepath.IsAbs(path) || clean == ".." || strings.HasPrefix(clean, "../") {
			return errors.New("changed files must be repository-relative")
		}
	}
	if request.LastActivity != "" {
		if _, err := time.Parse(time.RFC3339Nano, request.LastActivity); err != nil {
			return errors.New("last_activity must be RFC3339")
		}
	}
	return nil
}

func validateText(name, value string, minimum, maximum int) error {
	value = strings.TrimSpace(value)
	if len(value) < minimum {
		return fmt.Errorf("%s is required", name)
	}
	if len(value) > maximum {
		return fmt.Errorf("%s exceeds %d bytes", name, maximum)
	}
	if strings.ContainsAny(value, "\x00\r\n") {
		return fmt.Errorf("%s contains control characters", name)
	}
	return nil
}

func randomValue(prefix string, bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate credential: %w", err)
	}
	return prefix + base64.RawURLEncoding.EncodeToString(buffer), nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func tokenMatches(token, expectedHash string) bool {
	actual, err := hex.DecodeString(tokenHash(token))
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(expectedHash)
	if err != nil || len(actual) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare(actual, expected) == 1
}
