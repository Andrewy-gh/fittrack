Create a component in the same directory as `add-exercise-field2.tsx` called `workouts-focus-combobox.tsx`. It should be scaffolded in the same way as `add-exercise-field2.tsx`. 

We definitely need to implement useState with selected workout focus and set select workout focus. 

For props:
exercises should be renamed to `workoutsFocus` and should be type of Array<{name: string}> which we is being called from `new-2.tsx`. There is a useSuspenseQuery being used as a prop being passed down to this newly created component. If you need more reference check `add-exercise-field2.tsx`. The type is of `GetWorkoutsFocusValuesResponse` from `client/types.gen.ts`


Render:
Should be GenericCombox with the Array<{name: string}> value as options. 

Replace exercise combo box with generic combo box.
Props should be an array with key 'name' and value of type'string'.
Selected = selected exercise name? ""
On change should be the same and on create should be the same.
The button will be mostly the same, it's just that the onclick handle will be slightly different.
And below the Plus it should be "Add workout focus value"


Think about your answer first before you respond.
Do you have any question before you implement this?