"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { getApiError } from "@/utils/get-api-error";
import { ResendVerification } from "./resend-verification";
import { useVerifyEmail } from "../hooks/mutations/use-verify-email";

const cardClassName =
  "w-full max-w-sm border border-border bg-card p-6 text-card-foreground shadow-sm";

function tokenFromUrl(): string | null {
  return new URLSearchParams(window.location.hash.slice(1)).get("token");
}

function InvalidLink() {
  return (
    <div className={cardClassName}>
      <h1 className="text-xl font-semibold">This link doesn&apos;t work</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        It may have expired, already been used, or been replaced by a newer
        link. Enter your email to get a new one.
      </p>
      <div className="mt-6">
        <ResendVerification />
      </div>
    </div>
  );
}

/**
 * Verifies the token in the URL fragment as soon as the page opens. Email
 * scanners that open links don't run JavaScript, so they can't use it up.
 * Reads window.location, so it's loaded without server rendering.
 */
export function VerifyEmailStatus() {
  const verify = useVerifyEmail();
  const [token] = useState(tokenFromUrl);
  // Development runs effects twice, and a token only works once.
  const sent = useRef(false);

  useEffect(() => {
    // Keeps the token out of the address bar and browser history.
    if (window.location.hash) {
      window.history.replaceState(null, "", window.location.pathname);
    }
    if (token && !sent.current) {
      sent.current = true;
      verify.mutate({ token });
    }
  }, [token, verify]);

  if (!token) {
    return <InvalidLink />;
  }

  if (verify.isError) {
    if (getApiError(verify.error)?.code === "INVALID_VERIFICATION_TOKEN") {
      return <InvalidLink />;
    }
    return (
      <div className={cardClassName}>
        <h1 className="text-xl font-semibold">Something went wrong</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          We couldn&apos;t verify your email just now. Your link still works.
        </p>
        <Button
          className="mt-6 w-full"
          onClick={() => verify.mutate({ token })}
          disabled={verify.isPending}
        >
          Try again
        </Button>
      </div>
    );
  }

  if (verify.isSuccess) {
    return (
      <div className={cardClassName}>
        <h1 className="text-xl font-semibold">Email verified</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Your account is ready. Sign in to continue.
        </p>
        <p className="mt-5 text-center text-sm">
          <Link
            className="font-medium underline underline-offset-4"
            href="/auth/login"
          >
            Sign in
          </Link>
        </p>
      </div>
    );
  }

  return (
    <div className={cardClassName} role="status">
      <h1 className="text-xl font-semibold">Verifying your email…</h1>
    </div>
  );
}
