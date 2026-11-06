"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useLogoutAll } from "../hooks/mutations/use-logout-all";

/** Ends every session, after asking once, since it also logs out this browser. */
export function LogoutAllDevices() {
  const logoutAll = useLogoutAll();
  const [confirming, setConfirming] = useState(false);

  return (
    <section className="space-y-3 border border-border p-5">
      <div className="space-y-1">
        <h2 className="text-sm font-semibold">Security</h2>
        <p className="text-sm text-muted-foreground">
          Lost a device, or signed in somewhere you shouldn&apos;t stay signed
          in? Log out everywhere, including this browser.
        </p>
      </div>

      {confirming ? (
        <div className="flex flex-wrap items-center gap-3">
          <p className="text-sm font-medium">Log out all devices?</p>
          <Button
            variant="destructive"
            disabled={logoutAll.isPending}
            onClick={() => logoutAll.mutate()}
          >
            {logoutAll.isPending ? "Logging out…" : "Confirm"}
          </Button>
          <Button
            variant="outline"
            disabled={logoutAll.isPending}
            onClick={() => setConfirming(false)}
          >
            Cancel
          </Button>
        </div>
      ) : (
        <Button variant="outline" onClick={() => setConfirming(true)}>
          Log out all devices
        </Button>
      )}
    </section>
  );
}
