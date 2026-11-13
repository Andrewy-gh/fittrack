import { defineRule, eslintCompatPlugin } from "@oxlint/plugins";
import type { ESTree } from "@oxlint/plugins";

function keyName(
  key: ESTree.PropertyKey,
  computed: boolean,
): string | undefined {
  if (!computed && key.type === "Identifier") return key.name;
  if (key.type === "Literal" && typeof key.value === "string") return key.value;
  return undefined;
}

// Exact boundary names, not a blanket ban on all query options or catalog data.
const mutableFactories = new Set([
  "getWorkoutsQueryOptions",
  "getWorkoutsByIdQueryOptions",
  "getWorkoutsNewWorkoutContextQueryOptions",
  "getWorkoutsFocusValuesQueryOptions",
  "getWorkoutsContributionDataQueryOptions",
  "getExercisesQueryOptions",
  "getExercisesByIdQueryOptions",
  "getExercisesByIdRecentSetsQueryOptions",
  "getExercisesByIdMetricsHistoryQueryOptions",
  "getWorkoutListQueryOptions",
  "getWorkoutByIdQueryOptions",
  "getNewWorkoutContextQueryOptions",
  "getWorkoutsFocusQueryOptions",
  "getWorkoutContributionQueryOptions",
  "getExerciseListQueryOptions",
  "getExerciseDetailQueryOptions",
  "getRecentExerciseSetsQueryOptions",
  "workoutsQueryOptions",
  "workoutQueryOptions",
  "newWorkoutContextQueryOptions",
  "workoutsFocusValuesQueryOptions",
  "contributionDataQueryOptions",
  "exercisesQueryOptions",
  "exerciseByIdQueryOptions",
  "recentExerciseSetsQueryOptions",
  "exerciseMetricsHistoryQueryOptions",
  "trainingAssessmentQueryOptions",
  "muscleContributionsQueryOptions",
  "getDemoExercisesQueryOptions",
  "getDemoExercisesByIdQueryOptions",
  "getDemoExercisesByIdRecentSetsQueryOptions",
  "getDemoWorkoutsQueryOptions",
  "getDemoWorkoutsByIdQueryOptions",
  "getDemoWorkoutsFocusValuesQueryOptions",
  "getDemoNewWorkoutContextQueryOptions",
  "getDemoContributionDataQueryOptions",
]);

function isMutableFactory(node: ESTree.Expression): boolean {
  if (node.type !== "CallExpression" || node.callee.type !== "Identifier")
    return false;
  return mutableFactories.has(node.callee.name);
}

function isMutableKey(node: ESTree.Expression): boolean {
  if (node.type === "CallExpression" && node.callee.type === "Identifier") {
    return mutableFactories.has(
      node.callee.name.replace(/QueryKey$/, "QueryOptions"),
    );
  }
  if (node.type !== "ArrayExpression") return false;
  const first = node.elements[0];
  return (
    first?.type === "Literal" &&
    (first.value === "training-assessment" ||
      first.value === "muscle-contributions" ||
      first.value === "exercise-recommendation")
  );
}

const noStaticTrainingCache = defineRule({
  meta: {
    type: "problem",
    docs: {
      description:
        "Keep mutable training queries responsive to invalidation (A005/A006).",
    },
    messages: {
      mutable:
        "Mutable training data must not use staleTime: 'static': it disables invalidation refetches. Omit staleTime or use a number/Infinity; retain the training-data invalidation contract.",
    },
  },
  createOnce(context) {
    return {
      ObjectExpression(node) {
        const mutable = node.properties.some((property) =>
          property.type === "SpreadElement"
            ? isMutableFactory(property.argument)
            : property.type === "Property" &&
              keyName(property.key, property.computed) === "queryKey" &&
              isMutableKey(property.value),
        );
        if (!mutable) return;
        for (const property of node.properties) {
          if (
            property.type === "Property" &&
            keyName(property.key, property.computed) === "staleTime" &&
            property.value.type === "Literal" &&
            property.value.value === "static"
          ) {
            context.report({ node: property, messageId: "mutable" });
          }
        }
      },
    };
  },
});

/** FitTrack-specific, syntactic training cache regression guard. */
export default eslintCompatPlugin({
  meta: { name: "training-cache" },
  rules: { "no-static-training-data": noStaticTrainingCache },
});
