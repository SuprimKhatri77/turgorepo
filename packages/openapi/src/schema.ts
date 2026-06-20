import "./zod-extend.ts";
import { OpenApiGeneratorV3 } from "@asteasolutions/zod-to-openapi";
import { registerPaths } from "./paths/index.ts";
import { registry } from "./registry.ts";
import { registerSchemas } from "./schemas/index.ts";

type OpenAPIDocument = ReturnType<
  InstanceType<typeof OpenApiGeneratorV3>["generateDocument"]
>;

registerSchemas();
registerPaths();

export function generateOpenAPIDocument(): OpenAPIDocument {
  const generator = new OpenApiGeneratorV3(registry.definitions);
  return generator.generateDocument({
    openapi: "3.0.3",
    info: {
      title: "Turgorepo API",
      version: "1.0.0",
      description:
        "API for the Turgorepo monorepo (Go/Gin backend). Auth uses HTTP-only cookies for access and refresh tokens.",
    },
    servers: [{ url: "/", description: "Current host" }],
  });
}
