# Inngest Chat Recovery Handoff

## Branch

- Working branch: `feat/inngest-chat-jobs`

## What is implemented

- Normal chat responses still use the existing SSE flow.
- If the browser refreshes or disconnects during an active assistant response, the run is no longer immediately marked failed.
- The client now calls a recovery endpoint and then polls persisted conversation state so the response can continue after refresh.
- The server now has a recovery endpoint at `POST /api/ai/conversations/{id}/messages/recover`.
- The server publishes a deduped Inngest event for the active run.
- A new Inngest function reloads the run from Postgres and finishes the assistant response in the background.
- The public Inngest handler is mounted at `/inngest`.
- Server and client regression tests were added for the recovery flow.

## Files to know about

- Client recovery flow:
  - `client/src/routes/_layout/chat.tsx`
  - `client/src/lib/api/ai-chat.ts`
- Server recovery endpoint and service logic:
  - `server/internal/aichat/handler.go`
  - `server/internal/aichat/service.go`
  - `server/internal/aichat/repository.go`
  - `server/internal/aichat/inngest.go`
- API wiring:
  - `server/cmd/api/main.go`
  - `server/cmd/api/routes.go`
- SQL and generated query updates:
  - `server/query.sql`
  - `server/internal/database/query.sql.go`
- Swagger output:
  - `server/docs/docs.go`
  - `server/docs/swagger.json`
  - `server/docs/swagger.yaml`

## What you need to unblock

- Create an Inngest account/project.
- Provide the server env vars:
  - `INNGEST_EVENT_KEY`
  - `INNGEST_SIGNING_KEY`
- Make sure the deployed app exposes `/inngest` publicly so Inngest can reach it.
- Register the app endpoint in Inngest for local/staging/prod as needed.
- Run one end-to-end staging check:
  - start a streaming response
  - refresh mid-stream
  - confirm the conversation continues updating and completes

## Notes on auth

- Browser requests still use the existing Stack Auth header flow.
- The Inngest callback does not use the browser auth header.
- Recovery runs are executed with trusted app-side context based on the stored run and user data.
- The Inngest handler is mounted outside `/api` so it does not get blocked by the current Stack Auth middleware.

## Verification already run

- `cd server && go test ./...`
- `cd client && bun run lint`
- `cd client && bun run tsc`
- `cd client && bun run test`
- `cd client && bun run build`
- `cd server && make swagger`

## Expected follow-up after env setup

- Add the new Inngest env vars to local and Fly.
- Deploy this branch to a staging target.
- Verify that normal SSE still feels immediate.
- Verify that recovery only kicks in on disconnect/reload paths.
- Confirm there are no duplicate assistant completions when recovery starts.
