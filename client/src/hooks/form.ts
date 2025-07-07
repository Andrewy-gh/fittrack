import { lazy } from 'react';
import { createFormHook, createFormHookContexts } from '@tanstack/react-form';

export const { fieldContext, useFieldContext, formContext, useFormContext } =
createFormHookContexts();

const DatePicker = lazy(() => import('../components/form/date-picker'));
const NotesTextarea = lazy(() => import('../components/form/notes-textarea'));
const SetTypeSelect = lazy(() => import('../components/form/set-type-select'));
const InputField = lazy(() => import('../components/form/input-field'));
const NumberInput = lazy(() => import('../components/form/number-input'));
const TextInput = lazy(() => import('../components/form/text-input'));

export const { useAppForm, withForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    DatePicker,
    NotesTextarea,
    NumberInput,
    SetTypeSelect,
    TextInput,
    InputField,
  },
  formComponents: {
    
  }
});
