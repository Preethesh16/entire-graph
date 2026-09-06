import { useState, type FormEvent } from 'react';
import { createInviteURL, readInviteURL, type BrowserTeamSession, type ConnectedAgent, type PresenceInput } from '../network';
import type { Mission } from '../domain';
import { StatusPill } from './StatusPill';

interface ConnectionHubProps {
  session?: BrowserTeamSession;
  agents: ConnectedAgent[];
  missions: Mission[];
  presence: PresenceInput;
  error?: string;
  busy: boolean;
  onCreate: (input: { teamName: string; objective: string; name: string; role: string }) => Promise<void>;
  onJoin: (input: { teamId: string; inviteCode: string; name: string; role: string }) => Promise<void>;
  onPresence: (presence: PresenceInput) => Promise<void>;
  onLeave: () => void;
}

export function ConnectionHub({ session, agents, missions, presence, error, busy, onCreate, onJoin, onPresence, onLeave }: ConnectionHubProps) {
  const sharedInvite = readInviteURL();
  const [mode, setMode] = useState<'create' | 'join'>(sharedInvite ? 'join' : 'create');
  const [teamName, setTeamName] = useState('HardCoders Graph Room');
  const [objective, setObjective] = useState('Coordinate changes across the Anchor platform with verifiable Graph evidence.');
  const [teamId, setTeamId] = useState(sharedInvite?.teamId ?? '');
  const [inviteCode, setInviteCode] = useState(sharedInvite?.inviteCode ?? '');
  const [name, setName] = useState('');
  const [role, setRole] = useState('');
  const [copied, setCopied] = useState(false);
  const [missionId, setMissionId] = useState(presence.missionId ?? '');
  const [blocker, setBlocker] = useState(presence.blocker ?? '');

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (mode === 'create') await onCreate({ teamName, objective, name, role });
    else await onJoin({ teamId, inviteCode, name, role });
  };

  if (!session) return (
    <main className="connection-page">
      <section className="connection-intro">
        <p className="eyebrow">Secure team presence</p>
        <h1>Enter the<br /><em>signal room.</em></h1>
        <p>Create the shared room as lead, or join with an invite. Identity and role are entered by you; the browser runtime is detected and reported exactly as browser presence.</p>
        <div className="privacy-contract">
          <span>Transmitted</span><strong>Name · role · mission · blocker · presence</strong>
          <span>Never collected</span><strong>Prompts · reasoning · terminal · file contents · secrets</strong>
        </div>
      </section>
      <section className="connection-card" aria-labelledby="connect-title">
        <div className="mode-switch" role="tablist" aria-label="Team connection mode">
          <button role="tab" aria-selected={mode === 'create'} className={mode === 'create' ? 'active' : ''} onClick={() => setMode('create')}>Create room</button>
          <button role="tab" aria-selected={mode === 'join'} className={mode === 'join' ? 'active' : ''} onClick={() => setMode('join')}>Join invite</button>
        </div>
        <form onSubmit={submit}>
          <div className="form-heading"><span>0{mode === 'create' ? '1' : '2'}</span><div><p className="eyebrow">Connection protocol</p><h2 id="connect-title">{mode === 'create' ? 'Open a team room' : 'Accept an invitation'}</h2></div></div>
          {mode === 'create' ? <>
            <label>Team name<input required maxLength={100} value={teamName} onChange={(event) => setTeamName(event.target.value)} /></label>
            <label>Mission objective<textarea maxLength={500} value={objective} onChange={(event) => setObjective(event.target.value)} /></label>
          </> : <div className="form-pair">
            <label>Team ID<input required value={teamId} onChange={(event) => setTeamId(event.target.value)} placeholder="team_…" /></label>
            <label>Invite code<input required value={inviteCode} onChange={(event) => setInviteCode(event.target.value)} placeholder="invite_…" /></label>
          </div>}
          <div className="form-pair">
            <label>Your name<input required maxLength={100} value={name} onChange={(event) => setName(event.target.value)} placeholder="Preethesh" /></label>
            <label>Your role<input maxLength={100} value={role} onChange={(event) => setRole(event.target.value)} placeholder="Engineering lead" /></label>
          </div>
          {error ? <p className="inline-error" role="alert">{error}</p> : null}
          <button className="connection-submit" disabled={busy}>{busy ? 'Establishing session…' : mode === 'create' ? 'Create team room' : 'Join team room'}<span aria-hidden="true">↗</span></button>
        </form>
      </section>
    </main>
  );

  const copyInvite = async () => {
    if (!session.inviteCode) return;
    try {
      await navigator.clipboard.writeText(createInviteURL({ teamId: session.teamId, inviteCode: session.inviteCode }));
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1800);
    } catch { setCopied(false); }
  };

  return (
    <main className="focused-page connection-console">
      <header className="connection-console-head">
        <div><p className="eyebrow">Shared signal room · live</p><h1>Your team is connected.</h1><p>Presence updates stream directly from the shared coordinator. Browser access never claims local Git or Entire metadata it cannot observe.</p></div>
        <button className="quiet-action" onClick={onLeave}>Leave this browser session</button>
      </header>
      {session.inviteCode ? <section className="invite-vault">
        <div><p className="eyebrow">Invite link</p><h2>Bring your team onto the web</h2><p>Share this link privately. Credentials stay in the browser fragment, so they are not sent in the page request. The admin credential remains isolated in this tab.</p></div>
        <div className="invite-values"><code>{session.teamId}</code><code>{session.inviteCode}</code><button onClick={copyInvite}>{copied ? 'Link copied' : 'Copy private link'}</button></div>
      </section> : null}
      <section className="presence-editor">
        <div><p className="eyebrow">Your broadcast</p><h2>Current focus</h2></div>
        <label>Mission<select value={missionId} onChange={(event) => setMissionId(event.target.value)}><option value="">Unassigned</option>{missions.map((mission) => <option key={mission.id} value={mission.id}>{mission.title}</option>)}</select></label>
        <label>Blocker<input value={blocker} maxLength={500} onChange={(event) => setBlocker(event.target.value)} placeholder="No blocker" /></label>
        <button disabled={busy} onClick={() => onPresence({ missionId, blocker })}>Update presence</button>
      </section>
      {error ? <p className="inline-error" role="alert">{error}</p> : null}
      <section className="live-roster">
        <div className="section-heading"><div><p className="eyebrow">Authenticated live roster</p><h2>People in this room</h2></div><span className="count-badge">{agents.length}</span></div>
        <div className="roster-table" role="table" aria-label="Connected team agents">
          {agents.map((agent) => <article role="row" key={agent.id}>
            <span className="presence-orb" data-online={agent.online} aria-hidden="true" />
            <div><strong>{agent.name}</strong><span>{agent.role || 'Team member'}</span></div>
            <div><span>Detected client</span><strong>{agent.agent} · {agent.provider}</strong></div>
            <div><span>Mission</span><strong>{missions.find((mission) => mission.id === agent.mission_id)?.title ?? agent.mission_id ?? 'Unassigned'}</strong></div>
            <div><span>Last signal</span><strong>{new Date(agent.last_heartbeat).toLocaleTimeString()}</strong></div>
            <StatusPill status={agent.online ? 'online' : 'offline'} />
          </article>)}
        </div>
      </section>
    </main>
  );
}
