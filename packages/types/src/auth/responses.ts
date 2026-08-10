import { z } from "zod";
import { createApiResponseSchema } from "../api/response";
import { UserSchema } from "../user/user";

export const HealthResponseSchema = createApiResponseSchema().extend({
  success: z.literal(true),
});

export type HealthResponse = z.infer<typeof HealthResponseSchema>;

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

export const MeResponseSchema = createApiResponseSchema(UserSchema).extend({
  success: z.literal(true),
  data: UserSchema,
});

export type MeResponse = z.infer<typeof MeResponseSchema>;
