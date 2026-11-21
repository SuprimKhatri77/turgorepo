import {
  ApiUrlNotConfiguredError,
  createServerApiClient,
  getApiV1Health,
} from "@/lib/api/server-client";

export type HealthStatus = {
  online: boolean;
  message: string;
};

/** Asks the API whether it's up, for the landing page's status line. */
export async function checkHealth(): Promise<HealthStatus> {
  try {
    const client = await createServerApiClient();

    const { data, error, response } = await getApiV1Health({
      client,
      cache: "no-store",
      signal: AbortSignal.timeout(5_000),
    });

    if (!response) {
      return { online: false, message: "API unreachable" };
    }

    if (error || !response.ok) {
      return {
        online: false,
        message: `API responded ${response.status}`,
      };
    }

    return {
      online: true,
      message: data?.message ?? "Server is up and running",
    };
  } catch (err) {
    if (err instanceof ApiUrlNotConfiguredError) {
      return { online: false, message: "API URL is not configured" };
    }
    // API may be offline during local/CI builds
    return { online: false, message: "API unreachable" };
  }
}
