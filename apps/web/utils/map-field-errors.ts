import { ValidationError } from "@repo/types";

export function mapFieldErrors(
  error: ValidationError[],
): Record<string, string> {
  return Object.fromEntries(
    (error ?? []).map(({ field, message }) => [field, message]),
  );
}
