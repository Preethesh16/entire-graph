import type { DashboardSnapshot, EvidenceClass, EvidenceNode, EvidenceRelation, RiskSignal } from './domain';

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
    warning_count?: number;
    partial_failure_count?: number;
    analysis_partial?: boolean;
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
    path?: Array<{ from: ReportEndpoint; to: ReportEndpoint; relation: string; confidence: number; resolution?: string; evidence_class?: EvidenceClass }>;
    test_targets?: string[]; recommendations: string[]; caveat?: string;
    evidence_class?: EvidenceClass; verification_required?: boolean; verification?: string[];
  }>;
  sessions?: Array<{ session_id: string; agent: string; status: string }>;
  session_health: { available: boolean; detail?: string };
  checkpoints?: Array<{ id: string; message: string; date?: string; session_id: string; condensation_id?: string }>;
  checkpoint_health?: { available: boolean; detail?: string };
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
        resolution: ['exact', 'resolved', 'package', 'import_resolved'].includes(step.resolution ?? '') ? 'resolved' : step.resolution ? 'heuristic' : 'partial',
        evidenceClass: step.evidence_class ?? (['exact', 'resolved', 'package', 'import_resolved'].includes(step.resolution ?? '') ? 'confirmed' : step.resolution ? 'heuristic' : 'incomplete'),
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
      evidenceClass: decision.evidence_class ?? 'incomplete',
      verificationRequired: decision.verification_required ?? true,
      verification: decision.verification ?? ['Inspect source and run focused tests before relying on this claim.'],
    };
  });
  const generatedAt = new Date().toISOString();
  const graphAnalysisPartial = report.graph.analysis_partial ?? (
    !['ok', 'complete'].includes(report.graph.completeness_level)
    || (report.graph.partial_failure_count ?? 0) > 0
  );
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
      { id: 'graph', name: `${report.graph.provider} ${report.graph.provider_version}`, status: graphAnalysisPartial ? 'degraded' : 'online', detail: `${report.graph.profile} profile · ${report.graph.completeness_level}`, lastCheckedAt: generatedAt },
      { id: 'entire', name: 'Entire sessions', status: report.session_health.available ? 'online' : 'offline', detail: report.session_health.detail ?? `${report.sessions?.length ?? 0} mapped sessions`, lastCheckedAt: generatedAt },
      { id: 'checkpoints', name: 'Entire checkpoints', status: report.checkpoint_health?.available ? 'online' : 'offline', detail: report.checkpoint_health?.detail ?? `${report.checkpoints?.length ?? 0} team checkpoints`, lastCheckedAt: generatedAt },
    ],
    git: { available: false, branch: '', head: '', dirtyFileCount: 0, ahead: 0, behind: 0, activity: [] },
    graph: {
      nodeCount: report.graph.node_count, relationCount: report.graph.relation_count, analyzedDepth: 2,
      complete: !graphAnalysisPartial,
      warning: !graphAnalysisPartial
        ? undefined
        : `Graph is ${report.graph.completeness_level}; ${report.graph.warning_count ?? 0} warnings, ${report.graph.partial_failure_count ?? 0} partial failures`,
      warningCount: report.graph.warning_count ?? 0,
      partialFailureCount: report.graph.partial_failure_count ?? 0,
    },
    checkpoints: (report.checkpoints ?? []).map((checkpoint) => ({ id: checkpoint.id, message: checkpoint.message, date: checkpoint.date, sessionId: checkpoint.session_id, condensationId: checkpoint.condensation_id })),
  };
}

export async function loadDashboard(signal?: AbortSignal): Promise<DashboardSnapshot> {
  const response = await fetch('/api/v1/report', { signal, headers: { Accept: 'application/json' } });
  if (!response.ok) throw new Error(`Mission API returned ${response.status}`);
  return toDashboardSnapshot(await response.json() as CoordinateReport);
}

export async function refreshDashboard(): Promise<DashboardSnapshot> {
  const response = await fetch('/api/v1/refresh', { method: 'POST', headers: { Accept: 'application/json' } });
  if (!response.ok) throw new Error(`Graph refresh returned ${response.status}`);
  return toDashboardSnapshot(await response.json() as CoordinateReport);
}

async function apiJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { ...init, headers: { Accept: 'application/json', ...init?.headers } });
  if (!response.ok) throw new Error(`Mission API returned ${response.status}`);
  return response.json() as Promise<T>;
}

export const loadPlan = async (signal?: AbortSignal) => {
  const plan = await apiJSON<CoordinatePlan>('/api/v1/plan', { signal });
  return {
    ...plan,
    team: { ...plan.team, members: plan.team.members ?? [] },
    missions: plan.missions ?? [],
  };
};

export const loadSessions = async (signal?: AbortSignal) => {
  const payload = await apiJSON<{ sessions: PublicSession[] | null }>('/api/v1/sessions', { signal });
  return { ...payload, sessions: payload.sessions ?? [] };
};
export const savePlan = (plan: CoordinatePlan) => apiJSON<CoordinatePlan>('/api/v1/plan', {
  method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(plan),
});
