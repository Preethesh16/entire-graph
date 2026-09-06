package coordinate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Checkpoint struct {
	ID             string `json:"id"`
	Message        string `json:"message"`
	Date           string `json:"date,omitempty"`
	CondensationID string `json:"condensation_id,omitempty"`
	SessionID      string `json:"session_id"`
	LogsOnly       bool   `json:"is_logs_only,omitempty"`
}

func LoadEntireCheckpoints(ctx context.Context, binary string) ([]Checkpoint, error) {
	if binary == "" {
		binary = "entire"
	}
	output, err := runCommand(ctx, binary, "checkpoint", "list", "--pending", "--json")
	if err != nil {
		return nil, fmt.Errorf("list Entire checkpoints: %w", err)
	}
	if len(output) > maxSessionListBytes {
		return nil, fmt.Errorf("Entire checkpoint output exceeds %d bytes", maxSessionListBytes)
	}
	var checkpoints []Checkpoint
	if err := json.NewDecoder(bytes.NewReader(output)).Decode(&checkpoints); err != nil {
		return nil, fmt.Errorf("decode Entire checkpoints: %w", err)
	}
	return checkpoints, nil
}

func filterMappedCheckpoints(team Team, checkpoints []Checkpoint) []Checkpoint {
	allowed := map[string]bool{}
	for _, member := range team.Members {
		for _, id := range member.SessionIDs {
			allowed[id] = true
		}
	}
	filtered := make([]Checkpoint, 0)
	for _, checkpoint := range checkpoints {
		if allowed[checkpoint.SessionID] {
			filtered = append(filtered, checkpoint)
		}
	}
	return filtered
}
