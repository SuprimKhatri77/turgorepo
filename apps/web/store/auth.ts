import { User } from "@repo/types";
import { create } from "zustand";

interface AuthStore {
  user: User | null;
  setUser: (user: User) => void;
  clearUser: () => void;
  /** The address waiting for verification, so it stays out of the URL. */
  pendingEmail: string | null;
  setPendingEmail: (email: string) => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
  clearUser: () => set({ user: null }),
  pendingEmail: null,
  setPendingEmail: (pendingEmail) => set({ pendingEmail }),
}));
