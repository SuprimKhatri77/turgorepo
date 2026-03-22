import "./zod-extend.js";
import {
  OpenAPIRegistry,
  OpenApiGeneratorV3,
} from "@asteasolutions/zod-to-openapi";
import { z } from "zod";
import {
  TodoSchema as TodoSchemaBase,
  CreateTodoBodySchema as CreateTodoBodySchemaBase,
  UpdateTodoBodySchema as UpdateTodoBodySchemaBase,
} from "@repo/types";

/** Portable return type for the OpenAPI document (avoids referencing openapi3-ts) */
type OpenAPIDocument = ReturnType<
  InstanceType<typeof OpenApiGeneratorV3>["generateDocument"]
>;

const registry = new OpenAPIRegistry();

// --- Todo schemas (from @repo/types, extended with OpenAPI metadata) ---
const TodoSchema = TodoSchemaBase.openapi("Todo", {
  example: {
    id: "550e8400-e29b-41d4-a716-446655440000",
    title: "Buy milk",
    completed: false,
    created_at: "2025-03-15T12:00:00Z",
    updated_at: "2025-03-15T12:00:00Z",
  },
});
const CreateTodoBodySchema = CreateTodoBodySchemaBase.openapi(
  "CreateTodoBody",
  {
    example: { title: "Buy milk", completed: false },
  },
);
const UpdateTodoBodySchema = UpdateTodoBodySchemaBase.openapi(
  "UpdateTodoBody",
  {
    example: { title: "Buy milk", completed: true },
  },
);

const TodoIdParamSchema = z.object({
  id: z.string().uuid().openapi({ description: "Todo ID" }),
});

// Register schemas for $ref
registry.register("Todo", TodoSchema);
registry.register("CreateTodoBody", CreateTodoBodySchema);
registry.register("UpdateTodoBody", UpdateTodoBodySchema);

// --- Paths ---
registry.registerPath({
  method: "get",
  path: "/api/todos",
  summary: "List todos",
  description: "Returns all todos with optional pagination",
  request: {
    query: z.object({
      limit: z.coerce
        .number()
        .min(1)
        .max(100)
        .optional()
        .openapi({ example: 20 }),
      offset: z.coerce.number().min(0).optional().openapi({ example: 0 }),
    }),
  },
  responses: {
    200: {
      description: "List of todos",
      content: {
        "application/json": {
          schema: z
            .object({
              items: z.array(TodoSchema),
              total: z.number(),
            })
            .openapi("TodoListResponse"),
        },
      },
    },
  },
});

registry.registerPath({
  method: "get",
  path: "/api/todos/{id}",
  summary: "Get todo by ID",
  request: { params: TodoIdParamSchema },
  responses: {
    200: {
      description: "Todo found",
      content: { "application/json": { schema: TodoSchema } },
    },
    404: { description: "Todo not found" },
  },
});

registry.registerPath({
  method: "post",
  path: "/api/todos",
  summary: "Create todo",
  request: {
    body: {
      content: {
        "application/json": { schema: CreateTodoBodySchema },
      },
    },
  },
  responses: {
    201: {
      description: "Todo created",
      content: { "application/json": { schema: TodoSchema } },
    },
    400: { description: "Invalid input" },
  },
});

registry.registerPath({
  method: "put",
  path: "/api/todos/{id}",
  summary: "Update todo",
  request: {
    params: TodoIdParamSchema,
    body: {
      content: {
        "application/json": { schema: UpdateTodoBodySchema },
      },
    },
  },
  responses: {
    200: {
      description: "Todo updated",
      content: { "application/json": { schema: TodoSchema } },
    },
    404: { description: "Todo not found" },
    400: { description: "Invalid input" },
  },
});

registry.registerPath({
  method: "delete",
  path: "/api/todos/{id}",
  summary: "Delete todo",
  request: { params: TodoIdParamSchema },
  responses: {
    204: { description: "Todo deleted" },
    404: { description: "Todo not found" },
  },
});

// --- Generator ---
export function generateOpenAPIDocument(): OpenAPIDocument {
  const generator = new OpenApiGeneratorV3(registry.definitions);
  return generator.generateDocument({
    openapi: "3.0.3",
    info: {
      title: "Turgorepo API",
      version: "1.0.0",
      description: "API for the Turgorepo monorepo (Go/Gin backend)",
    },
    servers: [{ url: "/", description: "Current host" }],
  });
}
