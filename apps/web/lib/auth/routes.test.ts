import { describe, expect, test } from "bun:test";
import {
  getRequiredRoles,
  homePathForRole,
  isUnauthenticatedOnlyRoute,
  matchesRoute,
} from "./routes";

describe("matchesRoute", () => {
  test("matches the route and its children", () => {
    expect(matchesRoute("/admin", "/admin")).toBe(true);
    expect(matchesRoute("/admin/users", "/admin")).toBe(true);
  });

  test("doesn't match a longer name with the same prefix", () => {
    expect(matchesRoute("/administrator", "/admin")).toBe(false);
  });
});

describe("isUnauthenticatedOnlyRoute", () => {
  test("auth pages are for logged-out users", () => {
    expect(isUnauthenticatedOnlyRoute("/auth/login")).toBe(true);
    expect(isUnauthenticatedOnlyRoute("/auth/register")).toBe(true);
  });

  test("email links stay open to logged-in users", () => {
    expect(isUnauthenticatedOnlyRoute("/auth/reset-password")).toBe(false);
    expect(isUnauthenticatedOnlyRoute("/auth/verify-email")).toBe(false);
  });

  test("other pages aren't auth pages", () => {
    expect(isUnauthenticatedOnlyRoute("/dashboard")).toBe(false);
    expect(isUnauthenticatedOnlyRoute("/authors")).toBe(false);
  });
});

describe("getRequiredRoles", () => {
  test("returns the roles for a protected page and its children", () => {
    expect(getRequiredRoles("/admin")).toEqual(["admin", "superadmin"]);
    expect(getRequiredRoles("/dashboard/settings")).toEqual([
      "member",
      "admin",
      "superadmin",
    ]);
  });

  test("public pages need no role", () => {
    expect(getRequiredRoles("/")).toBeUndefined();
    expect(getRequiredRoles("/auth/login")).toBeUndefined();
  });
});

test("homePathForRole sends admins to /admin and members to /dashboard", () => {
  expect(homePathForRole("admin")).toBe("/admin");
  expect(homePathForRole("superadmin")).toBe("/admin");
  expect(homePathForRole("member")).toBe("/dashboard");
});
