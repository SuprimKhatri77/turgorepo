import { z } from "zod";

const MAX_PASSWORD_BYTES = 72;

// bcrypt rejects passwords over 72 bytes, which can be far fewer than 72
// characters once non-ASCII text is encoded.
const fitsBcrypt = (password: string) =>
  new TextEncoder().encode(password).length <= MAX_PASSWORD_BYTES;

const passwordTooLong = `Password cannot exceed ${MAX_PASSWORD_BYTES} bytes`;

export const RegisterBodySchema = z.object({
  name: z
    .string()
    .trim()
    .min(2)
    .max(50)
    .regex(
      /^[\p{L}\p{M}]+(?:[ '’.\-]+[\p{L}\p{M}]+)*\.?$/u,
      "Name can only contain letters, spaces, and ' . -",
    ),
  email: z.email(),
  password: z.string().min(8).refine(fitsBcrypt, passwordTooLong),
});

export type RegisterBody = z.infer<typeof RegisterBodySchema>;

export const LoginBodySchema = z.object({
  email: z.email(),
  password: z.string().min(1).refine(fitsBcrypt, passwordTooLong),
});

export type LoginBody = z.infer<typeof LoginBodySchema>;
