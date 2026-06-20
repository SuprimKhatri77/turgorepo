import type { ZodType } from "zod";
import { schemas } from "../schemas/index.ts";

const json = (schema: ZodType, description: string) => ({
  description,
  content: {
    "application/json": { schema },
  },
});

export const errorResponse = (description: string) =>
  json(schemas.ApiErrorResponse, description);
