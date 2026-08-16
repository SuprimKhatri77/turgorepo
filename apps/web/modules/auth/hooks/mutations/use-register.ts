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
  const setUser = useAuthStore((state) => state.setUser);

  return useMutation({
    mutationFn: (body: RegisterBody) =>
      postApiV1AuthRegister({
        client: apiClient,
        body,
        throwOnError: true,
      }),
    onSuccess: (response) => {
      setUser(response.data.data);
      toast.success(response.data.message);
      router.replace("/");
      router.refresh();
    },
    onError: (error) => {
      toast.error(getApiError(error)?.message ?? "Failed to create account");
    },
  });
}
