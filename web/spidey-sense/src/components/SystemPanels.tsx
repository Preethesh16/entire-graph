import type { GitSnapshot, GraphSnapshot, ProviderHealth } from '../domain';
import { StatusPill } from './StatusPill';

interface ProviderPanelProps {
  providers: ProviderHealth[];
}

export function ProviderPanel({ providers }: ProviderPanelProps) {
  return (
    <section className="panel provider-panel" aria-labelledby="provider-health-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Signal network</p>
          <h2 id="provider-health-title">Provider health</h2>
        </div>
      </div>
      {providers.length === 0 ? (
        <p className="muted-copy">No providers configured.</p>
      ) : (
        <ul className="provider-list">
          {providers.map((provider) => (
            <li key={provider.id}>
              <span className="provider-pulse" aria-hidden="true" />
              <div>
                <strong>{provider.name}</strong>
                <span>{provider.detail}</span>
              </div>
              <div className="provider-status">
                <StatusPill status={provider.status} />
                <small>{provider.latencyMs ? `${provider.latencyMs}ms · ` : ''}{provider.lastCheckedAt}</small>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

interface GitPanelProps {
  git: GitSnapshot;
}

export function GitPanel({ git }: GitPanelProps) {
	if (git.available === false) {
		return (
			<section className="panel git-panel" aria-labelledby="git-activity-title">
				<div className="section-heading"><div><p className="eyebrow">Source control</p><h2 id="git-activity-title">Git activity</h2></div></div>
				<p className="muted-copy">Git adapter not connected in this slice.</p>
			</section>
		);
	}
  return (
    <section className="panel git-panel" aria-labelledby="git-activity-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Source control</p>
          <h2 id="git-activity-title">Git activity</h2>
        </div>
        <code>{git.branch}</code>
      </div>
      <div className="git-summary">
        <span>HEAD <strong>{git.head}</strong></span>
        <span>{git.dirtyFileCount} dirty</span>
        <span>↑{git.ahead} ↓{git.behind}</span>
      </div>
      {git.activity.length === 0 ? (
        <p className="muted-copy">No recent Git activity.</p>
      ) : (
        <ol className="activity-list">
          {git.activity.map((item) => (
            <li key={item.id}>
              <span className="timeline-dot" aria-hidden="true" />
              <div>
                <strong>{item.action}</strong>
                <span>{item.actor} · <code>{item.ref}</code> · {item.timestamp}</span>
              </div>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}

interface GraphPanelProps {
  graph: GraphSnapshot;
}

export function GraphPanel({ graph }: GraphPanelProps) {
  return (
    <section className="panel graph-panel" aria-labelledby="graph-integrity-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Structural scan</p>
          <h2 id="graph-integrity-title">Graph integrity</h2>
        </div>
        <span className={`graph-completeness ${graph.complete ? 'complete' : 'partial'}`}>
          {graph.complete ? 'Complete' : 'Partial'}
        </span>
      </div>
      <div className="metric-grid">
        <div><strong>{graph.nodeCount.toLocaleString()}</strong><span>Nodes</span></div>
        <div><strong>{graph.relationCount.toLocaleString()}</strong><span>Relations</span></div>
        <div><strong>{graph.analyzedDepth}</strong><span>Depth</span></div>
      </div>
      {graph.warning ? <p className="graph-warning">{graph.warning}</p> : null}
    </section>
  );
}
