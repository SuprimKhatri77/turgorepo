import { NextRequest, NextResponse } from "next/server";
import { User } from "@repo/types";
import { jwtDecode } from "jwt-decode";
import { get_api_url } from "@/utils/get-api-url";

function clearAuthCookies(response: NextResponse) {
  const cookieOptions = {
    maxAge: 0,
    path: "/",
    domain:
      process.env.NODE_ENV === "production" ? process.env.COOKIE_DOMAIN : "",
    secure: process.env.NODE_ENV === "production",
  };

  response.cookies.set("access_token", "", {
    ...cookieOptions,
    httpOnly: true,
  });
  response.cookies.set("refresh_token", "", {
    ...cookieOptions,
    httpOnly: true,
  });
  response.cookies.set("is_logged_in", "", {
    ...cookieOptions,
    httpOnly: false,
  });
}

export function getSession(token: string): User | null {
  try {
    const decoded = jwtDecode<User & { exp: number }>(token);
    if (decoded.exp * 1000 < Date.now()) return null;
    return decoded;
  } catch {
    return null;
  }
}

export async function attemptRefresh(
  refreshToken: string,
  req: NextRequest,
  requiredRoles: string[],
) {
  try {
    console.log("attemptRefresh called, path:", req.nextUrl.pathname);
    const res = await fetch(`${get_api_url()}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { Cookie: `refresh_token=${refreshToken}` },
    });

    console.log("refresh response status:", res.status);

    if (!res.ok) {
      console.log(
        "refresh failed with status:",
        res.status,
        "— clearing cookies:",
        res.status === 401 ? "YES" : "NO",
      );
      if (res.status === 401) {
        console.log("in !res.ok block clearing cookies.....");
        const redirect = NextResponse.redirect(new URL("/auth/login", req.url));
        clearAuthCookies(redirect);
        return redirect;
      }

      return NextResponse.redirect(new URL("/auth/login", req.url));
    }

    // get ALL cookies from refresh response
    const cookies = res.headers.getSetCookie();

    // still need the access token value to check role
    const newAccessToken = extractTokenFromCookie(cookies.join("; "));

    if (!newAccessToken) {
      return NextResponse.redirect(new URL("/auth/login", req.url));
    }

    const user = await getSession(newAccessToken);

    if (!user) {
      return NextResponse.redirect(new URL("/auth/login", req.url));
    }

    if (!requiredRoles.includes(user.role)) {
      return NextResponse.redirect(new URL("/", req.url));
    }

    const newRefreshToken = extractRefreshTokenFromCookie(cookies.join("; "));

    const response = NextResponse.next({
      request: {
        headers: new Headers({
          ...Object.fromEntries(req.headers),
          cookie: `access_token=${newAccessToken}; refresh_token=${newRefreshToken}`,
        }),
      },
    });

    cookies.forEach((cookie) => {
      response.headers.append("set-cookie", cookie);
    });

    return response;
  } catch (error) {
    console.log("attemptRefresh threw an error:", error);
    const redirect = NextResponse.redirect(new URL("/auth/login", req.url));
    clearAuthCookies(redirect);
    return redirect;
  }
}

export async function attemptRefreshForAuthRoute(
  refreshToken: string,
  req: NextRequest,
) {
  try {
    const res = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/refresh`,
      {
        method: "POST",
        headers: { Cookie: `refresh_token=${refreshToken}` },
      },
    );

    if (!res.ok) {
      const response = NextResponse.next();
      if (res.status === 401) {
        clearAuthCookies(response);
      }
      return response;
    }

    const cookies = res.headers.getSetCookie();
    const newAccessToken = extractTokenFromCookie(cookies.join("; "));

    if (!newAccessToken) {
      const response = NextResponse.next();
      clearAuthCookies(response);
      return response;
    }

    const user = await getSession(newAccessToken);

    if (!user) {
      const response = NextResponse.next();
      clearAuthCookies(response);
      return response;
    }

    if (user.role === "admin" || user.role === "superadmin") {
      return NextResponse.redirect(new URL("/admin", req.url));
    }

    const response = NextResponse.next();
    cookies.forEach((cookie) => {
      response.headers.append("set-cookie", cookie);
    });
    return response;
  } catch (error) {
    console.log("attemptRefreshForAuthRoute threw an error:", error);
    const response = NextResponse.next();
    clearAuthCookies(response);
    return response;
  }
}

export async function getSessionFromRequest(
  req: NextRequest,
): Promise<User | null> {
  const accessToken = req.cookies.get("access_token")?.value;
  const refreshToken = req.cookies.get("refresh_token")?.value;

  if (!refreshToken) return null;

  if (accessToken) {
    const user = await getSession(accessToken);
    if (user) return user;
  }

  const res = await fetch(`${get_api_url()}/api/v1/auth/refresh`, {
    method: "POST",
    headers: { Cookie: `refresh_token=${refreshToken}` },
  });

  if (!res.ok) return null;

  const cookies = res.headers.getSetCookie();
  const newAccessToken = extractTokenFromCookie(cookies.join("; "));
  if (!newAccessToken) return null;

  return getSession(newAccessToken);
}

// parses access_token value out of the set-cookie header string
export function extractTokenFromCookie(
  setCookie: string | null,
): string | null {
  if (!setCookie) return null;
  const match = setCookie.match(/access_token=([^;]+)/);
  return match?.[1] ?? null;
}

export function extractRefreshTokenFromCookie(
  setCookie: string | null,
): string | null {
  if (!setCookie) return null;
  const match = setCookie.match(/refresh_token=([^;]+)/);
  return match?.[1] ?? null;
}
