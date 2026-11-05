import { z } from "zod";

const MAX_PASSWORD_BYTES = 72;

// bcrypt rejects passwords over 72 bytes, which can be far fewer than 72
// characters once non-ASCII text is encoded.
const fitsBcrypt = (password: string) =>
  new TextEncoder().encode(password).length <= MAX_PASSWORD_BYTES;

const passwordTooLong = `Password cannot exceed ${MAX_PASSWORD_BYTES} bytes`;

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
  email: z.email("Enter a valid email address"),
  password: newPassword,
});

export type RegisterBody = z.infer<typeof RegisterBodySchema>;

export const LoginBodySchema = z.object({
  email: z.email("Enter a valid email address"),
  password: z
    .string()
    .min(1, "Password is required")
    .refine(fitsBcrypt, passwordTooLong),
});

export type LoginBody = z.infer<typeof LoginBodySchema>;

export const ForgotPasswordBodySchema = z.object({
  email: z.email("Enter a valid email address"),
});

export type ForgotPasswordBody = z.infer<typeof ForgotPasswordBodySchema>;

export const ResetPasswordBodySchema = z.object({
  token: z.string().min(1).max(128),
  password: newPassword,
});

export type ResetPasswordBody = z.infer<typeof ResetPasswordBodySchema>;
