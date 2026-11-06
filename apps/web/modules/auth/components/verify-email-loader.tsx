"use client";

import dynamic from "next/dynamic";

// The page reads the token from the URL fragment, which only the browser can see.
export const VerifyEmailLoader = dynamic(
  () =>
    import("./verify-email-status").then((module) => module.VerifyEmailStatus),
  { ssr: false },
);
