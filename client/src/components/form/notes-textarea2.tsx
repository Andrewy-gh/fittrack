import { useState } from 'react';
import { useFieldContext } from '@/hooks/form';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { FileText } from 'lucide-react';
import { Textarea } from '../ui/textarea';

export default function NotesTextarea2() {
  const field = useFieldContext<string>();
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Card className="p-4">
          <div className="flex items-center gap-2 mb-2">
            <FileText className="w-5 h-5 text-primary" />
            <span className="font-semibold text-sm tracking-tight">Notes</span>
          </div>
          <div className="text-card-foreground font-semibold">
            {field.state.value ?? 'Enter notes'}
          </div>
        </Card>
      </DialogTrigger>
      <DialogContent className="w-[90vw] max-w-md sm:max-w-lg mx-auto my-8">
        <DialogHeader>
          <DialogTitle>Notes</DialogTitle>
          <DialogDescription>
            Enter any notes, focus areas, or observations for this workout.
          </DialogDescription>
        </DialogHeader>
        <Textarea
          id={field.name}
          name={field.name}
          value={field.state.value}
          onBlur={field.handleBlur}
          onChange={(e) => field.handleChange(e.target.value)}
          className="min-h-[80px]"
        />
        <DialogFooter className="sm:justify-start">
          <DialogClose asChild>
            <Button type="button" variant="outline">
              Close
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
