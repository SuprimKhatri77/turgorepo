export type { CreateClientConfig } from "./generated-server/client.gen";
export {
  createClient,
  createConfig,
  type Client,
  type Config,
} from "./generated-server/client";
// SDK ops accept an optional client; always pass a request-scoped one from
// createServerApiClient() so Cookie headers are forwarded. The generated
// module singleton is intentionally not re-exported.
export * from "./generated-server/sdk.gen";
export type * from "./generated-server/types.gen";
