"use client";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth";
import type { User } from "@repo/types";
import { apiClient, getApiV1AuthMe } from "@/lib/api/client";

export function AuthProvider({ user }: { user: User | null }) {
  const setUser = useAuthStore((state) => state.setUser);
  const clearUser = useAuthStore((state) => state.clearUser);

  useEffect(() => {
    if (user) {
      setUser(user);
      return;
    }

    const isLoggedIn = document.cookie.includes("is_logged_in=true");
    if (!isLoggedIn) return;

    getApiV1AuthMe({ client: apiClient, throwOnError: true })
      .then((res) => setUser(res.data.data))
      .catch(() => clearUser());
    // eslint-disable-next-line react-hooks/exhaustive-deps -- hydrate auth store once on mount
  }, []);

  return null;
}
