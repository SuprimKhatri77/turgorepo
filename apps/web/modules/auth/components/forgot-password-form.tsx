"use client";

import { useState, type SubmitEvent } from "react";
import Link from "next/link";
import { ForgotPasswordBodySchema } from "@repo/types";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { inputClassName, toFieldErrors } from "./auth-form";
import { useForgotPassword } from "../hooks/mutations/use-forgot-password";

const cardClassName =
  "w-full max-w-sm border border-border bg-card p-6 text-card-foreground shadow-sm";

export function ForgotPasswordForm() {
  const forgotPassword = useForgotPassword();
  const [email, setEmail] = useState("");
  const [emailError, setEmailError] = useState<string>();

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();

    const parsed = ForgotPasswordBodySchema.safeParse({ email });
    if (!parsed.success) {
      setEmailError(toFieldErrors(parsed.error.issues).email);
      return;
    }

    setEmailError(undefined);
    forgotPassword.mutate(parsed.data);
  }

  if (forgotPassword.isSuccess) {
    return (
      <div className={cardClassName}>
        <h1 className="text-xl font-semibold">Check your email</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          If {email} has an account, we&apos;ve sent a link to reset the
          password. It expires in 1 hour.
        </p>
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

  return (
    <form className={cardClassName} onSubmit={handleSubmit} noValidate>
      <div className="mb-6 space-y-1">
        <h1 className="text-xl font-semibold">Forgot your password?</h1>
        <p className="text-sm text-muted-foreground">
          Enter your email and we&apos;ll send you a link to reset it.
        </p>
      </div>

      <FieldGroup>
        <Field data-invalid={Boolean(emailError)}>
          <FieldLabel htmlFor="email">Email</FieldLabel>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            className={inputClassName}
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            aria-invalid={Boolean(emailError)}
          />
          <FieldError>{emailError}</FieldError>
        </Field>

        <Button
          className="w-full"
          type="submit"
          disabled={forgotPassword.isPending}
        >
          {forgotPassword.isPending ? "Sending…" : "Send reset link"}
        </Button>
      </FieldGroup>

      <p className="mt-5 text-center text-sm text-muted-foreground">
        <Link
          className="font-medium text-foreground underline underline-offset-4"
          href="/auth/login"
        >
          Back to sign in
        </Link>
      </p>
    </form>
  );
}
