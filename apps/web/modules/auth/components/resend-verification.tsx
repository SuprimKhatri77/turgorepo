"use client";

import { useState, type SubmitEvent } from "react";
import { ResendVerificationBodySchema } from "@repo/types";
import { Button } from "@/components/ui/button";
import { FieldGroup } from "@/components/ui/field";
import { TextField } from "@/components/form/text-field";
import { toFieldErrors } from "@/modules/auth/lib/form-errors";
import { useResendVerification } from "../hooks/mutations/use-resend-verification";

/**
 * Sends a new verification link. With a known address it's just a button,
 * otherwise it asks for the address.
 */
export function ResendVerification({ email }: { email?: string | null }) {
  const resend = useResendVerification();
  const [typedEmail, setTypedEmail] = useState("");
  const [emailError, setEmailError] = useState<string>();

  function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();

    const parsed = ResendVerificationBodySchema.safeParse({
      email: email ?? typedEmail,
    });
    if (!parsed.success) {
      setEmailError(toFieldErrors(parsed.error.issues).email);
      return;
    }

    setEmailError(undefined);
    resend.mutate(parsed.data);
  }

  if (resend.isSuccess) {
    return (
      <p className="text-sm text-muted-foreground" role="status">
        {resend.data.data.message}. Only the newest link works.
      </p>
    );
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <FieldGroup>
        {!email && (
          <TextField
            id="email"
            label="Email"
            type="email"
            autoComplete="email"
            value={typedEmail}
            onChange={setTypedEmail}
            error={emailError}
          />
        )}
        <Button
          className="w-full"
          type="submit"
          variant={email ? "outline" : "default"}
          disabled={resend.isPending}
        >
          {resend.isPending ? "Sending…" : "Send a new link"}
        </Button>
      </FieldGroup>
    </form>
  );
}
