import { Dashboard } from "./Dashboard";
import { Login } from "./Login";

// Minimal path switch; swap for a router when the app grows more pages.
export function App({ path = window.location.pathname }: { path?: string }) {
  switch (path) {
    case "/dashboard":
      return <Dashboard />;
    case "/":
    case "/login":
    default:
      return <Login />;
  }
}
