import { z } from "zod";
import { createApiResponseSchema } from "../api/response";
import { UserSchema } from "../user/user";

export const AuthUserResponseSchema = createApiResponseSchema(
  UserSchema,
).extend({
  success: z.literal(true),
  data: UserSchema,
});

export type AuthUserResponse = z.infer<typeof AuthUserResponseSchema>;

export const AuthSuccessResponseSchema = createApiResponseSchema().extend({
  success: z.literal(true),
});

export type AuthSuccessResponse = z.infer<typeof AuthSuccessResponseSchema>;

// The same shape as AuthUserResponseSchema. It keeps its own name so the
// OpenAPI spec documents /auth/me separately.
export const MeResponseSchema = AuthUserResponseSchema;

export type MeResponse = z.infer<typeof MeResponseSchema>;
