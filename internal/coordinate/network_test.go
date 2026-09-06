package coordinate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNetworkStoreCrossMachineLifecyclePersistsHashedCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network.json")
	store, err := OpenNetworkStore(path)
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.CreateTeam(CreateTeamRequest{Name: "Web Guard", Objective: "Coordinate safely", LeaderName: "Preethesh", LeaderRole: "Lead"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.JoinTeam(created.TeamID, JoinTeamRequest{InviteCode: "wrong", Name: "Deepthi"}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("wrong invite error = %v", err)
	}
	joined, err := store.JoinTeam(created.TeamID, JoinTeamRequest{InviteCode: created.InviteCode, Name: "Deepthi", Role: "UI engineer"})
	if err != nil {
		t.Fatal(err)
	}
	connected, err := store.ConnectAgent(joined.ConnectorToken, ConnectAgentRequest{TeamID: created.TeamID, MemberID: joined.MemberID, SessionID: "session-remote", AgentType: "Codex", Provider: "Entire", Model: "gpt-test"})
	if err != nil {
		t.Fatal(err)
	}
	events, cancel, err := store.Subscribe(created.TeamID, created.AdminToken)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()
	agent, err := store.Heartbeat(connected.AgentID, connected.AgentToken, HeartbeatRequest{MissionID: "connection-ui", Branch: "feature/ui", ChangedFiles: []string{"web/spidey-sense/src/App.tsx"}, CheckpointID: "abc123", Blocker: "waiting for API", LastActivity: "2026-09-06T12:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	if !agent.Online || agent.Name != "Deepthi" || agent.Role != "UI engineer" || agent.AgentType != "Codex" || agent.Provider != "Entire" {
		t.Fatalf("agent = %#v", agent)
	}
	select {
	case event := <-events:
		if event.Type != "agent.heartbeat" || event.AgentID != connected.AgentID {
			t.Fatalf("event = %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("heartbeat event not published")
	}
	if _, err := store.Agents(created.TeamID, "wrong"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("wrong admin token error = %v", err)
	}

	reopened, err := OpenNetworkStore(path)
	if err != nil {
		t.Fatal(err)
	}
	agents, err := reopened.Agents(created.TeamID, created.AdminToken)
	if err != nil || len(agents) != 1 || agents[0].SessionID != "session-remote" {
		t.Fatalf("reopened agents = %#v, err = %v", agents, err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{created.InviteCode, created.AdminToken, created.ConnectorToken, joined.ConnectorToken, connected.AgentToken} {
		if strings.Contains(string(content), secret) {
			t.Fatalf("plaintext credential persisted: %s", secret)
		}
	}
}

func TestNetworkStoreBoundsHeartbeatAndExpiresOnlineStatus(t *testing.T) {
	store, err := OpenNetworkStore(filepath.Join(t.TempDir(), "network.json"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	created, err := store.CreateTeam(CreateTeamRequest{Name: "Team", LeaderName: "Lead"})
	if err != nil {
		t.Fatal(err)
	}
	connected, err := store.ConnectAgent(created.ConnectorToken, ConnectAgentRequest{TeamID: created.TeamID, MemberID: created.MemberID, SessionID: "session", AgentType: "Claude Code", Provider: "Entire"})
	if err != nil {
		t.Fatal(err)
	}
	tooMany := make([]string, maxChangedFiles+1)
	for index := range tooMany {
		tooMany[index] = "file.go"
	}
	if _, err := store.Heartbeat(connected.AgentID, connected.AgentToken, HeartbeatRequest{ChangedFiles: tooMany}); err == nil {
		t.Fatal("oversized changed_files accepted")
	}
	if _, err := store.Heartbeat(connected.AgentID, connected.AgentToken, HeartbeatRequest{ChangedFiles: []string{"../secret"}}); err == nil {
		t.Fatal("traversal path accepted")
	}
	now = now.Add(AgentOnlineWindow + time.Second)
	agents, err := store.Agents(created.TeamID, created.AdminToken)
	if err != nil {
		t.Fatal(err)
	}
	if agents[0].Online {
		t.Fatal("stale agent reported online")
	}
}

func TestNetworkStoreHeartbeatRollsBackWhenPersistenceFails(t *testing.T) {
	store, err := OpenNetworkStore(filepath.Join(t.TempDir(), "network.json"))
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.CreateTeam(CreateTeamRequest{Name: "Team", LeaderName: "Lead"})
	if err != nil {
		t.Fatal(err)
	}
	connected, err := store.ConnectAgent(created.ConnectorToken, ConnectAgentRequest{TeamID: created.TeamID, MemberID: created.MemberID, SessionID: "session", AgentType: "Codex", Provider: "Entire"})
	if err != nil {
		t.Fatal(err)
	}
	store.path = t.TempDir() // Renaming an atomic state file over a directory must fail.
	if _, err := store.Heartbeat(connected.AgentID, connected.AgentToken, HeartbeatRequest{Branch: "must-not-stick"}); err == nil {
		t.Fatal("heartbeat unexpectedly persisted")
	}
	agents, err := store.Agents(created.TeamID, created.AdminToken)
	if err != nil {
		t.Fatal(err)
	}
	if agents[0].Branch != "" {
		t.Fatalf("failed heartbeat changed in-memory state: %#v", agents[0])
	}
}
