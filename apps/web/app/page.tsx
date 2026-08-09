import { get_api_url } from "@/utils/get-api-url";

export default async function Home() {
  const api = get_api_url();
  let data: unknown = {
    success: false,
    message: "API unavailable",
  };

  if (api) {
    try {
      const response = await fetch(`${api}/api/v1/health`, {
        cache: "no-store",
      });
      if (response.ok) {
        data = await response.json();
      }
    } catch {
      // API may be offline during local/CI builds
    }
  }

  return (
    <div className="flex flex-col items-center justify-center h-screen bg-gray-100">
      <h1 className="text-2xl font-bold text-gray-900 mb-4">Health Check</h1>
      <pre className="text-sm text-gray-500 w-full max-w-2xl overflow-x-auto p-4 bg-white rounded-lg">
        {JSON.stringify(data, null, 2)}
      </pre>
    </div>
  );
}
