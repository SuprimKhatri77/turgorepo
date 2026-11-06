"use client";

import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  apiClient,
  postApiV1AuthRegister,
  type RegisterBody,
} from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";
import { getApiError } from "@/utils/get-api-error";

export function useRegister() {
  const router = useRouter();
  const setPendingEmail = useAuthStore((state) => state.setPendingEmail);

  return useMutation({
    mutationFn: (body: RegisterBody) =>
      postApiV1AuthRegister({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    // Registering doesn't log the user in. They sign in after verifying.
    onSuccess: (_response, body) => {
      setPendingEmail(body.email);
      router.push("/auth/check-email");
    },
    onError: (error) => {
      toast.error(getApiError(error)?.message ?? "Failed to create account");
    },
  });
}
