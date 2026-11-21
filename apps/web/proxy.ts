import { NextRequest, NextResponse } from "next/server";
import { contentSecurityPolicy, createNonce } from "@/lib/proxy/csp";
import { handleAuthRoute, handleProtectedRoute } from "@/lib/proxy/handlers";
import { continueRequest } from "@/lib/proxy/redirects";
import {
  getRequiredRoles,
  isUnauthenticatedOnlyRoute,
} from "@/lib/auth/routes";

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
  // Every page, so that ROLE_RULES in lib/auth/routes.ts is the only place
  // to protect one, and every page gets a CSP nonce. Skips API routes and static
  // files.
  matcher: [
    "/((?!api/|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
  ],
};
