package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/entireio/entire-graph/internal/coordinate"
	"github.com/entireio/entire-graph/internal/sem"
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

func TestCoordinateHandlerServesVersionedReport(t *testing.T) {
	plan := coordinate.Plan{SchemaVersion: coordinate.PlanSchemaVersion, Team: coordinate.Team{ID: "team", Name: "Team"}}
	store, err := coordinate.NewPlanStore(filepath.Join(t.TempDir(), "plan.json"), plan)
	if err != nil {
		t.Fatal(err)
	}
	handler := coordinateHandler(store, sem.ProviderSnapshot{}, nil, coordinate.ProviderHealth{}, nil, coordinate.ProviderHealth{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/report", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"schema_version":"spidey-sense/v1alpha1"`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("cache control = %q", response.Header().Get("Cache-Control"))
	}
}

func TestCoordinateListenRejectsNonLoopback(t *testing.T) {
	err := serveCoordinate(t.Context(), Options{}, "0.0.0.0:4317", "plan.json", coordinate.Plan{}, sem.ProviderSnapshot{}, nil, coordinate.ProviderHealth{}, nil, coordinate.ProviderHealth{})
	if err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("error = %v", err)
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
