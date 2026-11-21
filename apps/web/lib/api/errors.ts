import type { AxiosError } from "axios";
import type { ApiResponse } from "@repo/types";
import { toast } from "sonner";

/** The API's error body, if the request reached the API. */
export const getApiError = (err: unknown) => {
  const error = err as AxiosError<ApiResponse>;
  return error.response?.data;
};

/** Shows the API's error message, or fallback when there isn't one. */
export function toastApiError(err: unknown, fallback: string) {
  toast.error(getApiError(err)?.message ?? fallback);
}
