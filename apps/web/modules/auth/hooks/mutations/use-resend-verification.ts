"use client";

import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  apiClient,
  postApiV1AuthResendVerification,
  type ResendVerificationBody,
} from "@/lib/api/client";
import { getApiError } from "@/utils/get-api-error";

export function useResendVerification() {
  return useMutation({
    mutationFn: (body: ResendVerificationBody) =>
      postApiV1AuthResendVerification({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    onError: (error) => {
      toast.error(getApiError(error)?.message ?? "Failed to send the email");
    },
  });
}
