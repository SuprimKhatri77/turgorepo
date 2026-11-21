import type { User } from "@repo/types";
import type { NextRequest, NextResponse } from "next/server";
import { getApiUrl } from "@/lib/api/url";
import { userFromAccessToken } from "@/lib/auth/claims";
import {
  ACCESS_TOKEN_COOKIE,
  LOGGED_IN_COOKIE,
  REFRESH_TOKEN_COOKIE,
} from "@/lib/auth/cookies";
import { LOGIN_PATH } from "@/lib/auth/routes";
import { continueRequest, redirectForRole, redirectTo } from "./redirects";

export function clearAuthCookies(response: NextResponse) {
  const cookieOptions = {
    maxAge: 0,
    path: "/",
    domain:
      process.env.NODE_ENV === "production" ? process.env.COOKIE_DOMAIN : "",
    secure: process.env.NODE_ENV === "production",
  };

  response.cookies.set(ACCESS_TOKEN_COOKIE, "", {
    ...cookieOptions,
    httpOnly: true,
  });
  response.cookies.set(REFRESH_TOKEN_COOKIE, "", {
    ...cookieOptions,
    httpOnly: true,
  });
  response.cookies.set(LOGGED_IN_COOKIE, "", {
    ...cookieOptions,
    httpOnly: false,
  });
}

/** What the API's /auth/refresh did with the refresh token. */
type RefreshResult =
  | {
      status: "refreshed";
      user: User;
      accessToken: string;
      /** Missing for a grace-window refresh, which only issues an access token. */
      refreshToken: string | null;
      /** The API's Set-Cookie headers, to pass on to the browser. */
      setCookies: string[];
    }
  /** 401: the session is over. */
  | { status: "rejected" }
  /** Succeeded, but returned no usable access token. */
  | { status: "unusable" }
  /** An API error or no answer: the session may still be valid. */
  | { status: "failed" };

async function refreshSession(refreshToken: string): Promise<RefreshResult> {
  let res: Response;
  try {
    res = await fetch(`${getApiUrl()}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { Cookie: `${REFRESH_TOKEN_COOKIE}=${refreshToken}` },
    });
  } catch (error) {
    console.error("Refresh failed in proxy:", error);
    return { status: "failed" };
  }

  if (res.status === 401) return { status: "rejected" };
  if (!res.ok) return { status: "failed" };

  const setCookies = res.headers.getSetCookie();
  const accessToken = cookieValue(setCookies, ACCESS_TOKEN_COOKIE);
  const user = accessToken ? userFromAccessToken(accessToken) : null;
  if (!accessToken || !user) return { status: "unusable" };

  return {
    status: "refreshed",
    user,
    accessToken,
    refreshToken: cookieValue(setCookies, REFRESH_TOKEN_COOKIE),
    setCookies,
  };
}

/**
 * Refreshes the session for a protected page. On success the page renders with
 * the new tokens; otherwise the user goes to log in, and only a rejected
 * session clears the cookies.
 */
export async function refreshForProtectedRoute(
  refreshToken: string,
  req: NextRequest,
  requestHeaders: Headers,
  requiredRoles: string[],
): Promise<NextResponse> {
  const result = await refreshSession(refreshToken);

  if (result.status !== "refreshed") {
    const redirect = redirectTo(req, LOGIN_PATH);
    if (result.status === "rejected") {
      clearAuthCookies(redirect);
    }
    return redirect;
  }

  if (!requiredRoles.includes(result.user.role)) {
    const redirect = redirectTo(req, "/");
    forwardSetCookies(redirect, result.setCookies);
    return redirect;
  }

  // A grace-window refresh only sets the access token, so keep the refresh
  // token the browser already has.
  const headers = withSessionCookies(
    req,
    requestHeaders,
    result.accessToken,
    result.refreshToken ?? refreshToken,
  );
  const response = continueRequest(headers);
  forwardSetCookies(response, result.setCookies);
  return response;
}

/**
 * Refreshes the session for a logged-out-only page. On success the user goes to
 * their home page; otherwise the page shows, and only an outage keeps the
 * cookies, since the session may still be valid.
 */
export async function refreshForAuthRoute(
  refreshToken: string,
  req: NextRequest,
  requestHeaders: Headers,
): Promise<NextResponse> {
  const result = await refreshSession(refreshToken);

  if (result.status === "refreshed") {
    const redirect = redirectForRole(req, result.user.role);
    forwardSetCookies(redirect, result.setCookies);
    return redirect;
  }

  const response = continueRequest(requestHeaders);
  if (result.status !== "failed") {
    clearAuthCookies(response);
  }
  return response;
}

/** The request's cookie header with the session tokens replaced, so the page sees the new ones. */
function withSessionCookies(
  req: NextRequest,
  requestHeaders: Headers,
  accessToken: string,
  refreshToken: string,
): Headers {
  const otherCookies = req.cookies
    .getAll()
    .filter(
      ({ name }) =>
        name !== ACCESS_TOKEN_COOKIE && name !== REFRESH_TOKEN_COOKIE,
    )
    .map(({ name, value }) => `${name}=${value}`);

  const headers = new Headers(requestHeaders);
  headers.set(
    "cookie",
    [
      ...otherCookies,
      `${ACCESS_TOKEN_COOKIE}=${accessToken}`,
      `${REFRESH_TOKEN_COOKIE}=${refreshToken}`,
    ].join("; "),
  );
  return headers;
}

function forwardSetCookies(response: NextResponse, setCookies: string[]) {
  for (const cookie of setCookies) {
    response.headers.append("set-cookie", cookie);
  }
}

/** The value of the named cookie in a list of Set-Cookie headers, if it has one. */
function cookieValue(setCookies: string[], name: string): string | null {
  for (const cookie of setCookies) {
    const [pair = ""] = cookie.split(";");
    const separator = pair.indexOf("=");
    if (pair.slice(0, separator).trim() === name) {
      return pair.slice(separator + 1) || null;
    }
  }
  return null;
}
