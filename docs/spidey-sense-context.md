# Spidey Sense initial context and decisions

## Buildathon context

- Event: Bengaluru Tech Week Buildathon 2026
- Track: E2 / Problem Statement 2 — Build with Graph Intelligence
- Submission repository: `Preethesh16/entire-graph`
- Working clone: Entire India mirror
- Product implementation begins in this fork after the event start.

## Product intent

Spidey Sense helps a lead coordinate several humans and coding agents working on one repository. The lead creates a plan and divides it into missions. Team members connect Entire-enabled Codex or Claude Code sessions. The product visualizes declared progress and observable activity, detects coordination risks using Entire Graph, explains why the risk exists, and recommends the next safe action.

The gamified presentation uses missions, runners, web zones, strands, and danger signals. The engineering model remains explicit and scriptable underneath the theme.

## Decisions made before implementation

### Entire Graph is the graph engine

The prior prototype contains a lightweight Python/JavaScript/TypeScript dependency parser. Rebuilding or copying that parser would duplicate the selected track's central technology and make Entire optional. The Buildathon version instead consumes Entire Graph files, symbols, relations, evidence, semantic diffs, completeness, and partial failures.

### Monitoring is deterministic first

The core monitor is not another autonomous AI agent. It is a coordinator service that polls documented public interfaces, normalizes observations, and evaluates deterministic Graph rules. This makes warnings repeatable and judge-verifiable. A later AI layer may summarize evidence but cannot create unsupported risk claims.

### Public activity only

The system may display bounded statuses, timestamps, worktrees, and touched-file paths exposed by supported tools. It never transports prompt text or reasoning and never claims to know what a model is privately thinking.

### Plan plus observation

Session telemetry alone cannot reliably identify the human owner or intended scope. The lead's plan remains an explicit source of intent. Observed session/Git activity can confirm or challenge that plan but does not silently overwrite it.

### Shared coordinator with local-first metadata

A loopback dashboard remains available for local development. Real team membership, invites, browser presence, and live rosters use the shared deployed or explicitly LAN-bound coordinator directly from the website. A local connector is optional for richer Entire/Git metadata that browser sandboxing cannot inspect. Both paths use allowlisted wire contracts and never send prompts, reasoning, terminal output, file content, or secrets.

### Original visual identity

The requested Spider-Man feeling is implemented as an original spider-sense radar language. Copyrighted characters, names, logos, film stills, costumes, and third-party artwork are excluded. Dynamic runner archetypes provide visual differentiation without impersonating licensed characters.

## Prior work disclosure

The supplied `Track Builder.zip` contains a previously completed Spidey Sense prototype with Python backend modules, a React/Three.js frontend, tests, generated dependencies, documentation, and Git history. It predates the Buildathon implementation window.

It is reference material only. It contributed:

- problem framing and user stories;
- the mission/pathway terminology;
- same-file and directed dependency blocker concepts;
- local-first and safe-directive boundaries;
- visual ideas for zones, nodes, strands, and danger pulses;
- example data contracts and test cases to reconsider.

It will not be copied wholesale or presented as newly built work. In particular, its custom graph parser is rejected for the Buildathon implementation because Entire Graph must be essential. If a small component is later reused, `BUILDATHON.md` will identify the exact source and modification.

The general “codebase as a place” interaction pattern was studied in Claude Clan. No source code or artwork from that project is included.

## Must-have scope

1. Dynamic team and mission plan.
2. At least two Entire-tracked sessions in the demo or a clearly labeled recorded fixture.
3. Entire Graph target resolution and collision evidence.
4. Same-target and structural-risk decisions.
5. Exact paths and source locations.
6. Test/review/sequencing recommendation.
7. Local dashboard with accessible textual evidence.
8. End-to-end tests for the critical decision loop.

## Deferred scope

- production authentication and hosted collaboration;
- production identity federation and internet-scale hosting operations;
- direct messaging to every agent provider;
- automatic merge/conflict resolution;
- GitHub REST synchronization;
- arbitrary shell access;
- historical analytics;
- decorative 3D assets that do not improve evidence comprehension;
- unsupported AI risk scoring.

## Open questions to resolve during implementation

- Which stable plan-status vocabulary best maps to both technical and themed labels?
- Which relation families are blocking versus review-only in the first policy?
- How should a mission identify an ambiguous symbol without exposing unstable UI complexity?
- Which test recommendation types can be executed automatically versus presented for review?
- Which recorded demo fixture best proves graceful fallback without overstating liveness?

## Checkpoint strategy

1. Initial understanding, architecture, scope, prior-work disclosure, and risks.
2. Last stable runnable version before the Noon Curveball.
3. Minimal complete Curveball response with impact analysis and tests.
4. Final verified implementation with semantic diff and limitations.

Each checkpoint should capture the decision context and evidence, not merely increase the count.
