import type { UserRole } from "@repo/types";
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { homePathForRole } from "@/lib/auth/routes";

/**
 * Lets the request through with the given headers, which carry the CSP nonce
 * that Next adds to its scripts.
 */
export function continueRequest(requestHeaders: Headers): NextResponse {
  return NextResponse.next({ request: { headers: requestHeaders } });
}

export function redirectTo(req: NextRequest, path: string) {
  return NextResponse.redirect(new URL(path, req.url));
}

/** Where a logged-in user lands when they visit a logged-out-only page. */
export function redirectForRole(
  req: NextRequest,
  role: UserRole,
): NextResponse {
  return redirectTo(req, homePathForRole(role));
}
