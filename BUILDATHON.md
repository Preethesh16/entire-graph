# Spidey Sense

## One-sentence summary

Spidey Sense is an Entire-powered mission-control dashboard that turns live agent activity and code-graph relationships into explainable coordination decisions before parallel work collides.

## Problem, intended user, and why it matters

Vibe-coding teams increasingly run several humans, Codex sessions, and Claude Code sessions against the same repository. A lead may divide the work cleanly in a written plan while the implementation still collides at the file, symbol, caller, type, or test level. Git usually exposes that problem late, during rebase, review, or merge.

Spidey Sense is for the lead coordinating those parallel sessions. It answers:

- Who is working on what right now?
- How does actual activity compare with the plan?
- Are two active missions touching the same or structurally connected code?
- What exact graph path makes the overlap risky?
- Should the team continue in parallel, review together, run particular tests, or sequence the work?

## Selected Entire track and why Entire is essential

**Track E2 / Problem Statement 2 — Build with Graph Intelligence.**

Entire is essential in two independent ways:

1. Entire sessions provide the public, auditable activity context for supported coding agents without scraping private reasoning.
2. Entire Graph provides definitions, files, call relationships, type relationships, data flows, test links, and semantic changes used to justify coordination decisions.

Without Entire Graph, Spidey Sense could only report same-file overlap. It could not explain that two apparently separate tasks are connected through a caller, shared type, data flow, or affected test. Raw graph output is not the product: the product converts evidence into `CLEAR`, `REVIEW`, or `BLOCK` decisions and a recommended next action.

## Architecture and main workflow

The initial demonstrable workflow is deliberately narrow:

1. A lead creates a dynamic team plan with any number of members and missions.
2. Team members run Codex or Claude Code with Entire enabled in the repository.
3. A local coordinator reads public Entire session metadata and observable Git state.
4. Mission targets and touched files are resolved against an Entire Graph snapshot.
5. A deterministic risk engine detects exact overlap and bounded structural paths.
6. The dashboard shows progress, evidence, and an actionable recommendation.
7. The lead can reassign or sequence a mission and immediately rerun the analysis.

The system is split into replaceable boundaries:

- an Entire session adapter;
- a versioned plan/activity store;
- an Entire Graph adapter;
- a deterministic coordination engine;
- read-only Git telemetry;
- a local API and original spider-radar dashboard.

See `docs/spidey-sense-architecture.md` for the detailed design.

## Must-have demo path

- Create or load a team and plan without hardcoded users, tasks, files, or repository names.
- Discover at least two Entire-tracked coding sessions.
- Display each mission as queued, active, blocked, or complete.
- Demonstrate a same-file collision and a structural dependency collision.
- Show the exact relationship path, confidence/resolution information, and source location.
- Recommend one verifiable action: sequence work, review a dependent target, or run an affected test.
- Update the plan and show the risk decision change.
- Keep all critical information usable without the graphical visualization.

## Optional features after the stable core

- richer Codex and Claude provider-specific metadata;
- multi-machine synchronization;
- GitHub pull-request and merge adapters;
- audited messages to supported agents;
- checkpoint-intent versus semantic-change drift;
- historical coordination timelines;
- a larger 3D dependency world;
- AI-written summaries grounded in the deterministic evidence.

## Implementation progress

The first backend slice is now available on the `progress` branch:

- a versioned, dynamic team and mission-plan model;
- validation for member ownership, mission status, targets, and repository-relative paths;
- resolution of file and symbol targets against an Entire Graph full-profile snapshot;
- same-symbol, same-file, direct dependency, and bounded two-hop decisions;
- Graph evidence, confidence/resolution data, test targets, recommendations, and completeness caveats;
- a read-only Entire session adapter that exposes bounded public activity metadata while intentionally omitting prompt text;
- isolated session-provider health, so analysis remains available if Entire activity cannot be loaded;
- additive `entire graph coordinate` text and JSON output;
- a clearly labeled synthetic plan at `examples/spidey-plan.json`;
- focused engine and CLI tests covering dynamic participants, direct risk, same-file risk, two-hop review, clear results, invalid targets, and command output.

Run the current slice from source:

```bash
go run ./cmd/entire-graph coordinate \
  --repo . \
  --plan examples/spidey-plan.json \
  --format text
```

## Entire Graph findings and verification

Initial Graph searches located the repository's safe extension surfaces:

- `sem.BuildProviderSnapshotWithOptions` exposes an in-memory snapshot containing files, symbols, relations, warnings, completeness, and partial failures;
- the existing `impact` implementation demonstrates bounded relationship traversal and source-linked evidence;
- CLI commands are registered through the hand-written dispatcher and matching help registry.

These findings were checked against focused source reads. Spidey Sense will consume the existing snapshot model through a separate package and output schema. It will not change the frozen provider `1.x` record meanings or stable `compound-v1` symbol IDs. Before any risky implementation change, the affected symbol will receive an `entire graph impact` analysis. Final submission will include `entire graph diff` semantic analysis plus test and runtime verification.

## Noon Curveball: what changed and how we adapted

Not announced yet. The pre-Curveball architecture isolates inputs, policy, and presentation so a new provider, rule, output constraint, or reliability requirement can be added without replacing the coordination engine.

## Checkpoint links and what each checkpoint proves

1. **Initial understanding and architecture:** this document, architecture, scope, prior-work disclosure, and Curveball seams.
2. **Stable pre-Curveball implementation:** pending.
3. **Curveball response:** pending.
4. **Final implementation and verification:** pending.

Checkpoint IDs and links will be added as they are created.

## Setup, run, and test instructions

The first runnable backend slice is implemented on `progress`. The repository verification surface remains:

```bash
mise run build
mise run test
mise run check
```

## Prior planning and reused components

Prior to this Buildathon, the team produced a separate Python/React Spidey Sense prototype and detailed planning documents. Those materials informed the problem statement, terminology, visual direction, activity states, and security considerations.

The submitted implementation is being built inside the required `Preethesh16/entire-graph` fork cloned through the Entire India mirror. The prior implementation will not be copied wholesale, presented as Buildathon work, or used to replace Entire Graph. Any individual component reused later will be named here with its origin and modifications.

The abstract “codebase as a place” idea was also studied in Claude Clan. No Claude Clan source code or artwork is included.

## Security, privacy, and responsible monitoring

- Observe only public session metadata and repository activity; never private chain-of-thought.
- Treat prompts, file paths, identities, repository URLs, and activity as sensitive.
- Keep the initial service local and bind to loopback.
- Never expose arbitrary shell execution through the dashboard.
- Keep any future agent action explicit, allowlisted, and auditable.
- Never commit tokens, credentials, live activity files, or private session data.
- Surface Graph incompleteness and heuristic edges instead of claiming certainty.

## Visual identity

The interface uses an original spider-inspired night-radar aesthetic: web zones, strands, danger pulses, mission paths, and distinct runner archetypes. It will not use copyrighted Spider-Man characters, names, film assets, logos, or copied artwork. The theme is presentation only; every warning remains backed by engineering evidence and accessible as text.

## Known limitations and next steps

- The first demo targets sessions visible to one local coordinator; cross-machine live synchronization is future work.
- Agent-to-human identity mapping remains explicit.
- Static analysis is heuristic and may miss dynamic dispatch or runtime wiring.
- `CLEAR` means no risk was found within the analyzed scope, not proof of independence.
- GitHub network integration and direct agent messaging are intentionally outside the first stable slice.
