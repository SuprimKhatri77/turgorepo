import type { Metadata } from "next";
import { ResetPasswordLoader } from "@/modules/auth/components/reset-password-loader";

export const metadata: Metadata = {
  title: "Reset password | Turgorepo",
};

export default function ResetPasswordPage() {
  return <ResetPasswordLoader />;
}
