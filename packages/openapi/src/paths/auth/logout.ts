import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerLogoutPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/logout",
    summary: "Log out",
    description:
      "Ends this device's session, found from the refresh token or else the access token, so its access token stops working straight away. Other devices stay logged in. Always clears auth cookies",
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
