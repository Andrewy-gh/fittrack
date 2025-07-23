FROM oven/bun:1.2-slim AS client-build
WORKDIR /app
ENV NODE_ENV=production

COPY --link client/package.json client/bun.lock ./
RUN bun install --frozen-lockfile

COPY --link client .
RUN bun run build

FROM golang:1.24.2-alpine AS server-build
WORKDIR /app
COPY --link server .
RUN go vet -v ./...
RUN go test -v ./...
RUN CGO_ENABLED=0 go build -o api ./cmd/api

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=client-build /app/dist ./dist
COPY --from=server-build /app/api .
EXPOSE 8080
CMD ["./api"]