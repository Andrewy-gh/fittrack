# Metrics History Bucketing (Decision)

## Context

We are moving exercise metric charts to a backend-powered `metrics-history` endpoint with **per-workout** points (session-level metrics).

For long ranges (`6M`, `Y`) we bucket points (weekly/monthly) to cap bar count and keep charts responsive.

## Decision

When bucketing multiple workouts into a period, **do not SUM across workouts**.

Rationale:
- SUM conflates *training frequency* with *session performance*.
  - Example: two lifters with identical sessions look "better" if they trained more often that week.
- These charts are intended to show "how strong/intense were my sessions" and "what did I hit", not "how many sessions did I do".
- SUM is only meaningful for "total workload over time"; we already have daily volume and can add a dedicated workload view later if we want.

## Reducers By Metric (Per Bucket)

- Volume (`total_volume_working`):
  - Per workout: SUM within the workout (working sets).
  - Bucketed: **AVG of workouts in the bucket**.
- Session Best e1RM: **MAX**
- Session Average e1RM: **AVG**
- Session Average Intensity: **AVG**
- Session Best Intensity: **MAX**

## Range Behavior (Hybrid)

- `W`, `M`: return raw **per-workout** points filtered by date cutoff.
- `6M`: return **weekly** buckets using reducers above.
- `Y`: return **monthly** buckets using reducers above.

## Intensity > 100%

Intensity may legitimately exceed 100% if a session beats the baseline (historical 1RM or session-best fallback).

Frontend charts should:
- use a dynamic Y-axis domain: `[0, dataMax]`
- optionally render a reference line at `100%`

