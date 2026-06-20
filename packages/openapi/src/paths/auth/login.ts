import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerLoginPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/login",
    summary: "Log in",
    description:
      "Authenticates a user and sets access, refresh, and session cookies",
    request: {
      body: {
        content: {
          "application/json": { schema: schemas.LoginBodySchema },
        },
      },
    },
    responses: {
      200: {
        description: "Logged in successfully",
        content: {
          "application/json": { schema: schemas.AuthUserResponse },
        },
      },
      400: errorResponse("Validation failed"),
      401: errorResponse("Invalid credentials"),
      500: errorResponse("Internal server error"),
    },
  });
}
