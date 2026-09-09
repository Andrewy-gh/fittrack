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

- GitHub's migration `DATABASE_URL` uses the owning/admin role for Goose. A retrievable backup is stored as masked `MIGRATION_DATABASE_URL` in Doppler `fittrack/prd_ci`; GitHub Actions still reads its own repository secret, not Doppler. Keep both copies synchronized when rotating the admin credential.
- Doppler `prd` and Fly's `DATABASE_URL` use the separately provisioned restricted `fittrack_app` runtime role. Fly secrets are separate: changing Doppler alone does not update the running application.

For an admin operation, use `doppler run --project fittrack --config prd_ci -- <command>` and have the command read `MIGRATION_DATABASE_URL`, not the inherited runtime `DATABASE_URL`. Never print either URL.

The runtime URL must use Supabase's session pooler on port `5432`. Port `6543` is the transaction pooler and cannot safely preserve FitTrack's per-connection RLS user setting. With `ENVIRONMENT=production`, RLS enforcement is mandatory: startup rejects transaction pooling and any role that is a superuser, has `BYPASSRLS`, owns an RLS table, or can assume such a role. `fly.toml` explicitly declares the production environment. Confirm no runtime secret overrides it with a different environment before deploying.

`RLS_ENFORCEMENT_REQUIRED` may remain `true` or be omitted in production; an explicit false value is a configuration error. Only literal lowercase `true` and `false` are accepted; aliases, uppercase, and whitespace-padded values fail startup in every environment. Unset or empty values use the environment default. Development and staging retain the optional flag (default false). This does not restrict separate admin/migration connections.

### Completed Cutover (2026-09-08)

PR #277 was deployed and the restricted runtime role activated in both Doppler and Fly with `RLS_ENFORCEMENT_REQUIRED=true`. Readiness, startup role checks, database-level isolation, two authenticated accounts (including cross-user workout denial), and the narrow Stripe customer lookup were verified. The admin credential was subsequently rotated and the migration/deploy workflow rerun successfully. Temporary credential files were removed.

This records the completed rollout, not a substitute for checking current deploy state. The staged privileged-role rollback path is retired in the hardened code: production must keep the restricted role. Roll back only to application versions compatible with that role, without disabling enforcement. This hardening takes effect when deployed; #277 itself still supported the staged override. Do not repeat provisioning or rotate credentials merely to perform triage.

### Recovery Precautions

Coordinate any runtime credential or enforcement-setting change in both Doppler and Fly, then verify readiness and user isolation. Do not disable enforcement as a routine troubleshooting step. The pre-cutover admin password has been rotated; its old secret version is no longer a usable rollback credential. Use `server/scripts/verify-runtime-role.sql` to validate a candidate restricted connection before changing runtime secrets.

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
