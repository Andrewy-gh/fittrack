import { format } from 'date-fns';
import { Calendar as CalendarIcon } from 'lucide-react';

import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { useState } from 'react';

type DatePickerProps = {
  value?: Date;
  onChange?: (
    date: Date | undefined
  ) => void | ((updater: (prev: Date) => Date) => void);
  className?: string;
  placeholder?: string;
};

export function DatePicker({ value, onChange }: DatePickerProps) {
  const [open, setOpen] = useState(false);

  const handleSelect = (date: Date | undefined) => {
    if (!onChange) return;

    if (date) {
      // If onChange is a function that expects an updater function
      if (typeof onChange === 'function' && onChange.length === 1) {
        onChange(date);
        setOpen(false);
      }
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            'w-full justify-start text-left font-normal bg-neutral-800 border-neutral-700 text-white hover:bg-neutral-700 hover:text-white',
            !value && 'text-neutral-400'
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {value ? format(value, 'PPP') : <span>Pick a date</span>}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0 border-neutral-700 bg-neutral-800">
        <Calendar 
          mode="single" 
          selected={value} 
          onSelect={handleSelect} 
          className="bg-neutral-800 text-white"
          classNames={{
            day: 'text-neutral-200 hover:bg-neutral-700 aria-selected:bg-neutral-600',
            day_selected: 'bg-neutral-600 text-white',
            day_today: 'font-bold',
            day_outside: 'text-neutral-500',
            day_disabled: 'text-neutral-500',
            day_range_middle: 'bg-neutral-700',
            head_cell: 'text-neutral-400',
            caption_label: 'text-white',
            nav_button: 'text-neutral-200 hover:bg-neutral-700',
            dropdown: 'bg-neutral-800 border-neutral-700 text-white',
          }}
        />
      </PopoverContent>
    </Popover>
  );
}
