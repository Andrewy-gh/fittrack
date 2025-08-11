# API Type Generation Tutorial

This guide explains how to automatically generate TypeScript types and API clients from your Go backend's OpenAPI/Swagger documentation.

## Overview

Our setup automatically generates TypeScript types and API client code from the backend's Swagger JSON specification. This ensures type safety between frontend and backend, and eliminates manual type maintenance.

## Prerequisites

- Go backend with Swagger documentation (using swaggo/swag)
- Node.js/Bun frontend project
- OpenAPI TypeScript Codegen package

## Backend Setup

### 1. Generate Swagger Documentation

The backend uses `swaggo/swag` to generate OpenAPI documentation from Go code annotations.

```bash
# In server directory
swag init -g cmd/api/main.go -o docs/
```

This generates:
- `docs/swagger.json` - OpenAPI specification
- `docs/swagger.yaml` - YAML version  
- `docs/docs.go` - Go documentation

### 2. Serve Swagger Documentation

The backend serves the Swagger UI at `/swagger/` endpoint for interactive API exploration.

## Frontend Setup

### 1. Install Dependencies

```bash
# In client directory
bun add -D openapi-typescript-codegen
```

### 2. Create OpenAPI Configuration

Create `client/openapi-config.ts`:

```typescript
import type { Config } from '@hey-api/openapi-ts';

export default {
  input: '../server/docs/swagger.json',
  output: 'src/generated',
  client: 'axios',
  types: {
    enums: 'javascript'
  }
} satisfies Config;
```

### 3. Add Generation Script

Add to `client/package.json`:

```json
{
  "scripts": {
    "generate:api": "openapi --input ../server/docs/swagger.json --output src/generated --client axios"
  }
}
```

### 4. Generate Types

```bash
# Generate API types from backend
bun run generate:api
```

This creates:
- `src/generated/models/` - TypeScript interfaces for all API models
- `src/generated/services/` - API service classes with methods
- `src/generated/core/` - HTTP client configuration
- `src/generated/index.ts` - Main exports

## Using Generated Types

### 1. Import Generated Types and Services

```typescript
import { 
  WorkoutsService, 
  workout_CreateWorkoutRequest,
  workout_WorkoutResponse 
} from '@/generated';
```

### 2. Replace Manual API Calls

**Before (manual):**
```typescript
// Manual API call with custom types
interface CreateWorkoutRequest {
  name: string;
  exercises: ExerciseInput[];
}

const createWorkout = async (data: CreateWorkoutRequest) => {
  const response = await fetch('/api/workouts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  return response.json();
};
```

**After (generated):**
```typescript
// Generated API client with types
import { WorkoutsService, workout_CreateWorkoutRequest } from '@/generated';

const createWorkout = async (data: workout_CreateWorkoutRequest) => {
  return await WorkoutsService.createWorkout(data);
};
```

### 3. Use Generated Types in Components

```typescript
import { workout_WorkoutResponse } from '@/generated';

interface WorkoutListProps {
  workouts: workout_WorkoutResponse[];
}

export function WorkoutList({ workouts }: WorkoutListProps) {
  return (
    <div>
      {workouts.map(workout => (
        <div key={workout.id}>
          <h3>{workout.name}</h3>
          <p>{workout.exercises?.length} exercises</p>
        </div>
      ))}
    </div>
  );
}
```

### 4. Configure API Base URL

Set the base URL for the generated client:

```typescript
// In your app initialization
import { OpenAPI } from '@/generated';

OpenAPI.BASE = process.env.VITE_API_BASE_URL || 'http://localhost:8080';
```

## Automation

### 1. Automatic Regeneration on Build

Add a `prebuild` script to automatically regenerate types before building:

```json
{
  "scripts": {
    "prebuild": "bun run generate:api",
    "build": "bunx --bun vite build"
  }
}
```

Now `bun run build` will automatically regenerate API types first.

### 2. Development Workflow

```bash
# 1. Update backend API with Swagger annotations
# 2. Regenerate backend docs
cd server && swag init -g cmd/api/main.go -o docs/

# 3. Regenerate frontend types
cd ../client && bun run generate:api

# 4. Use updated types in frontend code
```

## Best Practices

### 1. Backend Swagger Annotations

Use proper Swagger annotations in your Go handlers:

```go
// @Summary Create a new workout
// @Description Create a new workout with exercises
// @Tags workouts
// @Accept json
// @Produce json
// @Param workout body workout.CreateWorkoutRequest true "Workout data"
// @Success 201 {object} response.SuccessResponse{data=workout.WorkoutResponse}
// @Failure 400 {object} response.ErrorResponse
// @Router /workouts [post]
func (h *WorkoutHandler) CreateWorkout(c *gin.Context) {
    // Implementation
}
```

### 2. Consistent Naming

- Use consistent naming conventions between backend and frontend
- Generated types follow the pattern: `{package}_{TypeName}`
- Services are named after the API tags: `{Tag}Service`

### 3. Error Handling

Generated services include proper error handling:

```typescript
try {
  const workout = await WorkoutsService.createWorkout(data);
  console.log('Created:', workout);
} catch (error) {
  if (error instanceof ApiError) {
    console.error('API Error:', error.message, error.status);
  }
}
```

### 4. Type Safety

The generated types ensure compile-time type safety:

```typescript
// ✅ Type-safe - will catch missing fields at compile time
const workoutData: workout_CreateWorkoutRequest = {
  name: "Morning Workout",
  exercises: [{
    name: "Push-ups",
    sets: [{ reps: 10, weight: 0 }]
  }]
};

// ❌ Type error - missing required fields
const invalidData: workout_CreateWorkoutRequest = {
  name: "Incomplete"  // Missing exercises field
};
```

## Troubleshooting

### 1. Generation Fails

- Check that `../server/docs/swagger.json` exists and is valid JSON
- Ensure the backend server is generating proper OpenAPI 3.0 spec
- Verify openapi-typescript-codegen is installed

### 2. Types Not Found

- Make sure to run `bun run generate:api` after backend changes
- Check import paths match the generated file structure
- Verify the generated files are in `src/generated/`

### 3. API Calls Fail

- Check that `OpenAPI.BASE` is set correctly
- Ensure backend server is running and accessible
- Verify the API endpoints match the generated service methods

## File Structure

```
project/
├── server/
│   ├── docs/
│   │   ├── swagger.json      # Generated OpenAPI spec
│   │   ├── swagger.yaml
│   │   └── docs.go
│   └── ...
├── client/
│   ├── src/
│   │   ├── generated/        # Generated API client
│   │   │   ├── models/       # TypeScript interfaces
│   │   │   ├── services/     # API service classes
│   │   │   ├── core/         # HTTP client core
│   │   │   └── index.ts      # Main exports
│   │   └── ...
│   ├── openapi-config.ts     # Generation configuration
│   └── package.json
└── docs/
    └── API_TYPE_GENERATION.md  # This file
```

## Next Steps

- Set up Husky pre-commit hooks to ensure types stay in sync
- Add API generation to CI/CD pipelines
- Consider using generated types in testing
- Explore advanced OpenAPI features like response examples
