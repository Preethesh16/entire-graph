import { useMemo, useState } from 'react';
import type { DashboardSnapshot, RiskLevel } from '../domain';
import { EvidencePath } from './EvidencePath';
import { MissionPathways } from './MissionPathways';
import { RiskConsole } from './RiskConsole';
import { GitPanel, GraphPanel, ProviderPanel } from './SystemPanels';
import { TeamLobby } from './TeamLobby';

interface DashboardProps {
  snapshot?: DashboardSnapshot;
  loading?: boolean;
  error?: string;
}

const riskPriority: Record<RiskLevel, number> = { BLOCK: 0, REVIEW: 1, CLEAR: 2 };

export function Dashboard({ snapshot, loading = false, error }: DashboardProps) {
  const orderedRisks = useMemo(
    () => [...(snapshot?.risks ?? [])].sort((left, right) => riskPriority[left.level] - riskPriority[right.level]),
    [snapshot?.risks],
  );
  const [selectedRiskId, setSelectedRiskId] = useState<string | undefined>();
  const selectedRisk = orderedRisks.find((risk) => risk.id === selectedRiskId) ?? orderedRisks[0];

  if (loading) {
    return (
      <main className="state-shell" aria-busy="true" aria-live="polite">
        <div className="loading-radar" aria-hidden="true"><span /></div>
        <p className="eyebrow">Calibrating signal network</p>
        <h1>Mapping the mission web…</h1>
        <p>Resolving sessions, Git activity, and bounded Graph relationships.</p>
      </main>
    );
  }

  if (error || !snapshot) {
    return (
      <main className="state-shell error-shell" role="alert">
        <span className="error-signal" aria-hidden="true">!</span>
        <p className="eyebrow">Provider signal lost</p>
        <h1>Mission control is temporarily offline</h1>
        <p>{error ?? 'No dashboard snapshot is available.'}</p>
        <p className="state-detail">The interface remains stable while providers recover.</p>
      </main>
    );
  }

  const activeMissions = snapshot.missions.filter((mission) => mission.status === 'active').length;
  const blockedMissions = snapshot.missions.filter((mission) => mission.status === 'blocked').length;
  const liveRunners = snapshot.runners.filter((runner) => runner.status === 'active').length;
  const providerOnline = snapshot.providers.filter((provider) => provider.status === 'online').length;

  return (
    <div className="app-shell">
      <header className="topbar">
        <a className="brand" href="#mission-control" aria-label="Spidey Sense mission control home">
          <span className="brand-mark" aria-hidden="true"><i /></span>
          <span><small>Graph intelligence</small>SPIDEY SENSE</span>
        </a>
        <div className="topbar-status">
          <span className="live-dot" aria-hidden="true" />
          <span>{providerOnline}/{snapshot.providers.length} providers online</span>
          <span className="timestamp">Updated {new Date(snapshot.generatedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
        </div>
      </header>

      <main id="mission-control">
        <section className="hero" aria-labelledby="dashboard-title">
          <div className="hero-web" aria-hidden="true"><span /><span /><span /></div>
          <div className="hero-copy">
            <p className="eyebrow">{snapshot.team.callSign} · Live coordination</p>
            <h1 id="dashboard-title">Sense the collision<br /><em>before it lands.</em></h1>
            <p>{snapshot.team.objective}</p>
          </div>
          <div className="signal-orbit" aria-hidden="true">
            <span className="orbit orbit-one" />
            <span className="orbit orbit-two" />
            <span className="orbit-core"><i /></span>
            <span className="orbit-alert" />
          </div>
        </section>

        <section className="summary-grid" aria-label="Mission control summary">
          <article><span>Live runners</span><strong>{liveRunners.toString().padStart(2, '0')}</strong><small>of {snapshot.runners.length} connected</small></article>
          <article><span>Active missions</span><strong>{activeMissions.toString().padStart(2, '0')}</strong><small>{snapshot.missions.length} total pathways</small></article>
          <article className={blockedMissions ? 'alert-metric' : ''}><span>Blocked</span><strong>{blockedMissions.toString().padStart(2, '0')}</strong><small>need coordination</small></article>
          <article><span>Graph relations</span><strong>{snapshot.graph.relationCount.toLocaleString()}</strong><small>depth {snapshot.graph.analyzedDepth} analysis</small></article>
        </section>

        <div className="dashboard-grid">
          <div className="primary-column">
            <TeamLobby team={snapshot.team} runners={snapshot.runners} missions={snapshot.missions} />
            <MissionPathways missions={snapshot.missions} runners={snapshot.runners} />
          </div>
          <aside className="insight-column" aria-label="Coordination intelligence">
            <RiskConsole risks={orderedRisks} selectedRiskId={selectedRisk?.id} onSelectRisk={setSelectedRiskId} />
            <EvidencePath risk={selectedRisk} />
            <ProviderPanel providers={snapshot.providers} />
            <GitPanel git={snapshot.git} />
            <GraphPanel graph={snapshot.graph} />
          </aside>
        </div>
      </main>

      <footer>
        <span>Local-first coordination</span>
        <span>Evidence, not oracle · CLEAR is bounded, not guaranteed</span>
      </footer>
    </div>
  );
}
