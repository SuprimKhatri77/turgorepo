export function get_api_url() {
  return typeof window === "undefined" && process.env.INTERNAL_API_URL
    ? process.env.INTERNAL_API_URL
    : process.env.NEXT_PUBLIC_API_URL;
}
