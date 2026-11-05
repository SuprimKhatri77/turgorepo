"use client";

import { Button } from "@/components/ui/button";
import { useLogout } from "../hooks/mutations/use-logout";

export function LogoutButton() {
  const logout = useLogout();

  return (
    <Button
      variant="outline"
      disabled={logout.isPending}
      onClick={() => logout.mutate()}
    >
      {logout.isPending ? "Logging out…" : "Log out"}
    </Button>
  );
}
