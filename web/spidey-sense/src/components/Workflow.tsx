import { useState } from 'react';
import type { CoordinatePlan, PublicSession } from '../api';
import type { BrowserTeamSession, ConnectedAgent, PresenceInput } from '../network';
import type { DashboardSnapshot } from '../domain';
import { ConnectionHub } from './ConnectionHub';
import { Dashboard } from './Dashboard';

type View = 'connect' | 'plan' | 'tracker' | 'activity';

interface WorkflowProps {
  snapshot: DashboardSnapshot;
  plan: CoordinatePlan;
  sessions: PublicSession[];
  browserSession?: BrowserTeamSession;
  liveAgents: ConnectedAgent[];
  connectionBusy: boolean;
  connectionError?: string;
  presence: PresenceInput;
  onCreateTeam: (input: { teamName: string; objective: string; name: string; role: string }) => Promise<void>;
  onJoinTeam: (input: { teamId: string; inviteCode: string; name: string; role: string }) => Promise<void>;
  onPresence: (presence: PresenceInput) => Promise<void>;
  onLeaveTeam: () => void;
  onSave: (plan: CoordinatePlan) => Promise<void>;
}

export function Workflow({ snapshot, plan, sessions, browserSession, liveAgents, connectionBusy, connectionError, presence, onCreateTeam, onJoinTeam, onPresence, onLeaveTeam, onSave }: WorkflowProps) {
  const [view, setView] = useState<View>('connect');
  const [draft, setDraft] = useState(plan);
  const [saving, setSaving] = useState(false);
  const save = async () => { setSaving(true); try { await onSave(draft); } finally { setSaving(false); } };
  const addMission = () => setDraft({ ...draft, missions: [...draft.missions, { id: `mission-${crypto.randomUUID()}`, title: 'New task', owner: draft.team.members[0]?.id ?? '', status: 'queued', targets: [{ file: 'README.md' }] }] });

  return <div className="workflow-shell">
    <nav className="workflow-nav" aria-label="Product workflow">
      {(['connect', 'plan', 'tracker', 'activity'] as View[]).map((item) => <button className={view === item ? 'active' : ''} onClick={() => setView(item)} key={item}>{item === 'connect' ? '1. Connect team' : item === 'plan' ? '2. Plan & assign' : item === 'tracker' ? '3. Graph space' : 'Live activity'}</button>)}
    </nav>
    {view === 'tracker' ? <Dashboard snapshot={snapshot} /> : null}
    {view === 'connect' ? <ConnectionHub session={browserSession} agents={liveAgents} missions={snapshot.missions} presence={presence} error={connectionError} busy={connectionBusy} onCreate={onCreateTeam} onJoin={onJoinTeam} onPresence={onPresence} onLeave={onLeaveTeam} /> : null}
    {view === 'plan' ? <main className="focused-page"><header><p className="eyebrow">Step two</p><h1>Plan and assign</h1><p>The leader assigns tasks and declares the files or symbols each task expects to change.</p></header><div className="editor-list">{draft.missions.map((mission, index) => <article key={mission.id}><input aria-label="Task title" value={mission.title} onChange={(event) => { const missions = [...draft.missions]; missions[index] = { ...mission, title: event.target.value }; setDraft({ ...draft, missions }); }} /><select aria-label={`Owner for ${mission.title}`} value={mission.owner} onChange={(event) => { const missions = [...draft.missions]; missions[index] = { ...mission, owner: event.target.value }; setDraft({ ...draft, missions }); }}>{draft.team.members.map((member) => <option value={member.id} key={member.id}>{member.name}</option>)}</select><input aria-label="Target file" value={mission.targets[0]?.file ?? ''} onChange={(event) => { const missions = [...draft.missions]; missions[index] = { ...mission, targets: [{ ...mission.targets[0], file: event.target.value }] }; setDraft({ ...draft, missions }); }} /></article>)}</div><div className="editor-actions"><button disabled={!draft.team.members.length} onClick={addMission}>Add task</button><button className="primary-action" disabled={saving} onClick={save}>Save plan</button></div></main> : null}
    {view === 'activity' ? <main className="focused-page"><header><p className="eyebrow">Bounded activity telemetry</p><h1>Live presence.</h1><p>Browser presence and local Entire sessions stay explicitly distinct. Missing local metadata is shown as unavailable, never guessed.</p></header><div className="editor-list activity-editor">{liveAgents.map((agent) => <article key={agent.id}><strong>{agent.name}</strong><span>{agent.online ? 'Online' : 'Offline'} · {agent.role || 'Member'}</span><code>{agent.agent} · {agent.provider}</code><small>{agent.mission_id || 'No mission'} · {agent.last_heartbeat}</small></article>)}{sessions.map((session) => <article key={session.session_id}><strong>{session.agent}</strong><span>{session.status}</span><code>{session.session_id}</code><small>{session.worktree}</small></article>)}</div></main> : null}
  </div>;
}
