/**
 * The Content-Security-Policy for every page. proxy.ts creates a nonce for each
 * request, and Next adds it to its own scripts, so only scripts the app rendered
 * can run, and anything they load ('strict-dynamic').
 *
 * To allow a third-party script, give its <Script> the nonce from the x-nonce
 * request header. To call another API or load images from another site, add its
 * origin to connect-src or img-src below.
 */
export function createNonce(): string {
  return btoa(crypto.randomUUID());
}

export function contentSecurityPolicy(nonce: string): string {
  const isDev = process.env.NODE_ENV === "development";

  const directives: Record<string, string[]> = {
    "default-src": ["'self'"],
    // React needs eval in development for its debugging features.
    "script-src": [
      "'self'",
      `'nonce-${nonce}'`,
      "'strict-dynamic'",
      ...(isDev ? ["'unsafe-eval'"] : []),
    ],
    // Next, next/font and the toaster add inline styles. Styles can't run code,
    // so allowing them is a much smaller risk than allowing inline scripts.
    "style-src": ["'self'", "'unsafe-inline'"],
    "img-src": ["'self'", "data:", "blob:", "https://res.cloudinary.com"],
    "font-src": ["'self'"],
    "connect-src": ["'self'", ...originOf(process.env.NEXT_PUBLIC_API_URL)],
    "object-src": ["'none'"],
    "base-uri": ["'self'"],
    "form-action": ["'self'"],
    "frame-ancestors": ["'none'"],
  };

  return Object.entries(directives)
    .map(([name, values]) => `${name} ${values.join(" ")}`)
    .join("; ");
}

function originOf(url: string | undefined): string[] {
  if (!url) return [];
  try {
    return [new URL(url).origin];
  } catch {
    return [];
  }
}
