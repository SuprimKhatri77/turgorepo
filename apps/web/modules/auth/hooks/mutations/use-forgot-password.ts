"use client";

import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  apiClient,
  postApiV1AuthForgotPassword,
  type ForgotPasswordBody,
} from "@/lib/api/client";
import { getApiError } from "@/lib/api/errors";

export function useForgotPassword() {
  return useMutation({
    mutationFn: (body: ForgotPasswordBody) =>
      postApiV1AuthForgotPassword({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    onError: (error) => {
      toast.error(getApiError(error)?.message ?? "Failed to send the request");
    },
  });
}
