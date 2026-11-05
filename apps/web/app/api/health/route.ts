// For load balancer health checks. It doesn't call the API, so an API problem
// can't make the balancer replace healthy web containers.
export function GET() {
  return Response.json({ status: "ok" });
}
