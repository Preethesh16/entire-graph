export type MissionStatus = 'queued' | 'active' | 'blocked' | 'complete';
export type RiskLevel = 'BLOCK' | 'REVIEW' | 'CLEAR';
export type ProviderStatus = 'online' | 'degraded' | 'offline';
export type RunnerStatus = 'active' | 'idle' | 'offline';

export interface Team {
  id: string;
  name: string;
  callSign: string;
  objective: string;
}

export interface Runner {
  id: string;
  displayName: string;
  role: string;
  archetype: string;
  accent: string;
  initials: string;
  status: RunnerStatus;
  provider: string;
  currentMissionId?: string;
}

export interface Mission {
  id: string;
  title: string;
  intent: string;
  ownerId?: string;
  status: MissionStatus;
  targetCount: number;
  updatedAt: string;
  targets: string[];
}

export interface EvidenceNode {
  id: string;
  label: string;
  kind: 'file' | 'symbol' | 'test';
  location: string;
}

export interface EvidenceRelation {
  id: string;
  from: string;
  to: string;
  type: string;
  confidence: number;
  resolution: 'resolved' | 'heuristic' | 'partial';
}

export interface RiskSignal {
  id: string;
  level: RiskLevel;
  title: string;
  summary: string;
  missionIds: string[];
  evidenceNodes: EvidenceNode[];
  relations: EvidenceRelation[];
  recommendation: string;
  reviewTargets: string[];
}

export interface ProviderHealth {
  id: string;
  name: string;
  status: ProviderStatus;
  detail: string;
  latencyMs?: number;
  lastCheckedAt: string;
}

export interface GitActivity {
  id: string;
  actor: string;
  action: string;
  ref: string;
  timestamp: string;
}

export interface GitSnapshot {
	available?: boolean;
  branch: string;
  head: string;
  dirtyFileCount: number;
  ahead: number;
  behind: number;
  activity: GitActivity[];
}

export interface GraphSnapshot {
  nodeCount: number;
  relationCount: number;
  analyzedDepth: number;
  complete: boolean;
  warning?: string;
}

export interface DashboardSnapshot {
  generatedAt: string;
  team: Team;
  runners: Runner[];
  missions: Mission[];
  risks: RiskSignal[];
  providers: ProviderHealth[];
  git: GitSnapshot;
  graph: GraphSnapshot;
  checkpoints?: Array<{ id: string; message: string; date?: string; sessionId: string; condensationId?: string }>;
}
