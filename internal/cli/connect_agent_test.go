package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/entireio/entire-graph/internal/coordinate"
)

func TestConnectAgentFlagsRequireExplicitInsecureLANOptIn(t *testing.T) {
	if err := validateConnectorServer("http://192.168.1.20:4317", false); err == nil || !strings.Contains(err.Error(), "--allow-insecure-http") {
		t.Fatalf("LAN HTTP error = %v", err)
	}
	if err := validateConnectorServer("http://192.168.1.20:4317", true); err != nil {
		t.Fatal(err)
	}
	if err := validateConnectorServer("https://spidey.example", false); err != nil {
		t.Fatal(err)
	}
	if _, err := parseConnectAgentFlags([]string{"--server", "https://spidey.example", "--team", "team", "--invite", "invite", "--name"}); err == nil {
		t.Fatal("missing --name value accepted")
	}
}

func TestConnectorSessionAllowlistDropsPrivateEntireFields(t *testing.T) {
	raw := `{"session_id":"session","agent":"Codex","model":"gpt-test","branch":"feature","last_active":"2026-09-06T12:00:00Z","last_prompt":"secret prompt","reasoning":"secret reasoning","terminal":"secret terminal"}`
	var session connectorSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(coordinate.ConnectAgentRequest{TeamID: "team", MemberID: "member", SessionID: session.SessionID, AgentType: session.Agent, Provider: "Entire", Model: session.Model})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"secret prompt", "secret reasoning", "secret terminal", "last_prompt", "reasoning", "terminal"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("private field leaked into connector payload: %s", payload)
		}
	}
}
