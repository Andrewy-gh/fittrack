I want to creat a new component based on `exercise-combobox.tsx` to accept generics. The new component will be in the same directory as `exercise-combobox.tsx` called `generic-combobox.tsx`. In combo box, props, options should be a type of `T[]` and selected should be `T["name"]`. The only requirement of T is that it has a name property of type string. 

We should do the same for `onChange` and `onCreate`. Convert `onCreate` to accept a string instead of T['name'].

We should probably refactor exercise lists to accept the same generics of type T. We should also refactor the exercise combo box to accept generics. `ExerciseComboBox` also has a component called `ExerciseList`, so we should refactor that to be some sort of generic component. The goal is to make that component accept any kind of value. The first step I want to do is just to make sure it can accept `getWorkoutsFocusValuesQueryOptions` from the new query that I just created.

**QUESTIONS ASKED**

Before I implement this, I have a few questions to ensure I understand the requirements correctly:

1.  The goal is to make ExerciseCombobox and ExerciseList generic so they can accept any type T, not just
    ExerciseOption. Is that correct?

    **ANSWER**
    Yes, the exercise option should be a generic type of T that extends:

    ```typescript
    {
    id: number | null; // null for manually created items, number for DB exercises
    name: string;
    }
    ```

2.  You mentioned that options should be of type T, selected should be T['name'], onChange should accept T,
    and onCreate should accept T['name']. However, for onCreate, it seems like it should accept a string (the
    label) to create a new item. Should onCreate accept string instead of T['name']?

    **ANSWER**
    onCreate should accept string instead of T['name']. So if you look at `AddExerciseField2`, it should accept a string for the label.

3.  You want to make sure it can accept `getWorkoutsFocusValuesQueryOptions` from the new query. What is the
    data structure returned by `getWorkoutsFocusValuesQueryOptions`? I see it's defined in
    client/src/client/@tanstack/react-query.gen.ts, but I couldn't find the exact type. Is it an array of
    strings?

    **ANSWER**
    It is an array of strings, check `GetWorkoutsFocusValuesData`

4.  Should the refactored component still use ExerciseOption as the default type for T to maintain backward
    compatibility?

5.  Are there any specific requirements for how the filtering should work with the generic type T? For
    example, should it always filter by a name property, or should that be configurable?


