import {
  afterEach,
  beforeEach,
  describe,
  expect,
  mock,
  spyOn,
  test,
} from "bun:test";
import { NextRequest, type NextResponse } from "next/server";
import { testAccessToken } from "@/lib/auth/test-tokens";
import { handleAuthRoute, handleProtectedRoute } from "./handlers";

const API = "http://api.test";
const originalFetch = globalThis.fetch;
let refreshCalls: Request[] = [];

beforeEach(() => {
  process.env.NEXT_PUBLIC_API_URL = API;
  refreshCalls = [];
  spyOn(console, "error").mockImplementation(() => {});
});
afterEach(() => {
  globalThis.fetch = originalFetch;
  mock.restore();
});

/** Makes the API's /auth/refresh answer with status and these Set-Cookie headers. */
function refreshAnswers(status: number, setCookies: string[] = []) {
  globalThis.fetch = mock(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      refreshCalls.push(new Request(input, init));
      const headers = new Headers();
      for (const cookie of setCookies) headers.append("set-cookie", cookie);
      return new Response("{}", { status, headers });
    },
  ) as unknown as typeof fetch;
}

function refreshFails() {
  globalThis.fetch = mock(async () => {
    throw new TypeError("fetch failed");
  }) as unknown as typeof fetch;
}

function request(path: string, cookies: Record<string, string>) {
  const cookie = Object.entries(cookies)
    .map(([name, value]) => `${name}=${value}`)
    .join("; ");
  return new NextRequest(`http://localhost:3000${path}`, {
    headers: { cookie },
  });
}

const newAccess = testAccessToken({ role: "member" });
const rotated = [
  `access_token=${newAccess}; Path=/; HttpOnly`,
  "refresh_token=new-refresh; Path=/; HttpOnly",
  "is_logged_in=true; Path=/",
];

const location = (res: NextResponse) => res.headers.get("location");
const continues = (res: NextResponse) =>
  res.headers.get("x-middleware-next") === "1";
const forwardedCookies = (res: NextResponse) =>
  res.headers.get("x-middleware-request-cookie") ?? "";
const setCookieNames = (res: NextResponse) =>
  res.headers.getSetCookie().map((c) => c.split("=")[0]);
const clearsSession = (res: NextResponse) =>
  ["access_token", "refresh_token", "is_logged_in"].every((name) =>
    res.headers
      .getSetCookie()
      .some((c) => c.startsWith(`${name}=;`) && /Max-Age=0/i.test(c)),
  );

describe("handleProtectedRoute", () => {
  const roles = ["member", "admin", "superadmin"] as const;

  test("no refresh token goes to login without calling the API", async () => {
    refreshAnswers(200, rotated);
    const res = await handleProtectedRoute(
      request("/dashboard", {}),
      new Headers(),
      [...roles],
    );
    expect(location(res)).toBe("http://localhost:3000/auth/login");
    expect(refreshCalls).toHaveLength(0);
  });

  test("a valid access token passes without refreshing", async () => {
    refreshAnswers(200, rotated);
    const res = await handleProtectedRoute(
      request("/dashboard", {
        access_token: testAccessToken(),
        refresh_token: "r",
      }),
      new Headers(),
      [...roles],
    );
    expect(continues(res)).toBe(true);
    expect(refreshCalls).toHaveLength(0);
  });

  test("a valid token with the wrong role goes home", async () => {
    const res = await handleProtectedRoute(
      request("/admin", {
        access_token: testAccessToken({ role: "member" }),
        refresh_token: "r",
      }),
      new Headers(),
      ["admin"],
    );
    expect(location(res)).toBe("http://localhost:3000/");
  });

  test("an expired token is refreshed and the page sees the new tokens", async () => {
    refreshAnswers(200, rotated);
    const res = await handleProtectedRoute(
      request("/dashboard", {
        access_token: testAccessToken({ expiresIn: -60 }),
        refresh_token: "old-refresh",
        theme: "dark",
      }),
      new Headers(),
      [...roles],
    );

    expect(refreshCalls).toHaveLength(1);
    expect(refreshCalls[0]?.url).toBe(`${API}/api/v1/auth/refresh`);
    expect(refreshCalls[0]?.headers.get("cookie")).toBe(
      "refresh_token=old-refresh",
    );
    expect(continues(res)).toBe(true);
    expect(forwardedCookies(res)).toBe(
      `theme=dark; access_token=${newAccess}; refresh_token=new-refresh`,
    );
    expect(setCookieNames(res)).toEqual([
      "access_token",
      "refresh_token",
      "is_logged_in",
    ]);
  });

  test("a grace-window refresh keeps the refresh token the browser has", async () => {
    refreshAnswers(200, [rotated[0]!]);
    const res = await handleProtectedRoute(
      request("/dashboard", { refresh_token: "old-refresh" }),
      new Headers(),
      [...roles],
    );
    expect(forwardedCookies(res)).toBe(
      `access_token=${newAccess}; refresh_token=old-refresh`,
    );
  });

  test("a refreshed user without the role goes home, keeping the new cookies", async () => {
    refreshAnswers(200, rotated);
    const res = await handleProtectedRoute(
      request("/admin", { refresh_token: "r" }),
      new Headers(),
      ["admin"],
    );
    expect(location(res)).toBe("http://localhost:3000/");
    expect(setCookieNames(res)).toEqual([
      "access_token",
      "refresh_token",
      "is_logged_in",
    ]);
  });

  test("a rejected refresh goes to login and clears the session", async () => {
    refreshAnswers(401);
    const res = await handleProtectedRoute(
      request("/dashboard", { refresh_token: "r" }),
      new Headers(),
      [...roles],
    );
    expect(location(res)).toBe("http://localhost:3000/auth/login");
    expect(clearsSession(res)).toBe(true);
  });

  test("an API error goes to login but keeps the session", async () => {
    refreshAnswers(503);
    const res = await handleProtectedRoute(
      request("/dashboard", { refresh_token: "r" }),
      new Headers(),
      [...roles],
    );
    expect(location(res)).toBe("http://localhost:3000/auth/login");
    expect(res.headers.getSetCookie()).toEqual([]);
  });

  test("an unreachable API goes to login but keeps the session", async () => {
    refreshFails();
    const res = await handleProtectedRoute(
      request("/dashboard", { refresh_token: "r" }),
      new Headers(),
      [...roles],
    );
    expect(location(res)).toBe("http://localhost:3000/auth/login");
    expect(res.headers.getSetCookie()).toEqual([]);
  });

  test("a refresh without a usable access token goes to login", async () => {
    for (const cookies of [
      [rotated[1]!],
      [`access_token=${testAccessToken({ expiresIn: -60 })}; Path=/`],
    ]) {
      refreshAnswers(200, cookies);
      const res = await handleProtectedRoute(
        request("/dashboard", { refresh_token: "r" }),
        new Headers(),
        [...roles],
      );
      expect(location(res)).toBe("http://localhost:3000/auth/login");
    }
  });
});

describe("handleAuthRoute", () => {
  test("a logged-out visitor sees the page", async () => {
    const res = await handleAuthRoute(
      request("/auth/login", {}),
      new Headers(),
    );
    expect(continues(res)).toBe(true);
  });

  test("?expired=1 shows the page and clears the session, so it can't loop", async () => {
    const res = await handleAuthRoute(
      request("/auth/login?expired=1", {
        access_token: testAccessToken(),
        refresh_token: "r",
      }),
      new Headers(),
    );
    expect(continues(res)).toBe(true);
    expect(clearsSession(res)).toBe(true);
  });

  test("a logged-in user goes to their home page", async () => {
    const res = await handleAuthRoute(
      request("/auth/login", {
        access_token: testAccessToken({ role: "admin" }),
        refresh_token: "r",
      }),
      new Headers(),
    );
    expect(location(res)).toBe("http://localhost:3000/admin");
  });

  test("an expired token is refreshed, then the user goes home", async () => {
    refreshAnswers(200, rotated);
    const res = await handleAuthRoute(
      request("/auth/login", { refresh_token: "r" }),
      new Headers(),
    );
    expect(location(res)).toBe("http://localhost:3000/dashboard");
    expect(setCookieNames(res)).toEqual([
      "access_token",
      "refresh_token",
      "is_logged_in",
    ]);
  });

  test("a rejected refresh shows the page and clears the session", async () => {
    refreshAnswers(401);
    const res = await handleAuthRoute(
      request("/auth/login", { refresh_token: "r" }),
      new Headers(),
    );
    expect(continues(res)).toBe(true);
    expect(clearsSession(res)).toBe(true);
  });

  test("an API error shows the page and keeps the session", async () => {
    refreshAnswers(500);
    const res = await handleAuthRoute(
      request("/auth/login", { refresh_token: "r" }),
      new Headers(),
    );
    expect(continues(res)).toBe(true);
    expect(res.headers.getSetCookie()).toEqual([]);
  });

  test("an unreachable API shows the page and keeps the session", async () => {
    refreshFails();
    const res = await handleAuthRoute(
      request("/auth/login", { refresh_token: "r" }),
      new Headers(),
    );
    expect(continues(res)).toBe(true);
    expect(res.headers.getSetCookie()).toEqual([]);
  });

  test("a refresh without a usable access token shows the page and clears the session", async () => {
    refreshAnswers(200, [rotated[1]!]);
    const res = await handleAuthRoute(
      request("/auth/login", { refresh_token: "r" }),
      new Headers(),
    );
    expect(continues(res)).toBe(true);
    expect(clearsSession(res)).toBe(true);
  });
});
