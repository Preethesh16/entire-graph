package coordinate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const maxSessionListBytes = 4 << 20

type Session struct {
	SessionID      string        `json:"session_id"`
	Agent          string        `json:"agent"`
	Model          string        `json:"model,omitempty"`
	Status         string        `json:"status"`
	Worktree       string        `json:"worktree,omitempty"`
	StartedAt      string        `json:"started_at,omitempty"`
	LastActive     string        `json:"last_active,omitempty"`
	Turns          int           `json:"turns,omitempty"`
	Checkpoints    int           `json:"checkpoints,omitempty"`
	LastCheckpoint string        `json:"last_checkpoint_id,omitempty"`
	Tokens         SessionTokens `json:"tokens,omitempty"`
}

type SessionTokens struct {
	Total int64 `json:"total,omitempty"`
}

type commandRunner func(context.Context, string, ...string) ([]byte, error)

func LoadEntireSessions(ctx context.Context, binary string) ([]Session, error) {
	if strings.TrimSpace(binary) == "" {
		binary = "entire"
	}
	return loadEntireSessions(ctx, binary, runCommand)
}

func loadEntireSessions(ctx context.Context, binary string, run commandRunner) ([]Session, error) {
	output, err := run(ctx, binary, "session", "list", "--json")
	if err != nil {
		return nil, fmt.Errorf("list Entire sessions: %w", err)
	}
	if len(output) > maxSessionListBytes {
		return nil, fmt.Errorf("Entire session output exceeds %d bytes", maxSessionListBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	var sessions []Session
	if err := decoder.Decode(&sessions); err != nil {
		return nil, fmt.Errorf("decode Entire sessions: %w", err)
	}
	return sessions, nil
}

func runCommand(ctx context.Context, binary string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, binary, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return nil, fmt.Errorf("%w: %s", err, detail)
		}
		return nil, err
	}
	return output, nil
}
