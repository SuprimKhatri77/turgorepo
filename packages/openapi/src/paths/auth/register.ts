import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerRegisterPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/register",
    summary: "Register a new user",
    description:
      "Creates a new member account and sets access, refresh, and session cookies",
    request: {
      body: {
        content: {
          "application/json": { schema: schemas.RegisterBodySchema },
        },
      },
    },
    responses: {
      201: {
        description: "Registration successful",
        content: {
          "application/json": { schema: schemas.AuthUserResponse },
        },
      },
      400: errorResponse("Validation failed"),
      409: errorResponse("User already exists"),
      500: errorResponse("Internal server error"),
    },
  });
}
