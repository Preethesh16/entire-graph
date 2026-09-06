import { Dashboard } from './components/Dashboard';
import { syntheticDashboardFixture } from './fixtures/syntheticDashboardFixture';

export default function App() {
  return <Dashboard snapshot={syntheticDashboardFixture} />;
}
