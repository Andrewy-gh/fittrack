import { useState } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { WorkoutNotesCard } from '@/components/workouts/workout-notes-card';
import { Button } from '@/components/ui/button';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';

type LastWorkoutNoteSectionProps = {
  title: string;
  note?: string | null;
  dateLabel?: string;
};

export function LastWorkoutNoteSection({
  title,
  note,
  dateLabel,
}: LastWorkoutNoteSectionProps) {
  const [isOpen, setIsOpen] = useState(false);
  const trimmedNote = note?.trim();

  if (!trimmedNote) {
    return null;
  }

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger asChild>
        <Button
          type="button"
          variant="outline"
          className="h-auto w-full justify-between px-4 py-3 text-left"
        >
          <span className="text-sm font-medium text-foreground">
            {isOpen ? 'Hide last workout note' : 'Show last workout note'}
          </span>
          {isOpen ? (
            <ChevronUp className="h-4 w-4" />
          ) : (
            <ChevronDown className="h-4 w-4" />
          )}
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent className="pt-3">
        <WorkoutNotesCard title={title} note={trimmedNote} dateLabel={dateLabel} />
      </CollapsibleContent>
    </Collapsible>
  );
}
