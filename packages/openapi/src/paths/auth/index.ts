import { registerForgotPasswordPath } from "./forgot-password.ts";
import { registerLoginPath } from "./login.ts";
import { registerLogoutAllPath } from "./logout-all.ts";
import { registerLogoutPath } from "./logout.ts";
import { registerMePath } from "./me.ts";
import { registerRefreshPath } from "./refresh.ts";
import { registerRegisterPath } from "./register.ts";
import { registerResendVerificationPath } from "./resend-verification.ts";
import { registerResetPasswordPath } from "./reset-password.ts";
import { registerVerifyEmailPath } from "./verify-email.ts";

export function registerAuthPaths() {
  registerRegisterPath();
  registerLoginPath();
  registerLogoutPath();
  registerLogoutAllPath();
  registerRefreshPath();
  registerMePath();
  registerForgotPasswordPath();
  registerResetPasswordPath();
  registerVerifyEmailPath();
  registerResendVerificationPath();
}
