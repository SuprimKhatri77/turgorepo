import type { Metadata } from "next";
import { CheckEmail } from "@/modules/auth/components/check-email";

export const metadata: Metadata = {
  title: "Verify your email | Turgorepo",
};

export default function CheckEmailPage() {
  return <CheckEmail />;
}
