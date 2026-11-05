import axios from "axios";
import type { AxiosInstance, InternalAxiosRequestConfig } from "axios";
import { AUTH_ENDPOINTS, isProtectedPath } from "./constants";
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

      // Auth endpoints skip refresh retry — including /auth/refresh itself
      if (
        AUTH_ENDPOINTS.some((endpoint) =>
          originalRequest.url?.includes(endpoint),
        )
      ) {
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
        processQueue(refreshError); // reject all queued requests

        // A 401 means the session is over, and the API has already cleared the
        // cookies. Anything else (e.g. a 503 during an outage) isn't a logout.
        const sessionEnded =
          axios.isAxiosError(refreshError) &&
          refreshError.response?.status === 401;

        if (
          sessionEnded &&
          typeof window !== "undefined" &&
          isProtectedPath(window.location.pathname)
        ) {
          // Interceptors run outside React, so the Next router is not available here.
          // eslint-disable-next-line @next/next/no-location-assign-relative-destination
          window.location.href = "/auth/login";
        }

        return Promise.reject(refreshError);
      } finally {
        setIsRefreshing(false);
      }
    },
  );
}
