import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerResendVerificationPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/resend-verification",
    summary: "Resend the verification email",
    description:
      "Emails a new verification link if the address has an account that isn't verified yet, which makes older links stop working. The response is the same either way, so it can't be used to discover accounts.",
    request: {
      body: {
        content: {
          "application/json": { schema: schemas.ResendVerificationBodySchema },
        },
      },
    },
    responses: {
      200: {
        description: "Request accepted",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      400: errorResponse("Validation failed"),
      413: errorResponse("Request body is too large"),
      429: errorResponse("Too many requests"),
    },
  });
}
