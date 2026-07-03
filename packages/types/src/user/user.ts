import { z } from "zod";

export const UserRoleSchema = z.enum([
  "superadmin",
  "admin",
  "staff",
  "member",
]);

export type UserRole = z.infer<typeof UserRoleSchema>;

export const UserSchema = z.object({
  id: z.uuid(),
  name: z.string(),
  email: z.email(),
  role: UserRoleSchema,
  image_url: z.url().optional(),
});

export type User = z.infer<typeof UserSchema>;
