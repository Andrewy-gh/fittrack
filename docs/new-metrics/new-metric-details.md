Here’s a compact spec you can drop into an LLM or issue to implement the five metrics.

### Inputs per set

- Required: weight W (lbs/kg), reps R, is_working (bool).
- Optional: RPE (1–10).
- Reference values per lift: historical_1RM, training_max (typically 0.85–0.90 × historical_1RM).

### Per-set calculations

- Volume: volume = W × R.
- e1RM (estimated 1RM): use Epley e1RM = W × (1 + R/30).
- Percent 1RM vs historical: pct_hist = 100 × W / historical_1RM.
- Percent 1RM vs session best e1RM: defined after session_best_e1RM is known; for per-set reporting, compute later or store W for recompute.

### Session aggregates (working sets only unless specified)

1. Total volume

- total_volume_all = sum(volume for all sets).
- total_volume_working = sum(volume where is_working = true).

2. Session best e1RM

- session_best_e1RM = max(e1RM over working sets).

3. Session average e1RM

- session_avg_e1RM = mean(e1RM over working sets).

4. Average working-set intensity

- Choose a reference for consistency:
  - vs historical: avg_intensity_vs_hist = mean(100 × W / historical_1RM over working sets).
  - optionally also vs session best: avg_intensity_vs_session_best = mean(100 × W / session_best_e1RM over working sets).

5. Session best intensity

- Define relative to the chosen reference (store both if needed):
  - session_best_intensity_vs_hist = max(100 × W / historical_1RM over working sets).
  - session_best_intensity_vs_session_best = max(100 × W / session_best_e1RM over working sets).

### Update rules

- historical_1RM: update at cycle milestones (≈8–12 weeks) or when a clearly higher reliable e1RM/tested 1RM is achieved; avoid frequent downgrades due to short-term variance.
- training_max: maintain separately (e.g., 85–90% of historical_1RM) for prescriptions; adjust modestly by cycle.

### Data model (suggested)

- Set: {lift_id, date, weight, reps, is_working, rpe?, volume, e1rm}
- Session: {
  lift_id, date,
  total_volume_all, total_volume_working,
  session_best_e1RM, session_avg_e1RM,
  avg_intensity_vs_hist, avg_intensity_vs_session_best?,
  session_best_intensity_vs_hist, session_best_intensity_vs_session_best?
  }
- Lift state: {lift_id, historical_1RM, training_max}

### Notes

- Compute session averages as the arithmetic mean of per-set values; do not average weight/reps first before 1RM formulas.
- If RPE is logged, you can optionally compute RPE-aware e1RM per set (via RIR tables) and use it in place of Epley for the same aggregates.

[1](https://github.com/EesunMoon/AML_Project)
[2](https://www.tredict.com/blog/custom_fields/)
[3](https://www.youtube.com/watch?v=uzJXqx8G8gc)
[4](https://www.youtube.com/watch?v=cCONIdrM2VI)
[5](http://docs.allennlp.org/v0.9.0/api/allennlp.training.metrics.html)
[6](https://keras.io/guides/training_with_built_in_methods/)
[7](https://scikit-learn.org/stable/modules/model_evaluation.html)
[8](https://pmc.ncbi.nlm.nih.gov/articles/PMC10222347/)

## Recommendedations on items 4 and 5:

Short answer: use historical 1RM as the default reference for both average working-set intensity and session best intensity, and calculate the session-referenced versions as optional readiness metrics in reports, not for long-term comparisons.[1][2]

### Why historical as default

- Historical 1RM gives stable, cross-session comparability so intensity trends reflect real progress rather than day-to-day fluctuations in performance or estimation error.[2][1]
- Coaches commonly prescribe and evaluate training by percent of a stable 1RM or training max, precisely to standardize load and track progression over time.[3][1]

### When to use session-based

- Session-based percent (vs session_best_e1RM) is useful as a “readiness” view: it normalizes the day’s work against that day’s capability, helping interpret whether work sets were heavy relative to current daily form.[2][3]
- This is valuable for post-session analysis and autoregulation but less meaningful for longitudinal comparisons because the denominator changes each session.[1][2]

### Recommendation for implementation

- Store both, but surface historical-referenced metrics in primary progress charts and weekly/monthly summaries; show session-referenced metrics in session detail screens or readiness widgets.[1][2]
- Naming: keep “avg_intensity_vs_hist” and “session_best_intensity_vs_hist” as default KPIs; mark the session-referenced versions as “readiness-normalized” to avoid confusion.[2][1]

[1](https://www.trainingweightlifting.com/training-methodology/training-intensity-percentages/)
[2](https://barbell-logic.com/intensity-vs-volume-explained/)
[3](https://pmc.ncbi.nlm.nih.gov/articles/PMC7927075/)
