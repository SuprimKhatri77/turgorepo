import { z } from "zod";

/** Todo item schema (shared shape for API and OpenAPI) */
export const TodoSchema = z.object({
  id: z.string().uuid(),
  title: z.string().min(1),
  completed: z.boolean(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});

/** Body schema for creating a todo */
export const CreateTodoBodySchema = z.object({
  title: z.string().min(1),
  completed: z.boolean().optional(),
});

/** Body schema for updating a todo (all fields optional) */
export const UpdateTodoBodySchema = z.object({
  title: z.string().min(1).optional(),
  completed: z.boolean().optional(),
});

export type Todo = z.infer<typeof TodoSchema>;
export type CreateTodoBody = z.infer<typeof CreateTodoBodySchema>;
export type UpdateTodoBody = z.infer<typeof UpdateTodoBodySchema>;
