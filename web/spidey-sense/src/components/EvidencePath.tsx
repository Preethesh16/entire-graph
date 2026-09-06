import type { RiskSignal } from '../domain';
import { StatusPill } from './StatusPill';

interface EvidencePathProps {
  risk?: RiskSignal;
}

export function EvidencePath({ risk }: EvidencePathProps) {
  if (!risk) {
    return (
      <section className="panel evidence-panel" aria-labelledby="evidence-path-title">
        <div className="empty-state">
          <span className="radar-empty" aria-hidden="true" />
          <h2 id="evidence-path-title">Select a risk signal</h2>
          <p>Evidence paths, relation types, and source locations will appear here.</p>
        </div>
      </section>
    );
  }

  return (
    <section className="panel evidence-panel" aria-labelledby="evidence-path-title">
      <div className="section-heading evidence-heading">
        <div>
          <p className="eyebrow">Graph evidence path</p>
          <h2 id="evidence-path-title">{risk.title}</h2>
        </div>
        <StatusPill status={risk.level} />
      </div>

      <p className={`evidence-class evidence-${risk.evidenceClass}`}>
        Evidence: <strong>{risk.evidenceClass}</strong>
        {risk.verificationRequired ? ' · source or test verification required' : ' · structurally confirmed'}
      </p>

      {risk.evidenceNodes.length === 0 ? (
        <div className="clear-evidence">
          <span aria-hidden="true">✓</span>
          <div>
            <strong>No bounded relationship found</strong>
            <p>{risk.summary}</p>
          </div>
        </div>
      ) : (
        <ol className="evidence-path" aria-label="Risk evidence relationship path">
          {risk.evidenceNodes.map((node, index) => {
            const relation = risk.relations[index];
            return (
              <li key={node.id}>
                <div className="evidence-node">
                  <span className={`node-kind node-${node.kind}`}>{node.kind}</span>
                  <strong>{node.label}</strong>
                  <code>{node.location}</code>
                </div>
                {relation ? (
                  <div className="relation-chip">
                    <span>{relation.type}</span>
                    <span>{Math.round(relation.confidence * 100)}% · {relation.evidenceClass}</span>
                  </div>
                ) : null}
              </li>
            );
          })}
        </ol>
      )}

      <div className="recommendation">
        <p className="eyebrow">Recommended action</p>
        <strong>{risk.recommendation}</strong>
      </div>

      <div className="review-targets">
        <p className="eyebrow">Affected test / review targets</p>
        {risk.reviewTargets.length === 0 ? (
          <span className="muted-copy">No additional targets identified.</span>
        ) : (
          risk.reviewTargets.map((target) => <code key={target}>{target}</code>)
        )}
      </div>

      <div className="review-targets verification-path">
        <p className="eyebrow">Verification path</p>
        {risk.verification.map((step) => <span key={step}>{step}</span>)}
      </div>
    </section>
  );
}
