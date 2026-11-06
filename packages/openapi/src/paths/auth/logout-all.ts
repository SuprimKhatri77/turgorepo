import { registry } from "../../registry.ts";
import { schemas } from "../../schemas/index.ts";
import { errorResponse } from "../shared.ts";

export function registerLogoutAllPath() {
  registry.registerPath({
    method: "post",
    path: "/api/v1/auth/logout-all",
    summary: "Log out on all devices",
    description:
      "Ends every one of the user's sessions, including this one, so their access tokens stop working straight away and none can be refreshed. Clears auth cookies",
    security: [{ accessTokenCookie: [] }],
    responses: {
      200: {
        description: "Logged out on all devices",
        content: {
          "application/json": { schema: schemas.AuthSuccessResponse },
        },
      },
      400: errorResponse("Invalid user ID format"),
      401: errorResponse("Missing or invalid access token"),
      500: errorResponse("Sessions could not be revoked, so nothing changed"),
    },
  });
}
