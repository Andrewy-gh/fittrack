# Local Storage Date Handling in FitTrack

## Overview

This document explains the type compatibility issues encountered when working with dates in the FitTrack application, specifically between the form components, local storage, and API requirements.

## Component Requirements

### 1. Date Picker Components
- **Expect**: `Date` objects
- **Reason**: UI components like calendars need JavaScript Date objects to function properly
- **Location**: `DatePicker2` component uses `useFieldContext<Date>()`

### 2. Local Storage
- **Stores**: JSON strings
- **Requirement**: All data must be serializable to JSON strings
- **Date Handling**: Dates must be converted to ISO strings for storage

### 3. API Layer
- **Expects**: ISO date strings
- **Type Definition**: `WorkoutCreateWorkoutRequest` has `date: string`
- **Format**: `"2023-12-25T10:30:00.000Z"`

### 4. TanStack Form
- **Handles**: The type specified in the field context
- **Date Fields**: Initialized with `Date` objects for proper UI interaction

## The Problem

We encountered a TypeScript error on line 18 of `local-storage.ts`:
```
The left-hand side of an 'instanceof' expression must be of type 'any', an object type or a type parameter.ts(2358)
```

This occurred because:
1. The `WorkoutCreateWorkoutRequest` type defines `date` as a `string`
2. But our form components work with `Date` objects
3. The `saveToLocalStorage` function was trying to check `data.date instanceof Date`
4. TypeScript correctly identified this as an error since `data.date` is typed as a string

## The Solution

We created a flexible type system that handles conversions between all formats:

### Type Definitions

```typescript
// API and Storage Type (from generated client)
type WorkoutCreateWorkoutRequest = {
  date: string; // ISO date string
  exercises: Array<WorkoutExerciseInput>;
  notes?: string;
};

// Flexible Form Type (our custom type)
type FormDataType = Omit<WorkoutCreateWorkoutRequest, 'date'> & {
  date: Date | string; // Accepts both Date objects and strings
};
```

### Data Flow

1. **Form Initialization**: 
   - Start with `WorkoutCreateWorkoutRequest` (date as string)
   - Convert to `Date` object for form components

2. **Form Usage**:
   - Components work with `Date` objects
   - Form state contains `Date` objects

3. **Saving to Local Storage**:
   - Convert `Date` objects to ISO strings
   - Store as JSON

4. **Loading from Local Storage**:
   - Parse JSON
   - Convert ISO strings back to `Date` objects

5. **API Submission**:
   - Convert `Date` objects to ISO strings
   - Submit as `WorkoutCreateWorkoutRequest`

### Implementation

```typescript
// Save to localStorage - converts Date objects to strings
export const saveToLocalStorage = (
  data: FormDataType, // Accepts both Date objects and strings
  userId?: string
) => {
  try {
    const serializedData = {
      ...data,
      date: data.date instanceof Date ? data.date.toISOString() : data.date,
    };
    localStorage.setItem(getStorageKey(userId), JSON.stringify(serializedData));
  } catch (error) {
    console.warn('Failed to save to localStorage:', error);
  }
};

// Load from localStorage - converts strings back to Date objects
export const loadFromLocalStorage = (
  userId?: string
): WorkoutCreateWorkoutRequest | null => {
  try {
    const saved = localStorage.getItem(getStorageKey(userId));
    if (saved) {
      const parsed = JSON.parse(saved);
      if (parsed.date && typeof parsed.date === 'string') {
        parsed.date = new Date(parsed.date);
      }
      return parsed as WorkoutCreateWorkoutRequest;
    }
  } catch (error) {
    console.warn('Failed to load from localStorage:', error);
  }
  return null;
};
```

## Key Takeaways

1. **Type Flexibility**: We needed to create a flexible type that can handle both `Date` objects and strings
2. **Conversion Layer**: Local storage functions act as a conversion layer between different date representations
3. **Type Safety**: Despite the flexibility, we maintain TypeScript type safety through careful casting and type guards
4. **Component Compatibility**: Form components get the `Date` objects they need while storage and API get the strings they require