import { useEffect, useState } from 'react';
import { loadDashboard } from './api';
import { Dashboard } from './components/Dashboard';
import type { DashboardSnapshot } from './domain';

export default function App() {
  const [snapshot, setSnapshot] = useState<DashboardSnapshot>();
  const [error, setError] = useState<string>();
  useEffect(() => {
    const controller = new AbortController();
    loadDashboard(controller.signal).then(setSnapshot).catch((reason: unknown) => {
      if (!controller.signal.aborted) setError(reason instanceof Error ? reason.message : 'Unable to load mission data');
    });
    return () => controller.abort();
  }, []);
  return <Dashboard snapshot={snapshot} loading={!snapshot && !error} error={error} />;
}
