import { lazy } from 'react';
import { createFormHook, createFormHookContexts } from '@tanstack/react-form';

export const { fieldContext, useFieldContext, formContext, useFormContext } =
  createFormHookContexts();

const DatePicker = lazy(() => import('../components/form/date-picker'));
const NotesTextarea = lazy(() => import('../components/form/notes-textarea'));
const SetTypeSelect = lazy(() => import('../components/form/set-type-select'));
const InputField = lazy(() => import('../components/form/input-field'));
const AddExerciseField = lazy(
  () => import('../components/form/add-exercise-field')
);
const WorkoutFocusCombobox = lazy(
  () => import('../components/form/workout-focus-combobox')
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
