"use client";

import { useState, type SubmitEvent } from "react";
import Link from "next/link";
import {
  LoginBodySchema,
  RegisterBodySchema,
  type LoginBody,
  type RegisterBody,
} from "@repo/types";

import { getApiError } from "@/lib/api/errors";
import { mapFieldErrors, toFieldErrors } from "@/modules/auth/lib/form-errors";
import { TextField } from "@/components/form/text-field";
import { AuthCardForm } from "./auth-card";
import { Button } from "@/components/ui/button";
import { FieldGroup } from "@/components/ui/field";

type AuthFormProps =
  | {
      mode: "login";
      isPending: boolean;
      onSubmit: (body: LoginBody) => Promise<unknown>;
    }
  | {
      mode: "register";
      isPending: boolean;
      onSubmit: (body: RegisterBody) => Promise<unknown>;
    };

type FormValues = {
  name: string;
  email: string;
  password: string;
};

const submitLabel = {
  login: { idle: "Sign in", pending: "Signing in…" },
  register: { idle: "Create account", pending: "Creating account…" },
} as const;

export function AuthForm(props: AuthFormProps) {
  const [values, setValues] = useState<FormValues>({
    name: "",
    email: "",
    password: "",
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  function updateField(field: keyof FormValues, value: string) {
    setValues((current) => ({ ...current, [field]: value }));
    setFieldErrors((current) => {
      if (!current[field]) return current;
      const next = { ...current };
      delete next[field];
      return next;
    });
  }

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    setFieldErrors({});

    try {
      if (props.mode === "register") {
        const parsed = RegisterBodySchema.safeParse(values);
        if (!parsed.success) {
          setFieldErrors(toFieldErrors(parsed.error.issues));
          return;
        }
        await props.onSubmit(parsed.data);
        return;
      }

      const parsed = LoginBodySchema.safeParse({
        email: values.email,
        password: values.password,
      });
      if (!parsed.success) {
        setFieldErrors(toFieldErrors(parsed.error.issues));
        return;
      }
      await props.onSubmit(parsed.data);
    } catch (error) {
      const apiError = getApiError(error);
      if (apiError?.errors?.length) {
        setFieldErrors(mapFieldErrors(apiError.errors));
      }
    }
  }

  const isRegister = props.mode === "register";

  return (
    <AuthCardForm onSubmit={handleSubmit}>
      <div className="mb-6 space-y-1">
        <p className="text-xs uppercase tracking-widest text-muted-foreground">
          Turgorepo
        </p>
        <h1 className="text-xl font-semibold">
          {isRegister ? "Create your account" : "Welcome back"}
        </h1>
        <p className="text-sm text-muted-foreground">
          {isRegister
            ? "Register with the typed OpenAPI client."
            : "Sign in with the typed OpenAPI client."}
        </p>
      </div>

      <FieldGroup>
        {isRegister && (
          <TextField
            id="name"
            label="Name"
            autoComplete="name"
            value={values.name}
            onChange={(value) => updateField("name", value)}
            error={fieldErrors.name}
          />
        )}

        <TextField
          id="email"
          label="Email"
          type="email"
          autoComplete="email"
          value={values.email}
          onChange={(value) => updateField("email", value)}
          error={fieldErrors.email}
        />

        <TextField
          id="password"
          label="Password"
          type="password"
          autoComplete={isRegister ? "new-password" : "current-password"}
          value={values.password}
          onChange={(value) => updateField("password", value)}
          error={fieldErrors.password}
          labelAside={
            isRegister ? undefined : (
              <Link
                className="text-xs text-muted-foreground underline underline-offset-4"
                href="/auth/forgot-password"
              >
                Forgot password?
              </Link>
            )
          }
        />

        <Button className="w-full" type="submit" disabled={props.isPending}>
          {submitLabel[props.mode][props.isPending ? "pending" : "idle"]}
        </Button>
      </FieldGroup>

      <p className="mt-5 text-center text-sm text-muted-foreground">
        {isRegister ? "Already have an account?" : "New to Turgorepo?"}{" "}
        <Link
          className="font-medium text-foreground underline underline-offset-4"
          href={isRegister ? "/auth/login" : "/auth/register"}
        >
          {isRegister ? "Sign in" : "Create an account"}
        </Link>
      </p>
    </AuthCardForm>
  );
}
