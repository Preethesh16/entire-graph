// Package coordinate turns an Entire Graph snapshot and a team mission plan
// into deterministic, explainable coordination decisions.
package coordinate

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/entireio/entire-graph/internal/sem"
)

const (
	PlanSchemaVersion   = "1.0"
	ReportSchemaVersion = "spidey-sense/v1alpha1"
)

type Plan struct {
	SchemaVersion string    `json:"schema_version"`
	Revision      int       `json:"revision"`
	Team          Team      `json:"team"`
	Missions      []Mission `json:"missions"`
}

type Team struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CallSign  string   `json:"call_sign,omitempty"`
	Objective string   `json:"objective,omitempty"`
	Members   []Member `json:"members"`
}

type Member struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Role       string   `json:"role,omitempty"`
	Archetype  string   `json:"archetype,omitempty"`
	Accent     string   `json:"accent,omitempty"`
	SessionIDs []string `json:"session_ids,omitempty"`
}

type Mission struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Intent  string   `json:"intent,omitempty"`
	Owner   string   `json:"owner"`
	Status  string   `json:"status"`
	Targets []Target `json:"targets"`
}

type Target struct {
	File   string `json:"file,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Line   int    `json:"line,omitempty"`
}

type Report struct {
	SchemaVersion    string                `json:"schema_version"`
	Graph            GraphProvenance       `json:"graph"`
	Team             Team                  `json:"team"`
	Missions         []Mission             `json:"missions"`
	Decisions        []Decision            `json:"decisions"`
	Diagnostics      []Diagnostic          `json:"diagnostics,omitempty"`
	Summary          Summary               `json:"summary"`
	Warnings         []sem.ProviderWarning `json:"graph_warnings,omitempty"`
	Failures         []sem.PartialFailure  `json:"graph_partial_failures,omitempty"`
	Sessions         []Session             `json:"sessions,omitempty"`
	SessionHealth    ProviderHealth        `json:"session_health"`
	Checkpoints      []Checkpoint          `json:"checkpoints,omitempty"`
	CheckpointHealth ProviderHealth        `json:"checkpoint_health"`
}

type ProviderHealth struct {
	Available bool   `json:"available"`
	Detail    string `json:"detail,omitempty"`
}

type GraphProvenance struct {
	Provider            string `json:"provider"`
	ProviderVersion     string `json:"provider_version"`
	SchemaVersion       string `json:"schema_version"`
	RepoRoot            string `json:"repo_root"`
	Commit              string `json:"commit,omitempty"`
	Tree                string `json:"tree,omitempty"`
	Profile             string `json:"profile"`
	CompletenessLevel   string `json:"completeness_level"`
	NodeCount           int    `json:"node_count"`
	RelationCount       int    `json:"relation_count"`
	WarningCount        int    `json:"warning_count"`
	PartialFailureCount int    `json:"partial_failure_count"`
	AnalysisPartial     bool   `json:"analysis_partial"`
}

type Summary struct {
	Missions int `json:"missions"`
	Pairs    int `json:"pairs"`
	Block    int `json:"block"`
	Review   int `json:"review"`
	Clear    int `json:"clear"`
}

type Diagnostic struct {
	MissionID string `json:"mission_id,omitempty"`
	Target    Target `json:"target,omitempty"`
	Code      string `json:"code"`
	Detail    string `json:"detail"`
}

type Decision struct {
	Level                string   `json:"level"`
	Reason               string   `json:"reason"`
	MissionA             string   `json:"mission_a"`
	MissionB             string   `json:"mission_b"`
	Path                 []Step   `json:"path,omitempty"`
	TestTargets          []string `json:"test_targets,omitempty"`
	Recommendations      []string `json:"recommendations"`
	Caveat               string   `json:"caveat,omitempty"`
	EvidenceClass        string   `json:"evidence_class"`
	VerificationRequired bool     `json:"verification_required"`
	Verification         []string `json:"verification"`
}

type Step struct {
	From            Endpoint   `json:"from"`
	To              Endpoint   `json:"to"`
	Relation        string     `json:"relation"`
	Direction       string     `json:"direction"`
	Confidence      float64    `json:"confidence"`
	Resolution      string     `json:"resolution,omitempty"`
	EvidenceClass   string     `json:"evidence_class"`
	Reason          string     `json:"reason,omitempty"`
	Evidence        []Evidence `json:"evidence,omitempty"`
	WarningCodes    []string   `json:"warning_codes,omitempty"`
	EvidenceDropped int        `json:"evidence_dropped,omitempty"`
}

type Endpoint struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	File      string `json:"file,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
}

type Evidence struct {
	Kind      string `json:"kind"`
	File      string `json:"file,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

var allowedStatuses = map[string]bool{
	"queued": true, "active": true, "blocked": true, "complete": true,
}

func ValidatePlan(plan Plan) error {
	if plan.SchemaVersion != PlanSchemaVersion {
		return fmt.Errorf("plan schema_version must be %q", PlanSchemaVersion)
	}
	if plan.Revision < 0 {
		return errors.New("plan revision cannot be negative")
	}
	if strings.TrimSpace(plan.Team.ID) == "" || strings.TrimSpace(plan.Team.Name) == "" {
		return errors.New("team id and name are required")
	}
	members := make(map[string]bool, len(plan.Team.Members))
	for _, member := range plan.Team.Members {
		if strings.TrimSpace(member.ID) == "" || strings.TrimSpace(member.Name) == "" {
			return errors.New("every member needs an id and name")
		}
		if members[member.ID] {
			return fmt.Errorf("duplicate member id %q", member.ID)
		}
		members[member.ID] = true
	}
	missions := make(map[string]bool, len(plan.Missions))
	for _, mission := range plan.Missions {
		if strings.TrimSpace(mission.ID) == "" || strings.TrimSpace(mission.Title) == "" {
			return errors.New("every mission needs an id and title")
		}
		if missions[mission.ID] {
			return fmt.Errorf("duplicate mission id %q", mission.ID)
		}
		missions[mission.ID] = true
		if !members[mission.Owner] {
			return fmt.Errorf("mission %q has unknown owner %q", mission.ID, mission.Owner)
		}
		if !allowedStatuses[mission.Status] {
			return fmt.Errorf("mission %q has invalid status %q", mission.ID, mission.Status)
		}
		if len(mission.Targets) == 0 {
			return fmt.Errorf("mission %q needs at least one target", mission.ID)
		}
		for _, target := range mission.Targets {
			if strings.TrimSpace(target.File) == "" && strings.TrimSpace(target.Symbol) == "" {
				return fmt.Errorf("mission %q has an empty target", mission.ID)
			}
			if target.Line < 0 {
				return fmt.Errorf("mission %q has a negative target line", mission.ID)
			}
			if target.File != "" {
				clean := filepath.ToSlash(filepath.Clean(target.File))
				if filepath.IsAbs(target.File) || clean == ".." || strings.HasPrefix(clean, "../") {
					return fmt.Errorf("mission %q target file must be repository-relative", mission.ID)
				}
			}
		}
	}
	return nil
}

type resolvedMission struct {
	mission   Mission
	endpoints map[string]bool
	files     map[string]bool
	symbols   map[string]bool
}

type graphIndex struct {
	endpoints     map[string]Endpoint
	filesByPath   map[string]sem.FileRecord
	symbolsByFile map[string][]sem.SymbolRecord
	symbols       []sem.SymbolRecord
	adjacency     map[string][]adjacent
	relations     []sem.RelationRecord
}

type adjacent struct {
	next     string
	relation sem.RelationRecord
	reversed bool
}

func Analyze(plan Plan, snapshot sem.ProviderSnapshot) (Report, error) {
	return AnalyzeWithSessions(plan, snapshot, nil, ProviderHealth{})
}

func AnalyzeWithSessions(plan Plan, snapshot sem.ProviderSnapshot, sessions []Session, sessionHealth ProviderHealth) (Report, error) {
	return AnalyzeWithActivity(plan, snapshot, sessions, sessionHealth, nil, ProviderHealth{})
}

func AnalyzeWithActivity(plan Plan, snapshot sem.ProviderSnapshot, sessions []Session, sessionHealth ProviderHealth, checkpoints []Checkpoint, checkpointHealth ProviderHealth) (Report, error) {
	if err := ValidatePlan(plan); err != nil {
		return Report{}, err
	}
	index := newGraphIndex(snapshot)
	analysis := newGraphAnalysisState(snapshot)
	report := Report{
		SchemaVersion: ReportSchemaVersion,
		Graph: GraphProvenance{
			Provider: snapshot.Header.Provider, ProviderVersion: snapshot.Header.ProviderVersion,
			SchemaVersion: snapshot.Header.SchemaVersion, RepoRoot: snapshot.Header.RepoRoot,
			Commit: snapshot.Header.Commit, Tree: snapshot.Header.Tree, Profile: snapshot.Header.Profile,
			CompletenessLevel: snapshot.Header.Stats.CompletenessLevel,
			NodeCount:         len(snapshot.Symbols) + len(snapshot.Files), RelationCount: len(snapshot.Relations),
			WarningCount: len(snapshot.Header.Warnings), PartialFailureCount: len(snapshot.Header.PartialFailures),
			AnalysisPartial: analysis.partial(),
		},
		Team: plan.Team, Missions: append([]Mission(nil), plan.Missions...),
		Warnings: snapshot.Header.Warnings, Failures: snapshot.Header.PartialFailures,
		Sessions: filterMappedSessions(plan.Team, sessions), SessionHealth: sessionHealth,
		Checkpoints: filterMappedCheckpoints(plan.Team, checkpoints), CheckpointHealth: checkpointHealth,
	}
	resolved := make([]resolvedMission, 0, len(plan.Missions))
	for _, mission := range plan.Missions {
		if mission.Status == "complete" {
			continue
		}
		item, diagnostics := index.resolveMission(mission)
		report.Diagnostics = append(report.Diagnostics, diagnostics...)
		resolved = append(resolved, item)
	}
	for left := 0; left < len(resolved); left++ {
		for right := left + 1; right < len(resolved); right++ {
			decision := index.decide(resolved[left], resolved[right], graphAnalysisStateForMissions(snapshot, index, resolved[left], resolved[right]))
			report.Decisions = append(report.Decisions, decision)
			report.Summary.Pairs++
			switch decision.Level {
			case "BLOCK":
				report.Summary.Block++
			case "REVIEW":
				report.Summary.Review++
			default:
				report.Summary.Clear++
			}
		}
	}
	report.Summary.Missions = len(plan.Missions)
	sort.Slice(report.Diagnostics, func(i, j int) bool {
		left := report.Diagnostics[i].MissionID + report.Diagnostics[i].Code + report.Diagnostics[i].Detail
		right := report.Diagnostics[j].MissionID + report.Diagnostics[j].Code + report.Diagnostics[j].Detail
		return left < right
	})
	return report, nil
}

func filterMappedSessions(team Team, sessions []Session) []Session {
	allowed := make(map[string]bool)
	for _, member := range team.Members {
		for _, sessionID := range member.SessionIDs {
			if sessionID != "" {
				allowed[sessionID] = true
			}
		}
	}
	filtered := make([]Session, 0, len(allowed))
	for _, session := range sessions {
		if allowed[session.SessionID] {
			filtered = append(filtered, session)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].SessionID < filtered[j].SessionID })
	return filtered
}

func newGraphIndex(snapshot sem.ProviderSnapshot) graphIndex {
	index := graphIndex{
		endpoints: make(map[string]Endpoint), filesByPath: make(map[string]sem.FileRecord),
		symbolsByFile: make(map[string][]sem.SymbolRecord), adjacency: make(map[string][]adjacent),
		symbols: snapshot.Symbols, relations: snapshot.Relations,
	}
	for _, file := range snapshot.Files {
		path := filepath.ToSlash(file.Path)
		index.filesByPath[path] = file
		index.endpoints[file.ID] = Endpoint{ID: file.ID, Name: path, Kind: "file", File: path}
	}
	for _, symbol := range snapshot.Symbols {
		index.symbolsByFile[filepath.ToSlash(symbol.FilePath)] = append(index.symbolsByFile[filepath.ToSlash(symbol.FilePath)], symbol)
		index.endpoints[symbol.ID] = Endpoint{ID: symbol.ID, Name: symbol.QualifiedName, Kind: symbol.Kind, File: symbol.FilePath, StartLine: symbol.StartLine}
	}
	for _, external := range snapshot.Externals {
		index.endpoints[external.ID] = Endpoint{ID: external.ID, Name: external.Value, Kind: external.Kind, File: external.FilePath, StartLine: external.StartLine}
	}
	for _, relation := range snapshot.Relations {
		if !coordinationRelation(relation.Type) {
			continue
		}
		index.adjacency[relation.FromID] = append(index.adjacency[relation.FromID], adjacent{next: relation.ToID, relation: relation})
		index.adjacency[relation.ToID] = append(index.adjacency[relation.ToID], adjacent{next: relation.FromID, relation: relation, reversed: true})
	}
	for id := range index.adjacency {
		sort.Slice(index.adjacency[id], func(i, j int) bool {
			left, right := index.adjacency[id][i], index.adjacency[id][j]
			return left.relation.Type+left.next < right.relation.Type+right.next
		})
	}
	return index
}

func coordinationRelation(relation string) bool {
	switch relation {
	case "IMPORTS", "CALLS", "CONSTRUCTS", "ASYNC_CALLS", "USES_TYPE", "PARAM_TYPE", "RETURNS_TYPE",
		"READS_FIELD", "WRITES_FIELD", "ACCESSES", "DATA_FLOWS", "TESTS", "RESOURCE_DEPENDS_ON",
		"HANDLES_ROUTE", "HANDLES_GRPC", "HANDLES_GRAPHQL", "HANDLES_TRPC", "HTTP_CALLS", "FILE_CHANGES_WITH":
		return true
	default:
		return false
	}
}

func (index graphIndex) resolveMission(mission Mission) (resolvedMission, []Diagnostic) {
	resolved := resolvedMission{mission: mission, endpoints: map[string]bool{}, files: map[string]bool{}, symbols: map[string]bool{}}
	var diagnostics []Diagnostic
	for _, target := range mission.Targets {
		if target.Symbol != "" {
			matches := index.symbolMatches(target)
			switch len(matches) {
			case 0:
				diagnostics = append(diagnostics, Diagnostic{MissionID: mission.ID, Target: target, Code: "E_TARGET_NOT_FOUND", Detail: "symbol target did not resolve"})
			case 1:
				symbol := matches[0]
				resolved.endpoints[symbol.ID] = true
				resolved.symbols[symbol.ID] = true
				resolved.files[filepath.ToSlash(symbol.FilePath)] = true
				if file, ok := index.filesByPath[filepath.ToSlash(symbol.FilePath)]; ok {
					resolved.endpoints[file.ID] = true
				}
			default:
				diagnostics = append(diagnostics, Diagnostic{MissionID: mission.ID, Target: target, Code: "E_TARGET_AMBIGUOUS", Detail: fmt.Sprintf("symbol target matched %d definitions; add file, kind, or line", len(matches))})
			}
			continue
		}
		path := filepath.ToSlash(filepath.Clean(target.File))
		file, ok := index.filesByPath[path]
		if !ok {
			diagnostics = append(diagnostics, Diagnostic{MissionID: mission.ID, Target: target, Code: "E_TARGET_NOT_FOUND", Detail: "file target did not resolve"})
			continue
		}
		resolved.files[path] = true
		resolved.endpoints[file.ID] = true
		for _, symbol := range index.symbolsByFile[path] {
			resolved.endpoints[symbol.ID] = true
		}
	}
	return resolved, diagnostics
}

func (index graphIndex) symbolMatches(target Target) []sem.SymbolRecord {
	file := filepath.ToSlash(filepath.Clean(target.File))
	var matches []sem.SymbolRecord
	for _, symbol := range index.symbols {
		if symbol.Name != target.Symbol && symbol.QualifiedName != target.Symbol {
			continue
		}
		if target.File != "" && filepath.ToSlash(symbol.FilePath) != file {
			continue
		}
		if target.Kind != "" && symbol.Kind != target.Kind {
			continue
		}
		if target.Line > 0 && (target.Line < symbol.StartLine || target.Line > symbol.EndLine) {
			continue
		}
		matches = append(matches, symbol)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].FilePath != matches[j].FilePath {
			return matches[i].FilePath < matches[j].FilePath
		}
		return matches[i].StartLine < matches[j].StartLine
	})
	return matches
}

type graphAnalysisState struct {
	completeness string
	warnings     int
	failures     int
}

func (state graphAnalysisState) partial() bool {
	switch state.completeness {
	case "ok", "complete":
		return state.warnings > 0 || state.failures > 0
	default:
		return true
	}
}

func newGraphAnalysisState(snapshot sem.ProviderSnapshot) graphAnalysisState {
	warnings := 0
	for _, warning := range snapshot.Header.Warnings {
		// Worktree provenance identifies which source was analyzed; by itself it
		// does not mean relationships were dropped or resolved heuristically.
		if warning.Code != "W_WORKTREE_SNAPSHOT" {
			warnings++
		}
	}
	return graphAnalysisState{
		completeness: snapshot.Header.Stats.CompletenessLevel,
		warnings:     warnings,
		failures:     len(snapshot.Header.PartialFailures),
	}
}

func graphAnalysisStateForMissions(snapshot sem.ProviderSnapshot, index graphIndex, missions ...resolvedMission) graphAnalysisState {
	languages := map[string]bool{}
	for _, mission := range missions {
		for path := range mission.files {
			if language := index.filesByPath[path].Language; language != "" {
				languages[language] = true
			}
		}
	}
	relevant := func(path string) bool {
		if path == "" || len(languages) == 0 {
			return true
		}
		file, ok := index.filesByPath[filepath.ToSlash(path)]
		return !ok || file.Language == "" || languages[file.Language]
	}
	state := graphAnalysisState{completeness: snapshot.Header.Stats.CompletenessLevel}
	for _, warning := range snapshot.Header.Warnings {
		if warning.Code != "W_WORKTREE_SNAPSHOT" && relevant(warning.FilePath) {
			state.warnings++
		}
	}
	for _, failure := range snapshot.Header.PartialFailures {
		if relevant(failure.FilePath) {
			state.failures++
		}
	}
	return state
}

func (index graphIndex) decide(left, right resolvedMission, analysis graphAnalysisState) Decision {
	decision := Decision{MissionA: left.mission.ID, MissionB: right.mission.ID}
	if overlap := firstOverlap(left.symbols, right.symbols); overlap != "" {
		endpoint := index.endpoints[overlap]
		decision.Level, decision.Reason = "BLOCK", "same symbol is assigned to both missions"
		decision.Path = []Step{{From: endpoint, To: endpoint, Relation: "SAME_SYMBOL", Direction: "shared", Confidence: 1, EvidenceClass: "confirmed"}}
		decision.EvidenceClass = "confirmed"
		decision.Recommendations = []string{"Assign the symbol to one mission, or sequence the missions before editing."}
		decision.Verification = []string{"Verify the mission assignments against the plan before editing."}
		return decision
	}
	if overlap := firstOverlap(left.files, right.files); overlap != "" {
		endpoint := Endpoint{Name: overlap, Kind: "file", File: overlap}
		if file, ok := index.filesByPath[overlap]; ok {
			endpoint.ID = file.ID
		}
		decision.Level, decision.Reason = "BLOCK", "same file is assigned to both missions"
		decision.Path = []Step{{From: endpoint, To: endpoint, Relation: "SAME_FILE", Direction: "shared", Confidence: 1, EvidenceClass: "confirmed"}}
		decision.EvidenceClass = "confirmed"
		decision.Recommendations = []string{"Assign the file to one mission, or sequence the missions before editing."}
		decision.Verification = []string{"Verify the mission assignments against the plan before editing."}
		return decision
	}
	path := index.shortestPath(left.endpoints, right.endpoints, 2)
	if len(path) == 0 {
		decision.Level, decision.Reason = "CLEAR", "no connection found within two graph hops"
		decision.EvidenceClass = "incomplete"
		decision.VerificationRequired = true
		decision.Recommendations = []string{"Proceed in parallel, while treating this bounded static analysis as advisory."}
		decision.Verification = []string{"Inspect the source for runtime wiring, generated code, reflection, or dynamic dispatch, then run focused tests for both mission targets."}
		decision.Caveat = clearCaveat(analysis)
		return decision
	}
	decision.Path = path
	decision.EvidenceClass = pathEvidenceClass(path)
	if analysis.partial() {
		decision.EvidenceClass = "incomplete"
	}
	decision.VerificationRequired = decision.EvidenceClass != "confirmed"
	decision.TestTargets = index.testTargets(left.endpoints, right.endpoints)
	if len(path) == 1 && blockingRelation(path[0]) && !analysis.partial() {
		decision.Level, decision.Reason = "BLOCK", "direct structural dependency connects the missions"
		decision.Recommendations = []string{"Sequence the dependency-changing mission first, then refresh and review the dependent mission."}
	} else {
		decision.Level, decision.Reason = "REVIEW", "related code connects the missions and needs coordinated review"
		decision.Recommendations = []string{"Continue only with an explicit review handoff for the reported path."}
	}
	if len(decision.TestTargets) > 0 {
		decision.Recommendations = append(decision.Recommendations, "Run or inspect the reported test targets before merging.")
		decision.Verification = append(decision.Verification, "Run the reported test targets before merging.")
	}
	if decision.VerificationRequired {
		decision.Verification = append(decision.Verification, "Inspect the reported source locations and confirm runtime dispatch before relying on this relationship.")
	} else if len(decision.Verification) == 0 {
		decision.Verification = append(decision.Verification, "Review the reported source locations and run focused tests if the relationship affects runtime behavior.")
	}
	if analysis.partial() {
		decision.Caveat = partialCaveat(analysis)
	}
	return decision
}

func firstOverlap(left, right map[string]bool) string {
	var matches []string
	for value := range left {
		if right[value] {
			matches = append(matches, value)
		}
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}

type pathNode struct {
	id    string
	steps []Step
}

func (index graphIndex) shortestPath(starts, goals map[string]bool, maxDepth int) []Step {
	var startIDs []string
	for id := range starts {
		startIDs = append(startIDs, id)
	}
	sort.Strings(startIDs)
	queue := make([]pathNode, 0, len(startIDs))
	seen := map[string]bool{}
	for _, id := range startIDs {
		queue = append(queue, pathNode{id: id})
		seen[id] = true
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if len(current.steps) >= maxDepth {
			continue
		}
		for _, edge := range index.adjacency[current.id] {
			if seen[edge.next] {
				continue
			}
			step := index.step(current.id, edge)
			steps := append(append([]Step(nil), current.steps...), step)
			if goals[edge.next] {
				return steps
			}
			seen[edge.next] = true
			queue = append(queue, pathNode{id: edge.next, steps: steps})
		}
	}
	return nil
}

func (index graphIndex) step(current string, edge adjacent) Step {
	direction := "forward"
	if edge.reversed {
		direction = "reverse"
	}
	step := Step{
		From: index.endpoints[current], To: index.endpoints[edge.next], Relation: edge.relation.Type,
		Direction: direction, Confidence: edge.relation.Confidence, Resolution: edge.relation.Resolution,
		EvidenceClass: relationEvidenceClass(edge.relation), Reason: edge.relation.Reason,
		WarningCodes: append([]string(nil), edge.relation.WarningCodes...), EvidenceDropped: edge.relation.EvidenceDropped,
	}
	for _, evidence := range edge.relation.Evidence {
		step.Evidence = append(step.Evidence, Evidence{Kind: evidence.Kind, File: evidence.FilePath, StartLine: evidence.StartLine, EndLine: evidence.EndLine, Detail: evidence.Detail})
	}
	return step
}

func blockingRelation(step Step) bool {
	if step.EvidenceClass != "confirmed" {
		return false
	}
	switch step.Relation {
	case "IMPORTS", "CALLS", "CONSTRUCTS", "ASYNC_CALLS", "USES_TYPE", "PARAM_TYPE", "RETURNS_TYPE",
		"READS_FIELD", "WRITES_FIELD", "ACCESSES", "DATA_FLOWS", "RESOURCE_DEPENDS_ON":
		return true
	default:
		return false
	}
}

func (index graphIndex) testTargets(left, right map[string]bool) []string {
	all := map[string]bool{}
	for id := range left {
		all[id] = true
	}
	for id := range right {
		all[id] = true
	}
	tests := map[string]bool{}
	for _, relation := range index.relations {
		if relation.Type != "TESTS" || !all[relation.FromID] && !all[relation.ToID] {
			continue
		}
		for _, id := range []string{relation.FromID, relation.ToID} {
			if endpoint, ok := index.endpoints[id]; ok && conventionalTestPath(endpoint.File) {
				tests[endpoint.File] = true
			}
		}
	}
	var result []string
	for path := range tests {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func conventionalTestPath(path string) bool {
	lower := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(lower)
	return strings.Contains(lower, "/test/") || strings.Contains(lower, "/tests/") ||
		strings.Contains(lower, "/__tests__/") || strings.HasSuffix(base, "_test.go") ||
		strings.HasSuffix(base, "_test.py") || strings.HasPrefix(base, "test_") ||
		strings.Contains(base, ".test.") || strings.Contains(base, ".spec.")
}

func relationEvidenceClass(relation sem.RelationRecord) string {
	if len(relation.WarningCodes) > 0 || relation.EvidenceDropped > 0 {
		return "incomplete"
	}
	switch relation.Type {
	case "HANDLES_ROUTE", "HTTP_CALLS", "EMITS", "LISTENS_ON", "HANDLES_TOOL", "SIMILAR_TO", "TESTS", "FILE_CHANGES_WITH":
		return "heuristic"
	}
	switch relation.Resolution {
	case "exact", "package", "import_resolved", "resolved":
		return "confirmed"
	default:
		return "heuristic"
	}
}

func pathEvidenceClass(path []Step) string {
	result := "confirmed"
	for _, step := range path {
		if step.EvidenceClass == "incomplete" {
			return "incomplete"
		}
		if step.EvidenceClass != "confirmed" {
			result = "heuristic"
		}
	}
	return result
}

func clearCaveat(analysis graphAnalysisState) string {
	if !analysis.partial() {
		return "CLEAR means no connection was found within two graph hops; it is not proof of independence."
	}
	return partialCaveat(analysis) + " CLEAR is not proof of independence."
}

func partialCaveat(analysis graphAnalysisState) string {
	level := analysis.completeness
	if level == "" {
		level = "unknown"
	}
	return fmt.Sprintf("Graph analysis is partial (completeness=%s, warnings=%d, failures=%d); inspect diagnostics before relying on this decision.", level, analysis.warnings, analysis.failures)
}
