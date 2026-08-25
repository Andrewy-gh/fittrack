import { useSuspenseQuery } from "@tanstack/react-query";
import type { ApplicationUser } from "@/lib/application-user";

import { ExerciseList } from "@/features/exercises/components/exercise-list";
import { getExerciseListQueryOptions } from "@/features/exercises/api/exercise-query-options";

type ExercisesPageProps = {
  user: ApplicationUser | null;
};

export function ExercisesPage({ user }: ExercisesPageProps) {
  const { data: exercises } = useSuspenseQuery(
    getExerciseListQueryOptions(user),
  );

  return <ExerciseList exercises={exercises} />;
}
