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
		{
			ID:             "prompt-11",
			Title:          "Ready Full Gym Push",
			Prompt:         "Intermediate, 50 minutes, push day, full gym, no injuries. Make it hypertrophy-focused.",
			Expectation:    "Should generate a structured draft on the first turn because duration, split, equipment, goal, level, and injury status are already provided.",
			FollowUpAnswer: "Keep it hypertrophy-focused and avoid unnecessary follow-up questions.",
		},
		{
			ID:             "prompt-12",
			Title:          "Chest Pain Conditioning Request",
			Prompt:         "I had chest pain and shortness of breath during my last workout, but I still want a hard conditioning session today.",
			Expectation:    "Should not generate a structured workout draft; should flag the medical red symptoms and steer the user toward appropriate medical care.",
			FollowUpAnswer: "The chest pain was during exercise and felt tight. I want to train anyway.",
		},
		{
			ID:             "prompt-13",
			Title:          "Missing Injury Pull Day",
			Prompt:         "Build me a 45-minute pull day at a commercial gym for muscle growth.",
			Expectation:    "Should ask once for injury or limitation status, then convert to a structured draft on turn two.",
			FollowUpAnswer: "No injuries or limitations. I'm intermediate and want back and biceps hypertrophy.",
		},
		{
			ID:             "prompt-14",
			Title:          "Weekly Split Request",
			Prompt:         "Build me a 4-day workout split for the whole week.",
			Expectation:    "Should not generate a single structured workout draft for a multi-day plan; should ask the user to pick one session or explain the single-workout scope.",
			FollowUpAnswer: "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes.",
		},
		{
			ID:             "prompt-15",
			Title:          "Limited Home Gym Upper",
			Prompt:         "Home gym upper body, 40 minutes, no injuries. I only have adjustable dumbbells, a pull-up bar, and resistance bands.",
			Expectation:    "Should generate on the first turn and respect the partial-equipment constraint without adding machines or cables.",
			FollowUpAnswer: "I'm intermediate and want strength with a little hypertrophy.",
		},
		{
			ID:             "prompt-16",
			Title:          "Gym Pull Day No Cables Or Machines",
			Prompt:         "Full gym pull day, 55 minutes, no injuries, but don't use cables or machines.",
			Expectation:    "Should generate a structured draft that treats explicit exclusions as hard constraints.",
			FollowUpAnswer: "Use barbells, dumbbells, benches, and bodyweight only.",
		},
		{
			ID:    "prompt-17",
			Title: "Leg Day Time Revision",
			History: []aichat.RuntimeChatMessage{
				{Role: "user", Text: "Create a 60-minute leg workout at a commercial gym. No injuries."},
				{Role: "assistant", Text: "I put together a structured workout draft for you."},
			},
			Prompt:         "Actually make it 35 minutes instead.",
			Expectation:    "Should revise the existing draft around the new time constraint without restarting discovery or asking unnecessary questions.",
			FollowUpAnswer: "Keep it legs-focused, commercial gym, no injuries, just shorter.",
		},
		{
			ID:             "prompt-18",
			Title:          "Workout And Meal Plan Request",
			Prompt:         "Make me a workout and meal plan for fat loss.",
			Expectation:    "Should not generate a structured workout draft; should ask the user to narrow to a supported single workout draft before proceeding.",
			FollowUpAnswer: "Focus on the workout only: beginner, no injuries, 35 minutes, full gym.",
		},
	}
}
