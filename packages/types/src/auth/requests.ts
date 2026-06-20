import { z } from "zod";

export const RegisterBodySchema = z.object({
  name: z
    .string()
    .min(2)
    .max(50)
    .regex(/^[a-zA-Z\s]+$/, "Name must contain only letters and spaces"),
  email: z.email(),
  password: z.string().min(8).max(50),
});

export type RegisterBody = z.infer<typeof RegisterBodySchema>;

export const LoginBodySchema = z.object({
  email: z.email(),
  password: z.string().min(8).max(50),
});

export type LoginBody = z.infer<typeof LoginBodySchema>;
