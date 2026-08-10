import { z } from "zod";
import { ErrorCodeSchema } from "./error-codes";

export const ValidationErrorSchema = z.object({
  code: z.string(),
  field: z.string().optional(),
  message: z.string(),
});

export type ValidationError = z.infer<typeof ValidationErrorSchema>;

export const createApiResponseSchema = <T extends z.ZodType>(
  dataSchema?: T,
) => {
  const baseShape = {
    success: z.boolean(),
    message: z.string(),
    code: ErrorCodeSchema.optional(),
    errors: z.array(ValidationErrorSchema).optional(),
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

export const apiErrorResponseSchema = z.object({
  success: z.literal(false),
  message: z.string(),
  code: z.string(),
  errors: z.array(ValidationErrorSchema).optional(),
});
export type ApiErrorResponse = z.infer<typeof apiErrorResponseSchema>;

export type ApiResponse<T = unknown> = {
  success: boolean;
  message: string;
  code?: z.infer<typeof ErrorCodeSchema>;
  errors?: ValidationError[];
  data?: T;
  meta?: unknown;
};

const paginationMetaSchema = z.object({
  total: z.number(),
  page: z.number(),
  limit: z.number(),
  total_pages: z.number(),
  offset: z.number(),
});

export type PaginationMeta = z.infer<typeof paginationMetaSchema>;

export type ApiSuccessResponse<T = unknown> = {
  success: true;
  message: string;
  data?: T;
  meta?: PaginationMeta;
};
