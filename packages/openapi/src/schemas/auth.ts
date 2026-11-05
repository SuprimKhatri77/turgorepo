import {
  AuthSuccessResponseSchema,
  AuthUserResponseSchema,
  ForgotPasswordBodySchema as ForgotPasswordBodySchemaBase,
  HealthResponseSchema,
  LoginBodySchema as LoginBodySchemaBase,
  MeResponseSchema,
  RegisterBodySchema as RegisterBodySchemaBase,
  ResetPasswordBodySchema as ResetPasswordBodySchemaBase,
} from "@repo/types";

export const RegisterBodySchema = RegisterBodySchemaBase.openapi(
  "RegisterBody",
  {
    example: {
      name: "Jane Doe",
      email: "jane@example.com",
      password: "securepass123",
    },
  },
);

export const LoginBodySchema = LoginBodySchemaBase.openapi("LoginBody", {
  example: {
    email: "jane@example.com",
    password: "securepass123",
  },
});

export const ForgotPasswordBodySchema = ForgotPasswordBodySchemaBase.openapi(
  "ForgotPasswordBody",
  {
    example: {
      email: "jane@example.com",
    },
  },
);

export const ResetPasswordBodySchema = ResetPasswordBodySchemaBase.openapi(
  "ResetPasswordBody",
  {
    example: {
      token: "token-from-the-reset-email",
      password: "new-securepass123",
    },
  },
);

export const HealthResponse = HealthResponseSchema.openapi("HealthResponse", {
  example: {
    success: true,
    message: "Server is up and running",
  },
});

export const AuthUserResponse = AuthUserResponseSchema.openapi(
  "AuthUserResponse",
  {
    example: {
      success: true,
      message: "logged in successfully",
      data: {
        id: "550e8400-e29b-41d4-a716-446655440000",
        name: "Jane Doe",
        email: "jane@example.com",
        role: "member",
      },
    },
  },
);

export const MeResponse = MeResponseSchema.openapi("MeResponse", {
  example: {
    success: true,
    message: "Valid session",
    data: {
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Jane Doe",
      email: "jane@example.com",
      role: "member",
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
