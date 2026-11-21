import Link from "next/link";
import type { ComponentProps, ReactNode } from "react";
import { cn } from "@/lib/utils";

const cardClassName =
  "w-full max-w-sm border border-border bg-card p-6 text-card-foreground shadow-sm";

/** The bordered card each auth page sits in. */
export function AuthCard({ className, ...props }: ComponentProps<"div">) {
  return <div className={cn(cardClassName, className)} {...props} />;
}

/** AuthCard as a form. Validation runs in the component, not the browser. */
export function AuthCardForm({ className, ...props }: ComponentProps<"form">) {
  return (
    <form className={cn(cardClassName, className)} noValidate {...props} />
  );
}

/** A centred link at the bottom of a card, like "Back to sign in". */
export function AuthCardLink({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}) {
  return (
    <p className="mt-5 text-center text-sm">
      <Link className="font-medium underline underline-offset-4" href={href}>
        {children}
      </Link>
    </p>
  );
}
