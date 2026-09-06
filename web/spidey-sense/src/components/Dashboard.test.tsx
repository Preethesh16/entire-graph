import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DashboardSnapshot, Runner } from '../domain';
import { syntheticDashboardFixture } from '../fixtures/syntheticDashboardFixture';
import { Dashboard } from './Dashboard';

function snapshotWith(overrides: Partial<DashboardSnapshot>): DashboardSnapshot {
  return { ...syntheticDashboardFixture, ...overrides };
}

describe('Dashboard', () => {
  it('renders zero members and zero missions as useful empty states', () => {
    render(<Dashboard snapshot={snapshotWith({ runners: [], missions: [], risks: [] })} />);

    expect(screen.getByText('No runners connected')).toBeInTheDocument();
    expect(screen.getByText('No missions planned')).toBeInTheDocument();
  });

  it('renders one dynamically supplied member', () => {
    const runner: Runner = {
      id: 'solo-runner',
      displayName: 'Solo Signal',
      role: 'Release engineer',
      archetype: 'Night Watch',
      accent: '#44d7b6',
      initials: 'SS',
      status: 'idle',
      provider: 'Codex',
    };

    render(<Dashboard snapshot={snapshotWith({ runners: [runner], missions: [] })} />);

    expect(screen.getByRole('heading', { name: 'Solo Signal' })).toBeInTheDocument();
    expect(screen.getByText('Release engineer')).toBeInTheDocument();
    expect(screen.getByText('Awaiting assignment')).toBeInTheDocument();
  });

  it('renders multiple supplied members and missions without fixed counts', () => {
    render(<Dashboard snapshot={syntheticDashboardFixture} />);

    for (const runner of syntheticDashboardFixture.runners) {
      expect(screen.getByRole('heading', { name: runner.displayName })).toBeInTheDocument();
    }
    for (const mission of syntheticDashboardFixture.missions) {
      expect(screen.getByRole('heading', { name: mission.title })).toBeInTheDocument();
    }
    expect(screen.getByText(`${syntheticDashboardFixture.missions.length} total pathways`)).toBeInTheDocument();
  });

  it('shows blocker evidence, relation types, locations, and recommended targets', () => {
    const blocker = syntheticDashboardFixture.risks.find((risk) => risk.level === 'BLOCK')!;
    render(<Dashboard snapshot={syntheticDashboardFixture} />);

    expect(screen.getAllByText(blocker.title).length).toBeGreaterThan(0);
    expect(screen.getAllByText(blocker.evidenceNodes[0].location).length).toBeGreaterThan(0);
    expect(screen.getByText(blocker.relations[0].type)).toBeInTheDocument();
    expect(screen.getByText(blocker.recommendation)).toBeInTheDocument();
    expect(screen.getByText(blocker.reviewTargets[0])).toBeInTheDocument();
  });

  it('surfaces provider failure without crashing the dashboard', () => {
    const failingProvider = {
      ...syntheticDashboardFixture.providers[0],
      status: 'offline' as const,
      detail: 'Session provider returned an unavailable response',
    };
    render(<Dashboard snapshot={snapshotWith({ providers: [failingProvider] })} />);

    expect(screen.getByText('Session provider returned an unavailable response')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Work in motion' })).toBeInTheDocument();
  });

  it('distinguishes incomplete Graph evidence and exposes a verification path', () => {
    const incompleteRisk = {
      ...syntheticDashboardFixture.risks[1],
      evidenceClass: 'incomplete' as const,
      verificationRequired: true,
      verification: ['Inspect the generated registry and run the dispatcher integration test.'],
    };
    render(<Dashboard snapshot={snapshotWith({
      risks: [incompleteRisk],
      graph: { ...syntheticDashboardFixture.graph, complete: false, warningCount: 1, partialFailureCount: 1, warning: 'Generated dispatch targets were not indexed' },
    })} />);

    expect(screen.getByText(/Evidence:/)).toHaveTextContent('incomplete');
    expect(screen.getByText(/source or test verification required/)).toBeInTheDocument();
    expect(screen.getByText(incompleteRisk.verification[0])).toBeInTheDocument();
    expect(screen.getByText('1 warnings · 1 partial failures')).toBeInTheDocument();
  });

  it('renders stable loading and aggregate error states', () => {
    const { rerender } = render(<Dashboard loading />);
    expect(screen.getByText('Mapping the mission web…')).toBeInTheDocument();

    rerender(<Dashboard error="Graph provider did not respond" />);
    expect(screen.getByRole('alert')).toHaveTextContent('Graph provider did not respond');
  });
});
