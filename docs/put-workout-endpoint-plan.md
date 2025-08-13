# PUT Workout Endpoint - Complete Implementation Plan

**Date:** 2025-01-15  
**Status:** Planning  
**Context:** Implementing PUT /api/workouts/:id endpoint for workout updates with full type safety and documentation

---

## Executive Summary

This document outlines the complete step-by-step implementation plan for the PUT /api/workouts/:id endpoint, building upon the existing API documentation infrastructure and maintaining consistency with current patterns. The implementation includes backend development, frontend integration, testing, and documentation generation.

## Table of Contents

1. [Current State Analysis](#1-current-state-analysis)
2. [Technical Requirements](#2-technical-requirements)
3. [Implementation Phases](#3-implementation-phases)
4. [Phase 1: Database Layer Enhancement](#phase-1-database-layer-enhancement)
5. [Phase 2: Backend API Implementation](#phase-2-backend-api-implementation)
6. [Phase 3: Frontend Type Generation & Integration](#phase-3-frontend-type-generation--integration)
7. [Phase 4: Testing & Validation](#phase-4-testing--validation)
8. [Phase 5: Documentation & Deployment](#phase-5-documentation--deployment)
9. [Quality Assurance Checklist](#quality-assurance-checklist)
10. [Risk Mitigation](#risk-mitigation)
11. [Future Considerations](#future-considerations)

---

## 1. Current State Analysis

### ✅ **Strengths**
- **Backend Infrastructure:**
  - Go backend with structured handlers using `validator/v10`
  - sqlc-generated database models with PostgreSQL integration
  - Clean separation of concerns (handler/service/repository pattern)
  - Existing Swagger/OpenAPI documentation infrastructure via `swaggo/swag`
  - Row Level Security (RLS) policies implemented for data isolation
  - Comprehensive error handling with structured responses

- **Frontend Infrastructure:**
  - TypeScript frontend with TanStack Router + Query
  - Vite build system with optimized development workflow
  - Existing generated types using `openapi-typescript-codegen`
  - Form handling with `@tanstack/react-form`
  - Local storage integration for form state persistence

- **Documentation & Tooling:**
  - OpenAPI 3.0 specification generation
  - Automated type generation pipeline established
  - Comprehensive testing framework (unit + integration)
  - CI/CD integration ready

### ⚠️ **Current Gaps**
- No existing PUT endpoint for workout updates
- Missing `UpdateWorkoutRequest` type definitions
- Frontend edit workflow partially implemented but disabled
- Need validation rules for partial vs complete updates
- Missing optimistic updates and conflict resolution

### 📊 **Existing Endpoints for Reference**
- `GET /workouts` - List workouts (✅ implemented)
- `POST /workouts` - Create workout (✅ implemented)
- `GET /workouts/{id}` - Get workout with sets (✅ implemented)
- `PUT /workouts/{id}` - **Target implementation**

---

## 2. Technical Requirements

### **Functional Requirements**
- **FR1:** Update existing workout metadata (date, notes)
- **FR2:** Add, update, or remove exercises within a workout
- **FR3:** Add, update, or remove sets within exercises
- **FR4:** Maintain data consistency across related entities
- **FR5:** Preserve user ownership and RLS constraints
- **FR6:** Support partial updates (PATCH-like behavior via PUT)

### **Non-Functional Requirements**
- **NFR1:** Response time < 500ms for typical updates
- **NFR2:** Maintain backward compatibility with existing types
- **NFR3:** Comprehensive input validation and sanitization
- **NFR4:** Atomic operations with transaction rollback on failure
- **NFR5:** Detailed audit logging for all update operations
- **NFR6:** Complete test coverage (unit + integration + e2e)

### **API Contract Specification**
```http
PUT /api/workouts/{id}
Content-Type: application/json
Authorization: x-stack-access-token

Request Body: UpdateWorkoutRequest
Response: 200 OK with UpdatedWorkoutResponse
Error Responses: 400, 401, 403, 404, 409, 500
```

---

## 3. Implementation Phases

### **Phase Overview**
1. **Database Layer Enhancement** (2-3 days)
2. **Backend API Implementation** (3-4 days) 
3. **Frontend Type Generation & Integration** (2-3 days)
4. **Testing & Validation** (2-3 days)
5. **Documentation & Deployment** (1-2 days)

**Total Estimated Duration:** 10-15 days

---

## Phase 1: Database Layer Enhancement

### **1.1 SQL Query Development**

#### **1.1.1 Create Update Workout Query**
**File:** `server/query.sql`

```sql
-- name: UpdateWorkout :one
UPDATE workouts 
SET 
    date = COALESCE($2, date),
    notes = COALESCE($3, notes),
    updated_at = NOW()
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: UpdateWorkoutWithDefaults :one
UPDATE workouts 
SET 
    date = $2,
    notes = $3,
    updated_at = NOW()
WHERE id = $1 AND user_id = $4
RETURNING *;
```

#### **1.1.2 Create Batch Set Operations**
```sql
-- name: DeleteSetsByWorkout :exec
DELETE FROM sets 
WHERE workout_id = $1 
  AND exercise_id = ANY($2::int[])
  AND user_id = $3;

-- name: DeleteSetsByWorkoutAndExercise :exec
DELETE FROM sets 
WHERE workout_id = $1 
  AND exercise_id = $2 
  AND user_id = $3;

-- name: UpdateSet :one
UPDATE sets 
SET 
    weight = COALESCE($2, weight),
    reps = COALESCE($3, reps),
    set_type = COALESCE($4, set_type),
    updated_at = NOW()
WHERE id = $1 
  AND user_id = $5
RETURNING *;

-- name: BulkCreateSets :copyfrom
INSERT INTO sets (exercise_id, workout_id, weight, reps, set_type, user_id)
VALUES ($1, $2, $3, $4, $5, $6);
```

### **1.2 Generate Updated Database Code**
```bash
# Navigate to server directory
cd server

# Generate sqlc code
sqlc generate

# Verify generated functions in internal/database/query.sql.go
```

### **1.3 Update Database Models**
**File:** `server/internal/workout/models.go`

```go
// Add new request/response types
type UpdateWorkoutRequest struct {
    Date      *string         `json:"date,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
    Notes     *string         `json:"notes,omitempty" validate:"omitempty,max=256"`
    Exercises []ExerciseInput `json:"exercises,omitempty" validate:"omitempty,dive"`
}

// Enhanced exercise input for updates
type ExerciseUpdateInput struct {
    ID   *int32      `json:"id,omitempty"`        // nil for new exercises
    Name string      `json:"name" validate:"required,min=1,max=256"`
    Sets []SetUpdateInput `json:"sets" validate:"required,min=1,dive"`
}

type SetUpdateInput struct {
    ID      *int32  `json:"id,omitempty"`        // nil for new sets
    Weight  *int    `json:"weight,omitempty" validate:"omitempty,gte=0"`
    Reps    int     `json:"reps" validate:"required,gte=1"`
    SetType string  `json:"setType" validate:"required,oneof=warmup working"`
}

type UpdateWorkoutResponse struct {
    ID        int32      `json:"id" validate:"required"`
    Date      time.Time  `json:"date" validate:"required"`
    Notes     *string    `json:"notes,omitempty"`
    UpdatedAt time.Time  `json:"updated_at" validate:"required"`
    Success   bool       `json:"success"`
    Message   string     `json:"message"`
}
```

### **1.4 Repository Layer Updates**
**File:** `server/internal/workout/repository.go`

```go
// Add new interface methods
type WorkoutRepository interface {
    // Existing methods...
    UpdateWorkout(ctx context.Context, id int32, req *UpdateWorkoutRequest, userID string) (*UpdateWorkoutResponse, error)
    DeleteWorkoutSets(ctx context.Context, workoutID int32, exerciseIDs []int32, userID string) error
}

// Implementation
func (wr *workoutRepository) UpdateWorkout(ctx context.Context, id int32, req *UpdateWorkoutRequest, userID string) (*UpdateWorkoutResponse, error) {
    // Begin transaction
    tx, err := wr.conn.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    qtx := wr.queries.WithTx(tx)
    
    // Update workout metadata
    updatedWorkout, err := qtx.UpdateWorkout(ctx, db.UpdateWorkoutParams{
        ID:     id,
        Date:   convertTimePtr(req.Date),
        Notes:  convertStringPtr(req.Notes),
        UserID: userID,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to update workout: %w", err)
    }

    // Process exercise updates if provided
    if req.Exercises != nil {
        if err := wr.processExerciseUpdates(ctx, qtx, id, req.Exercises, userID); err != nil {
            return nil, fmt.Errorf("failed to process exercise updates: %w", err)
        }
    }

    // Commit transaction
    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return &UpdateWorkoutResponse{
        ID:        updatedWorkout.ID,
        Date:      updatedWorkout.Date.Time,
        Notes:     convertPGTextPtr(updatedWorkout.Notes),
        UpdatedAt: updatedWorkout.UpdatedAt.Time,
        Success:   true,
        Message:   "Workout updated successfully",
    }, nil
}

func (wr *workoutRepository) processExerciseUpdates(ctx context.Context, qtx *db.Queries, workoutID int32, exercises []ExerciseInput, userID string) error {
    // Complex logic for handling exercise and set updates
    // This includes create, update, and delete operations
    // Implementation details in Phase 2
}
```

---

## Phase 2: Backend API Implementation

### **2.1 Service Layer Implementation**
**File:** `server/internal/workout/service.go`

```go
// Add to WorkoutService interface
type WorkoutService interface {
    // Existing methods...
    UpdateWorkout(ctx context.Context, id int32, req UpdateWorkoutRequest) (*UpdateWorkoutResponse, error)
}

// Implementation with comprehensive validation
func (ws *WorkoutService) UpdateWorkout(ctx context.Context, id int32, req UpdateWorkoutRequest) (*UpdateWorkoutResponse, error) {
    userID, ok := user.FromContext(ctx)
    if !ok {
        return nil, &ErrUnauthorized{Message: "user not authenticated"}
    }

    // Validate workout exists and belongs to user
    existing, err := ws.repo.GetWorkoutWithSets(ctx, id, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch existing workout: %w", err)
    }
    if len(existing) == 0 {
        return nil, &ErrNotFound{Message: "workout not found"}
    }

    // Apply business logic validations
    if err := ws.validateUpdateRequest(req, existing); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Delegate to repository
    return ws.repo.UpdateWorkout(ctx, id, &req, userID)
}

func (ws *WorkoutService) validateUpdateRequest(req UpdateWorkoutRequest, existing []db.GetWorkoutWithSetsRow) error {
    // Business logic validation
    // Check for conflicts, validate exercise/set relationships, etc.
    return nil
}
```

### **2.2 HTTP Handler Implementation**
**File:** `server/internal/workout/handler.go`

```go
// UpdateWorkout godoc
// @Summary Update an existing workout
// @Description Update workout metadata, exercises, and sets for the authenticated user
// @Tags workouts
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Workout ID"
// @Param request body workout.UpdateWorkoutRequest true "Updated workout data"
// @Success 200 {object} workout.UpdateWorkoutResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request - Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Invalid token"
// @Failure 403 {object} response.ErrorResponse "Forbidden - RLS violation"  
// @Failure 404 {object} response.ErrorResponse "Not Found - Workout not found"
// @Failure 409 {object} response.ErrorResponse "Conflict - Concurrent modification"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /workouts/{id} [put]
func (h *WorkoutHandler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
    // Extract and validate workout ID
    workoutID := r.PathValue("id")
    if workoutID == "" {
        response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing workout ID", nil)
        return
    }

    workoutIDInt, err := strconv.Atoi(workoutID)
    if err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid workout ID", err)
        return
    }

    // Parse request body
    var req UpdateWorkoutRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to decode request body", err)
        return
    }

    // Validate request
    if err := h.validator.Struct(req); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "validation error occurred", err)
        return
    }

    // Process update
    result, err := h.workoutService.UpdateWorkout(r.Context(), int32(workoutIDInt), req)
    if err != nil {
        // Handle different error types
        var errUnauthorized *ErrUnauthorized
        var errNotFound *ErrNotFound
        var errConflict *ErrConflict
        
        switch {
        case errors.As(err, &errUnauthorized):
            response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
        case errors.As(err, &errNotFound):
            response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Message, nil)
        case errors.As(err, &errConflict):
            response.ErrorJSON(w, r, h.logger, http.StatusConflict, errConflict.Message, nil)
        default:
            response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to update workout", err)
        }
        return
    }

    // Success response
    if err := response.JSON(w, http.StatusOK, result); err != nil {
        response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
        return
    }
}
```

### **2.3 Router Integration**
**File:** `server/cmd/api/routes.go` (or equivalent routing file)

```go
// Add PUT route
mux.Handle("PUT /api/workouts/{id}", middleware.AuthMiddleware(workoutHandler.UpdateWorkout))
```

### **2.4 Error Types Definition**
**File:** `server/internal/workout/errors.go`

```go
type ErrNotFound struct {
    Message string
}

func (e *ErrNotFound) Error() string {
    return e.Message
}

type ErrConflict struct {
    Message string
}

func (e *ErrConflict) Error() string {
    return e.Message
}
```

---

## Phase 3: Frontend Type Generation & Integration

### **3.1 Swagger Documentation Generation**
```bash
# Navigate to server directory
cd server

# Generate updated Swagger documentation
swag init -g cmd/api/main.go -o ./docs

# Verify PUT endpoint is documented
# Check: docs/swagger.json for /workouts/{id} PUT operation
```

### **3.2 TypeScript Type Generation**
```bash
# Navigate to client directory
cd client

# Install/update dependencies if needed
npm install -D @apidevtools/swagger-cli openapi-typescript-codegen

# Generate TypeScript types from updated OpenAPI spec
npm run generate:api

# This should generate:
# - src/generated/models/workout_UpdateWorkoutRequest.ts
# - src/generated/models/workout_UpdateWorkoutResponse.ts
# - Updated WorkoutsService with updateWorkout method
```

### **3.3 API Client Integration**
**File:** `client/src/lib/api/workouts.ts`

```typescript
import { queryOptions, useMutation } from '@tanstack/react-query';
import { WorkoutsService, OpenAPI } from '@/generated';
import type {
  workout_UpdateWorkoutRequest,
  workout_UpdateWorkoutResponse,
} from '@/generated';

// Query option for workout editing
export function workoutForEditQueryOptions(
  workoutId: number,
  accessToken: string
) {
  return queryOptions<workout_WorkoutWithSetsResponse[], Error>({
    queryKey: ['workouts', 'edit', workoutId],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return WorkoutsService.getWorkouts1(workoutId);
    },
  });
}

// Update workout mutation
export function useUpdateWorkoutMutation(accessToken: string) {
  return useMutation({
    mutationFn: async ({ 
      id, 
      data 
    }: { 
      id: number; 
      data: workout_UpdateWorkoutRequest 
    }) => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return await WorkoutsService.putWorkouts(id, data);
    },
    onSuccess: (data, variables) => {
      // Invalidate and refetch relevant queries
      queryClient.invalidateQueries({
        queryKey: ['workouts', 'list'],
      });
      queryClient.invalidateQueries({
        queryKey: ['workouts', 'details', variables.id],
      });
      queryClient.invalidateQueries({
        queryKey: ['workouts', 'edit', variables.id],
      });
    },
    onError: (error) => {
      // Handle optimistic updates rollback if implemented
      console.error('Failed to update workout:', error);
    },
  });
}

// Transform function for editing
export function transformWorkoutForEditing(
  workouts: workout_WorkoutWithSetsResponse[]
): workout_UpdateWorkoutRequest {
  if (workouts.length === 0) {
    return {
      date: new Date().toISOString(),
      notes: '',
      exercises: [],
    };
  }

  // Group sets by exercise for editing
  const exercisesMap = new Map<number, {
    id: number;
    name: string;
    sets: Array<{
      id: number;
      weight: number | null;
      reps: number;
      setType: 'warmup' | 'working';
    }>;
  }>();

  // Process workout data into editable format
  for (const workout of workouts) {
    const exerciseId = workout.exercise_id || 0;
    if (!exercisesMap.has(exerciseId)) {
      exercisesMap.set(exerciseId, {
        id: exerciseId,
        name: workout.exercise_name || '',
        sets: [],
      });
    }

    const exercise = exercisesMap.get(exerciseId)!;
    exercise.sets.push({
      id: workout.set_id || 0,
      weight: workout.weight || null,
      reps: workout.reps || 0,
      setType: (workout.set_type as 'warmup' | 'working') || 'working',
    });
  }

  return {
    date: workouts[0].workout_date,
    notes: workouts[0].workout_notes || undefined,
    exercises: Array.from(exercisesMap.values()),
  };
}
```

### **3.4 Form Integration Update**
**File:** `client/src/routes/_auth/workouts/$workoutId/edit.tsx`

```typescript
import { useUpdateWorkoutMutation, transformWorkoutForEditing } from '@/lib/api/workouts';

// Enable the previously commented edit form
function EditWorkoutForm({
  accessToken,
  exercises,
  workout,
  workoutId,
}: {
  accessToken: string;
  exercises: exercise_ExerciseResponse[];
  workout: workout_UpdateWorkoutRequest;
  workoutId: number;
}) {
  const updateMutation = useUpdateWorkoutMutation(accessToken);
  const navigate = useNavigate();

  const form = useAppForm({
    defaultValues: workout,
    onSubmit: async ({ value }) => {
      try {
        await updateMutation.mutateAsync({
          id: workoutId,
          data: value,
        });
        
        // Success feedback
        toast.success('Workout updated successfully!');
        navigate({ to: '/workouts' });
      } catch (error) {
        // Error handling
        toast.error('Failed to update workout');
        console.error('Update failed:', error);
      }
    },
  });

  // Implement optimistic updates
  const handleOptimisticUpdate = (updates: Partial<workout_UpdateWorkoutRequest>) => {
    // Update form state immediately
    form.setFieldValue('date', updates.date || form.state.values.date);
    form.setFieldValue('notes', updates.notes || form.state.values.notes);
    
    // Update query cache optimistically
    queryClient.setQueryData(
      ['workouts', 'edit', workoutId],
      (old: workout_WorkoutWithSetsResponse[]) => {
        // Transform and update cached data
        return updateCachedWorkout(old, updates);
      }
    );
  };

  // Rest of component implementation...
}
```

---

## Phase 4: Testing & Validation

### **4.1 Backend Unit Tests**
**File:** `server/internal/workout/handler_test.go`

```go
func TestWorkoutHandler_UpdateWorkout(t *testing.T) {
    userID := "test-user-id"
    workoutID := int32(1)

    tests := []struct {
        name          string
        workoutID     string
        requestBody   interface{}
        setupMock     func(*MockWorkoutRepository)
        ctx           context.Context
        expectedCode  int
        expectedError string
    }{
        {
            name:      "successful full update",
            workoutID: "1",
            requestBody: UpdateWorkoutRequest{
                Date:  stringPtr("2023-01-15T10:00:00Z"),
                Notes: stringPtr("Updated workout notes"),
                Exercises: []ExerciseInput{
                    {
                        Name: "Updated Exercise",
                        Sets: []SetInput{
                            {Weight: intPtr(225), Reps: 8, SetType: "working"},
                        },
                    },
                },
            },
            setupMock: func(m *MockWorkoutRepository) {
                m.On("UpdateWorkout", mock.Anything, workoutID, mock.AnythingOfType("*workout.UpdateWorkoutRequest"), userID).
                  Return(&UpdateWorkoutResponse{
                      ID: workoutID,
                      Success: true,
                      Message: "Workout updated successfully",
                  }, nil)
            },
            ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
            expectedCode: http.StatusOK,
        },
        {
            name:      "partial update - notes only",
            workoutID: "1",
            requestBody: UpdateWorkoutRequest{
                Notes: stringPtr("Just updating notes"),
            },
            setupMock: func(m *MockWorkoutRepository) {
                m.On("UpdateWorkout", mock.Anything, workoutID, mock.AnythingOfType("*workout.UpdateWorkoutRequest"), userID).
                  Return(&UpdateWorkoutResponse{
                      ID: workoutID,
                      Success: true,
                      Message: "Workout updated successfully",
                  }, nil)
            },
            ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
            expectedCode: http.StatusOK,
        },
        {
            name:          "invalid workout ID",
            workoutID:     "invalid",
            setupMock:     func(m *MockWorkoutRepository) {},
            ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
            expectedCode:  http.StatusBadRequest,
            expectedError: "Invalid workout ID",
        },
        {
            name:      "workout not found",
            workoutID: "999",
            requestBody: UpdateWorkoutRequest{
                Notes: stringPtr("Updating non-existent workout"),
            },
            setupMock: func(m *MockWorkoutRepository) {
                m.On("UpdateWorkout", mock.Anything, int32(999), mock.AnythingOfType("*workout.UpdateWorkoutRequest"), userID).
                  Return(nil, &ErrNotFound{Message: "workout not found"})
            },
            ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
            expectedCode:  http.StatusNotFound,
            expectedError: "workout not found",
        },
        {
            name:          "unauthorized user",
            workoutID:     "1",
            requestBody:   UpdateWorkoutRequest{},
            setupMock:     func(m *MockWorkoutRepository) {},
            ctx:           context.Background(),
            expectedCode:  http.StatusUnauthorized,
            expectedError: "user not authenticated",
        },
        // Add more test cases for validation errors, conflicts, etc.
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation similar to existing patterns
        })
    }
}
```

### **4.2 Integration Tests**
**File:** `server/internal/workout/handler_integration_test.go`

```go
func TestWorkoutHandler_UpdateWorkout_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    pool, cleanup := setupTestDatabase(t)
    defer cleanup()

    // Setup test data
    userAID := "test-user-a"
    userBID := "test-user-b"
    
    // Create test workout for User A
    workoutID := setupTestWorkout(t, pool, userAID, "Original workout")
    
    t.Run("UserA_CanUpdateOwnWorkout", func(t *testing.T) {
        ctx := setTestUserContext(context.Background(), t, pool, userAID)
        ctx = user.WithContext(ctx, userAID)

        updateReq := UpdateWorkoutRequest{
            Notes: stringPtr("Updated notes"),
        }

        body, _ := json.Marshal(updateReq)
        req := httptest.NewRequest("PUT", fmt.Sprintf("/api/workouts/%d", workoutID), bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        req = req.WithContext(ctx)
        req.SetPathValue("id", strconv.Itoa(int(workoutID)))

        w := httptest.NewRecorder()
        handler.UpdateWorkout(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        
        var response UpdateWorkoutResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.True(t, response.Success)
        assert.Equal(t, workoutID, response.ID)
    })

    t.Run("UserB_CannotUpdateUserA_Workout", func(t *testing.T) {
        ctx := setTestUserContext(context.Background(), t, pool, userBID)
        ctx = user.WithContext(ctx, userBID)

        updateReq := UpdateWorkoutRequest{
            Notes: stringPtr("Malicious update attempt"),
        }

        body, _ := json.Marshal(updateReq)
        req := httptest.NewRequest("PUT", fmt.Sprintf("/api/workouts/%d", workoutID), bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        req = req.WithContext(ctx)
        req.SetPathValue("id", strconv.Itoa(int(workoutID)))

        w := httptest.NewRecorder()
        handler.UpdateWorkout(w, req)

        // Should fail due to RLS - workout not found for userB
        assert.Equal(t, http.StatusNotFound, w.Code)
    })

    // Add more RLS and concurrency tests
}
```

### **4.3 Frontend Component Tests**
**File:** `client/src/components/__tests__/workout-edit-form.test.tsx`

```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { EditWorkoutForm } from '../EditWorkoutForm';
import { useUpdateWorkoutMutation } from '@/lib/api/workouts';

// Mock the API hook
jest.mock('@/lib/api/workouts');

describe('EditWorkoutForm', () => {
  let queryClient: QueryClient;
  const mockMutateAsync = jest.fn();

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    
    (useUpdateWorkoutMutation as jest.Mock).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isLoading: false,
      error: null,
    });
  });

  const defaultProps = {
    accessToken: 'test-token',
    exercises: [],
    workout: {
      date: '2023-01-15T10:00:00Z',
      notes: 'Test workout',
      exercises: [],
    },
    workoutId: 1,
  };

  const renderWithProvider = (props = defaultProps) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <EditWorkoutForm {...props} />
      </QueryClientProvider>
    );
  };

  test('renders workout form with initial values', () => {
    renderWithProvider();
    
    expect(screen.getByDisplayValue('Test workout')).toBeInTheDocument();
    expect(screen.getByDisplayValue('2023-01-15T10:00:00Z')).toBeInTheDocument();
  });

  test('submits updated workout data', async () => {
    mockMutateAsync.mockResolvedValue({ success: true });
    
    renderWithProvider();
    
    const notesInput = screen.getByDisplayValue('Test workout');
    fireEvent.change(notesInput, { target: { value: 'Updated workout notes' } });
    
    const saveButton = screen.getByText('Save');
    fireEvent.click(saveButton);
    
    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        id: 1,
        data: expect.objectContaining({
          notes: 'Updated workout notes',
        }),
      });
    });
  });

  test('handles update errors gracefully', async () => {
    const errorMessage = 'Failed to update workout';
    mockMutateAsync.mockRejectedValue(new Error(errorMessage));
    
    renderWithProvider();
    
    const saveButton = screen.getByText('Save');
    fireEvent.click(saveButton);
    
    await waitFor(() => {
      expect(screen.getByText(/failed to update/i)).toBeInTheDocument();
    });
  });
});
```

### **4.4 End-to-End Tests**
**File:** `tests/e2e/workout-update.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Workout Update Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Setup authentication and navigation
    await page.goto('/login');
    await page.fill('[name="email"]', 'test@example.com');
    await page.fill('[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await page.waitForURL('/workouts');
  });

  test('user can update workout notes', async ({ page }) => {
    // Navigate to edit page
    await page.click('[data-testid="workout-item"]:first-child [data-testid="edit-button"]');
    await page.waitForURL(/\/workouts\/\d+\/edit/);

    // Update notes
    const notesInput = page.locator('[name="notes"]');
    await notesInput.clear();
    await notesInput.fill('Updated workout notes from E2E test');

    // Save changes
    await page.click('button:has-text("Save")');
    
    // Verify success
    await expect(page.locator('.toast-success')).toBeVisible();
    await page.waitForURL('/workouts');
    
    // Verify changes persisted
    await expect(page.locator(':text("Updated workout notes from E2E test")')).toBeVisible();
  });

  test('user can add new exercise to existing workout', async ({ page }) => {
    // Implementation for adding exercises during update
  });

  test('user can modify sets in existing exercise', async ({ page }) => {
    // Implementation for set modifications
  });

  test('validation prevents invalid updates', async ({ page }) => {
    // Test validation rules
  });
});
```

---

## Phase 5: Documentation & Deployment

### **5.1 API Documentation Updates**
**File:** `server/docs/swagger-implementation-summary.md`

```markdown
## PUT /workouts/{id} - Update Workout

### Overview
Updates an existing workout for the authenticated user. Supports partial updates for workout metadata and complete replacement of exercises and sets.

### Request Schema
```json
{
  "date": "2023-01-15T10:00:00Z",         // optional
  "notes": "Updated workout notes",        // optional  
  "exercises": [                          // optional, but if provided replaces all
    {
      "id": 123,                         // optional, omit for new exercises
      "name": "Bench Press",
      "sets": [
        {
          "id": 456,                     // optional, omit for new sets
          "weight": 225,
          "reps": 8,
          "setType": "working"
        }
      ]
    }
  ]
}
```

### Response Schema
```json
{
  "id": 123,
  "date": "2023-01-15T10:00:00Z",
  "notes": "Updated workout notes",
  "updated_at": "2023-01-15T10:05:00Z",
  "success": true,
  "message": "Workout updated successfully"
}
```

### Business Rules
1. **Partial Updates**: Any omitted fields remain unchanged
2. **Exercise Replacement**: If `exercises` array is provided, it completely replaces existing exercises
3. **Set Management**: Sets without `id` are created, sets with `id` are updated, missing sets are deleted
4. **Validation**: All validation rules from POST endpoint apply
5. **Concurrency**: Last-write-wins strategy, no optimistic locking

### Error Responses
- `400`: Invalid request format or validation failure
- `401`: Authentication required
- `403`: User doesn't own this workout (RLS)
- `404`: Workout not found
- `409`: Concurrent modification conflict (future enhancement)
- `500`: Internal server error
```

### **5.2 Frontend Documentation**
**File:** `client/docs/workout-editing.md`

```markdown
# Workout Editing Implementation

## Architecture Overview
The workout editing system uses TanStack Query for state management with optimistic updates and error rollback capabilities.

## Key Components
- `EditWorkoutForm`: Main editing interface
- `useUpdateWorkoutMutation`: API integration hook  
- `transformWorkoutForEditing`: Data transformation utility
- `WorkoutFormValues`: Shared type definitions

## Usage Patterns

### Basic Update
```typescript
const updateMutation = useUpdateWorkoutMutation(accessToken);

// Update just notes
await updateMutation.mutateAsync({
  id: workoutId,
  data: { notes: 'Updated notes' }
});
```

### Optimistic Updates
```typescript
// Update UI immediately, rollback on error
queryClient.setQueryData(['workouts', 'edit', id], optimisticData);
```

## Error Handling
- Network errors: Automatic retry with exponential backoff
- Validation errors: Form-level error display
- Conflict errors: User prompt for conflict resolution
- RLS violations: Redirect to unauthorized page
```

### **5.3 Deployment Configuration**
**File:** `server/deployment/update-deployment.md`

```markdown
# PUT Endpoint Deployment Guide

## Database Migrations
```sql
-- Verify RLS policies cover UPDATE operations
-- Add any new indexes for performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_workouts_updated_at ON workouts(updated_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sets_workout_exercise ON sets(workout_id, exercise_id);
```

## Environment Variables
- No new environment variables required
- Existing database and authentication configuration sufficient

## Monitoring & Alerting
- Add alerts for high error rates on PUT /workouts/{id}
- Monitor transaction rollback rates
- Track update operation latency

## Rollback Plan
1. Remove PUT route from routing configuration
2. Deploy previous version if needed
3. Database rollback not required (backwards compatible)
```

### **5.4 Team Communication**
**File:** `docs/put-endpoint-rollout-plan.md`

```markdown
# PUT Workout Endpoint Rollout

## Release Timeline
- **Week 1**: Backend implementation and testing
- **Week 2**: Frontend integration and testing
- **Week 3**: End-to-end testing and documentation
- **Week 4**: Staged deployment and monitoring

## Team Responsibilities
- **Backend Team**: API implementation, testing, deployment
- **Frontend Team**: UI integration, type generation, testing
- **DevOps Team**: Deployment pipeline, monitoring setup
- **QA Team**: End-to-end testing, edge case validation

## Success Criteria
- [ ] All tests passing (unit, integration, e2e)
- [ ] Performance benchmarks met (<500ms response time)
- [ ] Zero data corruption or RLS violations
- [ ] Complete type safety maintained
- [ ] Documentation updated and reviewed

## Risk Mitigation
- Feature flag for gradual rollout
- Database backup before deployment
- Monitoring dashboard for real-time metrics
- Rollback procedure tested and documented
```

---

## Quality Assurance Checklist

### **Backend QA**
- [ ] All database queries use parameterized statements
- [ ] Transaction boundaries properly defined
- [ ] RLS policies tested for all user scenarios
- [ ] Error handling covers all edge cases
- [ ] Input validation comprehensive and tested
- [ ] Logging provides sufficient debugging information
- [ ] Performance tests show acceptable response times
- [ ] Memory usage remains stable under load

### **Frontend QA**
- [ ] Types automatically generated and up-to-date
- [ ] Form validation matches backend constraints
- [ ] Optimistic updates with proper rollback
- [ ] Error states handled gracefully
- [ ] Loading states provide good UX
- [ ] Accessibility standards met (WCAG 2.1)
- [ ] Cross-browser compatibility tested
- [ ] Mobile responsive design verified

### **Integration QA**
- [ ] End-to-end workflows function correctly
- [ ] Authentication and authorization working
- [ ] Data consistency maintained across operations
- [ ] Concurrent user scenarios handled properly
- [ ] Network failure recovery implemented
- [ ] Cache invalidation working correctly

### **Documentation QA**
- [ ] API documentation accurate and complete
- [ ] Code comments explain complex logic
- [ ] Team runbooks updated
- [ ] Deployment procedures documented
- [ ] Troubleshooting guides created

---

## Risk Mitigation

### **Technical Risks**
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Data corruption during update | High | Low | Comprehensive transaction handling, rollback testing |
| RLS policy bypass | High | Low | Thorough integration testing, security review |
| Performance degradation | Medium | Medium | Load testing, query optimization, monitoring |
| Type drift between BE/FE | Medium | Low | Automated generation, CI/CD checks |
| Concurrent update conflicts | Low | Medium | Clear conflict resolution strategy |

### **Business Risks**
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| User data loss | High | Very Low | Backup procedures, transaction safety |
| Poor user experience | Medium | Low | Comprehensive UX testing, beta rollout |
| Feature adoption issues | Low | Medium | User training, gradual rollout |

### **Operational Risks**
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Deployment failures | Medium | Low | Staging environment, rollback procedures |
| Monitoring blind spots | Low | Medium | Comprehensive monitoring setup |
| Team knowledge gaps | Low | Low | Documentation, code reviews, pair programming |

---

## Future Considerations

### **Phase 2 Enhancements (Future Tickets)**
- **Optimistic Locking**: Add version fields for conflict detection
- **Bulk Operations**: Support updating multiple workouts simultaneously  
- **Audit Trail**: Track detailed change history for workouts
- **Partial Exercise Updates**: Allow updating individual sets without full replacement
- **Validation Webhooks**: External validation services for complex business rules

### **Scalability Improvements**
- **Caching Strategy**: Redis caching for frequently accessed workouts
- **Database Optimization**: Partitioning for large workout datasets
- **CDN Integration**: Static asset optimization for workout attachments
- **Microservice Migration**: Separate workout service for better scalability

### **User Experience Enhancements**
- **Real-time Collaboration**: Multiple users editing shared workouts
- **Offline Support**: PWA capabilities with sync on reconnection
- **Advanced Validation**: AI-powered form validation and suggestions
- **Undo/Redo**: Multi-level operation history
- **Templates**: Workout templates for common update patterns

### **Analytics & Monitoring**
- **User Behavior Tracking**: How users interact with edit functionality
- **Performance Metrics**: Detailed performance analytics per operation
- **Error Analysis**: Advanced error categorization and alerting
- **A/B Testing**: Framework for testing different UX approaches

---

## Conclusion

This comprehensive plan provides a robust foundation for implementing the PUT /api/workouts/:id endpoint while maintaining high code quality, type safety, and user experience standards. The phased approach ensures thorough testing and validation at each step, while the detailed QA checklist and risk mitigation strategies help ensure a successful deployment.

The implementation leverages existing patterns and infrastructure while introducing new capabilities in a backwards-compatible manner. The extensive testing strategy and documentation ensure long-term maintainability and team knowledge transfer.

**Next Steps:** 
1. Review and approve this plan with the development team
2. Create individual tickets for each phase in your project management system
3. Set up the development environment and begin Phase 1 implementation
4. Schedule regular check-ins to track progress against this plan

---

**Document Version:** 1.0  
**Last Updated:** 2025-01-15  
**Next Review:** 2025-02-15  
**Maintainers:** Backend Team, Frontend Team, DevOps Team
