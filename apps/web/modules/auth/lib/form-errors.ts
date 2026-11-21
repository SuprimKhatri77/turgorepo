import type { ValidationError } from "@repo/types";

export function mapFieldErrors(
  error: ValidationError[],
): Record<string, string> {
  return Object.fromEntries(
    (error ?? []).map(({ field, message }) => [field, message]),
  );
}

/** Maps zod issues to the first message for each top-level field. */
export function toFieldErrors(
  issues: ReadonlyArray<{ path: PropertyKey[]; message: string }>,
): Record<string, string> {
  const errors: Record<string, string> = {};
  for (const issue of issues) {
    const field = issue.path[0];
    if (typeof field === "string" && !errors[field]) {
      errors[field] = issue.message;
    }
  }
  return errors;
}
