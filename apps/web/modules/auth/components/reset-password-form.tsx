"use client";

import { useState, type SubmitEvent } from "react";
import { ResetPasswordBodySchema } from "@repo/types";
import { Button } from "@/components/ui/button";
import { FieldGroup } from "@/components/ui/field";
import { TextField } from "@/components/form/text-field";
import { getApiError } from "@/lib/api/errors";
import { mapFieldErrors, toFieldErrors } from "@/modules/auth/lib/form-errors";
import { AuthCard, AuthCardForm, AuthCardLink } from "./auth-card";
import { useResetPassword } from "../hooks/mutations/use-reset-password";
import { useFragmentToken } from "../hooks/use-fragment-token";

function InvalidLink() {
  return (
    <AuthCard>
      <h1 className="text-xl font-semibold">This link doesn&apos;t work</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        It may have expired or already been used. Request a new one to reset
        your password.
      </p>
      <AuthCardLink href="/auth/forgot-password">
        Request a new link
      </AuthCardLink>
    </AuthCard>
  );
}

/** Reads window.location, so it's loaded without server rendering. */
export function ResetPasswordForm() {
  const resetPassword = useResetPassword();
  const token = useFragmentToken();
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [linkInvalid, setLinkInvalid] = useState(false);

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
    <AuthCardForm onSubmit={handleSubmit}>
      <div className="mb-6 space-y-1">
        <h1 className="text-xl font-semibold">Set a new password</h1>
        <p className="text-sm text-muted-foreground">
          You&apos;ll be logged out on all devices, then you can sign in with
          your new password.
        </p>
      </div>

      <FieldGroup>
        <TextField
          id="password"
          label="New password"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={setPassword}
          error={fieldErrors.password}
        />

        <TextField
          id="confirmPassword"
          label="Confirm new password"
          type="password"
          autoComplete="new-password"
          value={confirmPassword}
          onChange={setConfirmPassword}
          error={fieldErrors.confirmPassword}
        />

        <Button
          className="w-full"
          type="submit"
          disabled={resetPassword.isPending}
        >
          {resetPassword.isPending ? "Saving…" : "Reset password"}
        </Button>
      </FieldGroup>
    </AuthCardForm>
  );
}
