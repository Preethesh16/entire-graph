import type { Mission, Runner, Team } from '../domain';
import { StatusPill } from './StatusPill';

interface TeamLobbyProps {
  team: Team;
  runners: Runner[];
  missions: Mission[];
}

export function TeamLobby({ team, runners, missions }: TeamLobbyProps) {
  const missionById = new Map(missions.map((mission) => [mission.id, mission]));

  return (
    <section className="panel team-panel" aria-labelledby="team-lobby-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Team lobby · {team.callSign}</p>
          <h2 id="team-lobby-title">Runners on the web</h2>
        </div>
        <span className="count-badge">{runners.length}</span>
      </div>

      {runners.length === 0 ? (
        <div className="empty-state compact-empty">
          <span className="empty-glyph" aria-hidden="true">○</span>
          <h3>No runners connected</h3>
          <p>Connect an Entire-enabled session to populate the team lobby.</p>
        </div>
      ) : (
        <div className="runner-grid">
          {runners.map((runner) => {
            const mission = runner.currentMissionId ? missionById.get(runner.currentMissionId) : undefined;
            return (
              <article className="runner-card" key={runner.id} style={{ '--runner-accent': runner.accent } as React.CSSProperties}>
                <div className="runner-identity">
                  <span className="runner-avatar" aria-hidden="true">{runner.initials}</span>
                  <div>
                    <h3>{runner.displayName}</h3>
                    <p>{runner.archetype}</p>
                  </div>
                  <StatusPill status={runner.status} />
                </div>
                <p className="runner-role">{runner.role}</p>
                <div className="runner-meta">
                  <span>{runner.provider}</span>
                  <span>{mission?.title ?? 'Awaiting assignment'}</span>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
