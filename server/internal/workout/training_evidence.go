package workout

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/exercisecatalog"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	evidenceDateLayout = "2006-01-02"

	// TrainingEvidenceCalculationPolicyVersion identifies the deterministic counting policy
	// implemented by EvidenceService. A changed meaning requires a new version.
	TrainingEvidenceCalculationPolicyVersion = "training-evidence-v1"
	currentClassificationBasis               = "current_exercise_catalog_link"
)

var (
	ErrInvalidTrainingEvidencePeriod = errors.New("invalid training evidence period")
	ErrFutureTrainingEvidencePeriod  = errors.New("training evidence period starts in the future")
	ErrInconsistentExerciseCatalog   = errors.New("inconsistent exercise catalog link")
	ErrInvalidTrainingEvidenceData   = errors.New("invalid training evidence data")
)

// EvidenceQuerier is the narrow read dependency required by EvidenceService.
type EvidenceQuerier interface {
	ListTrainingEvidence(context.Context, db.ListTrainingEvidenceParams) ([]db.ListTrainingEvidenceRow, error)
}

// TrainingEvidenceRequest names one seven-local-calendar-day period.
type TrainingEvidenceRequest struct {
	StartDate string `json:"startDate"`
	Timezone  string `json:"timezone"`
}

// TrainingEvidence is a descriptive count summary, not a target or progress verdict.
type TrainingEvidence struct {
	Period         EvidencePeriod         `json:"period"`
	Activity       EvidenceActivity       `json:"activity"`
	Classification EvidenceClassification `json:"classification"`
	Muscles        []MuscleEvidence       `json:"muscles"`
	Basis          EvidenceBasis          `json:"basis"`
	Limitations    []string               `json:"limitations"`
}

// EvidencePeriod records both the requested local dates and their timestamp bounds.
type EvidencePeriod struct {
	StartDate  string    `json:"startDate"`
	EndDate    string    `json:"endDate"`
	Timezone   string    `json:"timezone"`
	StartAt    time.Time `json:"startAt"`
	EndAt      time.Time `json:"endAt"`
	ObservedAt time.Time `json:"observedAt"`
	Partial    bool      `json:"partial"`
}

// EvidenceActivity keeps logging and working-set frequency separate.
type EvidenceActivity struct {
	LoggedWorkoutCount      int `json:"loggedWorkoutCount"`
	WorkingSetSessionCount  int `json:"workingSetSessionCount"`
	WorkingSetLocalDayCount int `json:"workingSetLocalDayCount"`
}

// EvidenceCoverageStatus describes how many working sets can be attributed from current links.
type EvidenceCoverageStatus string

const (
	EvidenceCoverageNotApplicable EvidenceCoverageStatus = "not_applicable"
	EvidenceCoverageUnclassified  EvidenceCoverageStatus = "unclassified"
	EvidenceCoveragePartial       EvidenceCoverageStatus = "partial"
	EvidenceCoverageComplete      EvidenceCoverageStatus = "complete"
)

// EvidenceClassification exposes classified and unclassified working-set coverage.
type EvidenceClassification struct {
	TotalWorkingSetCount        int                    `json:"totalWorkingSetCount"`
	ClassifiedWorkingSetCount   int                    `json:"classifiedWorkingSetCount"`
	UnclassifiedWorkingSetCount int                    `json:"unclassifiedWorkingSetCount"`
	CoverageStatus              EvidenceCoverageStatus `json:"coverageStatus"`
}

// MuscleEvidence describes counts for one catalog muscle label in one recorded role.
type MuscleEvidence struct {
	Muscle                  string `json:"muscle"`
	Role                    string `json:"role"`
	WorkingSetCount         int    `json:"workingSetCount"`
	WorkingSetSessionCount  int    `json:"workingSetSessionCount"`
	WorkingSetLocalDayCount int    `json:"workingSetLocalDayCount"`
}

// EvidenceBasis makes recalculation against current exercise links and catalog data explicit.
type EvidenceBasis struct {
	ClassificationBasis      string `json:"classificationBasis"`
	CalculationPolicyVersion string `json:"calculationPolicyVersion"`
	CatalogVersion           string `json:"catalogVersion"`
}

// EvidenceService validates, retrieves, and assembles descriptive training evidence.
type EvidenceService struct {
	queries EvidenceQuerier
	now     func() time.Time
}

// NewEvidenceService creates a callable service. The injected clock is sampled exactly once
// for each Summarize call; a nil clock uses the server clock.
func NewEvidenceService(queries EvidenceQuerier, now func() time.Time) *EvidenceService {
	if now == nil {
		now = time.Now
	}
	return &EvidenceService{queries: queries, now: now}
}

// Summarize returns current-catalog, descriptive evidence for one explicit local-date period.
func (s *EvidenceService) Summarize(ctx context.Context, request TrainingEvidenceRequest) (*TrainingEvidence, error) {
	observedAt := s.now().UTC()

	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "training evidence", UserID: ""}
	}

	period, location, err := newEvidencePeriod(request, observedAt)
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.ListTrainingEvidence(ctx, db.ListTrainingEvidenceParams{
		UserID:     userID,
		StartAt:    pgTimestamp(period.StartAt),
		EndAt:      pgTimestamp(period.EndAt),
		ObservedAt: pgTimestamp(period.ObservedAt),
	})
	if err != nil {
		return nil, fmt.Errorf("retrieve training evidence: %w", err)
	}

	evidence, err := assembleTrainingEvidence(rows, period, location)
	if err != nil {
		return nil, err
	}
	return evidence, nil
}

func newEvidencePeriod(request TrainingEvidenceRequest, observedAt time.Time) (EvidencePeriod, *time.Location, error) {
	if strings.TrimSpace(request.StartDate) != request.StartDate {
		return EvidencePeriod{}, nil, fmt.Errorf("%w: start date must be YYYY-MM-DD", ErrInvalidTrainingEvidencePeriod)
	}
	startDate, err := time.Parse(evidenceDateLayout, request.StartDate)
	if err != nil || startDate.Format(evidenceDateLayout) != request.StartDate {
		return EvidencePeriod{}, nil, fmt.Errorf("%w: start date must be YYYY-MM-DD", ErrInvalidTrainingEvidencePeriod)
	}

	if request.Timezone == "" || strings.TrimSpace(request.Timezone) != request.Timezone || strings.EqualFold(request.Timezone, "local") {
		return EvidencePeriod{}, nil, fmt.Errorf("%w: timezone must be an explicit IANA timezone", ErrInvalidTrainingEvidencePeriod)
	}
	location, err := time.LoadLocation(request.Timezone)
	if err != nil {
		return EvidencePeriod{}, nil, fmt.Errorf("%w: timezone must be an explicit IANA timezone: %v", ErrInvalidTrainingEvidencePeriod, err)
	}

	startAt, err := localMidnight(startDate, location)
	if err != nil {
		return EvidencePeriod{}, nil, err
	}
	endDate := startDate.AddDate(0, 0, 7)
	endAt, err := localMidnight(endDate, location)
	if err != nil {
		return EvidencePeriod{}, nil, err
	}

	if startAt.After(observedAt) {
		return EvidencePeriod{}, nil, fmt.Errorf("%w: start is after observedAt", ErrFutureTrainingEvidencePeriod)
	}

	return EvidencePeriod{
		StartDate:  request.StartDate,
		EndDate:    endDate.Format(evidenceDateLayout),
		Timezone:   request.Timezone,
		StartAt:    startAt,
		EndAt:      endAt,
		ObservedAt: observedAt,
		Partial:    endAt.After(observedAt),
	}, location, nil
}

func localMidnight(date time.Time, location *time.Location) (time.Time, error) {
	year, month, day := date.Date()
	local := time.Date(year, month, day, 0, 0, 0, 0, location)
	actual := local.In(location)
	if actual.Year() != year || actual.Month() != month || actual.Day() != day || actual.Hour() != 0 || actual.Minute() != 0 || actual.Second() != 0 || actual.Nanosecond() != 0 {
		return time.Time{}, fmt.Errorf("%w: local midnight on %s cannot be represented in %s", ErrInvalidTrainingEvidencePeriod, date.Format(evidenceDateLayout), location.String())
	}
	return local.UTC(), nil
}

func pgTimestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

type muscleRoleKey struct {
	muscle string
	role   string
}

type muscleAccumulator struct {
	workingSetCount int
	sessions        map[int32]struct{}
	localDays       map[string]struct{}
}

func assembleTrainingEvidence(rows []db.ListTrainingEvidenceRow, period EvidencePeriod, location *time.Location) (*TrainingEvidence, error) {
	workouts := make(map[int32]time.Time)
	workingSetsByWorkout := make(map[int32]int)
	muscles := make(map[muscleRoleKey]*muscleAccumulator)
	classification := EvidenceClassification{}

	for _, row := range rows {
		if !row.WorkoutDate.Valid {
			return nil, fmt.Errorf("%w: workout %d has no timestamp", ErrInvalidTrainingEvidenceData, row.WorkoutID)
		}
		workoutDate := row.WorkoutDate.Time.UTC()
		if previous, exists := workouts[row.WorkoutID]; exists && !previous.Equal(workoutDate) {
			return nil, fmt.Errorf("%w: workout %d has conflicting timestamps", ErrInvalidTrainingEvidenceData, row.WorkoutID)
		}
		workouts[row.WorkoutID] = workoutDate

		if row.WorkingSetCount < 0 {
			return nil, fmt.Errorf("%w: workout %d has a negative working-set count", ErrInvalidTrainingEvidenceData, row.WorkoutID)
		}
		if row.CatalogID.Valid && !row.ExerciseID.Valid {
			return nil, fmt.Errorf("%w: catalog link without an exercise", ErrInvalidTrainingEvidenceData)
		}

		var catalog *exercisecatalog.Entry
		if row.CatalogID.Valid {
			catalog = exercisecatalog.Find(row.CatalogID.String)
			if catalog == nil {
				return nil, fmt.Errorf("%w: catalog ID %q is not embedded", ErrInconsistentExerciseCatalog, row.CatalogID.String)
			}
		}

		workingSetCount := int(row.WorkingSetCount)
		if workingSetCount == 0 {
			continue
		}
		if !row.ExerciseID.Valid {
			return nil, fmt.Errorf("%w: working set without an owned exercise", ErrInvalidTrainingEvidenceData)
		}

		workingSetsByWorkout[row.WorkoutID] += workingSetCount
		classification.TotalWorkingSetCount += workingSetCount
		if catalog == nil {
			classification.UnclassifiedWorkingSetCount += workingSetCount
			continue
		}

		classification.ClassifiedWorkingSetCount += workingSetCount
		localDay := workoutDate.In(location).Format(evidenceDateLayout)
		addCatalogRoleEvidence(muscles, catalog.PrimaryMuscles, "primary", workingSetCount, row.WorkoutID, localDay)
		addCatalogRoleEvidence(muscles, catalog.SecondaryMuscles, "secondary", workingSetCount, row.WorkoutID, localDay)
	}

	workingSetDays := make(map[string]struct{})
	for workoutID := range workingSetsByWorkout {
		workingSetDays[workouts[workoutID].In(location).Format(evidenceDateLayout)] = struct{}{}
	}

	classification.CoverageStatus = coverageStatus(classification.TotalWorkingSetCount, classification.ClassifiedWorkingSetCount)
	return &TrainingEvidence{
		Period: period,
		Activity: EvidenceActivity{
			LoggedWorkoutCount:      len(workouts),
			WorkingSetSessionCount:  len(workingSetsByWorkout),
			WorkingSetLocalDayCount: len(workingSetDays),
		},
		Classification: classification,
		Muscles:        sortedMuscleEvidence(muscles),
		Basis: EvidenceBasis{
			ClassificationBasis:      currentClassificationBasis,
			CalculationPolicyVersion: TrainingEvidenceCalculationPolicyVersion,
			CatalogVersion:           exercisecatalog.Version(),
		},
		Limitations: append([]string(nil),
			"Muscle roles are descriptive upstream metadata, not validated stimulus.",
			"Unclassified working sets are unknown muscle training, not zero training for every muscle.",
			"Classified coverage describes current catalog links in logged data and is not complete real-world logging.",
			"This summary cannot infer progress, plateau, or target attainment.",
			"An empty secondary role means no secondary muscles are listed in the catalog; it does not mean no secondary muscles participate.",
			"Counts across muscle labels and roles are not additive because one working set can appear in multiple labels.",
			"Zero logged workouts means no workouts were logged in this period; it does not prove no training occurred.",
		),
	}, nil
}

func coverageStatus(totalWorkingSets, classifiedWorkingSets int) EvidenceCoverageStatus {
	switch {
	case totalWorkingSets == 0:
		return EvidenceCoverageNotApplicable
	case classifiedWorkingSets == 0:
		return EvidenceCoverageUnclassified
	case classifiedWorkingSets == totalWorkingSets:
		return EvidenceCoverageComplete
	default:
		return EvidenceCoveragePartial
	}
}

func addCatalogRoleEvidence(accumulators map[muscleRoleKey]*muscleAccumulator, listedMuscles []string, role string, workingSetCount int, workoutID int32, localDay string) {
	seen := make(map[string]struct{}, len(listedMuscles))
	for _, muscle := range listedMuscles {
		if _, duplicate := seen[muscle]; duplicate {
			continue
		}
		seen[muscle] = struct{}{}

		key := muscleRoleKey{muscle: muscle, role: role}
		accumulator := accumulators[key]
		if accumulator == nil {
			accumulator = &muscleAccumulator{
				sessions:  make(map[int32]struct{}),
				localDays: make(map[string]struct{}),
			}
			accumulators[key] = accumulator
		}
		accumulator.workingSetCount += workingSetCount
		accumulator.sessions[workoutID] = struct{}{}
		accumulator.localDays[localDay] = struct{}{}
	}
}

func sortedMuscleEvidence(accumulators map[muscleRoleKey]*muscleAccumulator) []MuscleEvidence {
	result := make([]MuscleEvidence, 0, len(accumulators))
	for key, accumulator := range accumulators {
		result = append(result, MuscleEvidence{
			Muscle:                  key.muscle,
			Role:                    key.role,
			WorkingSetCount:         accumulator.workingSetCount,
			WorkingSetSessionCount:  len(accumulator.sessions),
			WorkingSetLocalDayCount: len(accumulator.localDays),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		leftRole, rightRole := muscleRoleOrder(result[i].Role), muscleRoleOrder(result[j].Role)
		if leftRole != rightRole {
			return leftRole < rightRole
		}
		return result[i].Muscle < result[j].Muscle
	})
	return result
}

func muscleRoleOrder(role string) int {
	switch role {
	case "primary":
		return 0
	case "secondary":
		return 1
	default:
		return 2
	}
}

var _ EvidenceQuerier = (*db.Queries)(nil)
