import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerResetPasswordPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/reset-password",
    summary: "Set a new password",
    description:
      "Sets a new password using the token from a reset email, marks the email as verified, logs the user out on all devices and clears cookies. Tokens work once and expire after 1 hour.",
    request: {
      body: {
        content: {
          "application/json": { schema: schemas.ResetPasswordBodySchema },
        },
      },
    },
    responses: {
      200: {
        description: "Password reset",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      400: errorResponse(
        "Validation failed, or the token is invalid or expired",
      ),
      413: errorResponse("Request body is too large"),
      429: errorResponse("Too many requests"),
      500: errorResponse("Internal server error"),
    },
  });
}
