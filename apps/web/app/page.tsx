export default async function Home() {
  const api = process.env.INTERNAL_API_URL;

  console.log({ api });

  const response = await fetch(`${api}/api/v1/health`);

  const data = await response.json();
  return (
    <div className="flex flex-col items-center justify-center h-screen bg-gray-100">
      <h1 className="text-2xl font-bold text-gray-900 mb-4">Health Check</h1>
      <pre className="text-sm text-gray-500 w-full max-w-2xl overflow-x-auto p-4 bg-white rounded-lg">
        {JSON.stringify(data, null, 2)}
      </pre>
    </div>
  );
}
