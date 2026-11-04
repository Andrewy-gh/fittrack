import type { BrowserContext, Request, Response } from "@playwright/test";

/** Sanitized startup evidence, safe to retain before authentication begins. */
export type BrowserStartupResult = {
  readonly elapsedMs: number;
  readonly deadlineMs: number;
} & (
  | { readonly kind: "ready" }
  | {
      readonly kind: "failed";
      readonly reason: "deadline" | "navigation" | "module" | "bootstrap";
    }
);

/**
 * Load the real landing route once and wait for React's visible navigation.
 * Owns a fresh page, one total deadline, and all listeners; never bootstraps auth
 * or retries navigation. HTTP readiness remains Playwright webServer's job.
 */
export async function browserStartupReadiness(
  context: BrowserContext,
  deadlineMs: number,
): Promise<BrowserStartupResult> {
  const started = performance.now();
  let timer: ReturnType<typeof setTimeout> | undefined;
  const page = await context.newPage();
  type Outcome =
    | { kind: "ready" }
    | {
        kind: "failed";
        reason: "deadline" | "navigation" | "module" | "bootstrap";
      };
  let fail: (outcome: Outcome) => void = () => {};
  const failure = new Promise<Outcome>((resolve) => {
    fail = resolve;
    timer = setTimeout(
      () => resolve({ kind: "failed", reason: "deadline" }),
      deadlineMs,
    );
  });
  const onPageError = () => fail({ kind: "failed", reason: "bootstrap" });
  const onResponse = (response: Response) => {
    if (
      response.request().resourceType() === "script" &&
      response.status() >= 400
    ) {
      fail({ kind: "failed", reason: "module" });
    }
  };
  const onRequestFailed = (request: Request) => {
    if (request.resourceType() === "script") {
      fail({ kind: "failed", reason: "module" });
    }
  };
  page.on("pageerror", onPageError);
  page.on("crash", onPageError);
  page.on("response", onResponse);
  page.on("requestfailed", onRequestFailed);
  try {
    const load = async (): Promise<Outcome> => {
      try {
        // DOMContentLoaded itself may wait on the cold module graph. Commit only
        // establishes navigation; the visible React link below establishes mount.
        const response = await page.goto("/", {
          waitUntil: "commit",
          timeout: deadlineMs,
        });
        if (!response?.ok()) return { kind: "failed", reason: "navigation" };
        const remaining = Math.max(
          1,
          deadlineMs - (performance.now() - started),
        );
        return await Promise.race([
          page
            .getByRole("link", { name: "Browse Exercise History", exact: true })
            .waitFor({ state: "visible", timeout: remaining })
            .then((): Outcome => ({ kind: "ready" })),
          page
            .locator("vite-error-overlay")
            .waitFor({ state: "attached", timeout: remaining })
            .then((): Outcome => ({ kind: "failed", reason: "module" })),
          page
            .getByRole("heading", { name: "Something went wrong", exact: true })
            .waitFor({ state: "visible", timeout: remaining })
            .then((): Outcome => ({ kind: "failed", reason: "bootstrap" })),
        ]);
      } catch (error) {
        return {
          kind: "failed",
          reason:
            error instanceof Error && error.name === "TimeoutError"
              ? "deadline"
              : "navigation",
        };
      }
    };
    const outcome = await Promise.race([load(), failure]);
    return {
      ...outcome,
      elapsedMs: Math.round(performance.now() - started),
      deadlineMs,
    };
  } finally {
    clearTimeout(timer);
    page.off("pageerror", onPageError);
    page.off("crash", onPageError);
    page.off("response", onResponse);
    page.off("requestfailed", onRequestFailed);
    // Closing cancels pending locator/navigation work after an early failure.
    await page.close();
  }
}
