import {
  AuthSuccessResponseSchema,
  AuthUserResponseSchema,
  HealthResponseSchema,
  LoginBodySchema as LoginBodySchemaBase,
  MeResponseSchema,
  RegisterBodySchema as RegisterBodySchemaBase,
} from "@repo/types";

export const RegisterBodySchema = RegisterBodySchemaBase.openapi("RegisterBody", {
  example: {
    name: "Jane Doe",
    email: "jane@example.com",
    password: "securepass123",
  },
});

export const LoginBodySchema = LoginBodySchemaBase.openapi("LoginBody", {
  example: {
    email: "jane@example.com",
    password: "securepass123",
  },
});

export const HealthResponse = HealthResponseSchema.openapi("HealthResponse", {
  example: {
    success: true,
    message: "Server is up and running",
  },
});

export const AuthUserResponse = AuthUserResponseSchema.openapi("AuthUserResponse", {
  example: {
    success: true,
    message: "logged in successfully",
    data: {
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Jane Doe",
      email: "jane@example.com",
      role: "member",
      image_url: null,
    },
  },
});

export const MeResponse = MeResponseSchema.openapi("MeResponse", {
  example: {
    success: true,
    message: "Valid session",
    data: {
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Jane Doe",
      email: "jane@example.com",
      role: "member",
      image_url: null,
    },
  },
});

export const AuthSuccessResponse = AuthSuccessResponseSchema.openapi(
  "AuthSuccessResponse",
  {
    example: {
      success: true,
      message: "Logged out successfully",
    },
  },
);
