import { NextRequest, NextResponse } from "next/server";
import {
  handleAuthRoute,
  handleProtectedRoute,
} from "@/lib/middleware/handlers";
import {
  getRequiredRoles,
  isUnauthenticatedOnlyRoute,
} from "@/lib/middleware/routes";

export async function proxy(req: NextRequest) {
  const { pathname } = req.nextUrl;

  if (isUnauthenticatedOnlyRoute(pathname)) {
    return handleAuthRoute(req);
  }

  const requiredRoles = getRequiredRoles(pathname);
  if (!requiredRoles) {
    return NextResponse.next();
  }

  return handleProtectedRoute(req, requiredRoles);
}

export const config = {
  // Every page, so that ROLE_RULES in lib/middleware/config.ts is the only place
  // to protect one. Skips API routes and static files.
  matcher: [
    "/((?!api/|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
  ],
};
