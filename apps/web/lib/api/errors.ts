import { AxiosError } from "axios";
import { ApiResponse } from "@repo/types";

export const getApiError = (err: unknown) => {
  const error = err as AxiosError<ApiResponse>;
  return error.response?.data;
};
