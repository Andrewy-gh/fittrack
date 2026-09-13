import type { RecommendationSnapshot } from "@/client";
import { Card, CardContent } from "@/components/ui/card";

/** Displays the suggestions recorded when a workout was saved. */
export function SavedRecommendations({
  recommendations,
}: {
  recommendations: RecommendationSnapshot[];
}) {
  if (recommendations.length === 0) return null;

  return (
    <Card>
      <CardContent className="space-y-4 pt-6">
        <h2 className="font-semibold">Suggestions for this workout</h2>
        <p className="text-xs text-muted-foreground">
          These are the suggestions you saw. The sets you logged are above.
        </p>
        {recommendations.map(({ exerciseName, recommendation }) => (
          <section
            key={exerciseName}
            className="space-y-1 text-sm"
          >
            <h3 className="font-medium">{exerciseName}</h3>
            <p>How you felt: {recommendation.readiness}</p>
            <p>
              {recommendation.plan
                ? `Suggested: ${recommendation.plan.sets} × ${recommendation.plan.reps}${recommendation.plan.weight != null ? ` at ${recommendation.plan.weight} lb` : " · choose weight"}`
                : "No sets were suggested"}
            </p>
          </section>
        ))}
      </CardContent>
    </Card>
  );
}
