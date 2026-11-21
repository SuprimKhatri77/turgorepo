import { useRouter } from "next/navigation";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, postApiV1AuthLogout } from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";
import { toastApiError } from "@/lib/api/errors";

export function useLogout() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const clearUser = useAuthStore((state) => state.clearUser);

  return useMutation({
    mutationFn: () =>
      postApiV1AuthLogout({ client: apiClient, throwOnError: true }),
    onSuccess: () => {
      clearUser();
      queryClient.clear();
      router.replace("/");
      router.refresh();
    },
    // Only the API can clear httpOnly cookies, so on failure the user is still logged in.
    onError: (error) => {
      toastApiError(error, "Failed to log out");
    },
  });
}
