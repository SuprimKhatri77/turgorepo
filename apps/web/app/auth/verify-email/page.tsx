import type { Metadata } from "next";
import { VerifyEmailLoader } from "@/modules/auth/components/verify-email-loader";

export const metadata: Metadata = {
  title: "Verify email | Turgorepo",
};

export default function VerifyEmailPage() {
  return <VerifyEmailLoader />;
}
