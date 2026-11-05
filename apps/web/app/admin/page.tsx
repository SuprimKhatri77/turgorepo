import type { Metadata } from "next";
import { AppShell } from "@/components/layout/app-shell";
import { requireUser } from "@/lib/api/auth/require-user";
import { ROLE_RULES } from "@/lib/middleware/config";

export const metadata: Metadata = {
  title: "Admin | Turgorepo",
};

export default async function AdminPage() {
  const user = await requireUser(ROLE_RULES["/admin"]);

  return (
    <AppShell user={user} title="Admin">
      <p className="text-sm text-muted-foreground">
        Signed in as {user.name} with the {user.role} role. Only admins can see
        this page.
      </p>
    </AppShell>
  );
}
