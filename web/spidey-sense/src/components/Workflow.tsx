import { useEffect, useMemo, useState } from 'react';
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
  onRefresh: () => Promise<void>;
  onSave: (plan: CoordinatePlan) => Promise<void>;
}

export function Workflow({ snapshot, plan, sessions, browserSession, liveAgents, connectionBusy, connectionError, presence, onCreateTeam, onJoinTeam, onPresence, onLeaveTeam, onRefresh, onSave }: WorkflowProps) {
  const [view, setView] = useState<View>('connect');
  const [draft, setDraft] = useState(plan);
  const [saving, setSaving] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  useEffect(() => setDraft(plan), [plan]);
  const assignableMembers = useMemo(() => {
    const members = [...draft.team.members];
    for (const agent of liveAgents) {
      if (members.some((member) => member.id === agent.member_id)) continue;
      members.push({ id: agent.member_id, name: agent.name, role: agent.role, archetype: agent.agent });
    }
    return members;
  }, [draft.team.members, liveAgents]);
  const save = async () => {
    setSaving(true);
    try { await onSave({ ...draft, team: { ...draft.team, members: assignableMembers } }); }
    finally { setSaving(false); }
  };
  const addMission = () => setDraft({ ...draft, missions: [...draft.missions, { id: `mission-${crypto.randomUUID()}`, title: 'New task', intent: '', owner: assignableMembers[0]?.id ?? '', status: 'queued', targets: [{ file: '', symbol: '' }] }] });
  const updateMission = (index: number, changes: Partial<CoordinatePlan['missions'][number]>) => {
    const missions = [...draft.missions];
    missions[index] = { ...missions[index], ...changes };
    setDraft({ ...draft, missions });
  };

  return <div className="workflow-shell">
    <nav className="workflow-nav" aria-label="Product workflow">
      {(['connect', 'plan', 'tracker', 'activity'] as View[]).map((item) => <button className={view === item ? 'active' : ''} onClick={() => setView(item)} key={item}>{item === 'connect' ? '1. Connect team' : item === 'plan' ? '2. Plan & assign' : item === 'tracker' ? '3. Graph space' : 'Live activity'}</button>)}
      <button className="refresh-evidence" disabled={refreshing} onClick={async () => { setRefreshing(true); try { await onRefresh(); } finally { setRefreshing(false); } }}>{refreshing ? 'Rebuilding Graph…' : 'Refresh evidence'}</button>
    </nav>
    {view === 'tracker' ? <Dashboard snapshot={snapshot} /> : null}
    {view === 'connect' ? <ConnectionHub session={browserSession} agents={liveAgents} missions={snapshot.missions} presence={presence} error={connectionError} busy={connectionBusy} onCreate={onCreateTeam} onJoin={onJoinTeam} onPresence={onPresence} onLeave={onLeaveTeam} /> : null}
    {view === 'plan' ? <main className="focused-page"><header><p className="eyebrow">Step two</p><h1>Plan and assign</h1><p>Every authenticated teammate in the live room is assignable by stable member ID. Declare source targets precisely; Graph relationships remain evidence and verification stays visible.</p></header><div className="assignment-summary"><span>{assignableMembers.length} assignable people</span><span>{liveAgents.filter((agent) => agent.online).length} online now</span><span>{draft.missions.length} missions</span></div><div className="editor-list mission-editor">{draft.missions.map((mission, index) => <article key={mission.id}>
      <div className="mission-editor-main"><input aria-label="Task title" value={mission.title} onChange={(event) => updateMission(index, { title: event.target.value })} /><input aria-label={`Intent for ${mission.title}`} placeholder="What outcome must this mission produce?" value={mission.intent ?? ''} onChange={(event) => updateMission(index, { intent: event.target.value })} /></div>
      <div className="mission-editor-assignment"><select aria-label={`Owner for ${mission.title}`} value={mission.owner} onChange={(event) => updateMission(index, { owner: event.target.value })}><option value="">Unassigned</option>{assignableMembers.map((member) => <option value={member.id} key={member.id}>{member.name}{liveAgents.some((agent) => agent.member_id === member.id && agent.online) ? ' · online' : ''}</option>)}</select><select aria-label={`Status for ${mission.title}`} value={mission.status} onChange={(event) => updateMission(index, { status: event.target.value as CoordinatePlan['missions'][number]['status'] })}><option value="queued">Queued</option><option value="active">Active</option><option value="blocked">Blocked</option><option value="complete">Complete</option></select></div>
      <div className="mission-editor-target"><input aria-label="Target file" placeholder="apps/api/src/file.ts" value={mission.targets[0]?.file ?? ''} onChange={(event) => updateMission(index, { targets: [{ ...mission.targets[0], file: event.target.value }] })} /><input aria-label="Target symbol" placeholder="Optional symbol" value={mission.targets[0]?.symbol ?? ''} onChange={(event) => updateMission(index, { targets: [{ ...mission.targets[0], symbol: event.target.value }] })} /></div>
      <button className="remove-mission" aria-label={`Remove ${mission.title}`} onClick={() => setDraft({ ...draft, missions: draft.missions.filter((item) => item.id !== mission.id) })}>Remove</button>
    </article>)}</div><div className="editor-actions"><button disabled={!assignableMembers.length} onClick={addMission}>Add mission</button><button className="primary-action" disabled={saving || !assignableMembers.length} onClick={save}>{saving ? 'Analyzing…' : 'Save & analyze'}</button></div></main> : null}
    {view === 'activity' ? <main className="focused-page"><header><p className="eyebrow">Bounded activity telemetry</p><h1>Live presence.</h1><p>Browser presence and local Entire sessions stay explicitly distinct. Missing local metadata is shown as unavailable, never guessed.</p></header><div className="editor-list activity-editor">{liveAgents.map((agent) => <article key={agent.id}><strong>{agent.name}</strong><span>{agent.online ? 'Online' : 'Offline'} · {agent.role || 'Member'}</span><code>{agent.agent} · {agent.provider}</code><small>{agent.mission_id || 'No mission'} · {agent.last_heartbeat}</small></article>)}{sessions.map((session) => <article key={session.session_id}><strong>{session.agent}</strong><span>{session.status}</span><code>{session.session_id}</code><small>{session.worktree}</small></article>)}</div></main> : null}
  </div>;
}
