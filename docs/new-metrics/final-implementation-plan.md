# FitTrack New Metrics - Final Implementation Plan

*Last Updated: 2025-09-17*

## Executive Summary

This document outlines the finalized implementation plan for adding 5 new strength training metrics to the FitTrack application. The implementation features automatic PR detection, pre-calculated session metrics for performance, and a comprehensive auto-update system that will significantly enhance user experience while showcasing advanced system design patterns.

## Metrics Overview

### Target Metrics
1. **Total Volume** ✅ (already implemented)
2. **Session Best 1RM** 🆕 - Highest estimated 1RM in a session
3. **Session Average 1RM** 🆕 - Average estimated 1RM across working sets
4. **Session Average Intensity** 🆕 - Average intensity vs historical 1RM
5. **Session Best Intensity** 🆕 - Peak intensity vs historical 1RM

### Calculation Formulas
- **e1RM (Estimated 1RM)**: `weight × (1 + reps/30)` (Epley formula)
- **Volume**: `weight × reps` (already implemented)
- **Intensity**: `(working_weight / historical_1rm) × 100`

### Handling New Exercises (No Historical 1RM)
When `historical_1rm` is 0 or null (new exercises with no historical data):
- **Session Average Intensity**: `AVG(s.weight / session_best_e1rm * 100)` for working sets
- **Session Best Intensity**: `MAX(s.weight / session_best_e1rm * 100)` for working sets

This approach uses the session's best estimated 1RM as the baseline, providing meaningful intensity metrics even for brand new exercises.

### Key Design Decisions
- **Use existing `set_type` column** instead of adding redundant `is_working` column
- **No RPE column** - avoiding subjective metrics
- **Pre-calculated session metrics** for lightning-fast chart performance
- **Automatic PR detection** with conservative update thresholds
- **Comprehensive audit trail** for all historical 1RM changes

## Database Schema Design

### Phase 1: Core Metrics Columns

#### Exercise Table Enhancements
```sql
-- Migration: 00012_add_metrics_columns.sql
ALTER TABLE exercise ADD COLUMN historical_1rm DECIMAL(8,2);
ALTER TABLE exercise ADD COLUMN historical_1rm_updated_at TIMESTAMPTZ;
ALTER TABLE exercise ADD COLUMN historical_1rm_source_workout_id INTEGER REFERENCES workout(id);
```

#### Set Table Enhancements
```sql
ALTER TABLE "set" ADD COLUMN e1rm DECIMAL(8,2);
ALTER TABLE "set" ADD COLUMN volume DECIMAL(10,2);
```

**Note**: Using existing `set_type` column to filter working sets instead of adding redundant `is_working` column.

### Phase 2: Session Metrics Storage (Performance Optimization)

#### Session Metrics Table
```sql
-- Migration: 00013_add_session_metrics.sql
CREATE TABLE session_metrics (
    id SERIAL PRIMARY KEY,
    exercise_id INTEGER NOT NULL REFERENCES exercise(id),
    workout_id INTEGER NOT NULL REFERENCES workout(id),
    user_id TEXT NOT NULL,
    session_best_e1rm DECIMAL(8,2),
    session_avg_e1rm DECIMAL(8,2),
    avg_intensity_vs_hist DECIMAL(5,2),
    session_best_intensity_vs_hist DECIMAL(5,2),
    total_volume_working DECIMAL(10,2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(exercise_id, workout_id, user_id)
);

-- Indexes for fast chart queries
CREATE INDEX idx_session_metrics_exercise_date ON session_metrics(exercise_id, created_at DESC);
CREATE INDEX idx_session_metrics_user_exercise ON session_metrics(user_id, exercise_id);
```

**Benefits of this approach:**
- **Lightning-fast charts**: Pre-calculated metrics = instant chart rendering
- **Scalability**: Performance stays consistent as data grows
- **Portfolio value**: Demonstrates advanced system design and performance optimization thinking
- **Real-world pattern**: How companies like Strava handle analytics at scale

**Decision: Continue with current GetSessionMetrics query approach instead of session_metrics table**
Based on analysis, the current query-based approach is preferred because:
- **Real-time accuracy**: Calculates metrics directly from current set data
- **Simplicity**: No need to maintain/sync a separate metrics table
- **Flexibility**: Easy to modify calculations without data migration
- **Performance**: For single workout queries, the performance difference is negligible

The session metrics table adds complexity without significant benefit for this use case. The current `GetSessionMetrics` query is well-structured and efficient for the application's needs.

## Backend Implementation (Go)

### Repository Layer (`server/internal/`)

#### New SQL Queries (`query.sql`)
```sql
-- name: UpdateExerciseHistorical1RM :exec
UPDATE exercise SET historical_1rm = $3, historical_1rm_updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: GetSessionMetrics :many
SELECT
    e.id as exercise_id,
    e.name as exercise_name,
    e.historical_1rm,
    MAX(s.e1rm) FILTER (WHERE s.set_type = 'working') as session_best_e1rm,
    AVG(s.e1rm) FILTER (WHERE s.set_type = 'working') as session_avg_e1rm,
    -- For exercises with historical_1rm, use historical as baseline
    AVG(CASE WHEN e.historical_1rm > 0 THEN (s.weight::decimal / e.historical_1rm * 100) END)
        FILTER (WHERE s.set_type = 'working') as avg_intensity_vs_hist,
    MAX(CASE WHEN e.historical_1rm > 0 THEN (s.weight::decimal / e.historical_1rm * 100) END)
        FILTER (WHERE s.set_type = 'working') as session_best_intensity_vs_hist,
    -- For new exercises (no historical_1rm), use session_best_e1rm as baseline
    AVG(CASE WHEN e.historical_1rm <= 0 OR e.historical_1rm IS NULL
         THEN (s.weight::decimal / MAX(s.e1rm) OVER (PARTITION BY e.id) * 100) END)
        FILTER (WHERE s.set_type = 'working') as avg_intensity_vs_session_best,
    MAX(CASE WHEN e.historical_1rm <= 0 OR e.historical_1rm IS NULL
         THEN (s.weight::decimal / MAX(s.e1rm) OVER (PARTITION BY e.id) * 100) END)
        FILTER (WHERE s.set_type = 'working') as session_best_intensity_vs_session_best,
    SUM(s.volume) FILTER (WHERE s.set_type = 'working') as total_volume_working
FROM "set" s
JOIN exercise e ON s.exercise_id = e.id
WHERE s.workout_id = $1 AND s.user_id = $2
GROUP BY e.id, e.name, e.historical_1rm;
```

#### Automatic Historical 1RM Updates

```sql
-- name: CheckForHistoricalPRUpdate :one
SELECT
    e.id,
    e.historical_1rm,
    MAX(s.e1rm) FILTER (WHERE s.set_type = 'working') as new_max_e1rm
FROM exercise e
JOIN "set" s ON s.exercise_id = e.id
WHERE s.workout_id = $1 AND s.user_id = $2 AND e.id = $3
GROUP BY e.id, e.historical_1rm;

-- name: RecalculateHistoricalAfterDeletion :one
SELECT
    e.id,
    MAX(s.e1rm) FILTER (WHERE s.set_type = 'working') as recalculated_historical_1rm
FROM exercise e
LEFT JOIN "set" s ON s.exercise_id = e.id AND s.user_id = $2
WHERE e.id = $1 AND e.user_id = $2
GROUP BY e.id;

-- name: DeleteExerciseAndCascadeSets :exec
DELETE FROM exercise WHERE id = $1 AND user_id = $2;
-- Sets will cascade delete due to foreign key constraint
```

#### Repository Updates
- **`exercise/repository.go`**: Add historical 1RM CRUD operations, delete exercise with cascade
- **`workout/repository.go`**: Add session metrics calculation and storage queries
- **Database triggers**: Auto-calculate e1rm and volume on set insert/update
- **Auto-update logic**: Detect PRs and update historical_1rm after workout operations

#### Auto-Update Features (Key Portfolio Differentiator)
- **Automatic PR Detection**: System automatically recognizes new personal records
- **Smart Historical Updates**: Conservative updates only for significant improvements (5%+ threshold)
- **Cascade Handling**: Proper recalculation when workouts/exercises are deleted
- **Audit Trail**: Track which workout achieved each historical 1RM
- **Delete Protection**: Warning dialogs for exercise deletion with impact assessment

### Service Layer (`server/internal/`)

#### Exercise Service Enhancements
```go
// internal/exercise/service.go
func (s *Service) UpdateHistorical1RM(ctx context.Context, exerciseID int32, userID string, newHistorical1RM float64) error {
    // Validation: only allow updates if new value is significantly higher or at cycle milestones
    // Business logic for conservative historical 1RM updates
}

func (s *Service) GetExerciseMetricsHistory(ctx context.Context, exerciseID int32, userID string) ([]ExerciseMetricsHistory, error) {
    // Return time-series data for charts from session_metrics table
}

func (s *Service) DeleteExercise(ctx context.Context, exerciseID int32, userID string) error {
    // Cascade delete all sets and session_metrics
    // Frontend should show warning: "This will delete X sets across Y workouts"
}

func (s *Service) CheckAndUpdateHistoricalPR(ctx context.Context, exerciseID int32, workoutID int32, userID string) error {
    // Auto-detect PRs and update historical_1rm
    // Called after workout create/update

    current, err := s.repo.CheckForHistoricalPRUpdate(ctx, workoutID, userID, exerciseID)
    if err != nil {
        return err
    }

    // Update if new e1RM is significantly higher (5%+ threshold)
    if current.NewMaxE1RM > current.Historical1RM*1.05 {
        return s.repo.UpdateHistorical1RM(ctx, exerciseID, userID, current.NewMaxE1RM, workoutID)
    }

    return nil
}
```

#### Workout Service Enhancements
```go
// internal/workout/service.go
func (s *Service) CreateWorkout(ctx context.Context, params CreateWorkoutParams) (int32, error) {
    // 1. Create workout and sets with auto-calculated e1rm and volume
    workoutID, err := s.repo.CreateWorkout(ctx, params)
    if err != nil {
        return 0, err
    }

    // 2. Calculate and store session metrics
    err = s.calculateAndStoreSessionMetrics(ctx, workoutID, params.UserID)
    if err != nil {
        return 0, err
    }

    // 3. Check for historical 1RM updates
    err = s.checkAllExercisesForPRs(ctx, workoutID, params.UserID)
    if err != nil {
        return 0, err
    }

    return workoutID, nil
}

func (s *Service) UpdateWorkout(ctx context.Context, workoutID int32, params UpdateWorkoutParams) error {
    // 1. Update workout and sets
    // 2. Recalculate session metrics
    // 3. Check for new PRs
}

func (s *Service) DeleteWorkout(ctx context.Context, workoutID int32, userID string) error {
    // 1. Check if this workout contains any current historical 1RMs
    affectedExercises, err := s.repo.GetExercisesWithHistoricalFromWorkout(ctx, workoutID, userID)
    if err != nil {
        return err
    }

    // 2. Delete workout (cascades to sets and session_metrics)
    err = s.repo.DeleteWorkout(ctx, workoutID, userID)
    if err != nil {
        return err
    }

    // 3. Recalculate historical 1RMs for affected exercises
    for _, exerciseID := range affectedExercises {
        err = s.recalculateHistoricalAfterDeletion(ctx, exerciseID, userID)
        if err != nil {
            return err
        }
    }

    return nil
}

func (s *Service) calculateAndStoreSessionMetrics(ctx context.Context, workoutID int32, userID string) error {
    // Calculate session metrics and store in session_metrics table
    // This enables lightning-fast chart queries
}

func calculateE1RM(weight float64, reps int) float64 {
    return weight * (1 + float64(reps)/30)
}

func calculateIntensity(weight float64, historical1RM float64) float64 {
    if historical1RM <= 0 {
        return 0
    }
    return (weight / historical1RM) * 100
}
```

### Handler Layer (`server/internal/`)

#### New Endpoints
```go
// internal/exercise/handler.go
// PUT /api/exercises/{id}/historical-1rm
func (h *Handler) UpdateHistorical1RM(w http.ResponseWriter, r *http.Request) {
    // Manual update of historical 1RM with validation
}

// GET /api/exercises/{id}/metrics-history
func (h *Handler) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
    // Return time-series metrics data for charts from session_metrics table
    // Lightning-fast queries thanks to pre-calculated data
}

// DELETE /api/exercises/{id}
func (h *Handler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
    // Delete exercise with cascade warning
    // Frontend shows: "This will delete X sets across Y workouts"
}

// internal/workout/handler.go
// GET /api/workouts/{id}/session-metrics
func (h *Handler) GetSessionMetrics(w http.ResponseWriter, r *http.Request) {
    // Return session-level metrics from session_metrics table
}

// POST/PUT /api/workouts (enhanced)
// These endpoints now trigger automatic PR detection and session metric calculation
```

### Data Models (`server/internal/database/`)

#### Updated Structs (generated by sqlc)
```go
type Exercise struct {
    ID                         int32              `json:"id"`
    Name                       string             `json:"name"`
    Historical1RM              pgtype.Float8      `json:"historical_1rm"`
    Historical1RMUpdatedAt     pgtype.Timestamptz `json:"historical_1rm_updated_at"`
    Historical1RMSourceWorkoutID pgtype.Int4       `json:"historical_1rm_source_workout_id"`
    // ... existing fields
}

type Set struct {
    ID     int32         `json:"id"`
    E1RM   pgtype.Float8 `json:"e1rm"`
    Volume pgtype.Float8 `json:"volume"`
    // ... existing fields (including set_type for working/warmup distinction)
}

type SessionMetrics struct {
    ID                        int32         `json:"id"`
    ExerciseID               int32         `json:"exercise_id"`
    WorkoutID                int32         `json:"workout_id"`
    UserID                   string        `json:"user_id"`
    SessionBestE1RM          pgtype.Float8 `json:"session_best_e1rm"`
    SessionAvgE1RM           pgtype.Float8 `json:"session_avg_e1rm"`
    AvgIntensityVsHist       pgtype.Float8 `json:"avg_intensity_vs_hist"`
    SessionBestIntensityVsHist pgtype.Float8 `json:"session_best_intensity_vs_hist"`
    TotalVolumeWorking       pgtype.Float8 `json:"total_volume_working"`
    CreatedAt                pgtype.Timestamptz `json:"created_at"`
    UpdatedAt                pgtype.Timestamptz `json:"updated_at"`
}
```

## Frontend Implementation (React/TypeScript)

Note: bucketing/aggregation rules for long-range charts (`6M`, `Y`) are defined in `docs/new-metrics/metrics-history-bucketing.md`.

### API Client Updates (`client/src/`)

#### Generated Types
```typescript
// Update client types for new metrics
interface Exercise {
  id: number;
  name: string;
  historical_1rm?: number;
  historical_1rm_updated_at?: string;
  historical_1rm_source_workout_id?: number;
}

interface SessionMetrics {
  exercise_id: number;
  exercise_name: string;
  session_best_e1rm?: number;
  session_avg_e1rm?: number;
  avg_intensity_vs_hist?: number;
  session_best_intensity_vs_hist?: number;
  total_volume_working?: number;
}
```

### Exercise Detail Page Enhancements (`client/src/routes/_auth/exercises/$exerciseId.tsx`)

#### New Metric Tiles
```tsx
// Add 2 new summary tiles
<Card className="p-4">
  <div className="flex items-center gap-2 mb-2">
    <TrendingUp className="w-5 h-5 text-primary" />
    <span className="text-sm font-semibold">Historical 1RM</span>
  </div>
  <div className="text-2xl text-card-foreground font-bold">
    {exercise.historical_1rm ? `${exercise.historical_1rm} lbs` : 'Not set'}
  </div>
  <Button size="sm" onClick={() => setShowHistorical1RMDialog(true)}>
    <Edit className="w-4 h-4 mr-1" />
    Update
  </Button>
</Card>

<Card className="p-4">
  <div className="flex items-center gap-2 mb-2">
    <Activity className="w-5 h-5 text-primary" />
    <span className="text-sm font-semibold">Avg Intensity</span>
  </div>
  <div className="text-2xl text-card-foreground font-bold">
    {avgIntensity ? `${avgIntensity.toFixed(1)}%` : 'N/A'}
  </div>
</Card>
```

#### New Chart Components
```tsx
// Add 4 new bar charts after existing volume chart
<ChartBar1RM
  title="Session Best 1RM"
  data={sessionBest1RMData}
  exerciseName={exerciseName}
/>

<ChartBar1RM
  title="Session Average 1RM"
  data={sessionAvg1RMData}
  exerciseName={exerciseName}
/>

<ChartBarIntensity
  title="Session Average Intensity"
  data={sessionAvgIntensityData}
  exerciseName={exerciseName}
/>

<ChartBarIntensity
  title="Session Best Intensity"
  data={sessionBestIntensityData}
  exerciseName={exerciseName}
/>
```

### New Chart Components (`client/src/components/charts/`)

#### `chart-bar-1rm.tsx`
```tsx
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface ChartBar1RMProps {
  title: string;
  data: Array<{
    date: string;
    value: number;
    workout_id: number;
  }>;
  exerciseName: string;
}

export function ChartBar1RM({ title, data, exerciseName }: ChartBar1RMProps) {
  // Implementation similar to ChartBarVol but for 1RM data
}
```

#### `chart-bar-intensity.tsx`
```tsx
interface ChartBarIntensityProps {
  title: string;
  data: Array<{
    date: string;
    value: number; // percentage
    workout_id: number;
  }>;
  exerciseName: string;
}

export function ChartBarIntensity({ title, data, exerciseName }: ChartBarIntensityProps) {
  // Implementation for intensity percentage charts
  // Y-axis should allow >100% (dynamic [0, dataMax]); optionally show a 100% reference line.
}
```

### Data Fetching (`client/src/lib/api/`)

#### Updated Exercise Queries
```typescript
// lib/api/exercises.ts
export const exerciseMetricsHistoryQueryOptions = (exerciseId: string) =>
  queryOptions({
    queryKey: ['exercise', exerciseId, 'metrics-history'],
    queryFn: () => api.getExerciseMetricsHistory(parseInt(exerciseId)),
  });

export const updateHistorical1RMMutation = () =>
  useMutation({
    mutationFn: ({ exerciseId, historical1RM }: { exerciseId: number; historical1RM: number }) =>
      api.updateExerciseHistorical1RM(exerciseId, { historical_1rm: historical1RM }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['exercise'] });
    },
  });
```

## Implementation Timeline

### Phase 1: Database Foundation (Week 1)
- [ ] Create migration `00012_add_metrics_columns.sql`
- [ ] Create migration `00013_add_session_metrics.sql`
- [ ] Update `query.sql` with new queries including auto-update logic
- [ ] Add database triggers for e1RM and volume calculation
- [ ] Regenerate Go models with sqlc
- [ ] Test migrations on development database

### Phase 2: Backend Logic (Week 2)
- [ ] Implement repository layer changes
- [ ] Add service layer calculation functions with auto-update logic
- [ ] Create new API endpoints (including delete exercise)
- [ ] Implement automatic PR detection and historical_1rm updates
- [ ] Add session metrics calculation and storage
- [ ] Add comprehensive unit tests
- [ ] Update API documentation

### Phase 3: Frontend Implementation (Week 3)
- [ ] Update generated TypeScript client
- [ ] Create new chart components (using session_metrics for fast rendering)
- [ ] Enhance exercise detail page with new metric tiles
- [ ] Add historical 1RM update dialog
- [ ] Add delete exercise confirmation dialog with impact warning
- [ ] Implement optimized data fetching logic

### Phase 4: Integration & Polish (Week 4)
- [ ] End-to-end testing
- [ ] Performance optimization
- [ ] Mobile responsiveness testing
- [ ] User acceptance testing
- [ ] Documentation updates

## Technical Architecture Highlights

### Performance Optimizations
- **Pre-calculated Metrics**: Session metrics table eliminates heavy aggregations on chart loads
- **Strategic Indexing**: Optimized indexes for time-series chart queries
- **Lazy Loading**: Chart data loaded only when needed
- **Conservative Updates**: Historical 1RM updates only when significant (5%+ threshold)

### Data Integrity & Consistency
- **Automatic Calculations**: Database triggers ensure e1RM and volume are always current
- **Cascade Handling**: Proper cleanup when workouts/exercises are deleted
- **Audit Trail**: Complete history of when and why historical 1RMs were updated
- **Validation Logic**: Conservative business rules prevent erroneous downgrades

### System Design Patterns
- **Event-Driven Updates**: Workout operations trigger metric calculations
- **Denormalization for Performance**: Session metrics table trades storage for speed
- **Idempotent Operations**: Safe to rerun metric calculations
- **Graceful Degradation**: System works even with missing historical data

## Portfolio Value Proposition

This implementation demonstrates several advanced engineering concepts that will impress recruiters:

### 1. **Performance Engineering**
- Understanding of read vs write trade-offs
- Strategic denormalization for analytics workloads
- Index optimization for time-series queries

### 2. **System Design**
- Event-driven architecture for metric updates
- Automatic data consistency maintenance
- Scalable patterns used by major fitness apps

### 3. **Business Logic Sophistication**
- Automatic PR detection with configurable thresholds
- Conservative update policies to prevent data quality issues
- Comprehensive audit trails for transparency

### 4. **User Experience Focus**
- Lightning-fast chart rendering through pre-calculation
- Intelligent defaults and automatic data maintenance
- Proactive data protection (delete warnings)

### 5. **Real-World Patterns**
- Similar architecture to Strava, MyFitnessPal analytics
- Production-ready error handling and edge cases
- Comprehensive testing strategy

## Success Metrics

### User Engagement
- **Chart Interaction**: Track usage of new metric charts
- **Historical 1RM Updates**: Monitor frequency of manual vs automatic updates
- **PR Detection**: Track automatic PR recognition accuracy
- **Session Completion**: Measure impact on workout logging

### Technical Performance
- **Page Load Time**: Maintain <2s load time for exercise pages
- **API Response Time**: <500ms for metric calculations
- **Database Performance**: Monitor query execution times
- **Chart Rendering**: <200ms for chart data fetching

### User Feedback
- **Feature Adoption**: Track usage of new metrics vs existing features
- **User Satisfaction**: Collect feedback through surveys
- **Support Tickets**: Monitor for metric-related issues

---

*This final implementation plan represents a production-ready approach to advanced fitness metrics that balances performance, user experience, and technical sophistication. The automatic PR detection and pre-calculated metrics approach will significantly differentiate this project in a portfolio context.*
