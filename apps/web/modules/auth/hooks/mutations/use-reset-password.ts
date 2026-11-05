"use client";

import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  apiClient,
  postApiV1AuthResetPassword,
  type ResetPasswordBody,
} from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";

export function useResetPassword() {
  const router = useRouter();
  const clearUser = useAuthStore((state) => state.clearUser);

  return useMutation({
    mutationFn: (body: ResetPasswordBody) =>
      postApiV1AuthResetPassword({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    onSuccess: (response) => {
      // The reset logs the user out everywhere, including this browser.
      clearUser();
      toast.success(response.data.message);
      router.replace("/auth/login");
      router.refresh();
    },
  });
}
