import { z } from "zod";

// Mirrors the API's constants/error_codes.go, so change both together. Codes
// marked "starter" aren't returned yet; they're kept for features most projects
// add (pagination, ID params).
export const ErrorCodeSchema = z.enum([
  "INTERNAL_SERVER_ERROR",
  "VALIDATION_FAILED",
  "UNAUTHORIZED",
  "INVALID_PAGE_PARAMETER", // starter
  "INVALID_ID_FORMAT", // starter
  "INVALID_QUERY_PARAM", // starter
  "PAGE_NOT_FOUND", // starter
  "RATE_LIMITED",
  "SERVICE_UNAVAILABLE",
  "PAYLOAD_TOO_LARGE",
  "INVALID_CREDENTIALS",
  "INVALID_EMAIL", // starter
  "INVALID_PASSWORD", // starter
  "INVALID_TOKEN",
  "TOKEN_EXPIRED", // starter
  "TOKEN_NOT_PROVIDED",
  "TOKEN_INVALID",
  "FORBIDDEN",
  "USER_NOT_FOUND",
  "USER_ALREADY_EXISTS",
  "INVALID_RESET_TOKEN",
  "EMAIL_NOT_VERIFIED",
  "INVALID_VERIFICATION_TOKEN",
]);

export type ErrorCode = z.infer<typeof ErrorCodeSchema>;
