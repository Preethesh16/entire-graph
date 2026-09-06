import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { CoordinatePlan } from '../api';
import type { DashboardSnapshot } from '../domain';
import { Workflow } from './Workflow';

const plan: CoordinatePlan = {
  schema_version: 'spidey-sense/v1alpha1', revision: 1,
  team: { id: 'hardcoders', name: 'HardCoders', members: [{ id: 'lead', name: 'Preethesh' }] },
  missions: [{ id: 'quotes', title: 'Quote engine', intent: 'Verify quotes', owner: 'lead', status: 'active', targets: [{ file: 'apps/api/src/fx/quote.ts', symbol: 'buildQuote' }] }],
};

const snapshot: DashboardSnapshot = {
  generatedAt: new Date(0).toISOString(),
  team: { id: 'hardcoders', name: 'HardCoders', callSign: 'Anchor', objective: 'Ship safely' },
  runners: [], missions: [], risks: [], providers: [], checkpoints: [],
  git: { branch: '', head: '', dirtyFileCount: 0, ahead: 0, behind: 0, activity: [] },
  graph: { nodeCount: 1, relationCount: 1, analyzedDepth: 2, complete: false, warningCount: 1, partialFailureCount: 1 },
};

describe('production assignment workflow', () => {
  it('makes authenticated live members assignable and persists their stable member id', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<Workflow
      snapshot={snapshot} plan={plan} sessions={[]} liveAgents={[{
        id: 'agent-deepthi', team_id: 'team-live', member_id: 'member-deepthi', name: 'Deepthi', role: 'UI engineer',
        session_id: 'browser-1', agent: 'Browser presence', provider: 'Chrome', online: true,
        last_heartbeat: new Date(0).toISOString(),
      }]}
      connectionBusy={false} presence={{}} onCreateTeam={vi.fn()} onJoinTeam={vi.fn()} onPresence={vi.fn()}
      onLeaveTeam={vi.fn()} onRefresh={vi.fn()} onSave={onSave}
    />);

    fireEvent.click(screen.getByRole('button', { name: '2. Plan & assign' }));
    expect(screen.getByRole('option', { name: 'Deepthi · online' })).toHaveValue('member-deepthi');
    fireEvent.change(screen.getByLabelText('Owner for Quote engine'), { target: { value: 'member-deepthi' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save & analyze' }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
    const saved = onSave.mock.calls[0][0] as CoordinatePlan;
    expect(saved.missions[0].owner).toBe('member-deepthi');
    expect(saved.team.members).toContainEqual(expect.objectContaining({ id: 'member-deepthi', name: 'Deepthi' }));
  });
});
