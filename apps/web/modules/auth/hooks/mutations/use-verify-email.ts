"use client";

import { useMutation } from "@tanstack/react-query";
import {
  apiClient,
  postApiV1AuthVerifyEmail,
  type VerifyEmailBody,
} from "@/lib/api/client";

export function useVerifyEmail() {
  return useMutation({
    mutationFn: (body: VerifyEmailBody) =>
      postApiV1AuthVerifyEmail({
        client: apiClient,
        body,
        throwOnError: true,
      }),
  });
}
