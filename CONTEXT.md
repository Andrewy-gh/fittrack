# FitTrack Context

This context names the product concepts used by the client and server so code
can be grouped by user behavior instead of technical file type.

## Language

**Workout**:
A recorded training session containing one or more exercises.
_Avoid_: Training, session

**Exercise**:
A movement a user tracks across workouts.
_Avoid_: Lift, movement

**Set**:
One recorded effort for an exercise, with reps, weight, and set type.
_Avoid_: Entry, row

**Workout Draft**:
An in-progress workout stored before it is saved as a workout.
_Avoid_: Form state, unsaved workout

**Focus Area**:
The training emphasis assigned to a workout.
_Avoid_: Category, tag

**Metric Chart**:
A reusable visualization of workout or exercise measurements over time.
_Avoid_: Analytics widget, graph

## Relationships

- A **Workout** contains one or more **Exercises**.
- An **Exercise** contains one or more **Sets** inside a **Workout**.
- A **Workout Draft** can become exactly one **Workout**.
- A **Workout** can have zero or one **Focus Area**.
- A **Metric Chart** can show measurements for **Workouts** or **Exercises**.

## Example Dialogue

> **Dev:** "When a user opens the new workout route, should we load their latest **Workout Draft** or start from a blank **Workout**?"
> **Domain expert:** "Load the **Workout Draft** if it has content; saving it creates the **Workout**."

## Flagged Ambiguities

- "training" appears in UI copy, but the domain term is **Workout**.
