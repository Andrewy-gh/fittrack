# FitTrack Local Observability

This directory runs a local Prometheus and Grafana stack for inspecting the FitTrack backend metrics endpoint.

## Start Locally

1. Start the backend on port `8080`.

   ```sh
   cd server
   make dev
   ```

2. Start the observability stack.

   ```sh
   cd ops/observability
   docker compose up -d
   ```

3. Open Grafana at <http://localhost:3001>.

   - Username: `admin`
   - Password: `admin`
   - Dashboard: `FitTrack / FitTrack Local Observability`

Prometheus is available at <http://localhost:9090>. It scrapes the backend at `host.docker.internal:8080/metrics`.

## Metrics Auth

The local scrape config assumes the backend `/metrics` endpoint is not protected. If you set `METRICS_USERNAME` and `METRICS_PASSWORD` for the local backend, add matching Prometheus `basic_auth` values before starting this stack. Do not commit real credentials.

Example local-only shape:

```yaml
scrape_configs:
  - job_name: fittrack-backend
    metrics_path: /metrics
    basic_auth:
      username: replace-me
      password: replace-me
    static_configs:
      - targets:
          - host.docker.internal:8080
```

## Included Dashboards

Grafana provisions one dashboard from `grafana/dashboards/fittrack-local-observability.json`.

It includes:

- Backend HTTP request rate and p95 latency.
- Database active and idle connection gauges.
- AI chat rollout panels from `docs/ai-chat-observability.md`.
- A `cohort` variable for switching between `beta` and `non_beta` chat telemetry.

## Production Direction

Do not deploy this local stack to production.

For production, point Grafana Cloud Metrics Endpoint scraping at the deployed FitTrack `/metrics` endpoint and configure BasicAuth with the production `METRICS_USERNAME` and `METRICS_PASSWORD` values in Grafana Cloud. Keep those credentials in the hosting platform and Grafana Cloud secret settings, not in this repository.

Vercel, Cloudflare, Netlify, and other deployment wiring are intentionally out of scope for this scaffold.
