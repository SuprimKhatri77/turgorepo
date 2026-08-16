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
  matcher: ["/admin/:path*", "/auth/:path*"],
};
