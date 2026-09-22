import { useEffect, useRef, useState } from "react";
import { getExercisesByIdQueryOptions } from "@/client/@tanstack/react-query.gen";
import { ExerciseClassification } from "@/features/exercises/components/exercise-classification";
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
  const heading = useRef<HTMLHeadingElement>(null);
  const [classifying, setClassifying] = useState<{
    id: number;
    name: string;
  } | null>(null);
  const query = useQuery(
    muscleContributionsQueryOptions(userId, startDate, timezone),
  );
  return (
    <section
      aria-label="Muscle sets"
      className="space-y-4"
    >
      <div>
        <h3
          ref={heading}
          tabIndex={-1}
          className="font-semibold"
        >
          Muscle sets
        </h3>
        <p className="text-xs text-muted-foreground">
          All exercises · working sets
        </p>
      </div>
      {classifying && (
        <ClassificationPanel
          key={classifying.id}
          exercise={classifying}
          onClose={() => {
            setClassifying(null);
            heading.current?.focus();
          }}
        />
      )}
      {query.isPending ? (
        <p
          role="status"
          className="text-sm"
        >
          Loading sets…
        </p>
      ) : query.isError ? (
        <p
          role="alert"
          className="text-sm"
        >
          Could not load sets.{" "}
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
            <p className="text-sm">This week so far</p>
          )}
          {query.data.workingSets === 0 && (
            <p className="text-sm">No counts available for this week.</p>
          )}
          {query.data.invalidWorkingSets > 0 && (
            <p className="text-sm">
              {query.data.invalidWorkingSets}{" "}
              {query.data.invalidWorkingSets === 1 ? "set" : "sets"} excluded ·
              check reps or weight
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
                  <p className="text-lg font-semibold">
                    {muscle.directSets} direct · {muscle.indirectSets} indirect
                  </p>
                ) : (
                  <p className="text-sm">No counts available</p>
                )}
                {muscle.unresolvedSets > 0 && (
                  <p className="text-xs text-muted-foreground">
                    {muscle.unresolvedSets}{" "}
                    {muscle.unresolvedSets === 1 ? "set" : "sets"} not counted ·
                    muscle use not reviewed
                  </p>
                )}
                {muscle.invalidMappedSets + muscle.invalidUnresolvedSets >
                  0 && (
                  <p className="text-xs text-muted-foreground">
                    Excluded: {muscle.invalidMappedSets} for this muscle ·{" "}
                    {muscle.invalidUnresolvedSets} muscle unclear
                  </p>
                )}
                {query.data.exercises.some(
                  (exercise) => exercise.roles[muscle.muscle],
                ) && (
                  <details className="text-sm">
                    <summary className="cursor-pointer">Exercises</summary>
                    <ul className="space-y-1 pt-2 text-muted-foreground">
                      {query.data.exercises
                        .filter((exercise) => exercise.roles[muscle.muscle])
                        .map((exercise) => (
                          <li key={exercise.exerciseId}>
                            {exercise.name}: {exercise.workingSets}{" "}
                            {exercise.roles[muscle.muscle]}
                            {exercise.invalidWorkingSets > 0
                              ? ` · ${exercise.invalidWorkingSets} excluded`
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
              Exercise matches
            </summary>
            <div className="space-y-3 pt-3 text-muted-foreground">
              <p>
                {query.data.unclassifiedSets} sets need a match ·{" "}
                {query.data.unreviewedSets} sets await research review
              </p>
              <p>
                We cover some exercises and muscles so far. Missing counts do
                not mean no training.
              </p>
              <ul className="space-y-1">
                {query.data.exercises.map((exercise) => (
                  <li
                    key={exercise.exerciseId}
                    className="space-y-1"
                  >
                    <p>
                      {exercise.name}: {exercise.workingSets} working sets
                      {Object.keys(exercise.roles).length === 0
                        ? " · not counted yet"
                        : " · counted for some muscles"}
                      {exercise.invalidWorkingSets
                        ? ` · ${exercise.invalidWorkingSets} excluded`
                        : ""}
                    </p>
                    {!exercise.catalogId ? (
                      <Button
                        variant="link"
                        className="h-auto whitespace-normal p-0 text-left"
                        onClick={() =>
                          setClassifying({
                            id: exercise.exerciseId,
                            name: exercise.name,
                          })
                        }
                      >
                        Classify {exercise.name}
                      </Button>
                    ) : Object.keys(exercise.roles).length === 0 ? (
                      <p className="text-xs">
                        Matched · muscle use not reviewed yet.
                      </p>
                    ) : null}
                  </li>
                ))}
              </ul>
            </div>
          </details>
        </>
      )}
      <details className="text-sm">
        <summary className="cursor-pointer font-medium">
          About these counts
        </summary>
        <div className="space-y-3 pt-3 text-muted-foreground">
          <p>
            Direct sets work this muscle mainly; indirect sets work it as a
            helper.
          </p>
          <p>
            Research in healthy adults links weekly sets with muscle growth, but
            different counting methods mean these numbers are not a personal
            target.
          </p>
          <p>
            Only reviewed exercise matches count; effort and training outside
            FitTrack are unknown.
          </p>
          <p className="flex flex-wrap gap-x-4 gap-y-2">
            <a
              href="https://acsm.org/resistance-training-guidelines-update-2026/"
              target="_blank"
              rel="noreferrer"
              className="underline underline-offset-4"
            >
              ACSM guidance
            </a>
            <a
              href="https://doi.org/10.1007/s40279-025-02344-w"
              target="_blank"
              rel="noreferrer"
              className="underline underline-offset-4"
            >
              Counting research
            </a>
          </p>
        </div>
      </details>
    </section>
  );
}

function ClassificationPanel({
  exercise,
  onClose,
}: {
  exercise: { id: number; name: string };
  onClose: () => void;
}) {
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    heading.current?.focus();
  }, []);
  const detail = useQuery(
    getExercisesByIdQueryOptions({ path: { id: exercise.id } }),
  );
  const retry = useRef<HTMLButtonElement>(null);
  const hadError = useRef(false);
  useEffect(() => {
    if (detail.isError) {
      retry.current?.focus();
      hadError.current = true;
    } else if (hadError.current && detail.isSuccess) {
      heading.current?.focus();
      hadError.current = false;
    }
  }, [detail.isError, detail.isSuccess]);
  return (
    <section
      aria-label={`Classification for ${exercise.name}`}
      className="space-y-3 rounded-lg border p-3"
    >
      <h4
        ref={heading}
        tabIndex={-1}
        className="font-medium"
      >
        {exercise.name}
      </h4>
      <p className="text-xs text-muted-foreground">
        Choose the same movement. This also updates past weeks.
      </p>
      {detail.isPending ? (
        <p role="status">Loading classification…</p>
      ) : detail.isError ? (
        <div role="alert">
          Could not load classification.{" "}
          <Button
            ref={retry}
            variant="link"
            onClick={() => detail.refetch()}
          >
            Retry classification
          </Button>
        </div>
      ) : (
        <ExerciseClassification
          exerciseId={exercise.id}
          catalog={detail.data.exercise.catalog}
        />
      )}
      <Button
        variant="ghost"
        onClick={onClose}
      >
        Done
      </Button>
    </section>
  );
}
