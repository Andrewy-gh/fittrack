import { useState } from 'react';
import { useFieldContext } from '@/hooks/form';
import { format } from 'date-fns';
import { Calendar } from '@/components/ui/calendar';
import { Calendar as CalendarIcon } from 'lucide-react';
import { Card } from '@/components/ui/card';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

export default function DatePicker2() {
  const field = useFieldContext<Date>();
  const [open, setOpen] = useState(false);

  const handleSelect = (date: Date | undefined) => {
    if (date) {
      field.handleChange(date);
      setOpen(false);
    }
  };

  const hasErrors = field.state.meta.errors.length > 0;

  return (
    <div className="space-y-2">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <CalendarIcon className="w-5 h-5 text-primary" />
              <span className="font-semibold text-sm tracking-tight">Date</span>
            </div>
            <div className="text-card-foreground font-semibold">
              {field.state.value
                ? format(field.state.value, 'PPP')
                : 'Pick a date'}
            </div>
          </Card>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0">
          <Calendar
            mode="single"
            selected={field.state.value}
            onSelect={handleSelect}
          />
        </PopoverContent>
      </Popover>
      {hasErrors && (
        <p className="text-sm text-destructive">
          {field.state.meta.errors.join(', ')}
        </p>
      )}
    </div>
  );
}
