"use client";

import Link from "next/link";
import type { User } from "@repo/types";
import { homePathForRole } from "@/lib/middleware/config";
import { useAuthStore } from "@/store/auth";
import { LogoutButton } from "./logout-button";

const primaryLink =
  "inline-flex h-9 items-center border border-transparent bg-primary px-4 text-xs font-medium text-primary-foreground transition hover:bg-primary/80";
const secondaryLink =
  "inline-flex h-9 items-center border border-border px-4 text-xs font-medium transition hover:bg-muted";

/**
 * The server only sees a user whose access token hasn't expired. The store also
 * covers users restored by a refresh after the page loaded.
 */
export function LandingActions({ user: serverUser }: { user: User | null }) {
  const storeUser = useAuthStore((state) => state.user);
  const user = storeUser ?? serverUser;

  if (user) {
    return (
      <>
        <Link className={primaryLink} href={homePathForRole(user.role)}>
          Go to dashboard
        </Link>
        <LogoutButton />
      </>
    );
  }

  return (
    <>
      <Link className={primaryLink} href="/auth/register">
        Create an account
      </Link>
      <Link className={secondaryLink} href="/auth/login">
        Sign in
      </Link>
    </>
  );
}
