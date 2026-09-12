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
    return <p className="text-sm">Loading recorded guidance…</p>;
  if (!query.data?.ok)
    return <p className="text-sm">Recorded guidance could not be loaded.</p>;
  if (query.data.value.length === 0) return null;
  return (
    <Card>
      <CardContent className="space-y-4 pt-6">
        <h2 className="font-semibold">Guidance recorded with this workout</h2>
        <p className="text-xs text-muted-foreground">
          Actual sets above remain the record of work performed.
        </p>
        {query.data.value.map(
          ({ exerciseName, recommendation: result, feedback }) => (
            <section
              key={exerciseName}
              className="space-y-1 text-sm"
            >
              <h3 className="font-medium">{exerciseName}</h3>
              <p>Readiness: {result.readiness}</p>
              <p>
                {result.range
                  ? `Recommendation shown: ${result.range.min}–${result.range.max} working sets`
                  : "No range was recommended"}
              </p>
              <p>{result.explanation}</p>
              <p>
                Baseline:{" "}
                {result.source === "prescription"
                  ? "Your saved prescription"
                  : result.source === "history"
                    ? "Exercise history"
                    : "None"}
                {result.baseline
                  ? ` · ${result.baseline.min}–${result.baseline.max} sets`
                  : ""}
              </p>
              {feedback && (
                <p>
                  Feedback:{" "}
                  {feedback === "too_little"
                    ? "Too little"
                    : feedback === "too_much"
                      ? "Too much"
                      : "About right"}
                </p>
              )}
              <p className="text-xs text-muted-foreground">
                Policy: {result.policyVersion}
              </p>
            </section>
          ),
        )}
      </CardContent>
    </Card>
  );
}
