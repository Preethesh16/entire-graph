import type { Mission, MissionStatus, Runner } from '../domain';
import { StatusPill } from './StatusPill';

const stages: MissionStatus[] = ['queued', 'active', 'blocked', 'complete'];
const stageLabels: Record<MissionStatus, string> = {
  queued: 'Queued',
  active: 'Active',
  blocked: 'Blocked',
  complete: 'Complete',
};

interface MissionPathwaysProps {
  missions: Mission[];
  runners: Runner[];
}

export function MissionPathways({ missions, runners }: MissionPathwaysProps) {
  const runnerById = new Map(runners.map((runner) => [runner.id, runner]));

  return (
    <section className="panel mission-panel" aria-labelledby="mission-pathways-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Dynamic mission pathway</p>
          <h2 id="mission-pathways-title">Work in motion</h2>
        </div>
        <span className="count-badge">{missions.length}</span>
      </div>

      {missions.length === 0 ? (
        <div className="empty-state compact-empty">
          <span className="empty-glyph" aria-hidden="true">◇</span>
          <h3>No missions planned</h3>
          <p>Create a mission to begin coordination analysis.</p>
        </div>
      ) : (
        <div className="mission-list">
          {missions.map((mission) => {
            const owner = mission.ownerId ? runnerById.get(mission.ownerId) : undefined;
            return (
              <article className={`mission-card mission-${mission.status}`} key={mission.id}>
                <div className="mission-copy">
                  <div>
                    <p className="mission-owner">{owner?.displayName ?? 'Unassigned'} · {mission.updatedAt}</p>
                    <h3>{mission.title}</h3>
                    <p>{mission.intent}</p>
                  </div>
                  <StatusPill status={mission.status} />
                </div>

                <ol className="pathway" aria-label={`${mission.title} status: ${stageLabels[mission.status]}`}>
                  {stages.map((stage, index) => {
                    const current = stage === mission.status;
                    const reached = stages.indexOf(mission.status) >= index && mission.status !== 'blocked';
                    return (
                      <li className={current ? 'current' : reached ? 'reached' : ''} key={stage}>
                        <span className="path-dot">{index + 1}</span>
                        <span>{stageLabels[stage]}</span>
                      </li>
                    );
                  })}
                </ol>

                <div className="mission-targets">
                  <span>{mission.targetCount} graph targets</span>
                  <div className="target-chips" aria-label="Mission target files">
                    {mission.targets.map((target) => <code key={target}>{target}</code>)}
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
