"use client";

import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  apiClient,
  postApiV1AuthLogin,
  type LoginBody,
} from "@/lib/api/client";
import { homePathForRole } from "@/lib/middleware/config";
import { useAuthStore } from "@/store/auth";
import { getApiError } from "@/utils/get-api-error";

export function useLogin() {
  const router = useRouter();
  const setUser = useAuthStore((state) => state.setUser);
  const setPendingEmail = useAuthStore((state) => state.setPendingEmail);

  return useMutation({
    mutationFn: (body: LoginBody) =>
      postApiV1AuthLogin({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    onSuccess: (response) => {
      setUser(response.data.data);
      toast.success(response.data.message);
      router.replace(homePathForRole(response.data.data.role));
      router.refresh();
    },
    onError: (error, body) => {
      const apiError = getApiError(error);
      // The password was right, so send them to finish verifying instead.
      if (apiError?.code === "EMAIL_NOT_VERIFIED") {
        setPendingEmail(body.email);
        router.push("/auth/check-email");
        return;
      }
      toast.error(apiError?.message ?? "Failed to sign in");
    },
  });
}
