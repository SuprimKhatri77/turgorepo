import { NextRequest, NextResponse } from "next/server";

export function redirectTo(req: NextRequest, path: string) {
  return NextResponse.redirect(new URL(path, req.url));
}

export function redirectForRole(
  req: NextRequest,
  role: string,
): NextResponse | null {
  if (role === "admin" || role === "superadmin")
    return redirectTo(req, "/admin");
  return null;
}
