import { z } from "zod";
import { ErrorCodeSchema } from "./error-codes.js";

export const AppErrorSchema = z.object({
  code: z.string(),
  field: z.string().optional(),
  message: z.string(),
});

export type AppError = z.infer<typeof AppErrorSchema>;

export const createApiResponseSchema = <T extends z.ZodType>(dataSchema?: T) => {
  const baseShape = {
    success: z.boolean(),
    message: z.string().optional(),
    code: ErrorCodeSchema.optional(),
    errors: z.array(AppErrorSchema).optional(),
    meta: z.unknown().optional(),
  };

  if (dataSchema) {
    return z.object({
      ...baseShape,
      data: dataSchema,
    });
  }

  return z.object(baseShape);
};

export const ApiSuccessResponseSchema = createApiResponseSchema();
export const ApiErrorResponseSchema = createApiResponseSchema().extend({
  success: z.literal(false),
});

export type ApiResponse<T = unknown> = {
  success: boolean;
  message?: string;
  code?: z.infer<typeof ErrorCodeSchema>;
  errors?: AppError[];
  data?: T;
  meta?: unknown;
};
