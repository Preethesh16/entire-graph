import type { RiskSignal } from '../domain';
import { StatusPill } from './StatusPill';

interface RiskConsoleProps {
  risks: RiskSignal[];
  selectedRiskId?: string;
  onSelectRisk: (riskId: string) => void;
}

export function RiskConsole({ risks, selectedRiskId, onSelectRisk }: RiskConsoleProps) {
  const actionable = risks.filter((risk) => risk.level !== 'CLEAR');

  return (
    <section className="panel risk-panel" aria-labelledby="risk-console-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Early warning</p>
          <h2 id="risk-console-title">Risk signals</h2>
        </div>
        <span className="count-badge danger-count">{actionable.length}</span>
      </div>

      {risks.length === 0 ? (
        <div className="empty-state compact-empty">
          <span className="empty-glyph clear-glyph" aria-hidden="true">✓</span>
          <h3>No risk signals</h3>
          <p>No analyzed mission pairs are currently competing for attention.</p>
        </div>
      ) : (
        <div className="risk-list">
          {risks.map((risk) => (
            <button
              className={`risk-card risk-${risk.level.toLowerCase()} ${selectedRiskId === risk.id ? 'selected' : ''}`}
              key={risk.id}
              type="button"
              onClick={() => onSelectRisk(risk.id)}
              aria-pressed={selectedRiskId === risk.id}
            >
              <span className="risk-card-top">
                <StatusPill status={risk.level} />
                <span>{risk.relations.length} relations · {risk.evidenceClass}</span>
              </span>
              <strong>{risk.title}</strong>
              <span>{risk.summary}</span>
            </button>
          ))}
        </div>
      )}
    </section>
  );
}
