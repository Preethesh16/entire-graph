package coordinate

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCheckpointLoadingOmitsPromptAndFiltersByTeamSession(t *testing.T) {
	runner := commandRunner(func(context.Context, string, ...string) ([]byte, error) {
		return []byte(`[{"id":"commit","message":"architecture","session_id":"mapped","session_prompt":"sensitive"},{"id":"other","message":"other","session_id":"unmapped"}]`), nil
	})
	output, err := runner(t.Context(), "entire", "checkpoint", "list", "--pending", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var checkpoints []Checkpoint
	if err := json.Unmarshal(output, &checkpoints); err != nil {
		t.Fatal(err)
	}
	team := Team{Members: []Member{{SessionIDs: []string{"mapped"}}}}
	filtered := filterMappedCheckpoints(team, checkpoints)
	if len(filtered) != 1 || filtered[0].ID != "commit" {
		t.Fatalf("checkpoints = %#v", filtered)
	}
}
