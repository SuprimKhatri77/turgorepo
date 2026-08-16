"use client";

import { useState, type SubmitEvent } from "react";
import Link from "next/link";
import {
  LoginBodySchema,
  RegisterBodySchema,
  type LoginBody,
  type RegisterBody,
} from "@repo/types";

import { getApiError } from "@/utils/get-api-error";
import { mapFieldErrors } from "@/utils/map-field-errors";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";

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

const inputClassName =
  "h-9 w-full border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-1 focus:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-1 aria-invalid:ring-destructive/20";

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
    <form
      className="w-full max-w-sm border border-border bg-card p-6 text-card-foreground shadow-sm"
      onSubmit={handleSubmit}
      noValidate
    >
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
          <Field data-invalid={Boolean(fieldErrors.name)}>
            <FieldLabel htmlFor="name">Name</FieldLabel>
            <input
              id="name"
              name="name"
              autoComplete="name"
              className={inputClassName}
              value={values.name}
              onChange={(event) => updateField("name", event.target.value)}
              aria-invalid={Boolean(fieldErrors.name)}
            />
            <FieldError>{fieldErrors.name}</FieldError>
          </Field>
        )}

        <Field data-invalid={Boolean(fieldErrors.email)}>
          <FieldLabel htmlFor="email">Email</FieldLabel>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            className={inputClassName}
            value={values.email}
            onChange={(event) => updateField("email", event.target.value)}
            aria-invalid={Boolean(fieldErrors.email)}
          />
          <FieldError>{fieldErrors.email}</FieldError>
        </Field>

        <Field data-invalid={Boolean(fieldErrors.password)}>
          <FieldLabel htmlFor="password">Password</FieldLabel>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete={isRegister ? "new-password" : "current-password"}
            className={inputClassName}
            value={values.password}
            onChange={(event) => updateField("password", event.target.value)}
            aria-invalid={Boolean(fieldErrors.password)}
          />
          <FieldError>{fieldErrors.password}</FieldError>
        </Field>

        <Button className="w-full" type="submit" disabled={props.isPending}>
          {props.isPending
            ? isRegister
              ? "Creating account…"
              : "Signing in…"
            : isRegister
              ? "Create account"
              : "Sign in"}
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
    </form>
  );
}

function toFieldErrors(
  issues: ReadonlyArray<{ path: PropertyKey[]; message: string }>,
): Record<string, string> {
  const errors: Record<string, string> = {};
  for (const issue of issues) {
    const field = issue.path[0];
    if (typeof field === "string" && !errors[field]) {
      errors[field] = issue.message;
    }
  }
  return errors;
}
