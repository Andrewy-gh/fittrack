package workout

import (
	"context"
	"fmt"
	"math"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/exercisecatalog"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

const MuscleMappingVersion = "muscle-contributions-v1"

// Roles are reviewed exercise/outcome pairs, not transformations of catalog labels.
// Unlisted pairs remain unresolved, including other roles of a listed movement.
var reviewedMuscleRoles = map[string]map[string]string{
	"Barbell_Bench_Press_-_Medium_Grip": {"chest": "direct", "triceps": "indirect"},
	"Triceps_Pushdown":                  {"triceps": "direct"},
	"Barbell_Curl":                      {"elbow_flexors": "direct"},
	"Hammer_Curls":                      {"elbow_flexors": "direct"},
	"Bent_Over_Barbell_Row":             {"elbow_flexors": "indirect"},
	"Seated_Cable_Rows":                 {"elbow_flexors": "indirect"},
	"Wide-Grip_Lat_Pulldown":            {"elbow_flexors": "indirect"},
	"Barbell_Full_Squat":                {"quadriceps": "direct"},
	"Leg_Press":                         {"quadriceps": "direct"},
	"Dumbbell_Lunges":                   {"quadriceps": "direct"},
	"Leg_Extensions":                    {"quadriceps": "direct"},
	"Lying_Leg_Curls":                   {"hamstrings": "direct"},
	"Incline_Dumbbell_Press":            {"triceps": "indirect"},
	"Dumbbell_Shoulder_Press":           {"triceps": "indirect"},
}

type MuscleContributionQuerier interface {
	ListMuscleContributionSets(context.Context, db.ListMuscleContributionSetsParams) ([]db.ListMuscleContributionSetsRow, error)
}
type MuscleContribution struct {
	Muscle                string `json:"muscle"`
	DirectSets            int    `json:"directSets"`
	IndirectSets          int    `json:"indirectSets"`
	UnresolvedSets        int    `json:"unresolvedSets"`
	InvalidMappedSets     int    `json:"invalidMappedSets"`
	InvalidUnresolvedSets int    `json:"invalidUnresolvedSets"`
}
type ContributionExercise struct {
	ExerciseID         int32             `json:"exerciseId"`
	Name               string            `json:"name"`
	CatalogID          string            `json:"catalogId"`
	WorkingSets        int               `json:"workingSets"`
	InvalidWorkingSets int               `json:"invalidWorkingSets"`
	Roles              map[string]string `json:"roles"`
}
type MuscleContributions struct {
	Period              EvidencePeriod         `json:"period"`
	Muscles             []MuscleContribution   `json:"muscles"`
	Exercises           []ContributionExercise `json:"exercises"`
	WorkingSets         int                    `json:"workingSets"`
	InvalidWorkingSets  int                    `json:"invalidWorkingSets"`
	UnclassifiedSets    int                    `json:"unclassifiedSets"`
	UnreviewedSets      int                    `json:"unreviewedSets"`
	MappingVersion      string                 `json:"mappingVersion"`
	CatalogVersion      string                 `json:"catalogVersion"`
	ClassificationBasis string                 `json:"classificationBasis"`
}
type MuscleContributionService struct {
	queries MuscleContributionQuerier
	now     func() time.Time
}

func NewMuscleContributionService(queries MuscleContributionQuerier, now func() time.Time) *MuscleContributionService {
	if now == nil {
		now = time.Now
	}
	return &MuscleContributionService{queries, now}
}
func (s *MuscleContributionService) Summarize(ctx context.Context, request TrainingEvidenceRequest) (*MuscleContributions, error) {
	owner, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "muscle contributions"}
	}
	period, _, err := newEvidencePeriod(request, s.now().UTC())
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListMuscleContributionSets(ctx, db.ListMuscleContributionSetsParams{UserID: owner, StartAt: pgTimestamp(period.StartAt), EndAt: pgTimestamp(period.EndAt), ObservedAt: pgTimestamp(period.ObservedAt)})
	if err != nil {
		return nil, fmt.Errorf("load muscle contributions: %w", err)
	}
	result := &MuscleContributions{Period: period, Exercises: []ContributionExercise{}, MappingVersion: MuscleMappingVersion, CatalogVersion: exercisecatalog.Version(), ClassificationBasis: currentClassificationBasis}
	for _, muscle := range []string{"chest", "triceps", "elbow_flexors", "quadriceps", "hamstrings"} {
		result.Muscles = append(result.Muscles, MuscleContribution{Muscle: muscle})
	}
	indices := map[int32]int{}
	for _, row := range rows {
		if row.SetType != "working" {
			continue
		}
		if !row.WorkoutDate.Valid {
			return nil, ErrInvalidTrainingEvidenceData
		}
		if row.CatalogID.Valid && exercisecatalog.Find(row.CatalogID.String) == nil {
			return nil, ErrInconsistentExerciseCatalog
		}
		roles := reviewedMuscleRoles[row.CatalogID.String]
		index, exists := indices[row.ExerciseID]
		if !exists {
			index = len(result.Exercises)
			indices[row.ExerciseID] = index
			copied := map[string]string{}
			for muscle, role := range roles {
				copied[muscle] = role
			}
			result.Exercises = append(result.Exercises, ContributionExercise{ExerciseID: row.ExerciseID, Name: row.ExerciseName, CatalogID: row.CatalogID.String, Roles: copied})
		}
		exercise := &result.Exercises[index]
		weight, err := row.Weight.Float64Value()
		if err != nil {
			return nil, fmt.Errorf("read contribution weight: %w", err)
		}
		invalid := row.Reps <= 0 || (weight.Valid && (weight.Float64 < 0 || math.IsNaN(weight.Float64) || math.IsInf(weight.Float64, 0)))
		if invalid {
			result.InvalidWorkingSets++
			exercise.InvalidWorkingSets++
		} else {
			result.WorkingSets++
			exercise.WorkingSets++
			if !row.CatalogID.Valid {
				result.UnclassifiedSets++
			} else if len(roles) == 0 {
				result.UnreviewedSets++
			}
		}
		for i := range result.Muscles {
			muscle := &result.Muscles[i]
			role := roles[muscle.Muscle]
			switch {
			case invalid && role != "":
				muscle.InvalidMappedSets++
			case invalid:
				muscle.InvalidUnresolvedSets++
			case role == "direct":
				muscle.DirectSets++
			case role == "indirect":
				muscle.IndirectSets++
			default:
				muscle.UnresolvedSets++
			}
		}
	}
	return result, nil
}
