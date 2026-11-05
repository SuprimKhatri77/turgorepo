import Link from "next/link";
import type { User } from "@repo/types";
import { LogoutButton } from "@/modules/auth/components/logout-button";

export function AppShell({
  user,
  title,
  children,
}: {
  user: User;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-1 flex-col">
      <header className="border-b border-border">
        <div className="mx-auto flex w-full max-w-4xl items-center justify-between gap-4 px-6 py-3">
          <Link className="text-sm font-semibold tracking-tight" href="/">
            Turgorepo
          </Link>
          <div className="flex items-center gap-3">
            <span className="text-xs text-muted-foreground">{user.email}</span>
            <LogoutButton />
          </div>
        </div>
      </header>
      <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-6 py-10">
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        {children}
      </main>
    </div>
  );
}
