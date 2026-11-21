import { useEffect, useState } from "react";

/**
 * Reads the token from the URL fragment (#token=...) that email links carry,
 * then removes the fragment so the token stays out of the address bar and
 * browser history. Reads window.location, so the component that calls it must
 * be loaded without server rendering.
 */
export function useFragmentToken(): string | null {
  const [token] = useState(() =>
    new URLSearchParams(window.location.hash.slice(1)).get("token"),
  );

  useEffect(() => {
    if (window.location.hash) {
      window.history.replaceState(null, "", window.location.pathname);
    }
  }, []);

  return token;
}
