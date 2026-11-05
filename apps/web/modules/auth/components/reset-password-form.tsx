"use client";

import { useEffect, useState, type SubmitEvent } from "react";
import Link from "next/link";
import { ResetPasswordBodySchema } from "@repo/types";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { getApiError } from "@/utils/get-api-error";
import { mapFieldErrors } from "@/utils/map-field-errors";
import { inputClassName, toFieldErrors } from "./auth-form";
import { useResetPassword } from "../hooks/mutations/use-reset-password";

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
        It may have expired or already been used. Request a new one to reset
        your password.
      </p>
      <p className="mt-5 text-center text-sm">
        <Link
          className="font-medium underline underline-offset-4"
          href="/auth/forgot-password"
        >
          Request a new link
        </Link>
      </p>
    </div>
  );
}

/** Reads window.location, so it's loaded without server rendering. */
export function ResetPasswordForm() {
  const resetPassword = useResetPassword();
  const [token] = useState(tokenFromUrl);
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [linkInvalid, setLinkInvalid] = useState(false);

  // Keeps the token out of the address bar and browser history.
  useEffect(() => {
    if (window.location.hash) {
      window.history.replaceState(null, "", window.location.pathname);
    }
  }, []);

  if (!token || linkInvalid) {
    return <InvalidLink />;
  }

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();

    const parsed = ResetPasswordBodySchema.safeParse({ token, password });
    if (!parsed.success) {
      setFieldErrors(toFieldErrors(parsed.error.issues));
      return;
    }
    if (password !== confirmPassword) {
      setFieldErrors({ confirmPassword: "Passwords don't match" });
      return;
    }

    setFieldErrors({});
    resetPassword.mutate(parsed.data, {
      onError: (error) => {
        const apiError = getApiError(error);
        if (apiError?.code === "INVALID_RESET_TOKEN") {
          setLinkInvalid(true);
        } else if (apiError?.errors?.length) {
          setFieldErrors(mapFieldErrors(apiError.errors));
        } else {
          setFieldErrors({
            password:
              apiError?.message ?? "Something went wrong. Please try again",
          });
        }
      },
    });
  }

  return (
    <form className={cardClassName} onSubmit={handleSubmit} noValidate>
      <div className="mb-6 space-y-1">
        <h1 className="text-xl font-semibold">Set a new password</h1>
        <p className="text-sm text-muted-foreground">
          You&apos;ll be logged out on all devices, then you can sign in with
          your new password.
        </p>
      </div>

      <FieldGroup>
        <Field data-invalid={Boolean(fieldErrors.password)}>
          <FieldLabel htmlFor="password">New password</FieldLabel>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete="new-password"
            className={inputClassName}
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            aria-invalid={Boolean(fieldErrors.password)}
          />
          <FieldError>{fieldErrors.password}</FieldError>
        </Field>

        <Field data-invalid={Boolean(fieldErrors.confirmPassword)}>
          <FieldLabel htmlFor="confirmPassword">
            Confirm new password
          </FieldLabel>
          <input
            id="confirmPassword"
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
            className={inputClassName}
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
            aria-invalid={Boolean(fieldErrors.confirmPassword)}
          />
          <FieldError>{fieldErrors.confirmPassword}</FieldError>
        </Field>

        <Button
          className="w-full"
          type="submit"
          disabled={resetPassword.isPending}
        >
          {resetPassword.isPending ? "Saving…" : "Reset password"}
        </Button>
      </FieldGroup>
    </form>
  );
}
