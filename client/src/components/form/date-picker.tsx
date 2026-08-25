import { useState } from "react";
import { useFieldContext } from "@/hooks/form";
import { format } from "date-fns";
import { Calendar } from "@/components/ui/calendar";
import { Calendar as CalendarIcon } from "lucide-react";
import { Card } from "@/components/ui/card";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

function parseSelectedDate(value: string): Date | undefined {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? undefined : date;
}

export default function DatePicker() {
  const field = useFieldContext<string>();
  const [open, setOpen] = useState(false);
  const selectedDate = parseSelectedDate(field.state.value);

  const handleSelect = (date: Date | undefined) => {
    if (date) {
      field.handleChange(date.toISOString());
      setOpen(false);
    }
  };

  const hasErrors = field.state.meta.errors.length > 0;

  return (
    <div className="space-y-2 h-full">
      <Popover
        open={open}
        onOpenChange={setOpen}
      >
        <PopoverTrigger asChild>
          <Card className="p-4 h-full flex flex-col justify-between min-h-32">
            <div className="flex items-center gap-2 mb-2">
              <CalendarIcon className="w-5 h-5 text-primary" />
              <span className="font-semibold text-sm tracking-tight">Date</span>
            </div>
            <div className="text-card-foreground font-semibold">
              {selectedDate ? format(selectedDate, "PPP") : "Pick a date"}
            </div>
          </Card>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0">
          <Calendar
            mode="single"
            selected={selectedDate}
            onSelect={handleSelect}
          />
        </PopoverContent>
      </Popover>
      {hasErrors && (
        <p className="text-sm text-destructive">
          {field.state.meta.errors.join(", ")}
        </p>
      )}
    </div>
  );
}
