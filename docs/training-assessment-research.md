# Training assessment: research decisions

Reviewed 2026-09-21 for the first strength-reference assessment. This records the implementation proposal developed after the user authorized the first build slice; it does not imply that the user individually approved every numeric or presentation choice. No personalized prescription, health eligibility determination, or overall training-success verdict is established here.

## Sources and access

- [ACSM position stand, 2026](https://pubmed.ncbi.nlm.nih.gov/41843416/) and [official ACSM summary](https://acsm.org/resistance-training-guidelines-update-2026/): general healthy-adult guidance. The abstract and indexed full-text excerpts identify adult populations, interventions lasting at least six weeks, and strength-related references including 2–3 sets and heavier loading. Direct PMC full-text retrieval encountered a browser challenge; this is not a completed appraisal of every underlying trial. [PMC article](https://pmc.ncbi.nlm.nih.gov/articles/PMC12965823/).
- [Krieger, 2009](https://pubmed.ncbi.nlm.nih.gov/19661829/): abstract reviewed. Meta-regression found larger average strength gains for 2–3 sets per exercise than one. It does not establish an individual minimum, optimum, or upper limit.
- [Grgic et al., 2018](https://pubmed.ncbi.nlm.nih.gov/29470825/): abstract reviewed. Frequency was associated with strength gains overall, but its volume-equated subgroup did not find an independent significant effect. This does not establish that frequency never matters.
- [Nuzzo et al., 2024](https://pubmed.ncbi.nlm.nih.gov/37792272/): abstract reviewed. Repetition-to-maximum relationships were estimated from repetitions-to-failure tests and varied across individuals and exercises. Ordinary logged repetitions do not identify maximal effort.
- [Pelland et al., 2026, online 2025](https://link.springer.com/article/10.1007/s40279-025-02344-w): publisher abstract reviewed; full text subscription-limited. Direct/indirect sets were defined relative to the measured outcome. Half-credit for indirect sets had strongest relative model evidence among 0, 0.5 and 1 alternatives. Volume related to both outcomes with diminishing returns; frequency related more consistently to strength. This is not validation of FitTrack's catalog labels or individualized target formulas.
- [Currier et al., 2023](https://pubmed.ncbi.nlm.nih.gov/37414459/): abstract reviewed. Resistance-training prescriptions studied improved both strength and hypertrophy compared with no exercise. Higher loads favored strength and multi-set prescriptions featured prominently for hypertrophy. The same training can contribute to both goals.

## Decisions for the first slice

### Claim and population

The supported claim is a numerical comparison of logged working-set counts with a general multi-set reference. It does not establish that the user belongs to the studied healthy-adult population or that their training will produce a particular adaptation. Age and general health eligibility are not recorded; experience and absence of reported movement limitations cannot establish them.

Show this general-reference limitation visibly and expand the detailed explanation on demand. Do not infer age, sex, medical suitability, or personal training requirements. No new health questionnaire is needed for the narrow descriptive claim. Any future personalized recommendation must revisit applicability and necessary inputs.

### Set comparison

Use the 2–3 sets per exercise/session reference, with 2 as its lower edge. Comparing against that edge is a product convention, not a validated physiological classifier. A session with one qualifying logged working set is below the reference minimum; two or more reaches that minimum. Four or more is not excessive by this rule and must not be labelled optimal or within a scientifically bounded 2–3 range.

Compare sessions individually. A week with one set in one session and five in another must retain one below-reference session and one session reaching the minimum. An average of three hides that distribution. Summarize how many observed sessions reached the reference, without a combined weekly success status or an assertion that every session must do so.

### Eligibility and uncertainty

Assessment-qualified rows have exact working-set type and positive repetitions. Warmups do not count. Missing or zero external weight does not invalidate a set. Invalid legacy records remain visible as data-quality issues and prevent a confident comparison for the affected session. Do not change the descriptive evidence service's existing raw counting contract.

No observed qualifying training is unknown assessment evidence, not proof of inactivity or failed training. A catalog link is unnecessary for an exercise-specific count, but muscle attribution requires reviewed structured metadata. Do not merge exercise aliases or infer muscle roles from names.

Logging completeness, effort, range of motion, and deload intent are not established by catalog coverage or set counts. Findings concern recorded data only. Missing 1RM information does not invalidate an otherwise usable count comparison.

### Loading and frequency

Keep both descriptive in this slice. Do not enforce two weekly days for every named exercise. Research findings about program frequency do not provide that universal product rule.

Existing Session Metrics calculate Epley estimates, relative loading, and load-volume. Reuse the existing UI where appropriate, but do not use its display buckets, missing-data zeros, or session estimates as physiological ground truth. A manually entered historical maximum is not automatically a verified measured maximum. Its update time is not necessarily a test date. Auto-generated historical maxima and same-session fallback estimates cannot justify an 80% pass/fail assessment. No arbitrary 1RM freshness cutoff is introduced.

### Period

Weekly reporting is compatible with references stated per week; it is not a one-week adaptation test. A completed Monday–Sunday local week is a product default, not a uniquely validated biological window. Preserve explicit start date, IANA timezone, exclusive end, captured observation time, and future exclusion.

Current incomplete periods show accumulated facts with a period-in-progress reason and no comparison badges in this first version. No weekly extrapolation, missing-day zeros as inactivity, minimum two-week history requirement, plateau inference, or mandated week-over-week increase.

## Other goals and deferred questions

Hypertrophy comparisons require an assessment-specific exercise-to-muscle mapping and a benchmark expressed in compatible units. Retaining catalog primary/secondary counts is useful evidence, but neither primary-only counting nor universal half-credit solves mapping validity. The catalog's deadlift entry illustrates the issue: many listed secondary roles do not establish equal indirect stimulus. Do not silently combine an approximate ten-set reference with fractional totals from a different definition. This is a concrete deferred mapping/measurement task, not an AI dependency.

Endurance remains ambiguous between muscular and aerobic endurance; reps alone cannot resolve user intent or serve as a standardized outcome test. General fitness includes more than resistance training. Weight loss and mobility cannot receive set-based success judgments from the current logs. Explain unavailable assessment coverage for each selected goal rather than substituting strength. Relevant sources include [WHO activity guidance](https://www.who.int/news-room/fact-sheets/detail/physical-activity), [muscular endurance review](https://www.sciencedirect.com/science/article/pii/S0765159722000405), [CDC weight guidance](https://www.cdc.gov/healthy-weight-growth/physical-activity/), and [range-of-motion meta-analysis](https://pubmed.ncbi.nlm.nih.gov/36622555/).

Profile expansion is deferred until a supported rule requires it. Availability, goal clarification, effort recording, and period context are potential inputs with different storage scopes. Changes to collection or use should include an accurate privacy notice and verified data-control behavior; they are not prerequisites to the limited general-reference comparison.

## Acceptance examples

| Logged evidence | Expected interpretation |
| --- | --- |
| Completed week, one session with two valid working sets | Session reaches reference minimum; no frequency failure or overall goal-met claim |
| Completed week, sessions with one and five working sets | One below-reference row, one minimum-reached row; no average-based overall pass |
| Four sets in one session | Minimum reached; no excess-volume or optimality claim |
| Warmups only or no sessions | Cannot assess; no inferred inactivity |
| Two bodyweight working sets with positive reps and missing weight | Set comparison available; relative loading unavailable |
| Invalid legacy working-set row in a session | Counts/data issue visible; affected comparison unavailable |
| Current week with otherwise valid sets | Facts available, period in progress, comparison deferred |
| Unclassified custom exercise | Exercise-specific comparison available; no inferred muscle attribution |
| Strength absent from selected goals | Explain selected-goal coverage; do not silently substitute a strength assessment |
| Missing or estimated historical maximum | Set comparison unchanged; loading informational only |

These examples define observable product behavior and must be tested independently of helper implementation details. Ownership isolation and timezone/future boundaries also require validation.
