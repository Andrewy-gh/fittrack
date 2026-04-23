import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";
import type { AIWorkoutDraft, AIWorkoutSetInput } from "@/lib/api/ai-chat";
import { cn } from "@/lib/utils";

type ChatWorkoutDraftCardProps = {
  draft: AIWorkoutDraft;
  onEdit: () => void;
  className?: string;
};

export function ChatWorkoutDraftCard({
  draft,
  onEdit,
  className,
}: ChatWorkoutDraftCardProps) {
  const totalSets = draft.exercises.reduce(
    (count, exercise) => count + exercise.sets.length,
    0,
  );

  return (
    <div
      className={cn(
        "rounded-xl border border-primary/20 bg-primary/5 p-4",
        className,
      )}
    >
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-3">
          <div className="space-y-1">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-primary">
              Latest structured workout draft
            </p>
            <p className="text-sm text-muted-foreground">
              Review it here, then reprompt if needed or open it in the workout
              form.
            </p>
          </div>
          <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
            <DraftMetaPill>{formatWorkoutDate(draft.date)}</DraftMetaPill>
            {draft.workoutFocus ? (
              <DraftMetaPill>
                {labelizeWorkoutFocus(draft.workoutFocus)}
              </DraftMetaPill>
            ) : null}
            <DraftMetaPill>
              {draft.exercises.length} exercise
              {draft.exercises.length === 1 ? "" : "s"}
            </DraftMetaPill>
            <DraftMetaPill>
              {totalSets} total set{totalSets === 1 ? "" : "s"}
            </DraftMetaPill>
          </div>
        </div>
        <Button
          type="button"
          onClick={onEdit}
        >
          Edit in workout form
        </Button>
      </div>

      <div className="mt-4 space-y-3">
        {draft.exercises.map((exercise, exerciseIndex) => (
          <div
            key={`${exercise.name}-${exerciseIndex}`}
            className="rounded-lg border bg-background/80 p-3"
          >
            <div className="flex items-start justify-between gap-3">
              <p className="font-medium text-foreground">
                {exerciseIndex + 1}. {exercise.name}
              </p>
              <p className="shrink-0 text-xs text-muted-foreground">
                {exercise.sets.length} set
                {exercise.sets.length === 1 ? "" : "s"}
              </p>
            </div>
            <div className="mt-2 flex flex-wrap gap-2">
              {exercise.sets.map((set, setIndex) => (
                <span
                  key={`${exercise.name}-${setIndex}`}
                  className={cn(
                    "rounded-full border px-2.5 py-1 text-xs font-medium",
                    set.setType === "warmup"
                      ? "border-amber-200 bg-amber-50 text-amber-900"
                      : "border-primary/20 bg-primary/10 text-foreground",
                  )}
                >
                  {formatWorkoutSet(set)}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>

      {draft.notes ? (
        <div className="mt-4 rounded-lg border border-dashed bg-background/60 p-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Notes
          </p>
          <p className="mt-1 text-sm leading-relaxed text-foreground">
            {draft.notes}
          </p>
        </div>
      ) : null}
    </div>
  );
}

function DraftMetaPill({ children }: { children: ReactNode }) {
  return (
    <span className="rounded-full border bg-background/80 px-3 py-1">
      {children}
    </span>
  );
}

function formatWorkoutDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(date);
}

function labelizeWorkoutFocus(value: string): string {
  return value
    .trim()
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
    .replace(/\b\w/g, (match) => match.toUpperCase());
}

function formatWorkoutSet(set: AIWorkoutSetInput): string {
  const label = set.setType === "warmup" ? "Warm-up" : "Working";
  const weight =
    set.weight === undefined ? "" : ` @ ${formatWeight(set.weight)}`;
  return `${label}: ${set.reps} reps${weight}`;
}

function formatWeight(value: number): string {
  return Number.isInteger(value) ? `${value} lb` : `${value.toFixed(1)} lb`;
}
