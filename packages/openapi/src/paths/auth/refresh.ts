import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerRefreshPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/refresh",
    summary: "Refresh access token",
    description:
      "Issues a new access token using the refresh token cookie. Rotates the refresh token when the current one is older than 5 minutes.",
    security: [{ refreshTokenCookie: [] }],
    responses: {
      200: {
        description: "Tokens refreshed successfully",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      400: errorResponse("Missing refresh token"),
      401: errorResponse("Invalid refresh token"),
      500: errorResponse("Internal server error"),
    },
  });
}
