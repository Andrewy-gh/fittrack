package recommendation_test

import (
	"math"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/stretchr/testify/require"
)

var characterizationNow = time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

func float64Ptr(value float64) *float64 {
	return &value
}

func workingSets(reps int, weight *float64, count int) []recommendation.WorkingSet {
	sets := make([]recommendation.WorkingSet, count)
	for i := range sets {
		sets[i].Reps = reps
		if weight != nil {
			value := *weight
			sets[i].Weight = &value
		}
	}
	return sets
}

func TestRecommendTrainingCharacterizesReadiness(t *testing.T) {
	weight := float64Ptr(100)
	context := recommendation.TrainingContext{Sessions: []recommendation.Session{{
		Date: characterizationNow.Add(-time.Hour),
		Sets: workingSets(8, weight, 2),
	}}}
	cases := []struct {
		name            string
		readiness       string
		wantPlan        bool
		wantSets        int
		wantReps        int
		wantWeight      *float64
		wantExplanation string
	}{
		{
			name:            "sluggish",
			readiness:       "sluggish",
			wantPlan:        true,
			wantSets:        1,
			wantReps:        6,
			wantWeight:      float64Ptr(100),
			wantExplanation: "An easier day: fewer reps and, where possible, one fewer set.",
		},
		{
			name:            "normal",
			readiness:       "normal",
			wantPlan:        true,
			wantSets:        2,
			wantReps:        8,
			wantWeight:      float64Ptr(100),
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name:            "great",
			readiness:       "great",
			wantPlan:        true,
			wantSets:        2,
			wantReps:        9,
			wantWeight:      float64Ptr(100),
			wantExplanation: "Try one extra rep at the same weight.",
		},
		{
			name:            "invalid value",
			readiness:       "excellent",
			wantExplanation: "Choose how you feel today to get a suggestion.",
		},
		{
			name:            "empty value",
			readiness:       "",
			wantExplanation: "Choose how you feel today to get a suggestion.",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := recommendation.RecommendTraining(characterizationNow, context, tt.readiness)

			require.Equal(t, tt.readiness, result.Readiness)
			require.Equal(t, recommendation.PolicyVersion, result.PolicyVersion)
			require.Equal(t, tt.wantExplanation, result.Explanation)
			if !tt.wantPlan {
				require.Nil(t, result.Plan)
				return
			}

			require.NotNil(t, result.Plan)
			require.Equal(t, tt.wantSets, result.Plan.Sets)
			require.Equal(t, tt.wantReps, result.Plan.Reps)
			require.NotNil(t, result.Plan.Weight)
			require.Equal(t, *tt.wantWeight, *result.Plan.Weight)
		})
	}
}

func TestRecommendTrainingCharacterizesRepBoundaries(t *testing.T) {
	cases := []struct {
		name            string
		reps            int
		wantPlan        bool
		wantReps        int
		wantExplanation string
	}{
		{
			name:            "zero reps reject the plan",
			reps:            0,
			wantExplanation: "Log a working set with valid reps to get a suggestion.",
		},
		{
			name:            "one rep is preserved",
			reps:            1,
			wantPlan:        true,
			wantReps:        1,
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name:            "one hundred reps is preserved",
			reps:            100,
			wantPlan:        true,
			wantReps:        100,
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name:            "one hundred one reps reject the plan",
			reps:            101,
			wantExplanation: "Log a working set with valid reps to get a suggestion.",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := recommendation.RecommendTraining(characterizationNow, recommendation.TrainingContext{
				Sessions: []recommendation.Session{{
					Date: characterizationNow.Add(-time.Hour),
					Sets: workingSets(tt.reps, float64Ptr(100), 2),
				}},
			}, "normal")

			require.Equal(t, "normal", result.Readiness)
			require.Equal(t, tt.wantExplanation, result.Explanation)
			if !tt.wantPlan {
				require.Nil(t, result.Plan)
				return
			}

			require.NotNil(t, result.Plan)
			require.Equal(t, 2, result.Plan.Sets)
			require.Equal(t, tt.wantReps, result.Plan.Reps)
			require.NotNil(t, result.Plan.Weight)
			require.Equal(t, 100.0, *result.Plan.Weight)
		})
	}
}

func TestRecommendTrainingDoesNotPanicOnMalformedWeights(t *testing.T) {
	cases := []struct {
		name            string
		sets            []recommendation.WorkingSet
		wantWeight      *float64
		wantExplanation string
	}{
		{
			name:            "nil weight",
			sets:            workingSets(8, nil, 2),
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
		{
			name:            "NaN weight",
			sets:            workingSets(8, float64Ptr(math.NaN()), 2),
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
		{
			name:            "positive infinity weight",
			sets:            workingSets(8, float64Ptr(math.Inf(1)), 2),
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
		{
			name:            "negative infinity weight",
			sets:            workingSets(8, float64Ptr(math.Inf(-1)), 2),
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
		{
			name:            "negative weight",
			sets:            workingSets(8, float64Ptr(-1), 2),
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
		{
			name: "nonuniform weights",
			sets: []recommendation.WorkingSet{
				{Weight: float64Ptr(100), Reps: 8},
				{Weight: float64Ptr(110), Reps: 8},
			},
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
		{
			name:            "valid lower boundary",
			sets:            workingSets(8, float64Ptr(0), 2),
			wantWeight:      float64Ptr(0),
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name:            "valid upper boundary",
			sets:            workingSets(8, float64Ptr(10000), 2),
			wantWeight:      float64Ptr(10000),
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name:            "out of range below minimum",
			sets:            workingSets(8, float64Ptr(-0.1), 2),
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
		{
			name:            "out of range above maximum",
			sets:            workingSets(8, float64Ptr(10000.1), 2),
			wantExplanation: "Your last weights varied or weren’t recorded. Choose the weight for these reps.",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var result recommendation.Result
			require.NotPanics(t, func() {
				result = recommendation.RecommendTraining(characterizationNow, recommendation.TrainingContext{
					Sessions: []recommendation.Session{{
						Date: characterizationNow.Add(-time.Hour),
						Sets: tt.sets,
					}},
				}, "normal")
			})

			require.Equal(t, "normal", result.Readiness)
			require.NotNil(t, result.Plan)
			require.Equal(t, 2, result.Plan.Sets)
			require.Equal(t, 8, result.Plan.Reps)
			require.Equal(t, tt.wantExplanation, result.Explanation)
			if tt.wantWeight == nil {
				require.Nil(t, result.Plan.Weight)
				return
			}

			require.NotNil(t, result.Plan.Weight)
			require.Equal(t, *tt.wantWeight, *result.Plan.Weight)
		})
	}
}

func TestRecommendTrainingCharacterizesProgressionWeightBoundaries(t *testing.T) {
	cases := []struct {
		name            string
		weight          float64
		wantWeight      float64
		wantReps        int
		wantExplanation string
	}{
		{
			name:            "just below 50 does not progress",
			weight:          49.9,
			wantWeight:      49.9,
			wantReps:        12,
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name:            "50 progresses",
			weight:          50,
			wantWeight:      52.5,
			wantReps:        8,
			wantExplanation: "You reached the rep target in two workouts. Try a small weight increase, using the closest available weight.",
		},
		{
			name:            "2000 progresses",
			weight:          2000,
			wantWeight:      2002.5,
			wantReps:        8,
			wantExplanation: "You reached the rep target in two workouts. Try a small weight increase, using the closest available weight.",
		},
		{
			name:            "just above 2000 does not progress",
			weight:          2000.1,
			wantWeight:      2000.1,
			wantReps:        12,
			wantExplanation: "Repeat your last working sets.",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := recommendation.RecommendTraining(characterizationNow, recommendation.TrainingContext{
				Sessions: []recommendation.Session{
					{
						Date: characterizationNow.Add(-time.Hour),
						Sets: workingSets(12, float64Ptr(tt.weight), 2),
					},
					{
						Date: characterizationNow.Add(-2 * time.Hour),
						Sets: workingSets(12, float64Ptr(tt.weight), 2),
					},
				},
			}, "normal")

			require.Equal(t, "normal", result.Readiness)
			require.NotNil(t, result.Plan)
			require.Equal(t, 2, result.Plan.Sets)
			require.Equal(t, tt.wantReps, result.Plan.Reps)
			require.NotNil(t, result.Plan.Weight)
			require.Equal(t, tt.wantWeight, *result.Plan.Weight)
			require.Equal(t, tt.wantExplanation, result.Explanation)
		})
	}
}

func TestRecommendTrainingCharacterizesSessionSelection(t *testing.T) {
	weight := float64Ptr(100)
	recent := recommendation.Session{
		Date: characterizationNow.Add(-time.Hour),
		Sets: workingSets(8, weight, 2),
	}
	cases := []struct {
		name            string
		sessions        []recommendation.Session
		wantSets        int
		wantReps        int
		wantWeight      *float64
		wantExplanation string
	}{
		{
			name: "stale first session is ignored even with a recent later entry",
			sessions: []recommendation.Session{
				{
					Date: characterizationNow.Add(-recommendation.Recency - time.Second),
					Sets: workingSets(8, weight, 2),
				},
				recent,
			},
			wantSets:        2,
			wantReps:        8,
			wantExplanation: "Choose a comfortable starting weight; we don’t have recent working sets to use.",
		},
		{
			name: "future first session is ignored even with a past later entry",
			sessions: []recommendation.Session{
				{
					Date: characterizationNow.Add(time.Second),
					Sets: workingSets(8, weight, 2),
				},
				recent,
			},
			wantSets:        2,
			wantReps:        8,
			wantExplanation: "Choose a comfortable starting weight; we don’t have recent working sets to use.",
		},
		{
			name: "session exactly at now is recent",
			sessions: []recommendation.Session{{
				Date: characterizationNow,
				Sets: workingSets(8, weight, 2),
			}},
			wantSets:        2,
			wantReps:        8,
			wantWeight:      float64Ptr(100),
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name: "same-time sessions cannot establish progression order",
			sessions: []recommendation.Session{
				{
					Date: characterizationNow.Add(-time.Hour),
					Sets: workingSets(12, weight, 2),
				},
				{
					Date: characterizationNow.Add(-time.Hour),
					Sets: workingSets(12, weight, 2),
				},
			},
			wantSets:        2,
			wantReps:        12,
			wantWeight:      float64Ptr(100),
			wantExplanation: "Repeat your last working sets.",
		},
		{
			name: "out-of-order history uses the first entry as latest",
			sessions: []recommendation.Session{
				{
					Date: characterizationNow.Add(-2 * time.Hour),
					Sets: workingSets(8, weight, 2),
				},
				{
					Date: characterizationNow.Add(-time.Hour),
					Sets: workingSets(12, weight, 2),
				},
			},
			wantSets:        2,
			wantReps:        8,
			wantWeight:      float64Ptr(100),
			wantExplanation: "Repeat your last working sets.",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := recommendation.RecommendTraining(characterizationNow, recommendation.TrainingContext{Sessions: tt.sessions}, "normal")

			require.Equal(t, "normal", result.Readiness)
			require.NotNil(t, result.Plan)
			if tt.wantWeight == nil {
				require.Nil(t, result.Plan.Weight)
			} else {
				require.NotNil(t, result.Plan.Weight)
				require.Equal(t, *tt.wantWeight, *result.Plan.Weight)
			}
			require.Equal(t, tt.wantSets, result.Plan.Sets)
			require.Equal(t, tt.wantReps, result.Plan.Reps)
			require.Equal(t, tt.wantExplanation, result.Explanation)
		})
	}
}
