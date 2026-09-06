import { useMemo, useState } from 'react';
import type { DashboardSnapshot, EvidenceClass, RiskLevel } from '../domain';

interface SceneNode { id: string; label: string; detail: string; x: number; y: number; kind: 'mission' | 'evidence'; evidenceClass?: EvidenceClass }

const riskTone: Record<RiskLevel, string> = { BLOCK: 'block', REVIEW: 'review', CLEAR: 'clear' };

export function GraphExperience({ snapshot }: { snapshot: DashboardSnapshot }) {
  const [selected, setSelected] = useState<string>();
  const scene = useMemo(() => {
    const center = { x: 500, y: 285 };
    const missionNodes: SceneNode[] = snapshot.missions.map((mission, index) => {
      const angle = (Math.PI * 2 * index / Math.max(snapshot.missions.length, 1)) - Math.PI / 2;
      const radiusX = snapshot.missions.length < 3 ? 230 : 315;
      return { id: mission.id, label: mission.title, detail: mission.targets[0] ?? 'No target', x: center.x + Math.cos(angle) * radiusX, y: center.y + Math.sin(angle) * 190, kind: 'mission' };
    });
    const evidence = new Map<string, SceneNode>();
    snapshot.risks.forEach((risk, riskIndex) => risk.evidenceNodes.forEach((node, nodeIndex) => {
      if (evidence.has(node.id) || missionNodes.some((mission) => mission.id === node.id)) return;
      const angle = ((riskIndex * 2.1 + nodeIndex * .72) % (Math.PI * 2));
      const ring = 86 + ((riskIndex + nodeIndex) % 3) * 34;
      evidence.set(node.id, { id: node.id, label: node.label, detail: node.location, x: center.x + Math.cos(angle) * ring, y: center.y + Math.sin(angle) * ring * .62, kind: 'evidence', evidenceClass: risk.evidenceClass });
    }));
    return { center, nodes: [...missionNodes, ...evidence.values()], missionNodes };
  }, [snapshot]);
  const selectedNode = scene.nodes.find((node) => node.id === selected);

  return (
    <section className="graph-experience" aria-labelledby="graph-world-title">
      <div className="graph-experience-head">
        <div><p className="eyebrow">Repository intelligence · HardCoders_</p><h2 id="graph-world-title">Live dependency space</h2><p>Structural relationships are projected as evidence. Select a node to inspect its verified source location.</p></div>
        <div className="graph-legend" aria-label="Evidence legend"><span data-tone="confirmed">Confirmed</span><span data-tone="heuristic">Heuristic</span><span data-tone="incomplete">Incomplete</span></div>
      </div>
      <div className="graph-stage">
        <div className="graph-grid" aria-hidden="true" />
        <svg viewBox="0 0 1000 570" role="img" aria-label={`Spatial graph containing ${scene.nodes.length} visible nodes`}>
          <defs>
            <radialGradient id="core-glow"><stop offset="0" stopColor="#eaf4ff" stopOpacity=".95" /><stop offset="1" stopColor="#86b9e9" stopOpacity=".08" /></radialGradient>
            <filter id="soft-glow"><feGaussianBlur stdDeviation="4" result="blur" /><feMerge><feMergeNode in="blur" /><feMergeNode in="SourceGraphic" /></feMerge></filter>
          </defs>
          <g className="risk-strands">
            {snapshot.risks.map((risk) => {
              const from = scene.missionNodes.find((node) => node.id === risk.missionIds[0]);
              const to = scene.missionNodes.find((node) => node.id === risk.missionIds[1]);
              if (!from || !to) return null;
              const curve = `M ${from.x} ${from.y} Q ${scene.center.x} ${scene.center.y} ${to.x} ${to.y}`;
              return <path key={risk.id} d={curve} className={`risk-strand strand-${riskTone[risk.level]} evidence-${risk.evidenceClass}`}><title>{risk.level}: {risk.summary}</title></path>;
            })}
          </g>
          <g className="evidence-strands">
            {snapshot.risks.flatMap((risk) => risk.relations.map((relation) => {
              const from = scene.nodes.find((node) => node.id === relation.from);
              const to = scene.nodes.find((node) => node.id === relation.to);
              return from && to ? <line key={`${risk.id}-${relation.id}`} x1={from.x} y1={from.y} x2={to.x} y2={to.y} className={`evidence-${relation.evidenceClass}`}><title>{relation.type} · {Math.round(relation.confidence * 100)}%</title></line> : null;
            }))}
          </g>
          <g className="graph-core" transform={`translate(${scene.center.x} ${scene.center.y})`}>
            <circle r="70" fill="url(#core-glow)" /><circle r="32" /><circle r="5" filter="url(#soft-glow)" />
            <text y="-9">ENTIRE</text><text y="12">GRAPH</text>
          </g>
          {scene.nodes.map((node, index) => <g
            key={node.id} transform={`translate(${node.x} ${node.y})`} tabIndex={0} role="button"
            aria-label={`${node.kind}: ${node.label}, ${node.detail}`} aria-pressed={selected === node.id}
            className={`scene-node node-${node.kind} evidence-${node.evidenceClass ?? 'confirmed'} ${selected === node.id ? 'selected' : ''}`}
            onClick={() => setSelected(node.id)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') setSelected(node.id); }}
          >
            <circle r={node.kind === 'mission' ? 29 : 12} /><circle className="node-pulse" r={node.kind === 'mission' ? 37 : 18} />
            <text y={node.kind === 'mission' ? 51 : 31}>{node.label.length > 28 ? `${node.label.slice(0, 27)}…` : node.label}</text>
            <text className="node-index" y="4">{String(index + 1).padStart(2, '0')}</text>
          </g>)}
        </svg>
        <div className="graph-inspector" aria-live="polite">
          <span>{selectedNode ? selectedNode.kind : 'Graph scope'}</span>
          <strong>{selectedNode?.label ?? `${snapshot.graph.nodeCount.toLocaleString()} indexed nodes`}</strong>
          <code>{selectedNode?.detail ?? `${snapshot.graph.relationCount.toLocaleString()} relationships · depth ${snapshot.graph.analyzedDepth}`}</code>
        </div>
        <div className="graph-depth" aria-hidden="true"><span>z 02</span><span>semantic plane</span></div>
      </div>
      <div className="graph-health-strip">
        <span className={snapshot.graph.complete ? 'healthy' : 'partial'}>{snapshot.graph.complete ? 'Analysis complete' : 'Partial analysis detected'}</span>
        <span>{snapshot.graph.warningCount} warnings</span><span>{snapshot.graph.partialFailureCount} partial failures</span>
        <strong>Evidence is inspectable, never absolute.</strong>
      </div>
    </section>
  );
}
