import { z } from "zod";

const MAX_PASSWORD_BYTES = 72;

// bcrypt rejects passwords over 72 bytes, which can be far fewer than 72
// characters once non-ASCII text is encoded.
const fitsBcrypt = (password: string) =>
  new TextEncoder().encode(password).length <= MAX_PASSWORD_BYTES;

const passwordTooLong = `Password cannot exceed ${MAX_PASSWORD_BYTES} bytes`;

// Trimmed before the format check, so a pasted address with a trailing space
// passes. The API trims the same way before validating. preprocess keeps
// `format: email` in the generated OpenAPI spec, which .pipe() would drop.
const email = z.preprocess(
  (value) => (typeof value === "string" ? value.trim() : value),
  z.email("Enter a valid email address"),
);

const newPassword = z
  .string()
  .min(8, "Password must be at least 8 characters")
  .refine(fitsBcrypt, passwordTooLong);

export const RegisterBodySchema = z.object({
  name: z
    .string()
    .trim()
    .min(2, "Name must be at least 2 characters")
    .max(50, "Name cannot exceed 50 characters")
    .regex(
      /^[\p{L}\p{M}]+(?:[ '’.\-]+[\p{L}\p{M}]+)*\.?$/u,
      "Name can only contain letters, spaces, and ' . -",
    ),
  email,
  password: newPassword,
});

export type RegisterBody = z.infer<typeof RegisterBodySchema>;

export const LoginBodySchema = z.object({
  email,
  password: z
    .string()
    .min(1, "Password is required")
    .refine(fitsBcrypt, passwordTooLong),
});

export type LoginBody = z.infer<typeof LoginBodySchema>;

export const ForgotPasswordBodySchema = z.object({
  email,
});

export type ForgotPasswordBody = z.infer<typeof ForgotPasswordBodySchema>;

export const ResetPasswordBodySchema = z.object({
  token: z.string().min(1).max(128),
  password: newPassword,
});

export type ResetPasswordBody = z.infer<typeof ResetPasswordBodySchema>;

export const VerifyEmailBodySchema = z.object({
  token: z.string().min(1).max(128),
});

export type VerifyEmailBody = z.infer<typeof VerifyEmailBodySchema>;

export const ResendVerificationBodySchema = z.object({
  email,
});

export type ResendVerificationBody = z.infer<
  typeof ResendVerificationBodySchema
>;
