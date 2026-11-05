import { NextRequest, NextResponse } from "next/server";
import { homePathForRole } from "./config";

export function redirectTo(req: NextRequest, path: string) {
  return NextResponse.redirect(new URL(path, req.url));
}

/** Where a logged-in user lands when they visit a logged-out-only page. */
export function redirectForRole(req: NextRequest, role: string): NextResponse {
  return redirectTo(req, homePathForRole(role));
}
