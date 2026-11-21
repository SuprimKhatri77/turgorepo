import { expect, test } from "bun:test";
import { userFromAccessToken } from "./claims";
import { testAccessToken } from "./test-tokens";

test("reads the user from a valid token", () => {
  expect(userFromAccessToken(testAccessToken({ role: "admin" }))).toEqual({
    id: "01010101-0101-0101-0101-010101010101",
    name: "Ada Lovelace",
    email: "ada@example.com",
    role: "admin",
    image_url: undefined,
  });
});

test("an expired token has no user", () => {
  expect(userFromAccessToken(testAccessToken({ expiresIn: -60 }))).toBeNull();
});

test("a token without a session ID has no user, so the proxy refreshes it", () => {
  expect(userFromAccessToken(testAccessToken({ sid: undefined }))).toBeNull();
});

test("garbage has no user", () => {
  expect(userFromAccessToken("not-a-jwt")).toBeNull();
});
