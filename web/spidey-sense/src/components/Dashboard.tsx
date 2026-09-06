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
          <span className="brand-mark" aria-hidden="true"><i /><i /><i /></span>
          <span><strong>spidey sense</strong><small>graph intelligence</small></span>
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
          <div className="web-emblem" aria-hidden="true">
            <svg viewBox="0 0 320 420" fill="none" xmlns="http://www.w3.org/2000/svg">
              <g className="web-lines">
                <path d="M160 22V398M37 200H283M71 69L249 331M249 69L71 331" />
                <path d="M160 22C157 70 130 93 91 70C118 100 108 143 64 160C111 158 130 179 124 215C116 250 94 286 71 331C112 301 143 313 160 352C177 313 208 301 249 331C226 286 204 250 196 215C190 179 209 158 256 160C212 143 202 100 229 70C190 93 163 70 160 22Z" />
                <path d="M160 79C158 111 141 126 116 111C134 131 130 157 102 168C132 168 144 181 140 204C136 226 122 249 108 277C134 258 150 266 160 291C170 266 186 258 212 277C198 249 184 226 180 204C176 181 188 168 218 168C190 157 186 131 204 111C179 126 162 111 160 79Z" />
              </g>
              <circle className="web-node" cx="160" cy="200" r="5" />
            </svg>
            <span>signal network</span>
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
