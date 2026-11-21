import { createClient, createConfig, type Client } from "@repo/api-client";
import api from "@/lib/axios";

/**
 * Typed API client backed by the shared axios instance, so its requests go
 * through the refresh interceptor. Pass it to every SDK call:
 * postApiV1AuthLogin({ client: apiClient, body }).
 */
export const apiClient: Client = createClient(
  createConfig({
    axios: api,
    // The generated paths already start with /api/v1, so this is the host
    // only. The axios instance's own baseURL includes /api/v1 for the raw
    // calls the interceptor makes.
    baseURL: process.env.NEXT_PUBLIC_API_URL ?? "",
    withCredentials: true,
  }),
);
