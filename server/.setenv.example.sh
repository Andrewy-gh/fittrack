export DB_USER=postgres
export DB_NAME=fittrack
export DB_PORT=5432
: "${DB_PASSWORD:?Set DB_PASSWORD before sourcing this file}"
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@127.0.0.1:${DB_PORT}/${DB_NAME}?sslmode=disable"
export PROJECT_ID=your-stack-project-id
export SECRET_SERVER_KEY=your-stack-secret-server-key
export GEMINI_API_KEY=your-gemini-api-key
# export GOOGLE_API_KEY=your-google-api-key
export GEMINI_MODEL=googleai/gemini-2.5-flash
export INNGEST_EVENT_KEY=your-inngest-event-key
export INNGEST_SIGNING_KEY=your-inngest-signing-key
