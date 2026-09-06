package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/entireio/entire-graph/internal/coordinate"
	"github.com/entireio/entire-graph/internal/sem"
	"github.com/entireio/entire-graph/internal/termsafe"
)

const maxCoordinatePlanBytes = 1 << 20

type coordinateFlags struct {
	repo   string
	plan   string
	format string
	head   bool
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
	report, err := coordinate.AnalyzeWithSessions(plan, snapshot, sessions, sessionHealth)
	if err != nil {
		return err
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
	if _, err := fmt.Fprintf(out, "Graph: %s %s | profile=%s | completeness=%s\n", report.Graph.Provider, report.Graph.ProviderVersion, report.Graph.Profile, report.Graph.CompletenessLevel); err != nil {
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
		if _, err := fmt.Fprintf(out, "\n[%s] %s <-> %s: %s\n", decision.Level, decision.MissionA, decision.MissionB, decision.Reason); err != nil {
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
			if _, err := fmt.Fprintf(out, "  %s --%s/%s--> %s", step.From.Name, step.Relation, step.Direction, step.To.Name); err != nil {
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
