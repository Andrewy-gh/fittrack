import { useSuspenseQuery } from "@tanstack/react-query";

import { ExerciseList } from "@/features/exercises/components/exercise-list";
import { exercisesQueryOptions } from "@/features/exercises/api/exercises";
import { getDemoExercisesQueryOptions } from "@/lib/demo-data/query-options";

type ExercisesPageProps = {
  isDemoMode: boolean;
};

export function ExercisesPage({ isDemoMode }: ExercisesPageProps) {
  const { data: exercises } = isDemoMode
    ? useSuspenseQuery(getDemoExercisesQueryOptions())
    : useSuspenseQuery(exercisesQueryOptions());

  return <ExerciseList exercises={exercises} />;
}
