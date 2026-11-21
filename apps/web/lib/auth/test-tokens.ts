import type { UserRole } from "@repo/types";

/**
 * Builds an unsigned access token for tests. The web app only decodes tokens
 * (the API verifies them), so the signature doesn't matter here.
 */
export function testAccessToken(
  claims: {
    role?: UserRole;
    sid?: string;
    expiresIn?: number;
  } = {},
): string {
  const payload = {
    user_id: "01010101-0101-0101-0101-010101010101",
    name: "Ada Lovelace",
    email: "ada@example.com",
    role: claims.role ?? "member",
    sid: "sid" in claims ? claims.sid : "session-1",
    exp: Math.floor(Date.now() / 1000) + (claims.expiresIn ?? 900),
  };
  const encode = (value: object) =>
    btoa(JSON.stringify(value))
      .replace(/=+$/, "")
      .replace(/\+/g, "-")
      .replace(/\//g, "_");
  return `${encode({ alg: "HS256" })}.${encode(payload)}.signature`;
}
