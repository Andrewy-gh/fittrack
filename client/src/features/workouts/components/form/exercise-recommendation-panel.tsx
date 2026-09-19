import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ChevronDown, ChevronUp, Frown, Meh, Smile } from "lucide-react";
import type { RecommendationResult, RecommendationSnapshot } from "@/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  recommendationApi,
  type RecommendationApi,
} from "@/features/workouts/api/recommendations";
import type { Readiness } from "@/features/workouts/utils/recommendation-context";

const feelings = [
  { value: "sluggish", number: 1, label: "Low energy", Icon: Frown },
  { value: "normal", number: 2, label: "Okay", Icon: Meh },
  { value: "great", number: 3, label: "Feeling good", Icon: Smile },
] as const;

/** Displays guidance separately from the sets the user logs. */
export function ExerciseRecommendationPanel({
  exerciseId,
  exerciseName,
  userId,
  initialSnapshot,
  onChange,
  api = recommendationApi,
}: {
  exerciseId: number;
  exerciseName: string;
  userId: string;
  initialSnapshot?: RecommendationSnapshot;
  onChange: (
    exerciseName: string,
    recommendation: RecommendationResult | null,
  ) => void;
  api?: RecommendationApi;
}) {
  const initialReadiness = initialSnapshot?.recommendation.readiness;
  const [readiness, setReadiness] = useState<Readiness>(
    initialReadiness === "sluggish" || initialReadiness === "great"
      ? initialReadiness
      : "normal",
  );
  const [isOpen, setIsOpen] = useState(true);
  const query = useQuery({
    queryKey: ["exercise-recommendation", userId, exerciseId, readiness],
    queryFn: ({ signal }) => api.get(exerciseId, readiness, signal),
    staleTime: 0,
  });
  const result = query.data?.ok ? query.data.value : null;
  useEffect(() => {
    onChange(exerciseName, result);
  }, [exerciseName, result, onChange]);

  return (
    <Card className="p-4">
      <Collapsible
        open={isOpen}
        onOpenChange={setIsOpen}
      >
        <div className="flex items-center justify-between gap-3">
          <h2 className="font-semibold">Today’s suggestion</h2>
          <CollapsibleTrigger asChild>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="-my-2 -mr-2 size-9 shrink-0 text-muted-foreground"
              aria-label={
                isOpen
                  ? "Collapse today’s suggestion"
                  : "Expand today’s suggestion"
              }
            >
              {isOpen ? (
                <ChevronUp className="size-4" />
              ) : (
                <ChevronDown className="size-4" />
              )}
            </Button>
          </CollapsibleTrigger>
        </div>
        <CollapsibleContent className="pt-4">
          <CardContent className="space-y-4 p-0">
            <div
              aria-live="polite"
              className="space-y-2"
            >
              {query.isPending ? (
                <p>Loading…</p>
              ) : result ? (
                <>
                  <p className="text-lg font-semibold">
                    {result.plan
                      ? `${result.plan.sets} × ${result.plan.reps}${result.plan.weight != null ? ` at ${result.plan.weight} lb` : " · choose weight"}`
                      : "No suggestion yet"}
                  </p>
                  {!result.plan && (
                    <p className="text-sm text-muted-foreground">
                      {result.explanation}
                    </p>
                  )}
                </>
              ) : (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => void query.refetch()}
                >
                  Try again
                </Button>
              )}
            </div>
            <fieldset className="space-y-2">
              <legend className="text-sm">How do you feel?</legend>
              <div className="grid max-w-80 grid-cols-3 gap-2">
                {feelings.map(({ value, number, label, Icon }) => (
                  <label
                    key={value}
                    title={label}
                    className="flex min-h-11 cursor-pointer flex-col justify-center"
                  >
                    <input
                      className="peer sr-only"
                      type="radio"
                      name={`readiness-${exerciseId}`}
                      aria-label={`${number} · ${label}`}
                      value={value}
                      checked={readiness === value}
                      onChange={() => {
                        setReadiness(value);
                      }}
                    />
                    <span className="flex aspect-[16/7] items-center justify-center gap-2 rounded-md border text-sm text-muted-foreground transition-colors peer-checked:border-primary peer-checked:bg-primary/10 peer-checked:text-primary peer-focus-visible:outline-none peer-focus-visible:ring-2 peer-focus-visible:ring-ring peer-focus-visible:ring-offset-2">
                      <Icon
                        className="size-5"
                        aria-hidden="true"
                      />
                      <span aria-hidden="true">{number}</span>
                    </span>
                  </label>
                ))}
              </div>
            </fieldset>
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  );
}
