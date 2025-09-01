import { lazy } from 'react';
import { createFormHook, createFormHookContexts } from '@tanstack/react-form';

export const { fieldContext, useFieldContext, formContext, useFormContext } =
  createFormHookContexts();

const DatePicker = lazy(() => import('../components/form/date-picker'));
const DatePicker2 = lazy(() => import('../components/form/date-picker2'));
const NotesTextarea = lazy(() => import('../components/form/notes-textarea'));
const NotesTextarea2 = lazy(() => import('../components/form/notes-textarea2'));
const SetTypeSelect = lazy(() => import('../components/form/set-type-select'));
const InputField = lazy(() => import('../components/form/input-field'));
// const AddExerciseField = lazy(
//   () => import('../components/form/add-exercise-field')
// );
const AddExerciseField2 = lazy(
  () => import('../components/form/add-exercise-field2')
);
const WorkoutFocusCombobox = lazy(
  () => import('../components/form/workout-focus-combobox')
);

export const { useAppForm, withForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    DatePicker,
    DatePicker2,
    NotesTextarea,
    NotesTextarea2,
    SetTypeSelect,
    InputField,
    // AddExerciseField,
    AddExerciseField2,
    WorkoutFocusCombobox,
  },
  formComponents: {},
});

export type UseAppForm = ReturnType<typeof useAppForm>;