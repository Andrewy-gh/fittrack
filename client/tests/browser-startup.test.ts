import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { chromium, type Browser } from "@playwright/test";
import { createServer } from "node:http";
import { once } from "node:events";
import { spawn } from "node:child_process";
import { mkdtemp, readFile, writeFile, access, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { browserStartupReadiness } from "./e2e/browser-startup";

describe("browser startup boundary", () => {
  let browser: Browser;
  let baseURL: string;
  let mode:
    | "delayed"
    | "broken"
    | "exception"
    | "overlay"
    | "route-error"
    | "stalled" = "delayed";
  let navigations = 0;
  // Transport fixtures test the deadline and fatal signals, not real-app readiness.
  const server = createServer((request, response) => {
    if (request.url === "/entry.js") {
      if (mode === "stalled") return;
      if (mode === "broken") {
        response.writeHead(500).end("private compiler details");
        return;
      }
      response.writeHead(200, { "content-type": "text/javascript" });
      if (mode === "exception") {
        response.end("throw new Error('private session detail')");
        return;
      }
      const markup =
        mode === "overlay"
          ? "<vite-error-overlay>private compiler details</vite-error-overlay>"
          : mode === "route-error"
            ? "<h1>Something went wrong</h1>"
            : '<a href="/exercises">Browse Exercise History</a>';
      setTimeout(
        () =>
          response.end(`document.body.innerHTML = ${JSON.stringify(markup)}`),
        350,
      );
      return;
    }
    if (request.url === "/") navigations++;
    response.writeHead(200, { "content-type": "text/html" });
    response.end(
      '<div id="app"></div><script type="module" src="/entry.js"></script>',
    );
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
    server.closeAllConnections();
    await new Promise<void>((resolve, reject) =>
      server.close((error) => (error ? reject(error) : resolve())),
    );
  });

  it("waits beyond ordinary navigation timeout without reloading or changing session state", async () => {
    mode = "delayed";
    navigations = 0;
    const context = await browser.newContext({ baseURL });
    context.setDefaultNavigationTimeout(100);
    await context.addCookies([
      { name: "session", value: "preserve", url: baseURL },
    ]);
    try {
      const result = await browserStartupReadiness(context, 3000);
      expect(result.kind).toBe("ready");
      expect(result.elapsedMs).toBeGreaterThanOrEqual(300);
      expect(navigations).toBe(1);
      expect(context.pages()).toHaveLength(0);
      expect(
        (await context.cookies()).find((cookie) => cookie.name === "session")
          ?.value,
      ).toBe("preserve");
    } finally {
      await context.close();
    }
  });

  it.each([
    ["broken", "module"],
    ["exception", "bootstrap"],
    ["overlay", "module"],
    ["route-error", "bootstrap"],
    ["stalled", "deadline"],
  ] as const)(
    "rejects %s startup with a bounded, sanitized %s result",
    async (fixtureMode, reason) => {
      mode = fixtureMode;
      const context = await browser.newContext({ baseURL });
      try {
        const result = await browserStartupReadiness(context, 1200);
        expect(result).toMatchObject({
          kind: "failed",
          reason,
          deadlineMs: 1200,
        });
        expect(result.elapsedMs).toBeLessThan(3000);
        if (reason === "deadline")
          expect(result.elapsedMs).toBeGreaterThanOrEqual(1100);
        expect(JSON.stringify(result)).not.toContain("private");
        expect(context.pages()).toHaveLength(0);
      } finally {
        await context.close();
      }
    },
  );

  it("rejects an unreachable frontend and releases its page", async () => {
    const probe = createServer();
    probe.listen(0, "127.0.0.1");
    await once(probe, "listening");
    const address = probe.address();
    if (!address || typeof address === "string")
      throw new Error("Missing fixture port");
    const baseURL = `http://127.0.0.1:${address.port}`;
    await new Promise<void>((resolve) => probe.close(() => resolve()));
    const context = await browser.newContext({ baseURL });
    try {
      const result = await browserStartupReadiness(context, 1200);
      expect(result).toMatchObject({ kind: "failed", reason: "navigation" });
      expect(result.elapsedMs).toBeLessThan(3000);
      expect(context.pages()).toHaveLength(0);
    } finally {
      await context.close();
    }
  });
});

it("global setup fails before auth mutation, retains evidence and stops its owned webServer", async () => {
  const clientDir = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "..",
  );
  const scratch = await mkdtemp(path.join(tmpdir(), "fittrack-startup-"));
  const portProbe = createServer();
  portProbe.listen(0, "127.0.0.1");
  await once(portProbe, "listening");
  const address = portProbe.address();
  if (!address || typeof address === "string")
    throw new Error("Missing fixture port");
  const port = address.port;
  await new Promise<void>((resolve) => portProbe.close(() => resolve()));
  const baseURL = `http://127.0.0.1:${port}`;
  const outputDir = path.join(scratch, "results");
  const fixturePath = path.join(scratch, "server.mjs");
  const mutationPath = path.join(scratch, "bootstrap-called");
  const scenarioPath = path.join(scratch, "scenario-called");
  const importPath = path.join(
    clientDir,
    "node_modules/@playwright/test/index.mjs",
  );
  await writeFile(
    fixturePath,
    `
    import {createServer} from 'node:http';
    import {writeFileSync} from 'node:fs';
    createServer((req, res) => {
      if (req.method === 'POST') writeFileSync(${JSON.stringify(mutationPath)}, 'mutation');
      res.writeHead(200, {'content-type': 'text/html'});
      res.end('<script type="module">throw new Error("private bootstrap detail")</script>');
    }).listen(${port}, '127.0.0.1');
  `,
  );
  await writeFile(
    path.join(scratch, "scenario.test.ts"),
    `
    import {test} from ${JSON.stringify(importPath)};
    import {writeFileSync} from 'node:fs';
    test('scenario', () => writeFileSync(${JSON.stringify(scenarioPath)}, 'ran'));
  `,
  );
  const configPath = path.join(scratch, "playwright.config.mts");
  await writeFile(
    configPath,
    `
    export default {
      testDir: ${JSON.stringify(scratch)}, testMatch: 'scenario.test.ts',
      globalSetup: ${JSON.stringify(path.join(clientDir, "tests/e2e/global-setup.ts"))},
      outputDir: ${JSON.stringify(outputDir)}, reporter: 'line', workers: 1,
      globalTimeout: 30000, use: {baseURL: ${JSON.stringify(baseURL)}},
      webServer: {command: ${JSON.stringify(`node "${fixturePath}"`)},
        url: ${JSON.stringify(baseURL)}, reuseExistingServer: false, timeout: 10000}
    };
  `,
  );
  try {
    const child = spawn(
      process.execPath,
      [
        path.join(clientDir, "node_modules/@playwright/test/cli.js"),
        "test",
        "--config",
        configPath,
      ],
      {
        cwd: clientDir,
        windowsHide: true,
        env: {
          ...process.env,
          CI: "",
          E2E_LOCAL_AUTH_ENABLED: "true",
          VITE_E2E_LOCAL_AUTH_ENABLED: "true",
          PROJECT_ID: "00000000-0000-4000-8000-000000000001",
          VITE_PROJECT_ID: "00000000-0000-4000-8000-000000000001",
          VITE_PUBLISHABLE_CLIENT_KEY: "fixture-public-key",
          E2E_LOCAL_AUTH_API_BASE_URL: baseURL,
        },
      },
    );
    let logs = "";
    child.stdout.on("data", (chunk) => {
      logs += String(chunk);
    });
    child.stderr.on("data", (chunk) => {
      logs += String(chunk);
    });
    const [code] = await once(child, "exit");
    expect(code, logs).toBe(1);
    expect(
      await access(mutationPath).then(
        () => true,
        () => false,
      ),
    ).toBe(false);
    expect(logs).toContain("Browser startup failed (bootstrap)");
    expect(logs).not.toContain("private bootstrap detail");
    expect(
      await access(scenarioPath).then(
        () => true,
        () => false,
      ),
    ).toBe(false);
    expect(
      JSON.parse(
        await readFile(path.join(outputDir, "browser-startup.json"), "utf8"),
      ),
    ).toMatchObject({
      kind: "failed",
      reason: "bootstrap",
      deadlineMs: 120000,
    });
    await expect(fetch(baseURL)).rejects.toThrow();
  } finally {
    // All paths came from this test's fresh OS-temporary directory.
    await rm(scratch, { recursive: true, force: true });
  }
}, 45000);
