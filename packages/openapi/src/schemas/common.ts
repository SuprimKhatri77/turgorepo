import { ApiErrorResponseSchema } from "@repo/types";

export const ApiErrorResponse = ApiErrorResponseSchema.openapi("ApiErrorResponse", {
  example: {
    success: false,
    message: "Invalid request body",
    code: "VALIDATION_FAILED",
    errors: [
      {
        code: "VALIDATION_FAILED",
        field: "email",
        message: "email must be a valid email address",
      },
    ],
  },
});
