package coordinate

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPlanStorePersistsAtomicRevisionedUpdates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.json")
	plan := Plan{SchemaVersion: PlanSchemaVersion, Team: Team{ID: "team", Name: "Team", Members: []Member{{ID: "one", Name: "One"}}}}
	store, err := NewPlanStore(path, plan)
	if err != nil {
		t.Fatal(err)
	}
	next := store.Current()
	next.Team.Name = "Updated"
	updated, err := store.Update(next)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Revision != 1 {
		t.Fatalf("revision = %d", updated.Revision)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Update(next); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("error = %v", err)
	}
}
