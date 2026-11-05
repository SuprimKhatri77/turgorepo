import type { User } from "@repo/types";
import { jwtDecode } from "jwt-decode";

type AccessTokenClaims = Omit<User, "id"> & {
  user_id: string;
  exp: number;
};

/**
 * Reads the user from an access token. This only decodes the token, so use it
 * for UI and routing decisions; the API verifies the signature.
 */
export function userFromAccessToken(token: string): User | null {
  try {
    const { user_id, name, email, role, image_url, exp } =
      jwtDecode<AccessTokenClaims>(token);
    if (exp * 1000 < Date.now()) return null;
    return { id: user_id, name, email, role, image_url };
  } catch {
    return null;
  }
}
