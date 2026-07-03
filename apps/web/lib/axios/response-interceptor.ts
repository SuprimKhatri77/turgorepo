import type { AxiosInstance, InternalAxiosRequestConfig } from "axios";
import { AUTH_ENDPOINTS, COOKIE_DOMAIN, isProtectedPath } from "./constants";
import {
  getIsRefreshing,
  processQueue,
  pushToQueue,
  setIsRefreshing,
} from "./refresh-queue";

export function attachResponseInterceptor(api: AxiosInstance) {
  api.interceptors.response.use(
    (response) => response, // pass through 2xx as-is

    async (error) => {
      const originalRequest: InternalAxiosRequestConfig & { _retry?: boolean } =
        error.config;

      // we did this to send our actual error message from backend instead of the axios error message
      if (error.response?.data?.message) {
        error.message = error.response.data.message;
      }
      if (
        AUTH_ENDPOINTS.some((endpoint) =>
          originalRequest.url?.includes(endpoint),
        )
      ) {
        return Promise.reject(error);
      }

      // if the 401 came FROM /auth/refresh itself → don't retry, just reject
      if (originalRequest.url?.includes("/auth/refresh")) {
        return Promise.reject(error);
      }

      // not a 401, or already retried — just reject
      if (error.response?.status !== 401 || originalRequest._retry) {
        return Promise.reject(error);
      }

      // if a refresh is already in flight, queue this request
      if (getIsRefreshing()) {
        return new Promise((resolve, reject) => {
          pushToQueue({ resolve, reject });
        })
          .then(() => api(originalRequest))
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      setIsRefreshing(true);

      try {
        // refresh token is in the httpOnly cookie, sent automatically
        await api.post("/auth/refresh");

        processQueue(null); // unblock queued requests
        return api(originalRequest); // retry the original call
      } catch (refreshError) {
        // if (axios.isAxiosError(refreshError)) {
        //   console.log(
        //     "refresh failed:",
        //     refreshError.response?.status,
        //     refreshError.response?.data,
        //   );
        // }

        document.cookie = `is_logged_in=; max-age=0; path=/; domain=${COOKIE_DOMAIN}`;

        processQueue(refreshError); // reject all queued requests

        // on protected routes, send to login; on public pages stay put
        if (
          typeof window !== "undefined" &&
          isProtectedPath(window.location.pathname)
        ) {
          window.location.href = "/auth/login";
        }

        return Promise.reject(refreshError);
      } finally {
        setIsRefreshing(false);
      }
    },
  );
}
