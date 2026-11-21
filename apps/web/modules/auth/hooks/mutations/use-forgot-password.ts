import { useMutation } from "@tanstack/react-query";
import { postApiV1AuthForgotPassword } from "@repo/api-client";
import type { ForgotPasswordBody } from "@repo/types";
import { apiClient } from "@/lib/api/client";
import { toastApiError } from "@/lib/api/errors";

export function useForgotPassword() {
  return useMutation({
    mutationFn: (body: ForgotPasswordBody) =>
      postApiV1AuthForgotPassword({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    onError: (error) => {
      toastApiError(error, "Failed to send the request");
    },
  });
}
