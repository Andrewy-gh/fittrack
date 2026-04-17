export DB_USER=postgres
export DB_NAME=fittrack
export DB_PORT=5432
db_pass="${DB_PASSWORD:-}"
if [ -z "$db_pass" ]; then
	echo "Set your local database password in the shell before sourcing this file." >&2
	return 1 2>/dev/null || exit 1
fi
export DATABASE_URL="postgresql://${DB_USER}:${db_pass}@127.0.0.1:${DB_PORT}/${DB_NAME}?sslmode=disable"
unset db_pass
export PROJECT_ID=your-stack-project-id
export SECRET_SERVER_KEY=your-stack-secret-server-key
export GEMINI_API_KEY=your-gemini-api-key
# export GOOGLE_API_KEY=your-google-api-key
export GEMINI_MODEL=googleai/gemini-2.5-flash
export INNGEST_EVENT_KEY=your-inngest-event-key
export INNGEST_SIGNING_KEY=your-inngest-signing-key
