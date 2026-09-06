export interface ConnectedAgent {
  id: string;
  team_id: string;
  member_id: string;
  name: string;
  role?: string;
  session_id: string;
  agent: string;
  provider: string;
  model?: string;
  online: boolean;
  mission_id?: string;
  branch?: string;
  changed_files?: string[];
  checkpoint_id?: string;
  blocker?: string;
  last_activity?: string;
  last_heartbeat: string;
}

interface TeamCredentials {
  team_id: string;
  invite_code: string;
  admin_token: string;
  member_id: string;
  connector_token: string;
  viewer_token: string;
}

interface JoinCredentials {
  team_id: string;
  member_id: string;
  connector_token: string;
  viewer_token: string;
}

interface AgentCredentials { agent_id: string; agent_token: string }

export interface BrowserTeamSession {
  teamId: string;
  memberId: string;
  name: string;
  role: string;
  viewerToken: string;
  connectorToken: string;
  agentId: string;
  agentToken: string;
  sessionId: string;
  inviteCode?: string;
  adminToken?: string;
}

export interface PresenceInput { missionId?: string; blocker?: string }

const storageKey = 'spidey-sense.browser-team.v1';

async function requestJSON<T>(path: string, init: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { Accept: 'application/json', ...init.headers },
  });
  const payload = await response.json().catch(() => undefined) as { error?: string } | undefined;
  if (!response.ok) throw new Error(payload?.error ?? `Team service returned ${response.status}`);
  return payload as T;
}

function detectedBrowser(): string {
  const brands = (navigator as Navigator & { userAgentData?: { brands?: Array<{ brand: string }> } }).userAgentData?.brands;
  const named = brands?.map((entry) => entry.brand).find((brand) => !brand.toLowerCase().includes('not'));
  if (named) return named;
  if (/Firefox\//.test(navigator.userAgent)) return 'Firefox';
  if (/Edg\//.test(navigator.userAgent)) return 'Microsoft Edge';
  if (/Chrome\//.test(navigator.userAgent)) return 'Chrome';
  if (/Safari\//.test(navigator.userAgent)) return 'Safari';
  return 'Web browser';
}

async function connectBrowser(input: {
  teamId: string; memberId: string; connectorToken: string; viewerToken: string;
  name: string; role: string; inviteCode?: string; adminToken?: string;
}): Promise<BrowserTeamSession> {
  const sessionId = `browser_${crypto.randomUUID()}`;
  const connected = await requestJSON<AgentCredentials>('/api/v1/agents/connect', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${input.connectorToken}` },
    body: JSON.stringify({
      team_id: input.teamId,
      member_id: input.memberId,
      session_id: sessionId,
      agent: 'Browser presence',
      provider: detectedBrowser(),
    }),
  });
  const session: BrowserTeamSession = {
    ...input,
    sessionId,
    agentId: connected.agent_id,
    agentToken: connected.agent_token,
  };
  sessionStorage.setItem(storageKey, JSON.stringify(session));
  return session;
}

export async function createBrowserTeam(input: { teamName: string; objective: string; name: string; role: string }): Promise<BrowserTeamSession> {
  const created = await requestJSON<TeamCredentials>('/api/v1/teams', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: input.teamName, objective: input.objective, leader_name: input.name, leader_role: input.role }),
  });
  return connectBrowser({
    teamId: created.team_id, memberId: created.member_id, connectorToken: created.connector_token,
    viewerToken: created.viewer_token, inviteCode: created.invite_code, adminToken: created.admin_token,
    name: input.name, role: input.role,
  });
}

export async function joinBrowserTeam(input: { teamId: string; inviteCode: string; name: string; role: string }): Promise<BrowserTeamSession> {
  const joined = await requestJSON<JoinCredentials>(`/api/v1/teams/${encodeURIComponent(input.teamId)}/join`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ invite_code: input.inviteCode, name: input.name, role: input.role }),
  });
  return connectBrowser({
    teamId: joined.team_id, memberId: joined.member_id, connectorToken: joined.connector_token,
    viewerToken: joined.viewer_token, name: input.name, role: input.role,
  });
}

export function restoreBrowserTeam(): BrowserTeamSession | undefined {
  try {
    const value = sessionStorage.getItem(storageKey);
    return value ? JSON.parse(value) as BrowserTeamSession : undefined;
  } catch {
    sessionStorage.removeItem(storageKey);
    return undefined;
  }
}

export function leaveBrowserTeam(): void { sessionStorage.removeItem(storageKey); }

export async function sendBrowserHeartbeat(session: BrowserTeamSession, presence: PresenceInput = {}): Promise<ConnectedAgent> {
  return requestJSON<ConnectedAgent>(`/api/v1/agents/${encodeURIComponent(session.agentId)}/heartbeat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${session.agentToken}` },
    body: JSON.stringify({
      mission_id: presence.missionId ?? '', blocker: presence.blocker ?? '', last_activity: new Date().toISOString(),
    }),
  });
}

export async function loadTeamAgents(session: BrowserTeamSession, signal?: AbortSignal): Promise<ConnectedAgent[]> {
  const result = await requestJSON<{ agents: ConnectedAgent[] }>(`/api/v1/teams/${encodeURIComponent(session.teamId)}/agents`, {
    method: 'GET', signal, headers: { Authorization: `Bearer ${session.viewerToken}` },
  });
  return result.agents;
}

export async function watchTeamEvents(session: BrowserTeamSession, signal: AbortSignal, onEvent: () => void): Promise<void> {
  const response = await fetch(`/api/v1/teams/${encodeURIComponent(session.teamId)}/events`, {
    signal, headers: { Accept: 'text/event-stream', Authorization: `Bearer ${session.viewerToken}` },
  });
  if (!response.ok || !response.body) throw new Error(`Live team stream returned ${response.status}`);
  const reader = response.body.pipeThrough(new TextDecoderStream()).getReader();
  let buffered = '';
  while (!signal.aborted) {
    const { done, value } = await reader.read();
    if (done) return;
    buffered += value;
    let boundary = buffered.indexOf('\n\n');
    while (boundary >= 0) {
      const event = buffered.slice(0, boundary);
      buffered = buffered.slice(boundary + 2);
      if (/^event: agent\./m.test(event)) onEvent();
      boundary = buffered.indexOf('\n\n');
    }
  }
}
