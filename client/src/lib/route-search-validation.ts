import { valibotValidator } from "@tanstack/valibot-adapter";
import * as v from "valibot";

const optionalIntegerSearchParam = v.optional(
  v.pipe(v.unknown(), v.toNumber(), v.number(), v.integer()),
);

const optionalPositiveIntegerSearchParam = v.optional(
  v.pipe(v.unknown(), v.toNumber(), v.number(), v.integer(), v.minValue(1)),
);

const optionalSortOrderSearchParam = v.optional(v.picklist(["asc", "desc"]));

/** Search validator for workout creation and editing routes. */
export const workoutEditorSearchValidator = valibotValidator(
  v.object({
    addExercise: v.optional(v.boolean()),
    exerciseIndex: optionalIntegerSearchParam,
    newExercise: v.optional(v.boolean()),
  }),
);

/** Search validator for the workouts list route. */
export const workoutsSearchValidator = valibotValidator(
  v.object({
    focusArea: v.optional(v.string()),
    sortOrder: optionalSortOrderSearchParam,
    itemsPerPage: optionalPositiveIntegerSearchParam,
    page: optionalPositiveIntegerSearchParam,
  }),
);

/** Search validator for the exercise detail route. */
export const exerciseDetailSearchValidator = valibotValidator(
  v.object({
    sortOrder: optionalSortOrderSearchParam,
    itemsPerPage: optionalPositiveIntegerSearchParam,
    page: optionalPositiveIntegerSearchParam,
  }),
);

/** Search validator for the analytics route. */
export const analyticsSearchValidator = valibotValidator(
  v.object({
    exerciseId: optionalPositiveIntegerSearchParam,
  }),
);
