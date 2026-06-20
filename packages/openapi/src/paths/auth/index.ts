import { registerLoginPath } from "./login.ts";
import { registerLogoutPath } from "./logout.ts";
import { registerMePath } from "./me.ts";
import { registerRefreshPath } from "./refresh.ts";
import { registerRegisterPath } from "./register.ts";

export function registerAuthPaths() {
  registerRegisterPath();
  registerLoginPath();
  registerLogoutPath();
  registerRefreshPath();
  registerMePath();
}
