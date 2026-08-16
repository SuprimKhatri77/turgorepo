import type { Metadata } from "next";
import { RegisterForm } from "@/modules/auth/components/register-form";

export const metadata: Metadata = {
  title: "Register | Turgorepo",
};

export default function RegisterPage() {
  return <RegisterForm />;
}
