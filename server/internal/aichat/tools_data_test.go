package aichat

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

func TestResolveExerciseNameForChat(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		matches        []string
		wantResolved   string
		wantCandidates []string
		wantOK         bool
	}{
		{name: "exact fold wins", query: "bench press", matches: []string{"Incline Bench Press", "Bench Press"}, wantResolved: "Bench Press", wantOK: true},
		{name: "single partial auto selected", query: "dead", matches: []string{"Deadlift"}, wantResolved: "Deadlift", wantOK: true},
		{name: "ambiguous returns candidates", query: "row", matches: []string{"Barbell Row", "Cable Row"}, wantCandidates: []string{"Barbell Row", "Cable Row"}},
		{name: "none", query: "curl", matches: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, candidates, ok := resolveExerciseNameForChat(tt.query, tt.matches)
			if resolved != tt.wantResolved || ok != tt.wantOK || !reflect.DeepEqual(candidates, tt.wantCandidates) {
				t.Fatalf("resolveExerciseNameForChat() = %q, %#v, %v; want %q, %#v, %v", resolved, candidates, ok, tt.wantResolved, tt.wantCandidates, tt.wantOK)
			}
		})
	}
}

func TestParseChatToolDateForgiving(t *testing.T) {
	start, note := parseChatToolDate("2026-07-01", false)
	if note != "" || start == nil || start.Format(time.RFC3339Nano) != "2026-07-01T00:00:00Z" {
		t.Fatalf("start parse = %v note %q", start, note)
	}

	end, note := parseChatToolDate("July 3, 2026", true)
	if note != "" || end == nil || end.Format(time.RFC3339Nano) != "2026-07-03T23:59:59.999999999Z" {
		t.Fatalf("end parse = %v note %q", end, note)
	}

	invalid, note := parseChatToolDate("last someday-ish", false)
	if invalid != nil || !strings.Contains(note, `Ignored invalid date "last someday-ish".`) {
		t.Fatalf("invalid parse = %v note %q", invalid, note)
	}
}

func TestRunGetWorkoutsToolReturnsCandidatesWithoutFetchingWorkouts(t *testing.T) {
	reader := &stubChatDataReader{names: []string{"Barbell Row", "Cable Row"}}
	ctx := user.WithContext(context.Background(), "user-1")

	result := runGetWorkoutsTool(ctx, reader, GetWorkoutsToolInput{ExerciseName: "row"})

	if reader.listCalls != 0 {
		t.Fatalf("ListWorkoutsWithSets called %d times, want 0 for ambiguous exercise", reader.listCalls)
	}
	if !reflect.DeepEqual(result.CandidateExercises, []string{"Barbell Row", "Cable Row"}) {
		t.Fatalf("CandidateExercises = %#v", result.CandidateExercises)
	}
	if !strings.Contains(result.Message, "multiple matching exercises") {
		t.Fatalf("Message = %q, want ambiguity guidance", result.Message)
	}
}

func TestRunGetWorkoutsToolIgnoresBadDateAndReturnsReaderErrorAsMessage(t *testing.T) {
	reader := &stubChatDataReader{listErr: errors.New("database unavailable")}
	ctx := user.WithContext(context.Background(), "user-1")

	result := runGetWorkoutsTool(ctx, reader, GetWorkoutsToolInput{StartDate: "not-a-date"})

	if !strings.Contains(result.Message, `Ignored invalid date "not-a-date".`) {
		t.Fatalf("Message = %q, want invalid-date note", result.Message)
	}
	if !strings.Contains(result.Message, "couldn't read workout history") {
		t.Fatalf("Message = %q, want reader error message", result.Message)
	}
}

func TestRunGetWorkoutsToolEmptyResultMentionsLastWorkout(t *testing.T) {
	reader := &stubChatDataReader{snapshot: &TrainingSnapshot{LastWorkoutDate: "2026-07-03"}}
	ctx := user.WithContext(context.Background(), "user-1")

	result := runGetWorkoutsTool(ctx, reader, GetWorkoutsToolInput{LastN: 5})

	if !strings.Contains(result.Message, "No workouts found for that filter.") || !strings.Contains(result.Message, "2026-07-03") {
		t.Fatalf("Message = %q", result.Message)
	}
}

func TestRunGetExerciseStatsToolUsesResolvedExercise(t *testing.T) {
	reader := &stubChatDataReader{
		names: []string{"Bench Press"},
		stats: &ExerciseStatsView{
			ExerciseName:    "Bench Press",
			Window:          "3m",
			LastSessionDate: "2026-07-03",
			LastSessionSets: []string{"195x3 working"},
			SessionCount:    3,
		},
	}
	ctx := user.WithContext(context.Background(), "user-1")

	result := runGetExerciseStatsTool(ctx, reader, GetExerciseStatsToolInput{ExerciseName: "bench", Window: "3m"})

	if result.Stats == nil || result.Stats.ExerciseName != "Bench Press" {
		t.Fatalf("Stats = %#v, want resolved bench stats", result.Stats)
	}
	if reader.statsExerciseName != "Bench Press" {
		t.Fatalf("ExerciseStats exercise name = %q, want resolved exact name", reader.statsExerciseName)
	}
}

func TestRunGetExerciseStatsToolReturnsAmbiguousCandidates(t *testing.T) {
	reader := &stubChatDataReader{names: []string{"Barbell Row", "Cable Row"}}
	ctx := user.WithContext(context.Background(), "user-1")

	result := runGetExerciseStatsTool(ctx, reader, GetExerciseStatsToolInput{ExerciseName: "row"})

	if !reflect.DeepEqual(result.CandidateExercises, []string{"Barbell Row", "Cable Row"}) {
		t.Fatalf("CandidateExercises = %#v", result.CandidateExercises)
	}
	if !strings.Contains(result.Message, "multiple matching exercises") {
		t.Fatalf("Message = %q, want ambiguity guidance", result.Message)
	}
}

func TestRunUpdateTrainingProfileToolNormalizesAndSaves(t *testing.T) {
	reader := &stubChatDataReader{}
	ctx := user.WithContext(contextWithTrainingProfileSource(context.Background(), 12, 34), "user-1")
	duration := int32(45)
	goal := "muscle growth"
	location := "home gym"
	equipment := []string{" adjustable dumbbells ", "bench", "bench"}

	result := runUpdateTrainingProfileTool(ctx, reader, UpdateTrainingProfileToolInput{
		PrimaryGoal:                     &goal,
		PreferredSessionDurationMinutes: &duration,
		UsualTrainingLocation:           &location,
		AvailableEquipment:              &equipment,
	})

	if result.Profile == nil || result.Profile.PrimaryGoal != "hypertrophy" {
		t.Fatalf("Profile = %#v, want normalized hypertrophy profile", result.Profile)
	}
	if reader.profileUpdate.SourceConversationID == nil || *reader.profileUpdate.SourceConversationID != 12 {
		t.Fatalf("SourceConversationID = %#v, want 12", reader.profileUpdate.SourceConversationID)
	}
	if reader.profileUpdate.SourceMessageID == nil || *reader.profileUpdate.SourceMessageID != 34 {
		t.Fatalf("SourceMessageID = %#v, want 34", reader.profileUpdate.SourceMessageID)
	}
	if reader.profileUpdate.AvailableEquipment == nil || !reflect.DeepEqual(*reader.profileUpdate.AvailableEquipment, []string{"adjustable dumbbells", "bench"}) {
		t.Fatalf("AvailableEquipment update = %#v", reader.profileUpdate.AvailableEquipment)
	}
	if !strings.Contains(result.Message, "updated") {
		t.Fatalf("Message = %q, want update confirmation", result.Message)
	}
}

func TestRunUpdateTrainingProfileToolCanClearFields(t *testing.T) {
	reader := &stubChatDataReader{}
	ctx := user.WithContext(context.Background(), "user-1")
	goal := ""
	duration := int32(0)
	limitations := []string{}

	result := runUpdateTrainingProfileTool(ctx, reader, UpdateTrainingProfileToolInput{
		PrimaryGoal:                     &goal,
		PreferredSessionDurationMinutes: &duration,
		MovementLimitations:             &limitations,
	})

	if result.Profile == nil {
		t.Fatalf("Profile = nil, want saved cleared profile")
	}
	if reader.profileUpdate.PrimaryGoal == nil || *reader.profileUpdate.PrimaryGoal != "" {
		t.Fatalf("PrimaryGoal update = %#v, want clear", reader.profileUpdate.PrimaryGoal)
	}
	if reader.profileUpdate.PreferredSessionDurationMinutes == nil || *reader.profileUpdate.PreferredSessionDurationMinutes != 0 {
		t.Fatalf("PreferredSessionDurationMinutes update = %#v, want clear", reader.profileUpdate.PreferredSessionDurationMinutes)
	}
	if reader.profileUpdate.MovementLimitations == nil || len(*reader.profileUpdate.MovementLimitations) != 0 {
		t.Fatalf("MovementLimitations update = %#v, want empty replacement", reader.profileUpdate.MovementLimitations)
	}
}

type stubChatDataReader struct {
	names             []string
	workouts          []ChatWorkoutView
	snapshot          *TrainingSnapshot
	stats             *ExerciseStatsView
	profile           *TrainingProfile
	profileUpdate     TrainingProfileUpdate
	listErr           error
	listCalls         int
	statsExerciseName string
}

func (s *stubChatDataReader) ListWorkoutsWithSets(ctx context.Context, userID string, filter WorkoutHistoryFilter) ([]ChatWorkoutView, error) {
	_ = ctx
	_ = userID
	_ = filter
	s.listCalls++
	return s.workouts, s.listErr
}

func (s *stubChatDataReader) ResolveExerciseNames(ctx context.Context, userID string, query string) ([]string, error) {
	_ = ctx
	_ = userID
	_ = query
	return s.names, nil
}

func (s *stubChatDataReader) TrainingSnapshot(ctx context.Context, userID string) (*TrainingSnapshot, error) {
	_ = ctx
	_ = userID
	return s.snapshot, nil
}

func (s *stubChatDataReader) TrainingProfile(ctx context.Context, userID string) (*TrainingProfile, error) {
	_ = ctx
	_ = userID
	return s.profile, nil
}

func (s *stubChatDataReader) UpdateTrainingProfile(ctx context.Context, userID string, update TrainingProfileUpdate) (*TrainingProfile, error) {
	_ = ctx
	_ = userID
	s.profileUpdate = update
	profile := &TrainingProfile{}
	if update.PrimaryGoal != nil {
		profile.PrimaryGoal = *update.PrimaryGoal
	}
	if update.ExperienceLevel != nil {
		profile.ExperienceLevel = *update.ExperienceLevel
	}
	if update.PreferredSessionDurationMinutes != nil {
		profile.PreferredSessionDurationMinutes = *update.PreferredSessionDurationMinutes
	}
	if update.UsualTrainingLocation != nil {
		profile.UsualTrainingLocation = *update.UsualTrainingLocation
	}
	if update.AvailableEquipment != nil {
		profile.AvailableEquipment = *update.AvailableEquipment
	}
	if update.AvoidedExercises != nil {
		profile.AvoidedExercises = *update.AvoidedExercises
	}
	if update.MovementLimitations != nil {
		profile.MovementLimitations = *update.MovementLimitations
	}
	return profile, nil
}

func (s *stubChatDataReader) ExerciseStats(ctx context.Context, userID string, exerciseName string, window string) (*ExerciseStatsView, error) {
	_ = ctx
	_ = userID
	_ = window
	s.statsExerciseName = exerciseName
	return s.stats, nil
}
