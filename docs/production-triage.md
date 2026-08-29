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

## Production Database Roles And RLS

Keep two database credentials separate:

- GitHub's migration `DATABASE_URL` uses the owning/admin role for Goose.
- Doppler `prd` `DATABASE_URL` uses a separately provisioned restricted `fittrack_app` runtime role.

The runtime URL must use Supabase's session pooler on port `5432`. Port `6543` is the transaction pooler and cannot safely preserve FitTrack's per-connection RLS user setting. During the staged cutover, `RLS_ENFORCEMENT_REQUIRED=false` lets the new binary deploy while the old privileged URL is still active. The final runtime secret must set it to `true`; startup then rejects transaction pooling and any role that is a superuser, has `BYPASSRLS`, owns an RLS table, or can assume such a role.

Activation order:

1. Deploy the RLS-aware binary and migration 27 while the existing runtime URL remains active and `RLS_ENFORCEMENT_REQUIRED=false`. Migration 27 only adds the narrow Stripe customer-to-user lookup needed by unauthenticated webhooks; it does not enable or force RLS.
2. Run `server/scripts/provision-runtime-role.sql` with the same owning/admin role used by Goose after migration 27, then configure `fittrack_app` as `LOGIN` with a generated password. The script enforces role safety, removes earlier broad `PUBLIC` grants, and grants only the current runtime operations. It intentionally does not handle the password. Do not store that password in Git or a shell script.
3. Test the candidate `fittrack_app` session-pooler URL directly without changing Doppler. Run `server/scripts/verify-runtime-role.sql` with that URL, then verify two distinct user contexts and the Stripe lookup function.
4. Update Doppler's runtime `DATABASE_URL` to the tested URL and set `RLS_ENFORCEMENT_REQUIRED=true` in the same cutover. Set `ENVIRONMENT=production` as normal environment metadata.
5. Confirm readiness, run a two-user API isolation smoke test, and verify a metadata-free Stripe subscription webhook can resolve its customer. Keep the previous runtime secret version available for rollback; rollback must restore the old URL and set `RLS_ENFORCEMENT_REQUIRED=false` together.

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
