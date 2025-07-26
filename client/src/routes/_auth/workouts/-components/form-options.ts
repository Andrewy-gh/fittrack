import type { Exercise } from "@/lib/types";
import type { WorkoutFormValues } from "@/lib/types";
import { loadFromLocalStorage, saveToLocalStorage } from "@/lib/local-storage";
import { formOptions } from "@tanstack/react-form";

// MARK: Init values
export const MOCK_VALUES: WorkoutFormValues = {
  date: new Date(), // ! TODO: isoString or not?
  notes: '',
  exercises: [] as Exercise[],
};

const getMockValues = (): WorkoutFormValues => {
  return MOCK_VALUES;
};

export const getInitialValues = (userId: string): WorkoutFormValues => {
  const saved = loadFromLocalStorage(userId);
  return saved || MOCK_VALUES;
};

// MARK: Form opts
export const formOpts = formOptions({
  listeners: {
    onChange: ({ formApi }) => {
      console.log('Saving form data to localStorage');
      saveToLocalStorage(formApi.state.values);
    },
    onChangeDebounceMs: 500,
  },
});

export const formOptsMock = formOptions({
  defaultValues: getMockValues(),
  listeners: {
    onChange: ({ formApi }) => {
      console.log('Saving form data to localStorage');
      saveToLocalStorage(formApi.state.values);
    },
    onChangeDebounceMs: 500,
  },
});