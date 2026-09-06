package coordinate

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestLoadEntireSessionsNormalizesPublicMetadata(t *testing.T) {
	runner := func(_ context.Context, binary string, args ...string) ([]byte, error) {
		if binary != "entire-test" || strings.Join(args, " ") != "session list --json" {
			t.Fatalf("command = %s %v", binary, args)
		}
		return []byte(`[{"session_id":"session-1","agent":"Codex","model":"gpt-test","status":"active","worktree":"/repo","turns":3,"checkpoints":1,"tokens":{"total":42},"last_prompt":"private"}]`), nil
	}
	sessions, err := loadEntireSessions(t.Context(), "entire-test", runner)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].SessionID != "session-1" || sessions[0].Tokens.Total != 42 {
		t.Fatalf("sessions = %#v", sessions)
	}
	// Unknown fields such as last_prompt are intentionally not exposed.
}

func TestLoadEntireSessionsIsolatesProviderFailure(t *testing.T) {
	runner := func(context.Context, string, ...string) ([]byte, error) {
		return nil, errors.New("unavailable")
	}
	if _, err := loadEntireSessions(t.Context(), "entire-test", runner); err == nil {
		t.Fatal("provider failure was ignored")
	}
}
