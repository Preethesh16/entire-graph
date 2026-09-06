package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/entireio/entire-graph/internal/coordinate"
	"github.com/entireio/entire-graph/internal/sem"
	"github.com/entireio/entire-graph/internal/termsafe"
)

const maxCoordinatePlanBytes = 1 << 20

type coordinateFlags struct {
	repo         string
	plan         string
	format       string
	head         bool
	listen       string
	allowRemote  bool
	networkState string
}

type coordinateData struct {
	snapshot         sem.ProviderSnapshot
	sessions         []coordinate.Session
	sessionHealth    coordinate.ProviderHealth
	checkpoints      []coordinate.Checkpoint
	checkpointHealth coordinate.ProviderHealth
}

type coordinateRuntime struct {
	mu      sync.RWMutex
	data    coordinateData
	refresh func(context.Context) (coordinateData, error)
}

func (runtime *coordinateRuntime) current() coordinateData {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	return runtime.data
}

func (runtime *coordinateRuntime) reload(ctx context.Context) (coordinateData, error) {
	if runtime.refresh == nil {
		return coordinateData{}, errors.New("repository refresh is unavailable")
	}
	next, err := runtime.refresh(ctx)
	if err != nil {
		return coordinateData{}, err
	}
	runtime.mu.Lock()
	runtime.data = next
	runtime.mu.Unlock()
	return next, nil
}

func runCoordinate(ctx context.Context, opts Options, args []string) error {
	flags, err := parseCoordinateFlags(args)
	if err != nil {
		return err
	}
	repo, err := resolveRepo(ctx, opts.Env, flags.repo)
	if err != nil {
		return err
	}
	plan, err := readCoordinatePlan(flags.plan)
	if err != nil {
		return err
	}
	snapshot, err := sem.BuildProviderSnapshotWithOptions(ctx, repo, opts.Version, sem.ProviderSnapshotOptions{
		NoNetwork: true,
		Worktree:  !flags.head,
		Profile:   sem.ProfileFull,
	})
	if err != nil {
		return err
	}
	sessions, sessionErr := coordinate.LoadEntireSessions(ctx, "entire")
	sessionHealth := coordinate.ProviderHealth{Available: sessionErr == nil}
	if sessionErr != nil {
		sessionHealth.Detail = sessionErr.Error()
	}
	checkpoints, checkpointErr := coordinate.LoadEntireCheckpoints(ctx, "entire")
	checkpointHealth := coordinate.ProviderHealth{Available: checkpointErr == nil}
	if checkpointErr != nil {
		checkpointHealth.Detail = checkpointErr.Error()
	}
	report, err := coordinate.AnalyzeWithActivity(plan, snapshot, sessions, sessionHealth, checkpoints, checkpointHealth)
	if err != nil {
		return err
	}
	if flags.listen != "" {
		return serveCoordinate(ctx, opts, repo, flags.listen, flags.plan, plan, snapshot, sessions, sessionHealth, checkpoints, checkpointHealth, flags.head, flags.allowRemote, flags.networkState)
	}
	if flags.format == "json" {
		encoder := json.NewEncoder(termsafe.NewJSONWriter(opts.Stdout))
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	return writeCoordinateText(opts.Stdout, report)
}

func parseCoordinateFlags(args []string) (coordinateFlags, error) {
	flags := coordinateFlags{format: "text"}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--repo":
			index++
			if index >= len(args) {
				return flags, errors.New("coordinate --repo requires a path")
			}
			flags.repo = args[index]
		case "--plan":
			index++
			if index >= len(args) {
				return flags, errors.New("coordinate --plan requires a JSON file")
			}
			flags.plan = args[index]
		case "--format":
			index++
			if index >= len(args) {
				return flags, errors.New("coordinate --format requires text or json")
			}
			flags.format = args[index]
		case "--head":
			flags.head = true
		case "--listen":
			index++
			if index >= len(args) {
				return flags, errors.New("coordinate --listen requires host:port")
			}
			flags.listen = args[index]
		case "--allow-remote":
			flags.allowRemote = true
		case "--network-state":
			index++
			if index >= len(args) {
				return flags, errors.New("coordinate --network-state requires a path")
			}
			flags.networkState = args[index]
		default:
			return flags, unexpectedArgumentsError("coordinate", "", []string{args[index]})
		}
	}
	if strings.TrimSpace(flags.plan) == "" {
		return flags, errors.New("coordinate requires --plan <file>")
	}
	if flags.format != "text" && flags.format != "json" {
		return flags, fmt.Errorf("coordinate --format must be text or json, got %q", flags.format)
	}
	return flags, nil
}

func serveCoordinate(ctx context.Context, opts Options, repo, address, planPath string, plan coordinate.Plan, snapshot sem.ProviderSnapshot, sessions []coordinate.Session, health coordinate.ProviderHealth, checkpoints []coordinate.Checkpoint, checkpointHealth coordinate.ProviderHealth, head, allowRemote bool, networkStatePath string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("coordinate --listen requires host:port: %w", err)
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) && !allowRemote {
		return errors.New("coordinate --listen requires --allow-remote for a non-loopback address")
	}
	store, err := coordinate.NewPlanStore(planPath, plan)
	if err != nil {
		return err
	}
	if networkStatePath == "" {
		networkStatePath = filepath.Join(os.TempDir(), "spidey-sense-network-state.json")
	}
	network, err := coordinate.OpenNetworkStore(networkStatePath)
	if err != nil {
		return err
	}
	runtime := &coordinateRuntime{
		data: coordinateData{snapshot: snapshot, sessions: sessions, sessionHealth: health, checkpoints: checkpoints, checkpointHealth: checkpointHealth},
		refresh: func(refreshContext context.Context) (coordinateData, error) {
			nextSnapshot, buildErr := sem.BuildProviderSnapshotWithOptions(refreshContext, repo, opts.Version, sem.ProviderSnapshotOptions{
				NoNetwork: true, Worktree: !head, Profile: sem.ProfileFull,
			})
			if buildErr != nil {
				return coordinateData{}, buildErr
			}
			nextSessions, sessionErr := coordinate.LoadEntireSessions(refreshContext, "entire")
			nextSessionHealth := coordinate.ProviderHealth{Available: sessionErr == nil}
			if sessionErr != nil {
				nextSessionHealth.Detail = sessionErr.Error()
			}
			nextCheckpoints, checkpointErr := coordinate.LoadEntireCheckpoints(refreshContext, "entire")
			nextCheckpointHealth := coordinate.ProviderHealth{Available: checkpointErr == nil}
			if checkpointErr != nil {
				nextCheckpointHealth.Detail = checkpointErr.Error()
			}
			return coordinateData{snapshot: nextSnapshot, sessions: nextSessions, sessionHealth: nextSessionHealth, checkpoints: nextCheckpoints, checkpointHealth: nextCheckpointHealth}, nil
		},
	}
	handler := coordinateHandler(store, runtime, network)
	server := &http.Server{
		Addr: address, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	fmt.Fprintf(opts.Stderr, "Spidey Sense API listening on http://%s/api/v1/report\n", address)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func coordinateHandler(store *coordinate.PlanStore, runtime *coordinateRuntime, network *coordinate.NetworkStore) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(out http.ResponseWriter, _ *http.Request) {
		writeAPIJSON(out, http.StatusOK, map[string]any{"status": "ok", "schema_version": coordinate.ReportSchemaVersion})
	})
	mux.HandleFunc("GET /api/v1/report", func(out http.ResponseWriter, _ *http.Request) {
		data := runtime.current()
		report, err := coordinate.AnalyzeWithActivity(store.Current(), data.snapshot, data.sessions, data.sessionHealth, data.checkpoints, data.checkpointHealth)
		if err != nil {
			writeAPIJSON(out, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeAPIJSON(out, http.StatusOK, report)
	})
	mux.HandleFunc("GET /api/v1/plan", func(out http.ResponseWriter, _ *http.Request) {
		writeAPIJSON(out, http.StatusOK, store.Current())
	})
	mux.HandleFunc("GET /api/v1/sessions", func(out http.ResponseWriter, _ *http.Request) {
		data := runtime.current()
		writeAPIJSON(out, http.StatusOK, map[string]any{"sessions": data.sessions, "health": data.sessionHealth})
	})
	mux.HandleFunc("POST /api/v1/refresh", func(out http.ResponseWriter, request *http.Request) {
		data, err := runtime.reload(request.Context())
		if err != nil {
			writeAPIJSON(out, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
			return
		}
		report, err := coordinate.AnalyzeWithActivity(store.Current(), data.snapshot, data.sessions, data.sessionHealth, data.checkpoints, data.checkpointHealth)
		if err != nil {
			writeAPIJSON(out, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeAPIJSON(out, http.StatusOK, report)
	})
	registerNetworkRoutes(mux, network)
	mux.HandleFunc("PUT /api/v1/plan", func(out http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(out, request.Body, maxCoordinatePlanBytes)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		var next coordinate.Plan
		if err := decoder.Decode(&next); err != nil {
			writeAPIJSON(out, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		updated, err := store.Update(next)
		if errors.Is(err, coordinate.ErrRevisionConflict) {
			writeAPIJSON(out, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		if err != nil {
			writeAPIJSON(out, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeAPIJSON(out, http.StatusOK, updated)
	})
	return mux
}

func registerNetworkRoutes(mux *http.ServeMux, network *coordinate.NetworkStore) {
	if network == nil {
		return
	}
	mux.HandleFunc("POST /api/v1/teams", func(out http.ResponseWriter, request *http.Request) {
		var input coordinate.CreateTeamRequest
		if !decodeAPIRequest(out, request, &input) {
			return
		}
		created, err := network.CreateTeam(input)
		if err != nil {
			writeNetworkError(out, err)
			return
		}
		writeAPIJSON(out, http.StatusCreated, created)
	})
	mux.HandleFunc("POST /api/v1/teams/{id}/join", func(out http.ResponseWriter, request *http.Request) {
		var input coordinate.JoinTeamRequest
		if !decodeAPIRequest(out, request, &input) {
			return
		}
		joined, err := network.JoinTeam(request.PathValue("id"), input)
		if err != nil {
			writeNetworkError(out, err)
			return
		}
		writeAPIJSON(out, http.StatusCreated, joined)
	})
	mux.HandleFunc("POST /api/v1/agents/connect", func(out http.ResponseWriter, request *http.Request) {
		var input coordinate.ConnectAgentRequest
		if !decodeAPIRequest(out, request, &input) {
			return
		}
		connected, err := network.ConnectAgent(bearerToken(request), input)
		if err != nil {
			writeNetworkError(out, err)
			return
		}
		writeAPIJSON(out, http.StatusCreated, connected)
	})
	mux.HandleFunc("POST /api/v1/agents/{id}/heartbeat", func(out http.ResponseWriter, request *http.Request) {
		var input coordinate.HeartbeatRequest
		if !decodeAPIRequest(out, request, &input) {
			return
		}
		agent, err := network.Heartbeat(request.PathValue("id"), bearerToken(request), input)
		if err != nil {
			writeNetworkError(out, err)
			return
		}
		writeAPIJSON(out, http.StatusOK, agent)
	})
	mux.HandleFunc("GET /api/v1/teams/{id}/agents", func(out http.ResponseWriter, request *http.Request) {
		agents, err := network.Agents(request.PathValue("id"), bearerToken(request))
		if err != nil {
			writeNetworkError(out, err)
			return
		}
		writeAPIJSON(out, http.StatusOK, map[string]any{"agents": agents})
	})
	mux.HandleFunc("GET /api/v1/teams/{id}/events", func(out http.ResponseWriter, request *http.Request) {
		events, cancel, err := network.Subscribe(request.PathValue("id"), bearerToken(request))
		if err != nil {
			writeNetworkError(out, err)
			return
		}
		defer cancel()
		flusher, ok := out.(http.Flusher)
		if !ok {
			writeAPIJSON(out, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
			return
		}
		// The server's ordinary write deadline protects finite API responses, but an
		// authenticated event stream is intentionally long-lived.
		_ = http.NewResponseController(out).SetWriteDeadline(time.Time{})
		out.Header().Set("Content-Type", "text/event-stream")
		out.Header().Set("Cache-Control", "no-store")
		out.Header().Set("X-Content-Type-Options", "nosniff")
		out.Header().Set("X-Accel-Buffering", "no")
		fmt.Fprint(out, "event: ready\ndata: {}\n\n")
		flusher.Flush()
		for {
			select {
			case <-request.Context().Done():
				return
			case event, open := <-events:
				if !open {
					return
				}
				content, _ := json.Marshal(event)
				fmt.Fprintf(out, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Type, content)
				flusher.Flush()
			}
		}
	})
}

func decodeAPIRequest(out http.ResponseWriter, request *http.Request, value any) bool {
	request.Body = http.MaxBytesReader(out, request.Body, maxCoordinatePlanBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeAPIJSON(out, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		writeAPIJSON(out, http.StatusBadRequest, map[string]string{"error": "request must contain exactly one JSON value"})
		return false
	}
	return true
}

func bearerToken(request *http.Request) string {
	const prefix = "Bearer "
	value := request.Header.Get("Authorization")
	if !strings.HasPrefix(value, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}

func writeNetworkError(out http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, coordinate.ErrUnauthorized) {
		status = http.StatusUnauthorized
	}
	if errors.Is(err, coordinate.ErrTeamNotFound) || errors.Is(err, coordinate.ErrAgentNotFound) {
		status = http.StatusNotFound
	}
	writeAPIJSON(out, status, map[string]string{"error": err.Error()})
}

func writeAPIJSON(out http.ResponseWriter, status int, value any) {
	out.Header().Set("Content-Type", "application/json; charset=utf-8")
	out.Header().Set("Cache-Control", "no-store")
	out.Header().Set("X-Content-Type-Options", "nosniff")
	out.WriteHeader(status)
	_ = json.NewEncoder(out).Encode(value)
}

func readCoordinatePlan(path string) (coordinate.Plan, error) {
	file, err := os.Open(path)
	if err != nil {
		return coordinate.Plan{}, fmt.Errorf("open coordinate plan: %w", err)
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxCoordinatePlanBytes+1))
	if err != nil {
		return coordinate.Plan{}, fmt.Errorf("read coordinate plan: %w", err)
	}
	if len(content) > maxCoordinatePlanBytes {
		return coordinate.Plan{}, fmt.Errorf("coordinate plan exceeds %d bytes", maxCoordinatePlanBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var plan coordinate.Plan
	if err := decoder.Decode(&plan); err != nil {
		return coordinate.Plan{}, fmt.Errorf("decode coordinate plan: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return coordinate.Plan{}, errors.New("decode coordinate plan: multiple JSON values")
		}
		return coordinate.Plan{}, fmt.Errorf("decode coordinate plan: %w", err)
	}
	return plan, nil
}

func writeCoordinateText(out io.Writer, report coordinate.Report) error {
	if _, err := fmt.Fprintf(out, "Spidey Sense: %s\n", report.Team.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Graph: %s %s | profile=%s | completeness=%s | warnings=%d | partial_failures=%d\n", report.Graph.Provider, report.Graph.ProviderVersion, report.Graph.Profile, report.Graph.CompletenessLevel, report.Graph.WarningCount, report.Graph.PartialFailureCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Decisions: %d BLOCK | %d REVIEW | %d CLEAR\n", report.Summary.Block, report.Summary.Review, report.Summary.Clear); err != nil {
		return err
	}
	if report.SessionHealth.Available {
		if _, err := fmt.Fprintf(out, "Entire sessions: %d visible\n", len(report.Sessions)); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintf(out, "Entire sessions: unavailable (%s)\n", report.SessionHealth.Detail); err != nil {
		return err
	}
	for _, decision := range report.Decisions {
		if _, err := fmt.Fprintf(out, "\n[%s/%s] %s <-> %s: %s\n", decision.Level, strings.ToUpper(decision.EvidenceClass), decision.MissionA, decision.MissionB, decision.Reason); err != nil {
			return err
		}
		for _, step := range decision.Path {
			location := ""
			if step.From.File != "" {
				location = step.From.File
				if step.From.StartLine > 0 {
					location += fmt.Sprintf(":%d", step.From.StartLine)
				}
			}
			if _, err := fmt.Fprintf(out, "  %s --%s/%s/%s--> %s", step.From.Name, step.Relation, step.Direction, step.EvidenceClass, step.To.Name); err != nil {
				return err
			}
			if location != "" {
				if _, err := fmt.Fprintf(out, " (%s)", location); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
		for _, recommendation := range decision.Recommendations {
			if _, err := fmt.Fprintf(out, "  Action: %s\n", recommendation); err != nil {
				return err
			}
		}
		for _, target := range decision.TestTargets {
			if _, err := fmt.Fprintf(out, "  Test: %s\n", target); err != nil {
				return err
			}
		}
		for _, verification := range decision.Verification {
			label := "Verify"
			if decision.VerificationRequired {
				label = "Verify (required)"
			}
			if _, err := fmt.Fprintf(out, "  %s: %s\n", label, verification); err != nil {
				return err
			}
		}
		if decision.Caveat != "" {
			if _, err := fmt.Fprintf(out, "  Caveat: %s\n", decision.Caveat); err != nil {
				return err
			}
		}
	}
	for _, diagnostic := range report.Diagnostics {
		if _, err := fmt.Fprintf(out, "\n[%s] mission=%s: %s\n", diagnostic.Code, diagnostic.MissionID, diagnostic.Detail); err != nil {
			return err
		}
	}
	return nil
}
