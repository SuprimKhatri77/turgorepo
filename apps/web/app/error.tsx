"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function ErrorPage({ reset }: { reset: () => void }) {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4 px-6 py-16 text-center">
      <h1 className="text-2xl font-semibold tracking-tight">
        Something went wrong
      </h1>
      <p className="text-sm text-muted-foreground">
        This is on our side, and you&apos;re still logged in. Try again in a
        moment.
      </p>
      <div className="flex items-center gap-3">
        <Button onClick={reset}>Try again</Button>
        <Link className="text-sm underline underline-offset-4" href="/">
          Go home
        </Link>
      </div>
    </main>
  );
}
