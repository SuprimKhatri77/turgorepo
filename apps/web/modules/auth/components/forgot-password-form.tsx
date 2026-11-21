"use client";

import { useState, type SubmitEvent } from "react";
import { ForgotPasswordBodySchema } from "@repo/types";
import { Button } from "@/components/ui/button";
import { FieldGroup } from "@/components/ui/field";
import { TextField } from "@/components/form/text-field";
import { toFieldErrors } from "@/modules/auth/lib/form-errors";
import { AuthCard, AuthCardForm, AuthCardLink } from "./auth-card";
import { useForgotPassword } from "../hooks/mutations/use-forgot-password";

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
      <AuthCard>
        <h1 className="text-xl font-semibold">Check your email</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          If {email} has an account, we&apos;ve sent a link to reset the
          password. It expires in 1 hour.
        </p>
        <AuthCardLink href="/auth/login">Back to sign in</AuthCardLink>
      </AuthCard>
    );
  }

  return (
    <AuthCardForm onSubmit={handleSubmit}>
      <div className="mb-6 space-y-1">
        <h1 className="text-xl font-semibold">Forgot your password?</h1>
        <p className="text-sm text-muted-foreground">
          Enter your email and we&apos;ll send you a link to reset it.
        </p>
      </div>

      <FieldGroup>
        <TextField
          id="email"
          label="Email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={setEmail}
          error={emailError}
        />

        <Button
          className="w-full"
          type="submit"
          disabled={forgotPassword.isPending}
        >
          {forgotPassword.isPending ? "Sending…" : "Send reset link"}
        </Button>
      </FieldGroup>

      <AuthCardLink href="/auth/login">Back to sign in</AuthCardLink>
    </AuthCardForm>
  );
}
