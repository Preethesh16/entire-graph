import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createBrowserTeam, joinBrowserTeam, loadTeamAgents, restoreBrowserTeam } from './network';

const jsonResponse = (value: unknown, status = 200) => new Response(JSON.stringify(value), { status, headers: { 'Content-Type': 'application/json' } });

describe('browser team connectivity', () => {
  beforeEach(() => {
    sessionStorage.clear();
    vi.restoreAllMocks();
  });

  it('creates and connects a browser presence without private local metadata', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(jsonResponse({ team_id: 'team-1', invite_code: 'invite-1', admin_token: 'admin-1', member_id: 'member-1', connector_token: 'connector-1', viewer_token: 'viewer-1' }, 201))
      .mockResolvedValueOnce(jsonResponse({ agent_id: 'agent-1', agent_token: 'agent-token-1' }, 201));

    const session = await createBrowserTeam({ teamName: 'Production', objective: 'Coordinate', name: 'Preethesh', role: 'Lead' });

    expect(session).toMatchObject({ teamId: 'team-1', memberId: 'member-1', inviteCode: 'invite-1', viewerToken: 'viewer-1', agentId: 'agent-1' });
    const connectPayload = String(fetchMock.mock.calls[1][1]?.body);
    expect(connectPayload).toContain('Browser presence');
    for (const forbidden of ['prompt', 'reasoning', 'terminal', 'changed_files', 'checkpoint_id', 'branch']) expect(connectPayload).not.toContain(forbidden);
    expect(restoreBrowserTeam()).toMatchObject({ teamId: 'team-1', agentId: 'agent-1' });
  });

  it('uses a member-scoped viewer credential for the shared roster', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(jsonResponse({ team_id: 'team-2', member_id: 'member-2', connector_token: 'connector-2', viewer_token: 'viewer-2' }, 201))
      .mockResolvedValueOnce(jsonResponse({ agent_id: 'agent-2', agent_token: 'agent-token-2' }, 201))
      .mockResolvedValueOnce(jsonResponse({ agents: [{ id: 'agent-2', name: 'Deepthi', online: true }] }));
    const session = await joinBrowserTeam({ teamId: 'team-2', inviteCode: 'invite-2', name: 'Deepthi', role: 'UI engineer' });
    const agents = await loadTeamAgents(session);

    expect(agents[0]).toMatchObject({ name: 'Deepthi', online: true });
    expect(fetchMock.mock.calls[2][1]?.headers).toMatchObject({ Authorization: 'Bearer viewer-2' });
  });
});
