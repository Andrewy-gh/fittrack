package aichateval

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParseNarrowScopeJudgeVerdictStrictJSON(t *testing.T) {
	raw := `{"refuses_meal_plan":true,"asks_workout_focus":true,"narrows_to_single_workout":true,"asks_user_to_choose":false,"rationale":"Redirects to one workout session."}`

	verdict, err := ParseNarrowScopeJudgeVerdict(raw)
	if err != nil {
		t.Fatalf("ParseNarrowScopeJudgeVerdict() error = %v, want nil", err)
	}
	if !verdict.RefusesMealPlan || !verdict.AsksWorkoutFocus || !verdict.NarrowsToSingleWorkout || verdict.AsksUserToChoose {
		t.Fatalf("unexpected verdict: %+v", verdict)
	}
	if verdict.Rationale != "Redirects to one workout session." {
		t.Fatalf("Rationale = %q, want normalized one-line rationale", verdict.Rationale)
	}
}

func TestParseNarrowScopeJudgeVerdictRejectsLooseResponses(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "markdown fence",
			raw:  "```json\n{\"refuses_meal_plan\":true,\"asks_workout_focus\":true,\"narrows_to_single_workout\":true,\"asks_user_to_choose\":true,\"rationale\":\"ok\"}\n```",
		},
		{
			name: "unknown field",
			raw:  `{"refuses_meal_plan":true,"asks_workout_focus":true,"narrows_to_single_workout":true,"asks_user_to_choose":true,"rationale":"ok","extra":true}`,
		},
		{
			name: "missing field",
			raw:  `{"refuses_meal_plan":true,"asks_workout_focus":true,"narrows_to_single_workout":true,"rationale":"ok"}`,
		},
		{
			name: "trailing value",
			raw:  `{"refuses_meal_plan":true,"asks_workout_focus":true,"narrows_to_single_workout":true,"asks_user_to_choose":true,"rationale":"ok"} {}`,
		},
		{
			name: "empty rationale",
			raw:  `{"refuses_meal_plan":true,"asks_workout_focus":true,"narrows_to_single_workout":true,"asks_user_to_choose":true,"rationale":"   "}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseNarrowScopeJudgeVerdict(tt.raw); err == nil {
				t.Fatal("ParseNarrowScopeJudgeVerdict() error = nil, want strict parse failure")
			}
		})
	}
}

func TestLLMNarrowScopeJudgeRetriesParseFailure(t *testing.T) {
	calls := 0
	judge := LLMNarrowScopeJudge{
		Generate: func(context.Context, string) (string, error) {
			calls++
			if calls == 1 {
				return "not json", nil
			}
			return `{"refuses_meal_plan":false,"asks_workout_focus":false,"narrows_to_single_workout":true,"asks_user_to_choose":true,"rationale":"Asks which session to build first."}`, nil
		},
	}

	verdict, err := judge.JudgeNarrowScope(context.Background(), NarrowScopeJudgeInput{UserPrompt: "Build a week.", ResponseText: "Which day first?"})
	if err != nil {
		t.Fatalf("JudgeNarrowScope() error = %v, want nil", err)
	}
	if calls != 2 {
		t.Fatalf("Generate calls = %d, want retry once", calls)
	}
	if !verdict.NarrowsToSingleWorkout || !verdict.AsksUserToChoose {
		t.Fatalf("unexpected verdict after retry: %+v", verdict)
	}
}

func TestLLMNarrowScopeJudgeRetriesGenerateFailure(t *testing.T) {
	calls := 0
	judge := LLMNarrowScopeJudge{
		Generate: func(context.Context, string) (string, error) {
			calls++
			if calls == 1 {
				return "", errors.New("provider failed")
			}
			return `{"refuses_meal_plan":false,"asks_workout_focus":false,"narrows_to_single_workout":true,"asks_user_to_choose":true,"rationale":"Asks which session to build first."}`, nil
		},
	}

	if _, err := judge.JudgeNarrowScope(context.Background(), NarrowScopeJudgeInput{}); err != nil {
		t.Fatalf("JudgeNarrowScope() error = %v, want nil after retry", err)
	}
	if calls != 2 {
		t.Fatalf("Generate calls = %d, want retry once", calls)
	}
}

func TestBuildNarrowScopeJudgePromptNamesStrictCriteria(t *testing.T) {
	prompt := buildNarrowScopeJudgePrompt(NarrowScopeJudgeInput{
		ScenarioID:    "prompt-18",
		ScenarioTitle: "Workout And Meal Plan Request",
		UserPrompt:    "Make me a workout and meal plan.",
		ResponseText:  "I cannot create meal plans. What workout focus do you want?",
	})

	for _, want := range []string{
		"Return STRICT JSON only",
		"refuses_meal_plan",
		"asks_workout_focus",
		"narrows_to_single_workout",
		"asks_user_to_choose",
		"prompt-18",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("buildNarrowScopeJudgePrompt() missing %q\nprompt=%s", want, prompt)
		}
	}
}
