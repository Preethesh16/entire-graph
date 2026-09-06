package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	runtime := &coordinateRuntime{data: coordinateData{snapshot: sem.ProviderSnapshot{}}}
	handler := coordinateHandler(store, runtime, nil)
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

func TestCoordinateHandlerRefreshesRepositoryEvidence(t *testing.T) {
	plan := coordinate.Plan{SchemaVersion: coordinate.PlanSchemaVersion, Team: coordinate.Team{ID: "team", Name: "Team"}}
	store, err := coordinate.NewPlanStore(filepath.Join(t.TempDir(), "plan.json"), plan)
	if err != nil {
		t.Fatal(err)
	}
	refreshes := 0
	runtime := &coordinateRuntime{
		data: coordinateData{snapshot: sem.ProviderSnapshot{}},
		refresh: func(context.Context) (coordinateData, error) {
			refreshes++
			return coordinateData{snapshot: sem.ProviderSnapshot{}}, nil
		},
	}
	handler := coordinateHandler(store, runtime, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || refreshes != 1 {
		t.Fatalf("refresh = %d calls=%d body=%s", response.Code, refreshes, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"schema_version":"spidey-sense/v1alpha1"`) {
		t.Fatalf("response = %s", response.Body.String())
	}
}

func TestCoordinateListenRejectsNonLoopback(t *testing.T) {
	err := serveCoordinate(t.Context(), Options{}, t.TempDir(), "0.0.0.0:4317", "plan.json", coordinate.Plan{}, sem.ProviderSnapshot{}, nil, coordinate.ProviderHealth{}, nil, coordinate.ProviderHealth{}, false, false, "")
	if err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("error = %v", err)
	}
}

func TestCoordinateNetworkEndpointsAuthenticateAndRejectPromptFields(t *testing.T) {
	plan := coordinate.Plan{SchemaVersion: coordinate.PlanSchemaVersion, Team: coordinate.Team{ID: "plan-team", Name: "Plan Team"}}
	store, err := coordinate.NewPlanStore(filepath.Join(t.TempDir(), "plan.json"), plan)
	if err != nil {
		t.Fatal(err)
	}
	network, err := coordinate.OpenNetworkStore(filepath.Join(t.TempDir(), "network.json"))
	if err != nil {
		t.Fatal(err)
	}
	runtime := &coordinateRuntime{data: coordinateData{snapshot: sem.ProviderSnapshot{}}}
	handler := coordinateHandler(store, runtime, network)

	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	createdResponse := call(http.MethodPost, "/api/v1/teams", "", `{"name":"Shared Team","leader_name":"Preethesh","leader_role":"Lead"}`)
	if createdResponse.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", createdResponse.Code, createdResponse.Body.String())
	}
	var created coordinate.TeamCredentials
	if err := json.Unmarshal(createdResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	joinResponse := call(http.MethodPost, "/api/v1/teams/"+created.TeamID+"/join", "", fmt.Sprintf(`{"invite_code":%q,"name":"Deepthi","role":"UI engineer"}`, created.InviteCode))
	if joinResponse.Code != http.StatusCreated {
		t.Fatalf("join = %d %s", joinResponse.Code, joinResponse.Body.String())
	}
	var joined coordinate.JoinCredentials
	if err := json.Unmarshal(joinResponse.Body.Bytes(), &joined); err != nil {
		t.Fatal(err)
	}
	connectBody := fmt.Sprintf(`{"team_id":%q,"member_id":%q,"session_id":"remote-session","agent":"Codex","provider":"Entire"}`, created.TeamID, joined.MemberID)
	connectResponse := call(http.MethodPost, "/api/v1/agents/connect", joined.ConnectorToken, connectBody)
	if connectResponse.Code != http.StatusCreated {
		t.Fatalf("connect = %d %s", connectResponse.Code, connectResponse.Body.String())
	}
	var connected coordinate.AgentCredentials
	if err := json.Unmarshal(connectResponse.Body.Bytes(), &connected); err != nil {
		t.Fatal(err)
	}
	privacyResponse := call(http.MethodPost, "/api/v1/agents/"+connected.AgentID+"/heartbeat", connected.AgentToken, `{"branch":"feature/ui","raw_prompt":"must not leave laptop"}`)
	if privacyResponse.Code != http.StatusBadRequest || !strings.Contains(privacyResponse.Body.String(), "unknown field") {
		t.Fatalf("privacy response = %d %s", privacyResponse.Code, privacyResponse.Body.String())
	}
	heartbeatResponse := call(http.MethodPost, "/api/v1/agents/"+connected.AgentID+"/heartbeat", connected.AgentToken, `{"mission_id":"connection-ui","branch":"feature/ui","changed_files":["web/spidey-sense/src/App.tsx"]}`)
	if heartbeatResponse.Code != http.StatusOK {
		t.Fatalf("heartbeat = %d %s", heartbeatResponse.Code, heartbeatResponse.Body.String())
	}
	if response := call(http.MethodGet, "/api/v1/teams/"+created.TeamID+"/agents", "wrong", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized list = %d", response.Code)
	}
	agentsResponse := call(http.MethodGet, "/api/v1/teams/"+created.TeamID+"/agents", created.AdminToken, "")
	if agentsResponse.Code != http.StatusOK || strings.Contains(agentsResponse.Body.String(), "prompt") || !strings.Contains(agentsResponse.Body.String(), "remote-session") {
		t.Fatalf("agents = %d %s", agentsResponse.Code, agentsResponse.Body.String())
	}
	viewerResponse := call(http.MethodGet, "/api/v1/teams/"+created.TeamID+"/agents", joined.ViewerToken, "")
	if viewerResponse.Code != http.StatusOK {
		t.Fatalf("viewer agents = %d %s", viewerResponse.Code, viewerResponse.Body.String())
	}

	server := httptest.NewServer(handler)
	defer server.Close()
	streamContext, cancelStream := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancelStream()
	streamRequest, err := http.NewRequestWithContext(streamContext, http.MethodGet, server.URL+"/api/v1/teams/"+created.TeamID+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	streamRequest.Header.Set("Authorization", "Bearer "+joined.ViewerToken)
	streamResponse, err := server.Client().Do(streamRequest)
	if err != nil {
		t.Fatal(err)
	}
	defer streamResponse.Body.Close()
	reader := bufio.NewReader(streamResponse.Body)
	ready, err := reader.ReadString('\n')
	if err != nil || ready != "event: ready\n" {
		t.Fatalf("SSE ready = %q, %v", ready, err)
	}
	if response := call(http.MethodPost, "/api/v1/agents/"+connected.AgentID+"/heartbeat", connected.AgentToken, `{"branch":"feature/ui","checkpoint_id":"checkpoint-2"}`); response.Code != http.StatusOK {
		t.Fatalf("second heartbeat = %d %s", response.Code, response.Body.String())
	}
	foundHeartbeat := false
	for !foundHeartbeat {
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			t.Fatalf("SSE heartbeat: %v", readErr)
		}
		foundHeartbeat = line == "event: agent.heartbeat\n"
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
