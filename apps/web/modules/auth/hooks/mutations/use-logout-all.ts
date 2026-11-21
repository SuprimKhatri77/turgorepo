import { useRouter } from "next/navigation";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { postApiV1AuthLogoutAll } from "@repo/api-client";
import { apiClient } from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";
import { toastApiError } from "@/lib/api/errors";

export function useLogoutAll() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const clearUser = useAuthStore((state) => state.clearUser);

  return useMutation({
    mutationFn: () =>
      postApiV1AuthLogoutAll({ client: apiClient, throwOnError: true }),
    onSuccess: (response) => {
      clearUser();
      queryClient.clear();
      toast.success(response.data.message);
      router.replace("/");
      router.refresh();
    },
    // The API changes nothing when this fails, so the user is still logged in everywhere.
    onError: (error) => {
      toastApiError(error, "Failed to log out of all devices");
    },
  });
}
