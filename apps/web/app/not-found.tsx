import Link from "next/link";

export default function NotFound() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4 px-6 py-16 text-center">
      <h1 className="text-2xl font-semibold tracking-tight">Page not found</h1>
      <p className="text-sm text-muted-foreground">
        This page doesn&apos;t exist or has moved.
      </p>
      <Link className="text-sm underline underline-offset-4" href="/">
        Go home
      </Link>
    </main>
  );
}
