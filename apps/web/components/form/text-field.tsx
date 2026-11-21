import type { ReactNode } from "react";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";

const inputClassName =
  "h-9 w-full border border-input bg-background px-3 text-sm outline-none transition focus:border-ring focus:ring-1 focus:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-1 aria-invalid:ring-destructive/20";

type TextFieldProps = {
  /** Also used as the input's name. */
  id: string;
  label: ReactNode;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  type?: "email" | "password";
  autoComplete?: string;
  /** Shown at the end of the label row, like a "Forgot password?" link. */
  labelAside?: ReactNode;
};

/** A labelled text input with its validation message. */
export function TextField({
  id,
  label,
  value,
  onChange,
  error,
  type,
  autoComplete,
  labelAside,
}: TextFieldProps) {
  const fieldLabel = <FieldLabel htmlFor={id}>{label}</FieldLabel>;

  return (
    <Field data-invalid={Boolean(error)}>
      {labelAside ? (
        <div className="flex items-center justify-between gap-2">
          {fieldLabel}
          {labelAside}
        </div>
      ) : (
        fieldLabel
      )}
      <input
        id={id}
        name={id}
        type={type}
        autoComplete={autoComplete}
        className={inputClassName}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        aria-invalid={Boolean(error)}
      />
      <FieldError>{error}</FieldError>
    </Field>
  );
}
