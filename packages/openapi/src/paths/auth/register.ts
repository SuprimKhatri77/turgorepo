import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerRegisterPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/register",
    summary: "Register a new user",
    description:
      "Creates a new member account and emails a link to verify the address. It doesn't log the user in: login is refused until the email is verified.",
    request: {
      body: {
        content: {
          "application/json": { schema: schemas.RegisterBodySchema },
        },
      },
    },
    responses: {
      201: {
        description: "Account created and verification email queued",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      400: errorResponse("Validation failed"),
      409: errorResponse("User already exists"),
      413: errorResponse("Request body is too large"),
      429: errorResponse("Too many registration attempts"),
      500: errorResponse("Internal server error"),
    },
  });
}
