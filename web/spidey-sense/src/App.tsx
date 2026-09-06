import { useEffect, useState } from 'react';
import { loadDashboard, loadPlan, loadSessions, savePlan, type CoordinatePlan, type PublicSession } from './api';
import { Dashboard } from './components/Dashboard';
import { Workflow } from './components/Workflow';
import type { DashboardSnapshot } from './domain';

export default function App() {
  const [snapshot, setSnapshot] = useState<DashboardSnapshot>();
  const [error, setError] = useState<string>();
  const [plan, setPlan] = useState<CoordinatePlan>();
  const [sessions, setSessions] = useState<PublicSession[]>([]);
  useEffect(() => {
    const controller = new AbortController();
    Promise.all([loadDashboard(controller.signal), loadPlan(controller.signal), loadSessions(controller.signal)]).then(([nextSnapshot, nextPlan, activity]) => { setSnapshot(nextSnapshot); setPlan(nextPlan); setSessions(activity.sessions); }).catch((reason: unknown) => {
      if (!controller.signal.aborted) setError(reason instanceof Error ? reason.message : 'Unable to load mission data');
    });
    return () => controller.abort();
  }, []);
  if (!snapshot || !plan) return <Dashboard loading={!error} error={error} />;
  return <Workflow snapshot={snapshot} plan={plan} sessions={sessions} onSave={async (next) => { const saved = await savePlan(next); setPlan(saved); setSnapshot(await loadDashboard()); }} />;
}
