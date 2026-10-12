import type { Page } from "@playwright/test";

/** A safe, actionable failure of authenticated browser setup. */
export class AuthReadinessError extends Error {
  /** Stable failure category; messages never include session credentials. */
  readonly _tag = "AuthReadinessError";
}

/** Check local-auth prerequisites before the bootstrap can write fixtures. */
export function localAuthConfigurationError(options: {
  isCI: boolean;
  browserProjectId: string | undefined;
  publishableClientKey: string | undefined;
  clientAuthEnabled: boolean;
}): AuthReadinessError | undefined {
  if (options.isCI) {
    return new AuthReadinessError(
      "Local E2E auth requires the Vite development server. CI uses production preview, where local auth is disabled. Run authenticated local checks without CI; use real Stack auth for preview checks.",
    );
  }
  if (!options.browserProjectId || !options.publishableClientKey) {
    return new AuthReadinessError(
      "Local E2E auth requires VITE_PROJECT_ID and VITE_PUBLISHABLE_CLIENT_KEY in the frontend environment; otherwise the app renders as a guest.",
    );
  }
  if (!options.clientAuthEnabled) {
    return new AuthReadinessError(
      "Local E2E auth requires VITE_E2E_LOCAL_AUTH_ENABLED=true in the frontend environment. Restart Vite after changing it.",
    );
  }
}

/** Prove the app's own workout read uses a valid browser session, without mutating data. */
export async function browserAuthReadinessError(
  page: Page,
  options: {
    mode: { kind: "local"; userId: string } | { kind: "stack" };
    timeoutMs?: number;
  },
): Promise<AuthReadinessError | undefined> {
  const timeout = options.timeoutMs ?? 30000;
  try {
    // Observe the app request rather than synthesizing auth headers in the probe.
    const [response] = await Promise.all([
      page.waitForResponse(
        (candidate) =>
          new URL(candidate.url()).pathname === "/api/workouts" &&
          candidate.request().method() === "GET",
        { timeout },
      ),
      page.goto("/workouts", { waitUntil: "domcontentloaded", timeout }),
    ]);
    if (!response.ok()) {
      return new AuthReadinessError(
        `Authenticated browser preflight failed: GET /api/workouts returned HTTP ${response.status()}. Check the API, proxy and session before running authenticated tests.`,
      );
    }
    const headers = await response.request().allHeaders();
    const authenticated =
      options.mode.kind === "local"
        ? headers["x-fittrack-dev-e2e-user"] === options.mode.userId
        : Boolean(headers["x-stack-access-token"]);
    if (!authenticated) {
      return new AuthReadinessError(
        "Authenticated browser preflight failed: the app's workout request did not use the expected session. An auth state file alone is not proof of authentication.",
      );
    }
    const body: unknown = await response.json();
    if (!Array.isArray(body)) {
      return new AuthReadinessError(
        "Authenticated browser preflight failed: GET /api/workouts did not return a workout list. Check the API proxy target.",
      );
    }
  } catch {
    return new AuthReadinessError(
      "Authenticated browser preflight failed: the app could not complete its authenticated workout read. Check browser readiness, API/proxy availability and session validity. Local auth requires a development frontend with both public Stack settings and the client auth flag.",
    );
  }
}
