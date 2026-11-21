import type { Metadata } from "next";
import { Geist, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { ReactQueryProvider } from "@/providers/react-query-provider";
import { cn } from "@/lib/utils";
import { getCurrentUser } from "@/lib/auth/current-user";
import { AuthStoreSync } from "@/providers/auth-store-sync";
import { Toaster } from "@/components/ui/sonner";

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
});

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: { default: "Turgorepo", template: "%s | Turgorepo" },
  description:
    "Turborepo starter with a Go/Gin API, Next.js frontend, and a typed OpenAPI client.",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const user = await getCurrentUser();
  return (
    <html
      lang="en"
      className={cn(
        "h-full",
        "antialiased",
        geistSans.variable,
        "font-mono",
        jetbrainsMono.variable,
      )}
    >
      <body className="min-h-full flex flex-col">
        <ReactQueryProvider>
          <AuthStoreSync user={user} />
          {children}
          <Toaster richColors />
        </ReactQueryProvider>
      </body>
    </html>
  );
}
