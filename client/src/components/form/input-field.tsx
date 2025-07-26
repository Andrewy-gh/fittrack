import type { ChangeEvent } from 'react';
import { useFieldContext } from '@/hooks/form';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

type FormInputProps = {
  label: string;
  placeholder?: string;
  type?: 'text' | 'number';
  className?: string;
};

export default function InputField({
  label,
  placeholder,
  type = 'text',
  className,
}: FormInputProps) {
  const field = useFieldContext<string | number>();

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (type === 'number') {
      const numValue = Number(e.target.value);
      field.handleChange(isNaN(numValue) ? 0 : numValue);
    } else {
      field.handleChange(e.target.value);
    }
  };

  const getValue = () => field.state.value || '';

  const getDefaultPlaceholder = () => {
    if (placeholder) return placeholder;
    return type === 'number' ? '0' : undefined;
  };

  const getInputClassName = () => {
    const numberClasses = type === 'number' ? 'text-center h-9' : '';
    const customClasses = className || '';

    return `${numberClasses} ${customClasses}`.trim();
  };

  return (
    <div className="space-y-2">
      <Label className="text-xs tracking-wider">{label}</Label>
      <Input
        id={field.name}
        name={field.name}
        type={type}
        value={getValue()}
        onBlur={field.handleBlur}
        onChange={handleChange}
        className={getInputClassName()}
        placeholder={getDefaultPlaceholder()}
      />
    </div>
  );
}