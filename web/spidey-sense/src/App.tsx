import { useCallback, useEffect, useMemo, useState } from 'react';
import { loadDashboard, loadPlan, loadSessions, refreshDashboard, savePlan, type CoordinatePlan, type PublicSession } from './api';
import { Dashboard } from './components/Dashboard';
import { Workflow } from './components/Workflow';
import { createBrowserTeam, joinBrowserTeam, leaveBrowserTeam, loadTeamAgents, restoreBrowserTeam, sendBrowserHeartbeat, watchTeamEvents, type BrowserTeamSession, type ConnectedAgent, type PresenceInput } from './network';
import type { DashboardSnapshot } from './domain';

export default function App() {
  const [snapshot, setSnapshot] = useState<DashboardSnapshot>();
  const [error, setError] = useState<string>();
  const [plan, setPlan] = useState<CoordinatePlan>();
  const [sessions, setSessions] = useState<PublicSession[]>([]);
  const [browserSession, setBrowserSession] = useState<BrowserTeamSession | undefined>(() => restoreBrowserTeam());
  const [liveAgents, setLiveAgents] = useState<ConnectedAgent[]>([]);
  const [connectionBusy, setConnectionBusy] = useState(false);
  const [connectionError, setConnectionError] = useState<string>();
  const [presence, setPresence] = useState<PresenceInput>({});
  useEffect(() => {
    const controller = new AbortController();
    Promise.all([loadDashboard(controller.signal), loadPlan(controller.signal), loadSessions(controller.signal)]).then(([nextSnapshot, nextPlan, activity]) => { setSnapshot(nextSnapshot); setPlan(nextPlan); setSessions(activity.sessions); }).catch((reason: unknown) => {
      if (!controller.signal.aborted) setError(reason instanceof Error ? reason.message : 'Unable to load mission data');
    });
    return () => controller.abort();
  }, []);

  const refreshAgents = useCallback(async (active: BrowserTeamSession, signal?: AbortSignal) => {
    setLiveAgents(await loadTeamAgents(active, signal));
  }, []);

  useEffect(() => {
    if (!browserSession) return;
    const controller = new AbortController();
    const pulse = async () => {
      try {
        await sendBrowserHeartbeat(browserSession, presence);
        await refreshAgents(browserSession, controller.signal);
        setConnectionError(undefined);
      } catch (reason) {
        if (!controller.signal.aborted) setConnectionError(reason instanceof Error ? reason.message : 'Presence update failed');
      }
    };
    void pulse();
    const interval = window.setInterval(pulse, 15_000);
    void watchTeamEvents(browserSession, controller.signal, () => void refreshAgents(browserSession, controller.signal)).catch((reason: unknown) => {
      if (!controller.signal.aborted) setConnectionError(reason instanceof Error ? reason.message : 'Live team stream disconnected');
    });
    return () => { controller.abort(); window.clearInterval(interval); };
  }, [browserSession, presence, refreshAgents]);

  const runConnection = async (action: () => Promise<BrowserTeamSession>) => {
    setConnectionBusy(true); setConnectionError(undefined);
    try {
      const connected = await action();
      setBrowserSession(connected);
      await sendBrowserHeartbeat(connected, presence);
      await refreshAgents(connected);
    } catch (reason) {
      setConnectionError(reason instanceof Error ? reason.message : 'Unable to connect team');
    } finally { setConnectionBusy(false); }
  };

  const connectedSnapshot = useMemo(() => {
    if (!snapshot || liveAgents.length === 0) return snapshot;
    const runners = snapshot.runners.map((runner) => {
      const agent = liveAgents.find((candidate) => candidate.member_id === runner.id || candidate.name.toLowerCase() === runner.displayName.toLowerCase());
      return agent ? { ...runner, status: agent.online ? 'active' as const : 'offline' as const, provider: `${agent.agent} · ${agent.provider}`, currentMissionId: agent.mission_id || runner.currentMissionId } : runner;
    });
    for (const [index, agent] of liveAgents.entries()) {
      if (runners.some((runner) => runner.id === agent.member_id || runner.displayName.toLowerCase() === agent.name.toLowerCase())) continue;
      runners.push({
        id: agent.member_id, displayName: agent.name, role: agent.role || 'Team member', archetype: 'Browser presence',
        accent: ['#a9d6ff', '#f4be7c', '#c9b6ff', '#78e2ba'][index % 4], initials: agent.name.split(/\s+/).map((part) => part[0]).join('').slice(0, 2).toUpperCase(),
        status: agent.online ? 'active' : 'offline', provider: `${agent.agent} · ${agent.provider}`, currentMissionId: agent.mission_id,
      });
    }
    return { ...snapshot, runners };
  }, [snapshot, liveAgents]);

  if (!snapshot || !plan) return <Dashboard loading={!error} error={error} />;
  return <Workflow
    snapshot={connectedSnapshot ?? snapshot} plan={plan} sessions={sessions} browserSession={browserSession} liveAgents={liveAgents}
    connectionBusy={connectionBusy} connectionError={connectionError} presence={presence}
    onCreateTeam={(input) => runConnection(() => createBrowserTeam(input))}
    onJoinTeam={(input) => runConnection(async () => {
      const connected = await joinBrowserTeam(input);
      window.history.replaceState(null, '', `${window.location.pathname}${window.location.search}`);
      return connected;
    })}
    onPresence={async (next) => { setPresence(next); if (browserSession) { setConnectionBusy(true); setConnectionError(undefined); try { await sendBrowserHeartbeat(browserSession, next); await refreshAgents(browserSession); } catch (reason) { setConnectionError(reason instanceof Error ? reason.message : 'Presence update failed'); } finally { setConnectionBusy(false); } } }}
    onLeaveTeam={() => { leaveBrowserTeam(); setBrowserSession(undefined); setLiveAgents([]); setConnectionError(undefined); }}
    onRefresh={async () => { setSnapshot(await refreshDashboard()); setSessions((await loadSessions()).sessions); }}
    onSave={async (next) => { const saved = await savePlan(next); setPlan(saved); setSnapshot(await loadDashboard()); }}
  />;
}
