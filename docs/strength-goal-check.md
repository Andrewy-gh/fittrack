# Strength Goal Check

The exercise detail screen compares logged observations with general strength
references. It does not score training, predict gains, or prescribe the next workout.
Select strength anywhere in the training profile's goals to enable comparisons.
Demo mode explains availability without calculating a second frontend policy.

## Calculation (`strength-observations-v1`)

`GET /api/exercises/{id}/goal-check?timezone=America/New_York` reads only the
owner's exercise. Unknown/unowned IDs return 404; invalid IDs/timezones return
400. Missing timezone means UTC. No migration or persisted result is needed.

The window starts at local midnight six dates before today and ends at server
now, inclusive. Today is partial. Calendar arithmetic preserves the seven dates
across DST. Only working sets with positive reps and workout dates in the window
count. Warm-ups, future workouts, and sets outside the window are excluded.

- Sessions count distinct workout IDs containing eligible sets.
- Training days count distinct local workout dates; two sessions on one date are
  one training day. Separate exercise IDs/aliases are not combined.
- Working-set average is total sets divided by sessions; individual session
  counts remain available. No sessions means a null average.
- With observations, two training days and an average of at least two sets meet
  the respective soft reference. More than three sets is not flagged excessive.
- Without observations, comparisons report insufficient data. Missing logging
  is not evidence of inactivity. Non-strength profiles receive no comparisons.

Loading reports recorded reps and, if possible, the range of positive recorded
loads as a percentage of the saved historical 1RM. No same-session estimated-max
fallback is used. The saved reference may itself be an estimate derived from
these sets; a source workout identifies an estimate, while no source identifies
an unverified manual value. Its timestamp is when it was saved, not when tested.
Missing/zero loads never become a failed loading comparison. Percentages above
100 are retained. Loading never produces a sufficient/insufficient training verdict.
Bodyweight, assistance and machine changes make interpretation uncertain.

The client refreshes each minute while visible and on revisit. Workout mutations,
profile updates and saved-reference changes invalidate observed Goal Check queries.
Query keys include the account and timezone. Fetch errors stay inside the card.

## Evidence and boundaries

- [ACSM, 2026](https://pubmed.ncbi.nlm.nih.gov/41843416/): general strength references
  include heavier loading, 2–3 sets and at least two weekly sessions. Applying
  frequency to each named exercise is a product inference, not a validated rule.
- [Grgic et al., 2018](https://pubmed.ncbi.nlm.nih.gov/29470825/): volume-equated
  comparisons did not show a significant independent frequency effect.
- [Pelland et al., 2026](https://pubmed.ncbi.nlm.nih.gov/41343037/): dose-response
  models support diminishing returns, not a universal personal volume cap.
- [Nuzzo et al., 2024](https://pubmed.ncbi.nlm.nih.gov/37792272/): repetition-to-1RM
  relationships vary; logged reps alone do not establish maximal effort.

These population-level findings do not validate weekly pass/fail rules. Deloads,
short sessions, incomplete logs, mixed goals and accessory work remain legitimate
explanations for falling below a reference. Hypertrophy comparisons need muscle
mappings and are outside this feature.

## Verification

Calculation tests cover session/day distinctions, varied set counts, missing data,
ratios above 100%, mixed goals, boundaries and DST. HTTP tests cover invalid input,
missing authentication, ownership misses, database failure and empty legacy profiles.
The database query test exercises ownership, warm-up filtering and inclusive dates.
UI tests keep routing and queries real while mocking the HTTP client boundary.
