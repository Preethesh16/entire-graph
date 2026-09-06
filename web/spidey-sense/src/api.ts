import type { DashboardSnapshot, EvidenceNode, EvidenceRelation, RiskSignal } from './domain';

export interface CoordinatePlan {
  schema_version: string;
  revision: number;
  team: CoordinateReport['team'];
  missions: CoordinateReport['missions'];
}

export interface PublicSession { session_id: string; agent: string; model?: string; status: string; worktree?: string }

interface CoordinateReport {
  schema_version: string;
  graph: {
    provider: string;
    provider_version: string;
    commit?: string;
    profile: string;
    completeness_level: string;
    node_count: number;
    relation_count: number;
  };
  team: {
    id: string;
    name: string;
    call_sign?: string;
    objective?: string;
    members: Array<{ id: string; name: string; role?: string; archetype?: string; accent?: string; session_ids?: string[] }>;
  };
  missions: Array<{ id: string; title: string; intent?: string; owner: string; status: DashboardSnapshot['missions'][number]['status']; targets: Array<{ file?: string; symbol?: string }> }>;
  decisions: Array<{
    level: RiskSignal['level']; reason: string; mission_a: string; mission_b: string;
    path?: Array<{ from: ReportEndpoint; to: ReportEndpoint; relation: string; confidence: number; resolution?: string }>;
    test_targets?: string[]; recommendations: string[]; caveat?: string;
  }>;
  sessions?: Array<{ session_id: string; agent: string; status: string }>;
  session_health: { available: boolean; detail?: string };
}

interface ReportEndpoint { id: string; name: string; kind: string; file?: string; start_line?: number }

const accents = ['#ff395d', '#41d9ff', '#f5c451', '#ad7cff', '#55e69d'];

function endpointNode(endpoint: ReportEndpoint): EvidenceNode {
  return {
    id: endpoint.id || `${endpoint.file}:${endpoint.name}`,
    label: endpoint.name,
    kind: endpoint.kind === 'file' ? 'file' : 'symbol',
    location: `${endpoint.file ?? ''}${endpoint.start_line ? `:${endpoint.start_line}` : ''}`,
  };
}

export function toDashboardSnapshot(report: CoordinateReport): DashboardSnapshot {
  const sessions = new Map((report.sessions ?? []).map((session) => [session.session_id, session]));
  const risks = report.decisions.map((decision, riskIndex): RiskSignal => {
    const nodes = new Map<string, EvidenceNode>();
    const relations: EvidenceRelation[] = [];
    for (const [index, step] of (decision.path ?? []).entries()) {
      const from = endpointNode(step.from);
      const to = endpointNode(step.to);
      nodes.set(from.id, from);
      nodes.set(to.id, to);
      relations.push({
        id: `${riskIndex}-${index}`,
        from: from.id,
        to: to.id,
        type: step.relation,
        confidence: step.confidence,
        resolution: step.resolution === 'exact' || step.resolution === 'resolved' ? 'resolved' : step.resolution ? 'heuristic' : 'partial',
      });
    }
    return {
      id: `${decision.mission_a}:${decision.mission_b}`,
      level: decision.level,
      title: `${decision.mission_a} ↔ ${decision.mission_b}`,
      summary: decision.caveat ? `${decision.reason}. ${decision.caveat}` : decision.reason,
      missionIds: [decision.mission_a, decision.mission_b],
      evidenceNodes: [...nodes.values()],
      relations,
      recommendation: decision.recommendations.join(' '),
      reviewTargets: decision.test_targets ?? [],
    };
  });
  const generatedAt = new Date().toISOString();
  return {
    generatedAt,
    team: {
      id: report.team.id,
      name: report.team.name,
      callSign: report.team.call_sign ?? report.team.name,
      objective: report.team.objective ?? 'Coordinate repository missions using verifiable Graph evidence.',
    },
    runners: report.team.members.map((member, index) => {
      const linked = (member.session_ids ?? []).map((id) => sessions.get(id)).filter(Boolean);
      const currentMission = report.missions.find((mission) => mission.owner === member.id && mission.status === 'active');
      return {
        id: member.id,
        displayName: member.name,
        role: member.role ?? 'Builder',
        archetype: member.archetype ?? 'Web Runner',
        accent: member.accent ?? accents[index % accents.length],
        initials: member.name.split(/\s+/).map((part) => part[0]).join('').slice(0, 2).toUpperCase(),
        status: linked.some((session) => session?.status === 'active') ? 'active' : linked.length ? 'idle' : 'offline',
        provider: linked.map((session) => session?.agent).filter(Boolean).join(' + ') || 'Not connected',
        currentMissionId: currentMission?.id,
      };
    }),
    missions: report.missions.map((mission) => ({
      id: mission.id, title: mission.title, intent: mission.intent ?? '', ownerId: mission.owner,
      status: mission.status, targetCount: mission.targets.length, updatedAt: generatedAt,
      targets: mission.targets.map((target) => target.symbol ? `${target.symbol}${target.file ? ` · ${target.file}` : ''}` : target.file ?? ''),
    })),
    risks,
    providers: [
      { id: 'graph', name: `${report.graph.provider} ${report.graph.provider_version}`, status: report.graph.completeness_level === 'complete' ? 'online' : 'degraded', detail: `${report.graph.profile} profile · ${report.graph.completeness_level}`, lastCheckedAt: generatedAt },
      { id: 'entire', name: 'Entire sessions', status: report.session_health.available ? 'online' : 'offline', detail: report.session_health.detail ?? `${report.sessions?.length ?? 0} mapped sessions`, lastCheckedAt: generatedAt },
    ],
    git: { available: false, branch: '', head: '', dirtyFileCount: 0, ahead: 0, behind: 0, activity: [] },
    graph: { nodeCount: report.graph.node_count, relationCount: report.graph.relation_count, analyzedDepth: 2, complete: report.graph.completeness_level === 'complete', warning: report.graph.completeness_level === 'complete' ? undefined : `Graph is ${report.graph.completeness_level}` },
  };
}

export async function loadDashboard(signal?: AbortSignal): Promise<DashboardSnapshot> {
  const response = await fetch('/api/v1/report', { signal, headers: { Accept: 'application/json' } });
  if (!response.ok) throw new Error(`Mission API returned ${response.status}`);
  return toDashboardSnapshot(await response.json() as CoordinateReport);
}

async function apiJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { ...init, headers: { Accept: 'application/json', ...init?.headers } });
  if (!response.ok) throw new Error(`Mission API returned ${response.status}`);
  return response.json() as Promise<T>;
}

export const loadPlan = (signal?: AbortSignal) => apiJSON<CoordinatePlan>('/api/v1/plan', { signal });
export const loadSessions = (signal?: AbortSignal) => apiJSON<{ sessions: PublicSession[] }>('/api/v1/sessions', { signal });
export const savePlan = (plan: CoordinatePlan) => apiJSON<CoordinatePlan>('/api/v1/plan', {
  method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(plan),
});
