import { registry } from "../registry.ts";
import { schemas } from "../schemas/index.ts";

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
}
