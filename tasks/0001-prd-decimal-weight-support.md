# PRD: Decimal Weight Support

## Introduction/Overview

Currently, the FitTrack application only supports integer weights (e.g., 45, 135, 225 lbs) when logging workout sets. This limitation prevents users from accurately tracking workouts that use fractional weight plates or microloading techniques common in progressive overload programs.

This feature will enable users to log weights with one decimal place precision (e.g., 45.5, 135.5, 227.5 lbs), allowing for more accurate workout tracking and better progress monitoring. The feature will preserve all existing data and maintain backward compatibility with the current system.

**Problem Statement:** Users cannot accurately log weights when using fractional plates (2.5 lb, 1.25 lb) or when performing microloading progressions, leading to data inaccuracy and frustration.

**Goal:** Enable precise weight tracking with single decimal place support while maintaining data integrity and a clean user experience.

## Goals

1. **Enable decimal weight entry**: Allow users to log weights with one decimal place (e.g., 45.5 lbs)
2. **Maintain data integrity**: Preserve all existing integer weight data without loss or corruption
3. **Provide clean display**: Show whole numbers without unnecessary decimal points (45 not 45.0) while displaying decimals when needed (45.5)
4. **Ensure backward compatibility**: Existing API clients and integrations continue to work without modification
5. **Improve data accuracy**: Enable users to track small progressive overload increments accurately

## User Stories

1. **As a powerlifter**, I want to log weights with fractional plates (e.g., 227.5 lbs using 2.5 lb plates) so that I can accurately track my progressive overload program.

2. **As a beginner lifter**, I want to add small weight increments (1.25 lb microplates) to my lifts so that I can follow linear progression programs that recommend 2.5 lb total increases.

3. **As a strength athlete**, I want my historical integer weights (45, 135, 225) to display cleanly without decimal points so that my workout log remains readable and uncluttered.

4. **As a user reviewing workout history**, I want to see accurate weight values that reflect what I actually lifted so that I can track progress and plan future workouts.

5. **As an advanced lifter using periodization**, I want to log precise weights for percentage-based programming so that I can follow my training plan accurately.

## Functional Requirements

1. **FR-1**: The system must accept weight values with up to one decimal place (e.g., 45.0, 45.5, 135.5)

2. **FR-2**: The system must accept integer weight values without decimal points (e.g., 45, 135, 225) for backward compatibility

3. **FR-3**: The database must store weight values with exactly one decimal place precision using NUMERIC(10,1) data type

4. **FR-4**: The system must preserve all existing integer weight data during migration without loss

5. **FR-5**: The system must prevent negative weight values at both the database and API validation layers

6. **FR-6**: The API must accept both integer and decimal weight values in JSON requests (e.g., `"weight": 45` or `"weight": 45.5`)

7. **FR-7**: The frontend input field must guide users to enter single decimal increments using step="0.1"

8. **FR-8**: The display must show whole numbers without trailing decimals (display "45" for 45.0)

9. **FR-9**: The display must show decimal values with one decimal place (display "45.5" for 45.5)

10. **FR-10**: NULL weights (for bodyweight exercises) must continue to function correctly

11. **FR-11**: Volume calculations (weight × reps) must work correctly with decimal weights

12. **FR-12**: The system must support weights from 0.1 lbs up to 999,999,999.9 lbs

## Non-Goals (Out of Scope)

1. **Multiple decimal places**: This feature will NOT support more than one decimal place (e.g., 45.55 or 45.125 are not supported)

2. **Unit conversion**: This feature will NOT implement automatic conversion between lbs and kg - decimal support is independent of measurement units

3. **Historical data modification**: This feature will NOT automatically convert or "fix" historical integer data - existing data remains as-is

4. **Rounding logic**: This feature will NOT implement custom rounding - the database will handle rounding to one decimal place

5. **Volume precision**: This feature will NOT change volume calculation precision - volume will remain as integer (truncated)

6. **UI redesign**: This feature will NOT change the overall layout or design of workout logging screens

7. **Plate calculator**: This feature will NOT include a plate calculator to suggest which plates to use for a target weight

## Design Considerations

### User Interface

**Input Field:**
- Add `step="0.1"` attribute to weight input fields to guide decimal entry
- Add `min="0"` to prevent negative values
- Keep existing number input styling and layout

**Display:**
- Implement smart formatting: whole numbers appear as "45" not "45.0"
- Decimal values appear with precision: "45.5"
- Apply consistent formatting across all weight display locations:
  - Recent sets display
  - Workout detail views
  - Exercise history
  - Any other weight display components

**Example:**
```
Before: 225 lbs
After:  225 lbs (same display for whole numbers)

Before: N/A (couldn't log 227.5)
After:  227.5 lbs (new capability)
```

### Data Migration

- All existing integer weights (45, 135, 225) will be automatically converted to NUMERIC type (45.0, 135.0, 225.0 in storage)
- Display logic will strip the ".0" so users see "45" not "45.0"
- No manual data migration required - PostgreSQL ALTER COLUMN handles conversion

## Technical Considerations

### Database

- **Schema Change**: ALTER COLUMN weight from INTEGER to NUMERIC(10,1)
- **Constraint**: Add CHECK constraint to prevent negative weights
- **Migration Tool**: Use Goose for database migration with Up/Down scripts
- **Rollback**: Down migration will truncate decimals back to integers (data loss warning)

### Backend (Go)

- **Type Change**: Update weight fields from `*int` to `*float64` in business logic
- **Database Type**: Use `pgtype.Numeric` at database boundary for NULL handling
- **Validation**: Update validation tags to enforce range: `validate:"omitempty,gte=0,lte=999999999.9"`
- **Code Generation**: Re-run `sqlc generate` to update database models
- **Conversion**: Implement `pgtype.Numeric` ↔ `float64` conversion in repository layer using `math/big` for precision

### Frontend (React/TypeScript)

- **Type**: Already uses `number` type which supports decimals - no breaking changes
- **Formatting**: Create utility function `formatWeight()` to strip trailing ".0"
- **Input**: Update input fields with `step="0.1"` and `min="0"` attributes
- **Display**: Apply formatting function in all weight display components

### API

- **Backward Compatibility**: API must accept both `{"weight": 45}` and `{"weight": 45.5}`
- **Response Format**: Return weights as JSON numbers (45 or 45.5)
- **No Breaking Changes**: Existing API contract remains valid

## Success Metrics

### Adoption Metrics

1. **Decimal Weight Usage Rate**: Track percentage of workout sets logged with decimal weights vs integer weights
   - Target: 15-25% of sets use decimal weights within 3 months of launch
   - Measurement: Query database for `weight % 1 != 0`

2. **Feature Awareness**: Monitor how many unique users log at least one decimal weight
   - Target: 40%+ of active users try decimal weights within first month
   - Measurement: Count distinct users with decimal weight sets

### Quality Metrics

3. **Data Integrity**: Zero data loss during migration
   - Target: 100% of existing integer weights preserved
   - Measurement: Pre/post migration data comparison

4. **User Support Requests**: Reduction in support tickets related to weight precision
   - Target: No increase in support tickets; ideally 50% reduction in precision-related requests
   - Measurement: Track support ticket tags/keywords

5. **Display Accuracy**: Verification that whole numbers display without decimals
   - Target: 100% of whole number weights display as integers (no .0)
   - Measurement: Manual QA testing + user feedback

### Technical Metrics

6. **API Compatibility**: No breaking changes to existing API clients
   - Target: Zero regression bugs reported by API consumers
   - Measurement: Monitor error logs and client feedback

7. **Performance**: No degradation in query performance
   - Target: Query performance remains within 5% of baseline
   - Measurement: Database query timing before/after migration

## Open Questions

1. **Volume Display**: Should volume calculations display with decimal precision, or continue as integers?
   - Current decision: Keep as integers (truncated) - volume precision isn't critical
   - Revisit if users request decimal volume

2. **Input Validation**: Should the frontend strictly enforce single decimal, or let database handle rounding?
   - Current decision: Use `step="0.1"` to guide users, database rounds automatically
   - Revisit if users frequently enter invalid precision

3. **International Users**: Should this feature consider metric (kg) users differently?
   - Current decision: No - decimal support is unit-agnostic
   - Future consideration: Add unit selection/conversion as separate feature

4. **Plate Calculator Integration**: Would users benefit from a plate calculator to suggest combinations?
   - Current decision: Out of scope for this PRD
   - Future consideration: Create separate feature request if user feedback indicates demand

5. **Historical Data Display**: Should we retroactively add ".0" to historical weights or keep them as-is?
   - Current decision: Keep as-is in storage (45.0), format display to show "45"
   - Ensures consistency without modifying historical data

## Acceptance Criteria

This feature will be considered successfully implemented when:

- [ ] Users can enter decimal weights (e.g., 45.5) in the workout logging interface
- [ ] Users can enter integer weights (e.g., 45) without typing ".0"
- [ ] Whole number weights display as "45" not "45.0"
- [ ] Decimal weights display as "45.5" with one decimal place
- [ ] All existing integer weight data is preserved and displays correctly
- [ ] API accepts both integer and decimal weight values
- [ ] Database migration completes without errors or data loss
- [ ] Volume calculations work correctly with decimal weights
- [ ] Bodyweight exercises (NULL weight) continue to function
- [ ] No regression bugs in existing workout logging functionality
- [ ] All existing unit tests pass
- [ ] New tests verify decimal weight handling

## Implementation Notes for Developers

### Database Migration
- Create Goose migration file: `00013_change_weight_to_numeric.sql`
- Update schema.sql to reflect new type
- Run migration in development, staging, then production

### Backend Changes
- Update Go models from `*int` to `*float64`
- Update repository layer to handle `pgtype.Numeric` conversion
- Regenerate sqlc code after schema change
- Update validation rules and Swagger documentation

### Frontend Changes
- Add `formatWeight()` utility function in `client/src/lib/utils.ts`
- Update input components with `step` and `min` attributes
- Apply formatting in all weight display components
- Verify TypeScript types are compatible

### Testing Checklist
- Unit tests for weight conversion logic
- Integration tests for API endpoints with decimal weights
- Frontend tests for display formatting
- Database migration testing on copy of production data
- Regression testing for existing integer weight functionality
