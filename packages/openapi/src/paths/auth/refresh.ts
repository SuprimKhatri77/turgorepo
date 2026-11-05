import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerRefreshPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/refresh",
    summary: "Refresh access token",
    description:
      "Issues a new access token and rotates the refresh token. The previous refresh token keeps working for 5 minutes so concurrent requests don't fail, but it only gets a new access token.",
    security: [{ refreshTokenCookie: [] }],
    responses: {
      200: {
        description: "Tokens refreshed successfully",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      401: errorResponse("Missing or invalid refresh token"),
      500: errorResponse("Internal server error"),
      503: errorResponse("Session store unavailable; cookies are kept"),
    },
  });
}
