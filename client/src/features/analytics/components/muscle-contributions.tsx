import { useQuery } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { muscleContributionsQueryOptions } from "../api/muscle-contributions";

const names = new Map<string, string>([
  ["chest", "Chest"],
  ["triceps", "Triceps"],
  ["elbow_flexors", "Elbow flexors"],
  ["quadriceps", "Quadriceps"],
  ["hamstrings", "Hamstrings"],
]);
export function MuscleContributions({
  userId,
  startDate,
  timezone,
}: {
  userId: string;
  startDate: string;
  timezone: string;
}) {
  const query = useQuery(
    muscleContributionsQueryOptions(userId, startDate, timezone),
  );
  return (
    <section
      aria-label="Muscle contributions"
      className="space-y-4"
    >
      <div>
        <h3 className="font-semibold">Muscle contributions</h3>
        <p className="text-xs text-muted-foreground">
          All exercises · logged working sets
        </p>
      </div>
      {query.isPending ? (
        <p
          role="status"
          className="text-sm"
        >
          Loading muscle contributions…
        </p>
      ) : query.isError ? (
        <p
          role="alert"
          className="text-sm"
        >
          Could not load muscle contributions.{" "}
          <Button
            variant="link"
            onClick={() => query.refetch()}
          >
            Retry
          </Button>
        </p>
      ) : (
        <>
          {query.data.period.partial && (
            <p className="text-sm">Week in progress · contributions so far.</p>
          )}
          {query.data.workingSets === 0 && (
            <p className="text-sm">
              No valid working sets logged this week. This does not mean no
              training occurred.
            </p>
          )}
          {query.data.invalidWorkingSets > 0 && (
            <p className="text-sm">
              {query.data.invalidWorkingSets} invalid working sets excluded from
              counts.
            </p>
          )}
          <ul className="divide-y">
            {query.data.muscles.map((muscle) => (
              <li
                key={muscle.muscle}
                className="space-y-2 py-3"
              >
                <h4 className="font-medium">
                  {names.get(muscle.muscle) ?? muscle.muscle}
                </h4>
                {muscle.directSets + muscle.indirectSets > 0 ? (
                  <p className="text-sm">
                    {muscle.directSets} direct · {muscle.indirectSets} indirect
                  </p>
                ) : (
                  <p className="text-sm">No supported mapped contributions</p>
                )}
                {muscle.unresolvedSets > 0 && (
                  <p className="text-xs text-muted-foreground">
                    {muscle.unresolvedSets} sets not reviewed for this muscle
                  </p>
                )}
                {muscle.invalidMappedSets + muscle.invalidUnresolvedSets >
                  0 && (
                  <p className="text-xs text-muted-foreground">
                    Invalid entries: {muscle.invalidMappedSets} mapped ·{" "}
                    {muscle.invalidUnresolvedSets} unresolved
                  </p>
                )}
                {query.data.exercises.some(
                  (exercise) => exercise.roles[muscle.muscle],
                ) && (
                  <details className="text-sm">
                    <summary className="cursor-pointer">
                      Contributing exercises
                    </summary>
                    <ul className="space-y-1 pt-2 text-muted-foreground">
                      {query.data.exercises
                        .filter((exercise) => exercise.roles[muscle.muscle])
                        .map((exercise) => (
                          <li key={exercise.exerciseId}>
                            {exercise.name}: {exercise.workingSets}{" "}
                            {exercise.roles[muscle.muscle]}
                            {exercise.invalidWorkingSets > 0
                              ? ` · ${exercise.invalidWorkingSets} invalid`
                              : ""}
                          </li>
                        ))}
                    </ul>
                  </details>
                )}
              </li>
            ))}
          </ul>
          <details className="text-sm">
            <summary className="cursor-pointer font-medium">
              Coverage and counting
            </summary>
            <div className="space-y-3 pt-3 text-muted-foreground">
              <p>
                {query.data.unclassifiedSets} sets without catalog links ·{" "}
                {query.data.unreviewedSets} sets from unreviewed movements.
              </p>
              <p>
                Only listed exercise-to-muscle contributions have been reviewed.
                Other roles remain unknown, even for a mapped exercise. Other
                muscle groups are not covered yet.
              </p>
              <ul className="space-y-1">
                {query.data.exercises.map((exercise) => (
                  <li key={exercise.exerciseId}>
                    {exercise.name}: {exercise.workingSets} working sets
                    {Object.keys(exercise.roles).length === 0
                      ? " · mapping unavailable"
                      : " · some roles reviewed"}
                    {exercise.invalidWorkingSets
                      ? ` · ${exercise.invalidWorkingSets} invalid`
                      : ""}
                  </li>
                ))}
              </ul>
              <p>
                Direct means the main muscle worked; indirect means an assisting
                muscle in a reviewed mapping. Direct and indirect counts stay
                separate. These are logged contributions, not measured stimulus,
                a growth prediction, or a personal target. Muscle groups do not
                imply equal stimulus in every constituent muscle.
              </p>
              <p>
                Each logged set counts once per reviewed role. We do not double
                unilateral sets. Effort, range of motion and training outside
                FitTrack are unknown.
              </p>
              <p>
                These counts do not tell you whether you trained enough for
                muscle growth.
              </p>
              <a
                href="https://doi.org/10.1007/s40279-025-02344-w"
                target="_blank"
                rel="noreferrer"
                className="underline underline-offset-4"
              >
                Research and counting methods
              </a>
              <p>
                Recalculated from current catalog links. Mapping{" "}
                {query.data.mappingVersion} · catalog{" "}
                {query.data.catalogVersion}.
              </p>
            </div>
          </details>
        </>
      )}
    </section>
  );
}
