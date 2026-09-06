import { useState } from 'react';
import type { CoordinatePlan, PublicSession } from '../api';
import type { DashboardSnapshot } from '../domain';
import { Dashboard } from './Dashboard';

type View = 'team' | 'plan' | 'tracker' | 'activity';

export function Workflow({ snapshot, plan, sessions, onSave }: { snapshot: DashboardSnapshot; plan: CoordinatePlan; sessions: PublicSession[]; onSave: (plan: CoordinatePlan) => Promise<void> }) {
  const [view, setView] = useState<View>('tracker');
  const [draft, setDraft] = useState(plan);
  const [saving, setSaving] = useState(false);
  const save = async () => { setSaving(true); try { await onSave(draft); } finally { setSaving(false); } };
  const addMember = () => setDraft({ ...draft, team: { ...draft.team, members: [...draft.team.members, { id: `member-${crypto.randomUUID()}`, name: 'New teammate', session_ids: [] }] } });
  const addMission = () => setDraft({ ...draft, missions: [...draft.missions, { id: `mission-${crypto.randomUUID()}`, title: 'New task', owner: draft.team.members[0]?.id ?? '', status: 'queued', targets: [{ file: 'README.md' }] }] });

  return <div className="workflow-shell">
    <nav className="workflow-nav" aria-label="Product workflow">
      {(['team', 'plan', 'tracker', 'activity'] as View[]).map((item) => <button className={view === item ? 'active' : ''} onClick={() => setView(item)} key={item}>{item === 'team' ? '1. Team setup' : item === 'plan' ? '2. Plan & assign' : item === 'tracker' ? '3. Spidey tracker' : 'Agent activity'}</button>)}
    </nav>
    {view === 'tracker' ? <Dashboard snapshot={snapshot} /> : null}
    {view === 'team' ? <main className="focused-page"><header><p className="eyebrow">Step one</p><h1>Create your team</h1><p>Add any number of teammates, then map each person to a detected Entire agent session.</p></header><div className="editor-list">{draft.team.members.map((member, index) => <article key={member.id}><input aria-label="Member name" value={member.name} onChange={(event) => { const members = [...draft.team.members]; members[index] = { ...member, name: event.target.value }; setDraft({ ...draft, team: { ...draft.team, members } }); }} /><select aria-label={`Agent session for ${member.name}`} value={member.session_ids?.[0] ?? ''} onChange={(event) => { const members = [...draft.team.members]; members[index] = { ...member, session_ids: event.target.value ? [event.target.value] : [] }; setDraft({ ...draft, team: { ...draft.team, members } }); }}><option value="">Not connected</option>{sessions.map((session) => <option value={session.session_id} key={session.session_id}>{session.agent} · {session.status} · {session.session_id.slice(0, 8)}</option>)}</select></article>)}</div><div className="editor-actions"><button onClick={addMember}>Add teammate</button><button className="primary-action" disabled={saving} onClick={save}>Save team</button></div></main> : null}
    {view === 'plan' ? <main className="focused-page"><header><p className="eyebrow">Step two</p><h1>Plan and assign</h1><p>The leader assigns tasks and declares the files or symbols each task expects to change.</p></header><div className="editor-list">{draft.missions.map((mission, index) => <article key={mission.id}><input aria-label="Task title" value={mission.title} onChange={(event) => { const missions = [...draft.missions]; missions[index] = { ...mission, title: event.target.value }; setDraft({ ...draft, missions }); }} /><select aria-label={`Owner for ${mission.title}`} value={mission.owner} onChange={(event) => { const missions = [...draft.missions]; missions[index] = { ...mission, owner: event.target.value }; setDraft({ ...draft, missions }); }}>{draft.team.members.map((member) => <option value={member.id} key={member.id}>{member.name}</option>)}</select><input aria-label="Target file" value={mission.targets[0]?.file ?? ''} onChange={(event) => { const missions = [...draft.missions]; missions[index] = { ...mission, targets: [{ ...mission.targets[0], file: event.target.value }] }; setDraft({ ...draft, missions }); }} /></article>)}</div><div className="editor-actions"><button disabled={!draft.team.members.length} onClick={addMission}>Add task</button><button className="primary-action" disabled={saving} onClick={save}>Save plan</button></div></main> : null}
    {view === 'activity' ? <main className="focused-page"><header><p className="eyebrow">Separate signal view</p><h1>Agent activity</h1><p>Process monitoring stays here so the coordination graph remains clean.</p></header><div className="editor-list">{sessions.map((session) => <article key={session.session_id}><strong>{session.agent}</strong><span>{session.status}</span><code>{session.session_id}</code><small>{session.worktree}</small></article>)}</div></main> : null}
  </div>;
}
