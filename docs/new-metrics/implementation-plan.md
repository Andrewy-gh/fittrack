# FitTrack New Metrics Implementation Plan

## Executive Summary

This document outlines the implementation plan for adding 5 new strength training metrics to the FitTrack application. The metrics will enhance user insights into their training progress through advanced calculations including estimated 1RM, intensity percentages, and session-based aggregations.

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

## Database Schema Design

### Phase 1: Core Metrics Columns

#### Exercise Table Enhancements
```sql
-- Migration: 00012_add_metrics_columns.sql
ALTER TABLE exercise ADD COLUMN historical_1rm DECIMAL(8,2);
ALTER TABLE exercise ADD COLUMN training_max DECIMAL(8,2);
ALTER TABLE exercise ADD COLUMN historical_1rm_updated_at TIMESTAMPTZ;
```

**QUESTION** 
What is the 'Training Max' column?

#### Set Table Enhancements
```sql
ALTER TABLE "set" ADD COLUMN is_working BOOLEAN DEFAULT true;
ALTER TABLE "set" ADD COLUMN rpe INTEGER CHECK (rpe >= 1 AND rpe <= 10);
ALTER TABLE "set" ADD COLUMN e1rm DECIMAL(8,2);
ALTER TABLE "set" ADD COLUMN volume DECIMAL(10,2);
```

**QUESTION**
Is the is_working column necessary? I already have a "set_type" column which is an enum of "warmup" and "working".
I don't want the "rpe" column as that is subjective.

### Phase 2: Session Metrics Storage (Future)

#### Session Metrics Table
```sql
-- Migration: 00013_add_session_metrics.sql (optional optimization)
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
    UNIQUE(exercise_id, workout_id, user_id)
);
```

**QUESTION**
Basically, this adds a new session metric row for each exercise in a workout? Does this make it easier to calculate the metrics on the client side and render the charts?

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
    MAX(s.e1rm) FILTER (WHERE s.is_working = true) as session_best_e1rm,
    AVG(s.e1rm) FILTER (WHERE s.is_working = true) as session_avg_e1rm,
    AVG(CASE WHEN e.historical_1rm > 0 THEN (s.weight::decimal / e.historical_1rm * 100) END)
        FILTER (WHERE s.is_working = true) as avg_intensity_vs_hist,
    MAX(CASE WHEN e.historical_1rm > 0 THEN (s.weight::decimal / e.historical_1rm * 100) END)
        FILTER (WHERE s.is_working = true) as session_best_intensity_vs_hist,
    SUM(s.volume) FILTER (WHERE s.is_working = true) as total_volume_working
FROM "set" s
JOIN exercise e ON s.exercise_id = e.id
WHERE s.workout_id = $1 AND s.user_id = $2
GROUP BY e.id, e.name, e.historical_1rm;
```

**INPUT**
There should be a check to see if historial_1rm needs to be updated when a workout is created, updated, or deleted. I've been meaning to add a delete exercise endpoint too as users do not have the ability to delete an exercise. The frontend should have a dialog that pops up to warn against all sets belonging to the exercise being deleted. So if this endpoint is implemented, a historial_1rm check should be triggered.

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
    // Return time-series data for charts
}
```

#### Workout Service Enhancements
```go
// internal/workout/service.go
func (s *Service) GetSessionMetrics(ctx context.Context, workoutID int32, userID string) ([]SessionMetrics, error) {
    // Calculate and return session-level metrics for all exercises in workout
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
    // Update historical 1RM with validation
}

// GET /api/exercises/{id}/metrics-history
func (h *Handler) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
    // Return time-series metrics data for charts
}

// internal/workout/handler.go
// GET /api/workouts/{id}/session-metrics
func (h *Handler) GetSessionMetrics(w http.ResponseWriter, r *http.Request) {
    // Return session-level metrics for all exercises in workout
}
```

### Data Models (`server/internal/database/`)

#### Updated Structs (generated by sqlc)
```go
type Exercise struct {
    ID                    int32              `json:"id"`
    Name                  string             `json:"name"`
    Historical1RM         pgtype.Float8      `json:"historical_1rm"`
    TrainingMax           pgtype.Float8      `json:"training_max"`
    Historical1RMUpdatedAt pgtype.Timestamptz `json:"historical_1rm_updated_at"`
    // ... existing fields
}

type Set struct {
    ID        int32         `json:"id"`
    IsWorking bool          `json:"is_working"`
    RPE       pgtype.Int4   `json:"rpe"`
    E1RM      pgtype.Float8 `json:"e1rm"`
    Volume    pgtype.Float8 `json:"volume"`
    // ... existing fields
}
```

## Frontend Implementation (React/TypeScript)

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
  // Y-axis shows 0-100% scale
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

## Technical Considerations

### Performance Optimizations
- **Indexing**: Add database indexes on new columns used in aggregations
- **Caching**: Consider Redis caching for frequently accessed metrics
- **Lazy Loading**: Load chart data only when tabs are activated
- **Pagination**: Implement pagination for large exercise histories

### Data Integrity
- **Validation**: Strict validation on historical 1RM updates (only allow increases or cycle milestones)
- **Constraints**: Database constraints on RPE (1-10), positive weights, etc.
- **Audit Trail**: Log historical 1RM changes for transparency
- **Rollback**: Ability to revert incorrect historical 1RM updates

### User Experience
- **Progressive Disclosure**: Hide advanced metrics behind settings toggle initially
- **Onboarding**: Guide users through setting initial historical 1RM values
- **Mobile First**: Ensure all charts work well on mobile devices
- **Accessibility**: Proper ARIA labels and keyboard navigation

### Security Considerations
- **Authorization**: Ensure users can only update their own exercise data
- **Input Validation**: Sanitize all numerical inputs
- **Rate Limiting**: Prevent abuse of calculation-heavy endpoints
- **Audit Logging**: Log all metric updates for security monitoring

## Testing Strategy

### Database Testing
- [ ] Migration rollback tests
- [ ] Performance testing with large datasets
- [ ] Concurrent update scenarios
- [ ] Data integrity constraints

### Backend Testing
- [ ] Unit tests for calculation functions
- [ ] Integration tests for new endpoints
- [ ] Performance tests for metric aggregations
- [ ] Security tests for authorization

### Frontend Testing
- [ ] Component unit tests for new charts
- [ ] Integration tests for data flow
- [ ] Visual regression tests
- [ ] Mobile responsiveness tests

### End-to-End Testing
- [ ] Complete user workflows
- [ ] Cross-browser compatibility
- [ ] Performance under load
- [ ] Error handling scenarios

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

### User Feedback
- **Feature Adoption**: Track usage of new metrics
- **User Satisfaction**: Collect feedback through surveys
- **Support Tickets**: Monitor for metric-related issues

## Risk Mitigation

### Technical Risks
- **Performance Impact**: Gradual rollout with monitoring
- **Data Migration**: Comprehensive backup and rollback procedures
- **Calculation Accuracy**: Extensive testing of formulas

### User Experience Risks
- **Feature Complexity**: Optional features with clear onboarding
- **Information Overload**: Settings to hide/show advanced metrics
- **Mobile Performance**: Thorough mobile testing

### Business Risks
- **Development Timeline**: Buffer time for unexpected issues
- **Resource Allocation**: Clear ownership and responsibilities
- **User Adoption**: Phased rollout with feedback collection

## Future Enhancements

### Advanced Features
- **RPE-Based Calculations**: More sophisticated 1RM estimations using RPE
- **Periodization Tracking**: Automatic cycle detection and historical updates
- **AI Insights**: Machine learning for training recommendations
- **Competition Mode**: Meet planning and attempt selection

### Integrations
- **Wearable Devices**: Heart rate and HRV integration
- **External APIs**: Nutrition and sleep data correlation
- **Social Features**: Progress sharing and challenges
- **Export Options**: PDF reports and data exports

---

*This implementation plan serves as a comprehensive guide for adding advanced strength training metrics to FitTrack. Regular reviews and adjustments should be made based on development progress and user feedback.*