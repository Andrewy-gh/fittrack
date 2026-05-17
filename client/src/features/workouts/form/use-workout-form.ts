import { lazy } from "react";
import { createFormHook, createFormHookContexts } from "@tanstack/react-form";

export const { fieldContext, useFieldContext, formContext, useFormContext } =
  createFormHookContexts();

const DatePicker = lazy(() => import("./fields/date-picker"));
const NotesTextarea = lazy(() => import("./fields/notes-textarea"));
const SetTypeSelect = lazy(() => import("./fields/set-type-select"));
const InputField = lazy(() => import("./fields/input-field"));
const AddExerciseField = lazy(() => import("./fields/add-exercise-field"));
const WorkoutFocusCombobox = lazy(
  () => import("./fields/workout-focus-combobox"),
);

export const { useAppForm, withForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    DatePicker,
    NotesTextarea,
    SetTypeSelect,
    InputField,
    AddExerciseField,
    WorkoutFocusCombobox,
  },
  formComponents: {},
});

export type UseAppForm = ReturnType<typeof useAppForm>;
