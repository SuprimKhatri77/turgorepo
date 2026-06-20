import { z } from "zod";

export const UserRoleSchema = z.enum([
  "superadmin",
  "admin",
  "staff",
  "member",
]);

export type UserRole = z.infer<typeof UserRoleSchema>;

export const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  email: z.string().email(),
  role: UserRoleSchema,
  image_url: z.string().nullable().optional(),
});

export type User = z.infer<typeof UserSchema>;
