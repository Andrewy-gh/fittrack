import type { WorkoutWorkoutResponse } from "@/client";

export function getFocusAreas(
  workouts: WorkoutWorkoutResponse[],
): string[] {
  return Array.from(
    new Set(
      workouts
        .map((workout) => workout.workout_focus)
        .filter((focus): focus is string => !!focus),
    ),
  ).sort((a, b) => a.localeCompare(b));
}

export function filterWorkoutsByFocus(
  workouts: WorkoutWorkoutResponse[],
  focusArea: string,
): WorkoutWorkoutResponse[] {
  if (!focusArea || focusArea === "all") {
    return workouts;
  }

  return workouts.filter((workout) => workout.workout_focus === focusArea);
}

export function sortWorkoutsByCreatedAt(
  workouts: WorkoutWorkoutResponse[],
  sortOrder: "asc" | "desc",
): WorkoutWorkoutResponse[] {
  const direction = sortOrder === "asc" ? 1 : -1;

  return [...workouts].sort((a, b) => {
    const aTime = new Date(a.created_at).getTime();
    const bTime = new Date(b.created_at).getTime();
    return (aTime - bTime) * direction;
  });
}

export function paginateWorkouts(
  workouts: WorkoutWorkoutResponse[],
  itemsPerPage: number,
  page: number | undefined,
): {
  pagedWorkouts: WorkoutWorkoutResponse[];
  totalPages: number;
  currentPage: number;
} {
  const safeItemsPerPage = itemsPerPage > 0 ? itemsPerPage : 10;
  const totalPages = Math.max(
    1,
    Math.ceil(workouts.length / safeItemsPerPage),
  );
  const currentPage = Math.min(Math.max(1, page ?? 1), totalPages);
  const startIndex = (currentPage - 1) * safeItemsPerPage;

  return {
    pagedWorkouts: workouts.slice(startIndex, startIndex + safeItemsPerPage),
    totalPages,
    currentPage,
  };
}
