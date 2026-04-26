package aichateval

import "github.com/Andrewy-gh/fittrack/server/internal/aichat"

// DefaultScenarios returns realistic prompts that exercise workout draft readiness.
func DefaultScenarios() []Scenario {
	return []Scenario{
		{
			ID:             "prompt-01",
			Title:          "Home Dumbbell Pull",
			Prompt:         "I've got 45 minutes at home with just adjustable dumbbells and a bench. Give me a pull workout.",
			Expectation:    "Should ideally generate a structured draft with realistic dumbbell-and-bench pull volume.",
			FollowUpAnswer: "No injuries. I'm intermediate and want muscle growth.",
		},
		{
			ID:             "prompt-02",
			Title:          "Commercial Gym Hypertrophy Legs",
			Prompt:         "I want a 60-minute hypertrophy leg workout at a commercial gym.",
			Expectation:    "Should ask about injuries first or generate only if it can proceed after asking once; when it generates the draft should feel full enough for 60 minutes.",
			FollowUpAnswer: "No injuries. I'm intermediate, and hypertrophy is the goal.",
		},
		{
			ID:             "prompt-03",
			Title:          "Lower Body With Knee Pain",
			Prompt:         "I have knee pain. Build me a lower-body workout.",
			Expectation:    "Should gather missing readiness info first and avoid obvious knee-aggravating exercise choices if it generates.",
			FollowUpAnswer: "It's mild front-of-knee discomfort, no sharp pain. Keep it joint-friendly, 35 minutes, dumbbells and machines are available.",
		},
		{
			ID:             "prompt-04",
			Title:          "Hotel Room No Equipment",
			Prompt:         "Hotel room, 25 minutes, no equipment, full body.",
			Expectation:    "Should ask about injuries first or generate a concise feasible hotel-room bodyweight draft.",
			FollowUpAnswer: "No injuries. I'm beginner to intermediate and want a balanced full-body session.",
		},
		{
			ID:             "prompt-05",
			Title:          "Beginner Upper Body Full Gym",
			Prompt:         "Beginner, 40 minutes, upper body, full gym, no injuries.",
			Expectation:    "Should generate a structured beginner-appropriate draft without needing more info.",
			FollowUpAnswer: "Main goal is general strength and confidence with machines and dumbbells.",
		},
		{
			ID:             "prompt-06",
			Title:          "Shoulder Issue Push Workout",
			Prompt:         "I tore up my shoulder recently, but I still want a push workout with dumbbells.",
			Expectation:    "Should gather missing readiness info first and avoid obvious shoulder-conflict push exercises if it generates.",
			FollowUpAnswer: "It is a recent mild strain and overhead pressing bothers it. Keep the workout shoulder-friendly, 30 minutes, no sharp pain.",
		},
		{
			ID:             "prompt-07",
			Title:          "Bodyweight Strength Full Body",
			Prompt:         "Bodyweight only, 30 minutes, strength-focused full body.",
			Expectation:    "Should ask about injuries first or generate a realistic bodyweight strength-leaning draft.",
			FollowUpAnswer: "No injuries. I'm intermediate and can do push-ups, lunges, and planks.",
		},
		{
			ID:             "prompt-08",
			Title:          "Barbell-Only Pull Day",
			Prompt:         "No machines, no cables, just barbells and plates, 50 minutes, pull day.",
			Expectation:    "Should avoid cable and machine work if it generates.",
			FollowUpAnswer: "No injuries. I'm intermediate and want strength with some hypertrophy.",
		},
		{
			ID:             "prompt-09",
			Title:          "Mobility Rehab Back",
			Prompt:         "I only want 15 minutes of mobility and rehab work for my back.",
			Expectation:    "Should allow intentionally low-volume work instead of forcing a normal lifting session.",
			FollowUpAnswer: "No acute injury or numbness. This is gentle low-back stiffness, and I want easy mobility only.",
		},
		{
			ID:    "prompt-10",
			Title: "Elbow-Sensitive Revision Follow-Up",
			History: []aichat.RuntimeChatMessage{
				{Role: "user", Text: "Build me a 45 minute push workout with dumbbells and a bench. No injuries."},
				{Role: "assistant", Text: "I put together a structured workout draft for you."},
			},
			Prompt:         "Swap out anything that bothers my elbow.",
			Expectation:    "Should ask a focused revision follow-up or adapt the next draft around the new elbow constraint.",
			FollowUpAnswer: "It's mild elbow irritation during deep pressing and skull crushers. Keep dumbbells and bench, avoid elbow-aggravating moves.",
		},
	}
}
