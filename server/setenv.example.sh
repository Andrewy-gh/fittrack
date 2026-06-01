export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=fittrack
export DB_PORT=55432

# Local development and migration URL.
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@127.0.0.1:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Stack Auth project ID used by the API to validate JWTs.
export PROJECT_ID="your-stack-project-id"

# Common local defaults
export PORT=8080
export LOG_LEVEL=info
export ENVIRONMENT=development
export RATE_LIMIT_RPM=100

# Optional when the frontend is served from a different origin than the Vite proxy.
# export ALLOWED_ORIGINS="http://localhost:5173,http://localhost:3000"

# Optional database pool tuning
export DB_MAX_CONNS=15
export DB_MIN_CONNS=2
export DB_MAX_CONN_IDLE=30s
export DB_MAX_CONN_LIFE=30m
export DB_HEALTHCHECK=30s

# Optional metrics auth
# export METRICS_USERNAME="metrics-user"
# export METRICS_PASSWORD="metrics-password"

# Optional local Fly read-only token for inspecting app status/logs with flyctl.
# Keep deploy tokens in GitHub Actions secrets instead.
# export FLY_API_TOKEN="fly-readonly-token"

# Optional AI chat / Gemini smoke test
export GEMINI_API_KEY="your-gemini-api-key"
# export GOOGLE_API_KEY="your-google-api-key"
# export GEMINI_MODEL="googleai/gemini-2.5-flash"
# export AI_CHAT_TRACE_LOGS=true # enables verbose AI chat stream timing logs for local debugging

# Optional Inngest-backed AI chat recovery
# export INNGEST_DEV="1" # or a dev server URL such as http://127.0.0.1:8288
# export INNGEST_EVENT_KEY="your-inngest-event-key"
# export INNGEST_SIGNING_KEY="your-inngest-signing-key"

# Optional Stripe premium AI chat billing.
# Use test-mode values from the same Stripe account.
# STRIPE_SECRET_KEY must be a sk_test_... secret key.
# STRIPE_WEBHOOK_SECRET must be the whsec_... printed by `stripe listen`.
# STRIPE_PREMIUM_PRICE_ID must be a recurring test Price ID such as price_...,
# not a display label.
# APP_BASE_URL is the frontend origin used for Checkout redirects.
# AI_CHAT_TRIAL_PROMPT_CAP is the Stripe trial AI chat run limit.
# export STRIPE_SECRET_KEY="sk_test_..."
# export STRIPE_WEBHOOK_SECRET="whsec_..."
# export STRIPE_PREMIUM_PRICE_ID="price_..."
# export APP_BASE_URL="http://localhost:5173"
# export AI_CHAT_TRIAL_PROMPT_CAP=30
