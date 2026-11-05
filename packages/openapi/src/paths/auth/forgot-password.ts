import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerForgotPasswordPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/forgot-password",
    summary: "Request a password reset",
    description:
      "Emails a reset link if the address has an account. The response is the same whether or not it does, so it can't be used to discover accounts.",
    request: {
      body: {
        content: {
          "application/json": { schema: schemas.ForgotPasswordBodySchema },
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
