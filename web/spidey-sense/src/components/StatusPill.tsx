import type { MissionStatus, ProviderStatus, RiskLevel, RunnerStatus } from '../domain';

type Status = MissionStatus | ProviderStatus | RiskLevel | RunnerStatus;

const statusLabels: Record<Status, string> = {
  queued: 'Queued',
  active: 'Active',
  blocked: 'Blocked',
  complete: 'Complete',
  online: 'Online',
  degraded: 'Degraded',
  offline: 'Offline',
  idle: 'Idle',
  BLOCK: 'Block',
  REVIEW: 'Review',
  CLEAR: 'Clear',
};

interface StatusPillProps {
  status: Status;
}

export function StatusPill({ status }: StatusPillProps) {
  return <span className={`status-pill status-${status.toLowerCase()}`}>{statusLabels[status]}</span>;
}
