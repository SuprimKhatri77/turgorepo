import { useMutation } from "@tanstack/react-query";
import { postApiV1AuthVerifyEmail } from "@repo/api-client";
import type { VerifyEmailBody } from "@repo/types";
import { apiClient } from "@/lib/api/client";

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
