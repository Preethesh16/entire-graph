package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoordinateCommandReportsDynamicSameFileCollision(t *testing.T) {
	repo := t.TempDir()
	git(t, repo, "init")
	write(t, repo, "alpha.go", "package alpha\n\nfunc Alpha() {}\n")
	planPath := filepath.Join(repo, "plan.json")
	plan := `{
  "schema_version": "1.0",
  "team": {
    "id": "dynamic-team",
    "name": "Dynamic Team",
    "members": [
      {"id": "member-1", "name": "First"},
      {"id": "member-2", "name": "Second"}
    ]
  },
  "missions": [
    {"id": "mission-1", "title": "First mission", "owner": "member-1", "status": "active", "targets": [{"file": "alpha.go"}]},
    {"id": "mission-2", "title": "Second mission", "owner": "member-2", "status": "active", "targets": [{"file": "alpha.go"}]}
  ]
}`
	if err := os.WriteFile(planPath, []byte(plan), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	err := Run(t.Context(), Options{Version: "test", Env: EntireEnv{RepoRoot: repo}, Stdout: &stdout}, []string{
		"coordinate", "--repo", repo, "--plan", planPath, "--format", "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Spidey Sense: Dynamic Team", "1 BLOCK", "SAME_FILE", "mission-1 <-> mission-2"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("output missing %q:\n%s", want, stdout.String())
		}
	}
}

func TestCoordinateFlagsRequirePlanAndKnownFormat(t *testing.T) {
	if _, err := parseCoordinateFlags(nil); err == nil {
		t.Fatal("missing plan was accepted")
	}
	if _, err := parseCoordinateFlags([]string{"--plan", "plan.json", "--format", "yaml"}); err == nil {
		t.Fatal("unknown format was accepted")
	}
}
