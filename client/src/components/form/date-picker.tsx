import { Label } from '@/components/ui/label';
import { Calendar } from 'lucide-react';
import { useFieldContext } from '@/hooks/form';
import { DatePickerBase } from '@/components/date-picker-base';

export default function DatePicker() {
  const field = useFieldContext<Date>();
  return (
    <div className="space-y-3">
      <Label className="text-xs text-neutral-400 tracking-wider flex items-center gap-2">
        <Calendar className="w-3 h-3" />
        TRAINING DATE
      </Label>
      <DatePickerBase
        value={field.state.value}
        onChange={(date) => {
          if (date) {
            field.handleChange(date);
          }
        }}
      />
    </div>
  );
}
