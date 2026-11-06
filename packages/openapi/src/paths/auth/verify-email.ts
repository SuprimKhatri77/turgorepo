import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerVerifyEmailPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/verify-email",
    summary: "Verify an email address",
    description:
      "Marks the email as verified using the token from a verification email. It doesn't log the user in. Tokens work once, expire after 24 hours, and only the newest one for an account works.",
    request: {
      body: {
        content: {
          "application/json": { schema: schemas.VerifyEmailBodySchema },
        },
      },
    },
    responses: {
      200: {
        description: "Email verified",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      400: errorResponse(
        "Validation failed, or the token is invalid or expired (INVALID_VERIFICATION_TOKEN)",
      ),
      413: errorResponse("Request body is too large"),
      429: errorResponse("Too many requests"),
      500: errorResponse("Internal server error"),
    },
  });
}
