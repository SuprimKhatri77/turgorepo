import type { Metadata } from "next";
import { AppShell } from "@/components/layout/app-shell";
import { requireUser } from "@/lib/auth/require-user";
import { LogoutAllDevices } from "@/modules/auth/components/logout-all-devices";

export const metadata: Metadata = {
  title: "Dashboard",
};

export default async function DashboardPage() {
  const user = await requireUser();

  return (
    <AppShell user={user} title="Dashboard">
      <p className="text-sm text-muted-foreground">
        Signed in as {user.name}. Any logged-in user can see this page.
      </p>
      <LogoutAllDevices />
    </AppShell>
  );
}
