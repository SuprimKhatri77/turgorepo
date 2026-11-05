import { registry } from "../registry.ts";
import { schemas } from "../schemas/index.ts";
import { errorResponse } from "./shared.ts";

export function registerHealthPath() {
  registry.registerPath({
    method: "get",
    path: "/api/v1/health",
    summary: "Health check",
    description: "Returns whether the API server is running",
    responses: {
      200: {
        description: "Server is healthy",
        content: {
          "application/json": { schema: schemas.HealthResponse },
        },
      },
    },
  });

  registry.registerPath({
    method: "get",
    path: "/api/v1/ready",
    summary: "Readiness check",
    description: "Returns whether the API can reach its database and Redis",
    responses: {
      200: {
        description: "Service is ready",
        content: {
          "application/json": { schema: schemas.HealthResponse },
        },
      },
      503: errorResponse("A dependency is unavailable"),
    },
  });
}
