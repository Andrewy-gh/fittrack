# Production Triage

Use this checklist when production behavior differs from the expected app behavior, especially when the question crosses deploy state, Fly health, Stripe billing, or user-specific access.

## Deploy State

Confirm the code is actually running before debugging product behavior:

```bash
gh run list --workflow fly-deploy.yml --branch main --limit 5 --json databaseId,displayTitle,headSha,status,conclusion,createdAt,updatedAt,url
flyctl status -a fittrack
flyctl releases -a fittrack --json
```

Match the GitHub workflow `headSha` to the Fly release timestamp and image. A successful GitHub deploy is useful evidence, but Fly status tells you what is currently serving traffic.

## Health And Slowness

Start with health, machine state, and warm request timing:

```bash
flyctl checks list -a fittrack
flyctl machine list -a fittrack
curl.exe -sS -o NUL -w "ready_code=%{http_code} ttfb=%{time_starttransfer} total=%{time_total}\n" https://fittrack.fly.dev/ready
```

If the first browser load is slow but warm checks are fast, inspect `fly.toml` for `min_machines_running` and `auto_stop_machines`. A stopped machine with a later fast `/ready` response usually points to cold start rather than slow handler work.

Use request logs to separate app latency from network or proxy latency:

```bash
flyctl logs -a fittrack --no-tail
```

Look for structured `request completed` logs with `route`, `status`, `duration_ms`, and `request_id`.

## User Billing State

For AI chat billing questions, Stripe is the source of truth, but the app reads FitTrack's stored Stripe snapshot from `stripe_subscriptions`. Compare both sides before concluding the UI is wrong.

Check these in order:

1. Stripe customer/subscription state.
2. FitTrack `stripe_customers` mapping.
3. Latest `stripe_subscriptions` row.
4. Active `user_feature_access` row for `feature_key = 'ai_chatbot'`.
5. Recent `stripe_webhook_events` rows when event timing matters.

Use Doppler for production secrets:

```bash
doppler run --project fittrack --config prd -- <read-only command>
```

If a scheduled cancellation exists in Stripe but not in FitTrack, check whether the Stripe event was processed before the deploy that added the stored field. Processed Stripe event IDs are intentionally idempotent, so replaying the same event may not update the row.

## Stripe Backfills

When a billing change adds a new stored Stripe snapshot field, ask whether existing subscriptions need a one-time backfill. Prefer a dry run first, and avoid ad hoc production SQL when an app-owned command or support path exists.

For a backfill command, expect this shape:

```bash
cd server
doppler run --project fittrack --config prd -- go run ./cmd/<billing-backfill>
doppler run --project fittrack --config prd -- go run ./cmd/<billing-backfill> --subscription sub_...
doppler run --project fittrack --config prd -- go run ./cmd/<billing-backfill> --verify sub_...
```

Only run `--apply` after the dry-run output is scoped and expected.
