import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { chromium, type Browser } from "@playwright/test";
import { createServer } from "node:http";
import { once } from "node:events";
import {
  browserAuthReadinessError,
  localAuthConfigurationError,
} from "./e2e/auth-readiness";

describe("local authenticated setup", () => {
  const configured = {
    isCI: false,
    browserProjectId: "public-project",
    publishableClientKey: "public-key",
    clientAuthEnabled: true,
  };

  it("accepts configured development authentication", () => {
    expect(localAuthConfigurationError(configured)).toBeUndefined();
  });

  it.each([
    [{ isCI: true }, "development server"],
    [{ browserProjectId: undefined }, "VITE_PROJECT_ID"],
    [{ publishableClientKey: undefined }, "VITE_PUBLISHABLE_CLIENT_KEY"],
    [{ clientAuthEnabled: false }, "VITE_E2E_LOCAL_AUTH_ENABLED"],
  ])("refuses unusable local-auth configuration %j", (override, message) => {
    expect(
      localAuthConfigurationError({ ...configured, ...override })?.message,
    ).toContain(message);
  });
});

describe("authenticated browser readiness", () => {
  let browser: Browser;
  let baseURL: string;
  // This HTTP fixture exercises the browser/network boundary, not FitTrack's backend.
  // The probe must observe a request made by the page; it cannot supply its own headers.
  const server = createServer((request, response) => {
    if (request.url === "/api/workouts") {
      const denied = request.headers["x-fittrack-dev-e2e-user"] === "expired";
      response.writeHead(denied ? 401 : 200, {
        "content-type": "application/json",
      });
      response.end(denied ? '{"message":"unauthorized"}' : "[]");
      return;
    }
    response.writeHead(200, { "content-type": "text/html" });
    response.end(`<script>
      const session = JSON.parse(localStorage.getItem('session') || '{}');
      if (session.kind === 'guest') {
        document.body.textContent = 'Guest workouts';
      } else {
        fetch('/api/workouts', { headers: session.headers || {} });
      }
    </script>`);
  });

  beforeAll(async () => {
    server.listen(0, "127.0.0.1");
    await once(server, "listening");
    const address = server.address();
    if (!address || typeof address === "string")
      throw new Error("Missing fixture port");
    baseURL = `http://127.0.0.1:${address.port}`;
    browser = await chromium.launch();
  });

  afterAll(async () => {
    await browser?.close();
    await new Promise<void>((resolve, reject) => {
      server.close((error) => (error ? reject(error) : resolve()));
    });
  });

  it.each([
    {
      kind: "local" as const,
      headers: { "x-fittrack-dev-e2e-user": "current-user" },
    },
    {
      kind: "stack" as const,
      headers: { "x-stack-access-token": "fixture-session" },
    },
  ])(
    "accepts the page's authenticated $kind API read",
    async ({ kind, headers }) => {
      const context = await browser.newContext({ baseURL });
      try {
        await context.addInitScript(
          (session) => {
            localStorage.setItem("session", JSON.stringify(session));
          },
          { headers },
        );
        const page = await context.newPage();
        expect(
          await browserAuthReadinessError(page, {
            mode:
              kind === "local" ? { kind, userId: "current-user" } : { kind },
          }),
        ).toBeUndefined();
      } finally {
        await context.close();
      }
    },
  );

  it.each([
    [{ headers: { "x-fittrack-dev-e2e-user": "expired" } }, "HTTP 401"],
    [
      { headers: { "x-fittrack-dev-e2e-user": "old-user" } },
      "expected session",
    ],
    [{ headers: {} }, "expected session"],
    [{ kind: "guest" }, "could not complete"],
  ])(
    "rejects invalid or missing browser authentication %j",
    async (session, message) => {
      const context = await browser.newContext({ baseURL });
      try {
        await context.addInitScript((value) => {
          localStorage.setItem("session", JSON.stringify(value));
        }, session);
        const page = await context.newPage();
        const error = await browserAuthReadinessError(page, {
          mode: { kind: "local", userId: "current-user" },
          timeoutMs: 1500,
        });
        expect(error?.message).toContain(message);
        expect(error?.message).not.toContain("expired");
        expect(error?.message).not.toContain("old-user");
      } finally {
        await context.close();
      }
    },
  );

  it("rejects a successful proxy response that is not a workout list", async () => {
    const context = await browser.newContext({ baseURL });
    try {
      await context.addInitScript(() => {
        localStorage.setItem(
          "session",
          JSON.stringify({
            headers: { "x-fittrack-dev-e2e-user": "current-user" },
          }),
        );
      });
      await context.route(`${baseURL}/api/workouts`, (route) =>
        route.fulfill({
          status: 200,
          contentType: "application/json",
          body: '{"status":"healthy"}',
        }),
      );
      const error = await browserAuthReadinessError(await context.newPage(), {
        mode: { kind: "local", userId: "current-user" },
      });
      expect(error?.message).toContain("did not return a workout list");
    } finally {
      await context.close();
    }
  });
});
