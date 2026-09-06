import { describe, expect, it } from 'vitest';
import { toDashboardSnapshot } from './api';

describe('coordinate report adapter', () => {
  it('maps dynamic plan, sessions, and Graph evidence into dashboard data', () => {
    const snapshot = toDashboardSnapshot({
      schema_version: 'spidey-sense/v1alpha1',
      graph: { provider: 'entire-graph', provider_version: 'test', commit: '1234567890', profile: 'full', completeness_level: 'ok', node_count: 4, relation_count: 1, warning_count: 0, partial_failure_count: 0, analysis_partial: false },
      team: { id: 'dynamic', name: 'Dynamic team', members: [{ id: 'runner', name: 'Ada Lovelace', session_ids: ['session-1'] }] },
      missions: [
        { id: 'one', title: 'One', owner: 'runner', status: 'active', targets: [{ file: 'one.go', symbol: 'One' }] },
        { id: 'two', title: 'Two', owner: 'runner', status: 'queued', targets: [{ file: 'two.go', symbol: 'Two' }] },
      ],
      decisions: [{
        level: 'BLOCK', reason: 'direct dependency', mission_a: 'one', mission_b: 'two',
        path: [{ from: { id: 'one', name: 'One', kind: 'function', file: 'one.go', start_line: 2 }, to: { id: 'two', name: 'Two', kind: 'function', file: 'two.go', start_line: 3 }, relation: 'CALLS', confidence: 1, resolution: 'exact', evidence_class: 'confirmed' }],
        recommendations: ['Sequence missions.'], test_targets: ['one_test.go'], evidence_class: 'confirmed', verification_required: false, verification: ['Run one_test.go'],
      }],
      sessions: [{ session_id: 'session-1', agent: 'Codex', status: 'active' }],
      session_health: { available: true },
    });
    expect(snapshot.runners[0]).toMatchObject({ displayName: 'Ada Lovelace', status: 'active', provider: 'Codex' });
    expect(snapshot.risks[0]).toMatchObject({ level: 'BLOCK', reviewTargets: ['one_test.go'] });
    expect(snapshot.risks[0].relations[0]).toMatchObject({ type: 'CALLS', resolution: 'resolved' });
    expect(snapshot.risks[0]).toMatchObject({ evidenceClass: 'confirmed', verificationRequired: false });
    expect(snapshot.graph).toMatchObject({ nodeCount: 4, relationCount: 1, complete: true });
  });
});
