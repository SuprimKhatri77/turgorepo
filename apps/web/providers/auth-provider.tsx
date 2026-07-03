"use client";
import { useEffect } from "react";
import { useAuthStore } from "@/store/auth";
import type { User } from "@repo/types";
import api from "@/lib/axios";

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

    api
      .get("/auth/me")
      .then((res) => setUser(res.data.data))
      .catch(() => clearUser());
  }, []);

  return null;
}
