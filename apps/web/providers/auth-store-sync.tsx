"use client";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth";
import type { User } from "@repo/types";
import { getApiV1AuthMe } from "@repo/api-client";
import { apiClient } from "@/lib/api/client";

/**
 * Copies the signed-in user into the client auth store. The layout passes the
 * user it read from the access token; without one, an is_logged_in cookie means
 * the access token expired, so ask /auth/me. Renders nothing.
 */
export function AuthStoreSync({ user }: { user: User | null }) {
  const setUser = useAuthStore((state) => state.setUser);
  const clearUser = useAuthStore((state) => state.clearUser);

  useEffect(() => {
    if (user) {
      setUser(user);
      return;
    }

    const isLoggedIn = document.cookie.includes("is_logged_in=true");
    if (!isLoggedIn) {
      clearUser();
      return;
    }

    getApiV1AuthMe({ client: apiClient, throwOnError: true })
      .then((res) => setUser(res.data.data))
      .catch(() => clearUser());
  }, [user, setUser, clearUser]);

  return null;
}
