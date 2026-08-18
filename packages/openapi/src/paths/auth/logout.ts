import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerLogoutPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/logout",
    summary: "Log out",
    description:
      "Revokes the current refresh token when present and always clears auth cookies",
    security: [{ refreshTokenCookie: [] }],
    responses: {
      200: {
        description: "Logged out successfully",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      500: errorResponse("Internal server error"),
    },
  });
}
