import { registry } from "../registry.ts";
import * as auth from "./auth.ts";
import * as common from "./common.ts";
import * as user from "./user.ts";

export const schemas = {
  ...common,
  ...user,
  ...auth,
};

function registerSecuritySchemes() {
  registry.registerComponent("securitySchemes", "accessTokenCookie", {
    type: "apiKey",
    in: "cookie",
    name: "access_token",
    description: "Short-lived JWT access token (15 minutes)",
  });

  registry.registerComponent("securitySchemes", "refreshTokenCookie", {
    type: "apiKey",
    in: "cookie",
    name: "refresh_token",
    description: "Long-lived JWT refresh token (30 days)",
  });
}

export function registerSchemas() {
  registry.register("User", schemas.UserSchema);
  registry.register("RegisterBody", schemas.RegisterBodySchema);
  registry.register("LoginBody", schemas.LoginBodySchema);
  registry.register("HealthResponse", schemas.HealthResponse);
  registry.register("AuthUserResponse", schemas.AuthUserResponse);
  registry.register("MeResponse", schemas.MeResponse);
  registry.register("AuthSuccessResponse", schemas.AuthSuccessResponse);
  registry.register("ApiErrorResponse", schemas.ApiErrorResponse);

  registerSecuritySchemes();
}
