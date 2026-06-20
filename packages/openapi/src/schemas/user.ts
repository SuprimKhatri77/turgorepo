import { UserSchema as UserSchemaBase } from "@repo/types";

export const UserSchema = UserSchemaBase.openapi("User", {
  example: {
    id: "550e8400-e29b-41d4-a716-446655440000",
    name: "Jane Doe",
    email: "jane@example.com",
    role: "member",
    image_url: null,
  },
});
