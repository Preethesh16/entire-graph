package coordinate

import (
	"testing"

	"github.com/entireio/entire-graph/internal/sem"
)

func TestAnalyzeDynamicMissionsFindsDirectCallRisk(t *testing.T) {
	snapshot := testSnapshot()
	plan := Plan{
		SchemaVersion: PlanSchemaVersion,
		Team: Team{ID: "web-team", Name: "Web Team", Members: []Member{
			{ID: "alice", Name: "Alice"}, {ID: "bob", Name: "Bob"},
		}},
		Missions: []Mission{
			{ID: "cli", Title: "Change CLI", Owner: "alice", Status: "active", Targets: []Target{{Symbol: "runCheckpoint"}}},
			{ID: "analysis", Title: "Change analyzer", Owner: "bob", Status: "active", Targets: []Target{{Symbol: "AnalyzeCheckpoint"}}},
		},
	}
	report, err := Analyze(plan, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Block != 1 || len(report.Decisions) != 1 {
		t.Fatalf("summary = %#v, decisions = %#v", report.Summary, report.Decisions)
	}
	decision := report.Decisions[0]
	if decision.Level != "BLOCK" || len(decision.Path) != 1 || decision.Path[0].Relation != "CALLS" {
		t.Fatalf("decision = %#v", decision)
	}
	if len(decision.TestTargets) != 1 || decision.TestTargets[0] != "internal/sem/analyze_test.go" {
		t.Fatalf("test targets = %#v", decision.TestTargets)
	}
}

func TestAnalyzeSameFileAndClearDecisions(t *testing.T) {
	snapshot := testSnapshot()
	plan := Plan{
		SchemaVersion: PlanSchemaVersion,
		Team: Team{ID: "team", Name: "Team", Members: []Member{
			{ID: "one", Name: "One"}, {ID: "two", Name: "Two"}, {ID: "three", Name: "Three"},
		}},
		Missions: []Mission{
			{ID: "a", Title: "A", Owner: "one", Status: "active", Targets: []Target{{File: "internal/cli/root.go"}}},
			{ID: "b", Title: "B", Owner: "two", Status: "queued", Targets: []Target{{File: "internal/cli/root.go"}}},
			{ID: "c", Title: "C", Owner: "three", Status: "active", Targets: []Target{{File: "README.md"}}},
		},
	}
	report, err := Analyze(plan, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Pairs != 3 || report.Summary.Block != 1 || report.Summary.Clear != 2 {
		t.Fatalf("summary = %#v", report.Summary)
	}
	if report.Decisions[0].Path[0].Relation != "SAME_FILE" {
		t.Fatalf("same-file decision = %#v", report.Decisions[0])
	}
}

func TestAnalyzeTwoHopConnectionRequiresReview(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Symbols = append(snapshot.Symbols, sem.SymbolRecord{ID: "middle", Name: "middle", QualifiedName: "middle", Kind: "function", FilePath: "middle.go", StartLine: 1, EndLine: 3})
	snapshot.Files = append(snapshot.Files, sem.FileRecord{ID: "file-middle", Path: "middle.go"})
	snapshot.Relations = []sem.RelationRecord{
		{FromID: "run", ToID: "middle", Type: "CALLS", Confidence: 1, Resolution: "exact"},
		{FromID: "middle", ToID: "analyze", Type: "CALLS", Confidence: 1, Resolution: "exact"},
	}
	plan := Plan{
		SchemaVersion: PlanSchemaVersion,
		Team: Team{ID: "team", Name: "Team", Members: []Member{
			{ID: "one", Name: "One"}, {ID: "two", Name: "Two"},
		}},
		Missions: []Mission{
			{ID: "a", Title: "A", Owner: "one", Status: "active", Targets: []Target{{Symbol: "runCheckpoint"}}},
			{ID: "b", Title: "B", Owner: "two", Status: "active", Targets: []Target{{Symbol: "AnalyzeCheckpoint"}}},
		},
	}
	report, err := Analyze(plan, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	decision := report.Decisions[0]
	if decision.Level != "REVIEW" || len(decision.Path) != 2 {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestAnalyzeReportsAmbiguousAndMissingTargets(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Symbols = append(snapshot.Symbols, sem.SymbolRecord{ID: "duplicate", Name: "runCheckpoint", QualifiedName: "other.runCheckpoint", FilePath: "other.go", StartLine: 1, EndLine: 2})
	plan := Plan{
		SchemaVersion: PlanSchemaVersion,
		Team:          Team{ID: "team", Name: "Team", Members: []Member{{ID: "owner", Name: "Owner"}}},
		Missions: []Mission{{ID: "mission", Title: "Mission", Owner: "owner", Status: "active", Targets: []Target{
			{Symbol: "runCheckpoint"}, {File: "missing.go"},
		}}},
	}
	report, err := Analyze(plan, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Diagnostics) != 2 || report.Diagnostics[0].Code != "E_TARGET_AMBIGUOUS" || report.Diagnostics[1].Code != "E_TARGET_NOT_FOUND" {
		t.Fatalf("diagnostics = %#v", report.Diagnostics)
	}
}

func TestValidatePlanRejectsHardToInterpretInput(t *testing.T) {
	tests := []Plan{
		{},
		{SchemaVersion: PlanSchemaVersion, Team: Team{ID: "t", Name: "T"}, Missions: []Mission{{ID: "m", Title: "M", Owner: "missing", Status: "active", Targets: []Target{{File: "a.go"}}}}},
		{SchemaVersion: PlanSchemaVersion, Team: Team{ID: "t", Name: "T", Members: []Member{{ID: "x", Name: "X"}}}, Missions: []Mission{{ID: "m", Title: "M", Owner: "x", Status: "active", Targets: []Target{{File: "../secret"}}}}},
	}
	for index, plan := range tests {
		if err := ValidatePlan(plan); err == nil {
			t.Fatalf("case %d unexpectedly passed", index)
		}
	}
}

func testSnapshot() sem.ProviderSnapshot {
	return sem.ProviderSnapshot{
		Header: sem.SnapshotHeader{SchemaVersion: "1.1", Provider: "entire-graph", ProviderVersion: "test", RepoRoot: "/repo", Profile: "full", Stats: sem.ProviderStats{CompletenessLevel: "complete"}},
		Files: []sem.FileRecord{
			{ID: "file-cli", Path: "internal/cli/root.go"},
			{ID: "file-sem", Path: "internal/sem/analyze.go"},
			{ID: "file-test", Path: "internal/sem/analyze_test.go"},
			{ID: "file-readme", Path: "README.md"},
		},
		Symbols: []sem.SymbolRecord{
			{ID: "run", Name: "runCheckpoint", QualifiedName: "runCheckpoint", Kind: "function", FilePath: "internal/cli/root.go", StartLine: 794, EndLine: 817},
			{ID: "analyze", Name: "AnalyzeCheckpoint", QualifiedName: "AnalyzeCheckpoint", Kind: "function", FilePath: "internal/sem/analyze.go", StartLine: 1159, EndLine: 1177},
			{ID: "test", Name: "TestAnalyzeCheckpoint", QualifiedName: "TestAnalyzeCheckpoint", Kind: "function", FilePath: "internal/sem/analyze_test.go", StartLine: 1200, EndLine: 1220},
		},
		Relations: []sem.RelationRecord{
			{FromID: "run", ToID: "analyze", Type: "CALLS", Confidence: 0.98, Resolution: "exact", Evidence: []sem.Evidence{{Kind: "call_site", FilePath: "internal/cli/root.go", StartLine: 812}}},
			{FromID: "test", ToID: "analyze", Type: "TESTS", Confidence: 0.9, Resolution: "exact"},
		},
	}
}
