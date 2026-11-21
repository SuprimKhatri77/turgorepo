import { afterEach, expect, test } from "bun:test";
import { contentSecurityPolicy, createNonce } from "./csp";

const env = { ...process.env };
afterEach(() => {
  process.env = { ...env };
});

function directive(policy: string, name: string): string[] {
  const found = policy.split("; ").find((d) => d.startsWith(`${name} `));
  return found ? found.split(" ").slice(1) : [];
}

test("each request gets a new nonce", () => {
  expect(createNonce()).not.toBe(createNonce());
});

test("scripts need the nonce", () => {
  const policy = contentSecurityPolicy("abc123");
  expect(directive(policy, "script-src")).toEqual([
    "'self'",
    "'nonce-abc123'",
    "'strict-dynamic'",
  ]);
  expect(directive(policy, "frame-ancestors")).toEqual(["'none'"]);
});

test("development also allows eval, for React's debugging", () => {
  process.env = { ...env, NODE_ENV: "development" };
  expect(directive(contentSecurityPolicy("n"), "script-src")).toContain(
    "'unsafe-eval'",
  );
});

test("connect-src allows the API's origin only", () => {
  process.env = { ...env, NEXT_PUBLIC_API_URL: "https://api.example.com/v1" };
  expect(directive(contentSecurityPolicy("n"), "connect-src")).toEqual([
    "'self'",
    "https://api.example.com",
  ]);
});

test("a malformed API URL is left out", () => {
  process.env = { ...env, NEXT_PUBLIC_API_URL: "not a url" };
  expect(directive(contentSecurityPolicy("n"), "connect-src")).toEqual([
    "'self'",
  ]);
});
