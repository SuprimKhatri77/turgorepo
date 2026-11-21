import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";

import { postApiV1AuthRegister } from "@repo/api-client";
import type { RegisterBody } from "@repo/types";
import { apiClient } from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";
import { toastApiError } from "@/lib/api/errors";

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
      toastApiError(error, "Failed to create account");
    },
  });
}
