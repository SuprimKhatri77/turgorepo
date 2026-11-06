import { NextRequest, NextResponse } from "next/server";
import { contentSecurityPolicy, createNonce } from "@/lib/middleware/csp";
import {
  handleAuthRoute,
  handleProtectedRoute,
} from "@/lib/middleware/handlers";
import { continueRequest } from "@/lib/middleware/redirects";
import {
  getRequiredRoles,
  isUnauthenticatedOnlyRoute,
} from "@/lib/middleware/routes";

export async function proxy(req: NextRequest) {
  const nonce = createNonce();
  const csp = contentSecurityPolicy(nonce);

  // Next reads the nonce from the request's policy and adds it to its scripts.
  const requestHeaders = new Headers(req.headers);
  requestHeaders.set("x-nonce", nonce);
  requestHeaders.set("Content-Security-Policy", csp);

  const response = await route(req, requestHeaders);
  response.headers.set("Content-Security-Policy", csp);
  return response;
}

function route(
  req: NextRequest,
  requestHeaders: Headers,
): NextResponse | Promise<NextResponse> {
  const { pathname } = req.nextUrl;

  if (isUnauthenticatedOnlyRoute(pathname)) {
    return handleAuthRoute(req, requestHeaders);
  }

  const requiredRoles = getRequiredRoles(pathname);
  if (!requiredRoles) {
    return continueRequest(requestHeaders);
  }

  return handleProtectedRoute(req, requestHeaders, requiredRoles);
}

export const config = {
  // Every page, so that ROLE_RULES in lib/middleware/config.ts is the only place
  // to protect one, and every page gets a CSP nonce. Skips API routes and static
  // files.
  matcher: [
    "/((?!api/|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
  ],
};
