"use client";

import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { getApiError } from "@/lib/api/errors";
import { AuthCard, AuthCardLink } from "./auth-card";
import { ResendVerification } from "./resend-verification";
import { useVerifyEmail } from "../hooks/mutations/use-verify-email";
import { useFragmentToken } from "../hooks/use-fragment-token";

function InvalidLink() {
  return (
    <AuthCard>
      <h1 className="text-xl font-semibold">This link doesn&apos;t work</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        It may have expired, already been used, or been replaced by a newer
        link. Enter your email to get a new one.
      </p>
      <div className="mt-6">
        <ResendVerification />
      </div>
    </AuthCard>
  );
}

/**
 * Verifies the token in the URL fragment as soon as the page opens. Email
 * scanners that open links don't run JavaScript, so they can't use it up.
 * Reads window.location, so it's loaded without server rendering.
 */
export function VerifyEmailStatus() {
  const verify = useVerifyEmail();
  const token = useFragmentToken();
  // Development runs effects twice, and a token only works once.
  const sent = useRef(false);

  useEffect(() => {
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
      <AuthCard>
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
      </AuthCard>
    );
  }

  if (verify.isSuccess) {
    return (
      <AuthCard>
        <h1 className="text-xl font-semibold">Email verified</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Your account is ready. Sign in to continue.
        </p>
        <AuthCardLink href="/auth/login">Sign in</AuthCardLink>
      </AuthCard>
    );
  }

  return (
    <AuthCard role="status">
      <h1 className="text-xl font-semibold">Verifying your email…</h1>
    </AuthCard>
  );
}
