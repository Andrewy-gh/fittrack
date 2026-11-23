import type { ChangeEvent } from "react";
import { useFieldContext } from "@/hooks/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type FormInputProps = {
  label: string;
  placeholder?: string;
  type?: "text" | "number";
  className?: string;
  step?: string;
  min?: string;
  allowEmpty?: boolean;
};

export default function InputField({
  label,
  placeholder,
  type = "text",
  className,
  step,
  min,
  allowEmpty = false,
}: FormInputProps) {
  const field = useFieldContext<string | number | undefined>();

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (type === "number") {
      if (allowEmpty && e.target.value === "") {
        field.handleChange(undefined);
        return;
      }
      const numValue = Number(e.target.value);
      field.handleChange(isNaN(numValue) ? 0 : numValue);
    } else {
      field.handleChange(e.target.value);
    }
  };

  const getValue = () =>
    allowEmpty ? (field.state.value ?? "") : field.state.value || "";

  const getDefaultPlaceholder = () => {
    if (placeholder) return placeholder;
    return type === "number" ? "0" : undefined;
  };

  const getInputClassName = () => {
    const numberClasses = type === "number" ? "text-center h-9" : "";
    const customClasses = className || "";

    return `${numberClasses} ${customClasses}`.trim();
  };

  const hasErrors = field.state.meta.errors.length > 0;

  return (
    <div className="space-y-4">
      <Label
        htmlFor={field.name}
        className="tracking-wider"
      >
        {label}
      </Label>
      <Input
        id={field.name}
        name={field.name}
        type={type}
        value={getValue()}
        onBlur={field.handleBlur}
        onChange={handleChange}
        className={getInputClassName()}
        placeholder={getDefaultPlaceholder()}
        step={step}
        min={min}
        aria-invalid={hasErrors}
      />
      {hasErrors && (
        <p className="text-sm text-destructive">
          {field.state.meta.errors.join(", ")}
        </p>
      )}
    </div>
  );
}
