import {
  validateWorkoutText,
  workoutTextLength,
  WORKOUT_TEXT_LIMIT,
} from "@/lib/workout-text";
import { useState } from "react";
import { useFieldContext } from "@/hooks/form";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { FileText } from "lucide-react";
import { Textarea } from "../ui/textarea";

export default function NotesTextarea() {
  const field = useFieldContext<string>();
  const [open, setOpen] = useState(false);
  const text = field.state.value ?? "";
  const error = validateWorkoutText(text.trim(), "Notes");

  return (
    <Dialog
      open={open}
      onOpenChange={setOpen}
    >
      <DialogTrigger asChild>
        <Card
          asChild
          className="p-4"
        >
          <button
            type="button"
            aria-label="Notes"
            className="w-full cursor-pointer text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          >
            <div className="flex items-center gap-2">
              <FileText className="w-5 h-5 text-primary" />
              <span className="font-semibold text-sm tracking-tight">
                Notes
              </span>
            </div>
            <div className="text-xs font-semibold text-card-foreground [overflow-wrap:anywhere]">
              {field.state.value ||
                "Enter any notes, focus areas, or observations for this workout."}
            </div>
          </button>
        </Card>
      </DialogTrigger>
      <DialogContent className="w-[90vw] max-w-md sm:max-w-lg mx-auto">
        <DialogHeader>
          <DialogTitle>Notes</DialogTitle>
          <DialogDescription>
            Enter any notes, focus areas, or observations for this workout.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2">
          <Label
            htmlFor={field.name}
            className="sr-only"
          >
            Notes
          </Label>
          <Textarea
            id={field.name}
            name={field.name}
            value={text}
            onBlur={field.handleBlur}
            onChange={(e) => field.handleChange(e.target.value)}
            autoFocus
            className="min-h-[80px]"
            aria-invalid={Boolean(error) || field.state.meta.errors.length > 0}
            aria-describedby={`${field.name}-limit`}
          />
          <p
            id={`${field.name}-limit`}
            className={
              error
                ? "text-sm text-destructive"
                : "text-sm text-muted-foreground"
            }
            role={error ? "alert" : undefined}
          >
            {error ??
              `${workoutTextLength(text.trim())}/${WORKOUT_TEXT_LIMIT} characters`}
          </p>
          {field.state.meta.errors.length > 0 && (
            <p className="text-sm text-destructive">
              {field.state.meta.errors.join(", ")}
            </p>
          )}
        </div>
        <DialogFooter className="sm:justify-start">
          <DialogClose asChild>
            <Button
              type="button"
              variant="outline"
            >
              Close
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
