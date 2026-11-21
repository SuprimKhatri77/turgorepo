import { z } from "zod";
import { createApiResponseSchema } from "./response";

export const HealthResponseSchema = createApiResponseSchema().extend({
  success: z.literal(true),
});

export type HealthResponse = z.infer<typeof HealthResponseSchema>;
