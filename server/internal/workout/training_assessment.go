package workout

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
)

const TrainingAssessmentPolicyVersion = "strength-session-reference-v1"

type AssessmentQuerier interface {
	GetExercise(context.Context, db.GetExerciseParams) (db.GetExerciseRow, error)
	ListExerciseAssessmentSets(context.Context, db.ListExerciseAssessmentSetsParams) ([]db.ListExerciseAssessmentSetsRow, error)
}

type AssessmentSession struct {
	WorkoutID          int32     `json:"workoutId"`
	Date               time.Time `json:"date"`
	WorkingSets        int       `json:"workingSets"`
	InvalidWorkingSets int       `json:"invalidWorkingSets"`
	ReasonCode         string    `json:"reasonCode"`
	State              string    `json:"state"`
	Reason             string    `json:"reason"`
}

type StrengthAssessment struct {
	ExerciseID               int32               `json:"exerciseId"`
	Period                   EvidencePeriod      `json:"period"`
	Sessions                 []AssessmentSession `json:"sessions"`
	WorkingSets              int                 `json:"workingSets"`
	TrainingDays             int                 `json:"trainingDays"`
	SessionsMeetingReference int                 `json:"sessionsMeetingReference"`
	PolicyVersion            string              `json:"policyVersion"`
	Reference                string              `json:"reference"`
	SourceURL                string              `json:"sourceUrl"`
	Applicability            string              `json:"applicability"`
	Limitations              []string            `json:"limitations"`
}

type AssessmentService struct {
	queries AssessmentQuerier
	now     func() time.Time
}

func NewAssessmentService(queries AssessmentQuerier, now func() time.Time) *AssessmentService {
	if now == nil {
		now = time.Now
	}
	return &AssessmentService{queries: queries, now: now}
}

// Assess compares saved session set counts with a general reference. It never
// turns a weekly average or an estimated 1RM into an overall training verdict.
func (s *AssessmentService) Assess(ctx context.Context, exerciseID int32, request TrainingEvidenceRequest) (*StrengthAssessment, error) {
	owner, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "training assessment"}
	}
	period, location, err := newEvidencePeriod(request, s.now().UTC())
	if err != nil {
		return nil, err
	}
	_, err = s.queries.GetExercise(ctx, db.GetExerciseParams{ID: exerciseID, UserID: owner})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprint(exerciseID)}
	}
	if err != nil {
		return nil, fmt.Errorf("load assessment exercise: %w", err)
	}
	rows, err := s.queries.ListExerciseAssessmentSets(ctx, db.ListExerciseAssessmentSetsParams{UserID: owner, ExerciseID: exerciseID, StartAt: pgTimestamp(period.StartAt), EndAt: pgTimestamp(period.EndAt), ObservedAt: pgTimestamp(period.ObservedAt)})
	if err != nil {
		return nil, fmt.Errorf("load assessment sets: %w", err)
	}
	result := &StrengthAssessment{ExerciseID: exerciseID, Period: period, Sessions: []AssessmentSession{}, PolicyVersion: TrainingAssessmentPolicyVersion,
		Reference: "2–3 working sets per exercise per session", SourceURL: "https://acsm.org/resistance-training-guidelines-update-2026/",
		Applicability: "General reference for healthy adults. Personal applicability has not been verified.",
		Limitations:   []string{"This compares logged set counts, not effort, training quality, or expected strength gains.", "Two sets is the lower reference value, not a personal minimum. More than three is not classified as excessive.", "Training outside FitTrack and intentional deloads are unknown.", "Session Metrics loading may use an estimated or historical maximum; it is informational, not an 80% pass/fail test."},
	}
	days := map[string]bool{}
	indices := map[int32]int{}
	invalid := map[int32]bool{}
	for _, row := range rows {
		index, exists := indices[row.WorkoutID]
		if !exists {
			if !row.WorkoutDate.Valid {
				return nil, ErrInvalidTrainingEvidenceData
			}
			index = len(result.Sessions)
			indices[row.WorkoutID] = index
			result.Sessions = append(result.Sessions, AssessmentSession{WorkoutID: row.WorkoutID, Date: row.WorkoutDate.Time})
		}
		if row.SetType != "working" {
			continue
		}
		session := &result.Sessions[index]
		weight, err := row.Weight.Float64Value()
		if err != nil {
			return nil, fmt.Errorf("read assessment weight: %w", err)
		}
		if row.Reps <= 0 || (weight.Valid && (weight.Float64 < 0 || math.IsNaN(weight.Float64) || math.IsInf(weight.Float64, 0))) {
			invalid[row.WorkoutID] = true
			session.InvalidWorkingSets++
			continue
		}
		session.WorkingSets++
		result.WorkingSets++
		days[row.WorkoutDate.Time.In(location).Format(evidenceDateLayout)] = true
	}
	result.TrainingDays = len(days)
	for i := range result.Sessions {
		session := &result.Sessions[i]
		switch {
		case period.Partial:
			session.ReasonCode = "period_in_progress"
			session.State, session.Reason = "cannot_assess", "Week in progress. Counts so far only."
		case invalid[session.WorkoutID]:
			session.ReasonCode = "invalid_set_data"
			session.State, session.Reason = "cannot_assess", "Check invalid working-set entries."
		case session.WorkingSets == 0:
			session.ReasonCode = "no_working_sets"
			session.State, session.Reason = "cannot_assess", "No working sets logged."
		case session.WorkingSets < 2:
			session.ReasonCode = "below_lower_reference"
			session.State, session.Reason = "below_reference", "Fewer than the lower reference of 2 working sets."
		default:
			session.ReasonCode = "at_least_lower_reference"
			session.State, session.Reason = "reference_met", "At least the lower reference of 2 working sets logged."
			result.SessionsMeetingReference++
		}
	}
	return result, nil
}
