import { useMutation } from "@tanstack/react-query";
import { postApiV1AuthResendVerification } from "@repo/api-client";
import type { ResendVerificationBody } from "@repo/types";
import { apiClient } from "@/lib/api/client";
import { toastApiError } from "@/lib/api/errors";

export function useResendVerification() {
  return useMutation({
    mutationFn: (body: ResendVerificationBody) =>
      postApiV1AuthResendVerification({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    onError: (error) => {
      toastApiError(error, "Failed to send the email");
    },
  });
}
