import { useQuery } from "@tanstack/react-query";
import { Card, CardContent } from "@/components/ui/card";
import { recommendationApi } from "@/features/workouts/api/recommendations";

/** Displays the snapshot recorded at save time, without applying the current policy again. */
export function SavedRecommendations({
  workoutId,
  userId,
}: {
  workoutId: number;
  userId: string;
}) {
  const query = useQuery({
    queryKey: ["saved-recommendations", userId, workoutId],
    queryFn: ({ signal }) => recommendationApi.saved(workoutId, signal),
  });
  if (query.isPending)
    return <p className="text-sm">Loading saved suggestions…</p>;
  if (!query.data?.ok)
    return <p className="text-sm">We couldn’t load the saved suggestions.</p>;
  if (query.data.value.length === 0) return null;
  return (
    <Card>
      <CardContent className="space-y-4 pt-6">
        <h2 className="font-semibold">Suggestions for this workout</h2>
        <p className="text-xs text-muted-foreground">
          These are the suggestions you saw. The sets you logged are above.
        </p>
        {query.data.value.map(
          ({ exerciseName, recommendation: result, feedback }) => (
            <section
              key={exerciseName}
              className="space-y-1 text-sm"
            >
              <h3 className="font-medium">{exerciseName}</h3>
              <p>How you felt: {result.readiness}</p>
              <p>
                {result.range
                  ? `Suggested: ${result.range.min}–${result.range.max} working sets`
                  : "No sets were suggested"}
              </p>
              <p>
                Based on:{" "}
                {result.source === "prescription"
                  ? "Your saved range"
                  : result.source === "history"
                    ? "Your last workout"
                    : "None"}
                {result.baseline
                  ? ` · ${result.baseline.min}–${result.baseline.max} sets`
                  : ""}
              </p>
              {feedback && (
                <p>
                  How the amount felt:{" "}
                  {feedback === "too_little"
                    ? "Too little"
                    : feedback === "too_much"
                      ? "Too much"
                      : "About right"}
                </p>
              )}
            </section>
          ),
        )}
      </CardContent>
    </Card>
  );
}
