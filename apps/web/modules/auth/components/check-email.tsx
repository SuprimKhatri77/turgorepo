"use client";

import Link from "next/link";
import { useAuthStore } from "@/store/auth";
import { ResendVerification } from "./resend-verification";

const cardClassName =
  "w-full max-w-sm border border-border bg-card p-6 text-card-foreground shadow-sm";

/** Shown after registering, and when an unverified user tries to sign in. */
export function CheckEmail() {
  const email = useAuthStore((state) => state.pendingEmail);

  return (
    <div className={cardClassName}>
      <h1 className="text-xl font-semibold">Verify your email</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        {email ? (
          <>
            We sent a link to <span className="text-foreground">{email}</span>.
          </>
        ) : (
          "We sent a link to the address you registered with."
        )}{" "}
        Open it to finish setting up your account, then sign in. It expires in
        24 hours.
      </p>

      <div className="mt-6 space-y-3">
        <p className="text-sm font-medium">Didn&apos;t get it?</p>
        <ResendVerification email={email} />
      </div>

      <p className="mt-5 text-center text-sm">
        <Link
          className="font-medium underline underline-offset-4"
          href="/auth/login"
        >
          Back to sign in
        </Link>
      </p>
    </div>
  );
}
