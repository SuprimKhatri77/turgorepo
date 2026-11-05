"use client";

import dynamic from "next/dynamic";

// The form reads the token from the URL fragment, which only the browser can see.
export const ResetPasswordLoader = dynamic(
  () =>
    import("./reset-password-form").then((module) => module.ResetPasswordForm),
  { ssr: false },
);
