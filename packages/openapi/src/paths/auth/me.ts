import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerMePath() {
  registry.registerPath({
    method: "get",
    path: "/api/v1/auth/me",
    summary: "Get current user",
    description: "Returns the authenticated user's profile",
    security: [{ accessTokenCookie: [] }],
    responses: {
      200: {
        description: "Valid session",
        content: {
          "application/json": { schema: schemas.MeResponse },
        },
      },
      400: errorResponse("Invalid user ID format"),
      401: errorResponse("Missing or invalid access token"),
      403: errorResponse("Insufficient permissions"),
      404: errorResponse("User not found"),
      500: errorResponse("Internal server error"),
    },
  });
}
