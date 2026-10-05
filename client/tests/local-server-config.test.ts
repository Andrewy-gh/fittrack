import { afterEach, describe, expect, it, vi } from "vitest";
import {
  readPort,
  resolveE2EServerConfig,
  resolveViteServerConfig,
} from "../scripts/local-server-config";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.resetModules();
});

describe("local server ports and API proxy", () => {
  it("preserves default ports and the same-origin API proxy", () => {
    const config = resolveViteServerConfig({});
    expect(config.server).toEqual({
      host: "127.0.0.1",
      port: 5173,
      strictPort: true,
      proxy: {
        "/api": { target: "http://127.0.0.1:8080", changeOrigin: true },
      },
    });
    expect(config.preview).toEqual({
      host: "127.0.0.1",
      port: 4173,
      strictPort: true,
      proxy: config.server.proxy,
    });
  });

  it("uses separate development and preview ports with the selected backend", () => {
    const config = resolveViteServerConfig({
      FRONTEND_PORT: "24000",
      PREVIEW_PORT: "24006",
      API_PROXY_TARGET: "http://127.0.0.1:24001",
    });
    expect(config.server.port).toBe(24000);
    expect(config.preview.port).toBe(24006);
    expect(config.server.proxy["/api"].target).toBe("http://127.0.0.1:24001");
    expect(config.preview.proxy["/api"].target).toBe("http://127.0.0.1:24001");
  });

  it.each([
    "",
    "0",
    "-1",
    "65536",
    "1.5",
    "1e3",
    "0x1405",
    "5173oops",
    " 5173 ",
  ])(
    "rejects invalid port %j instead of falling back or finding a free port",
    (value) => {
      for (const name of ["FRONTEND_PORT", "PREVIEW_PORT"]) {
        expect(() => resolveViteServerConfig({ [name]: value })).toThrow(name);
      }
      expect(() => resolveE2EServerConfig({ E2E_PORT: value })).toThrow(
        "E2E_PORT",
      );
    },
  );

  it.each(["1", "65535"])("accepts port boundary %s", (value) => {
    expect(readPort("PORT", 5173, { PORT: value })).toBe(Number(value));
  });

  it.each([
    "",
    "localhost:8080",
    "file:///tmp/api",
    "http://user:secret@localhost:8080",
  ])("rejects an invalid API proxy URL without echoing it: %j", (target) => {
    expect(() => resolveViteServerConfig({ API_PROXY_TARGET: target })).toThrow(
      "API_PROXY_TARGET must be an absolute HTTP(S) URL without credentials",
    );
  });
});

describe("Playwright server ownership", () => {
  it("starts its own strict-port development server by default", () => {
    const config = resolveE2EServerConfig({});
    expect(config.baseURL).toBe("http://127.0.0.1:5173");
    expect(config.webServer).toMatchObject({
      command: "bun run dev --host 127.0.0.1 --port 5173 --strictPort",
      url: config.baseURL,
      reuseExistingServer: false,
    });
  });

  it.each([false, true])(
    "launches the selected E2E port when CI=%s",
    (isCI) => {
      const config = resolveE2EServerConfig({
        CI: isCI ? "true" : undefined,
        FRONTEND_PORT: "24000",
        PREVIEW_PORT: "24006",
        E2E_PORT: "24003",
      });
      expect(config.baseURL).toBe("http://127.0.0.1:24003");
      expect(config.webServer).toMatchObject({
        command: `bun run ${isCI ? "serve" : "dev"} --host 127.0.0.1 --port 24003 --strictPort`,
        url: config.baseURL,
        reuseExistingServer: false,
      });
    },
  );

  it("binds localhost when it was explicitly selected", () => {
    const config = resolveE2EServerConfig({
      E2E_PORT: "24003",
      E2E_BASE_URL: "http://localhost:24003",
    });
    expect(config.webServer?.command).toContain(
      "--host localhost --port 24003",
    );
    expect(config.webServer?.url).toBe(config.baseURL);
  });

  it.each([undefined, "false", "1", "yes"])(
    "does not reuse a server with an implicit or non-true opt-in: %j",
    (value) => {
      expect(
        resolveE2EServerConfig({ E2E_REUSE_EXISTING_SERVER: value }).webServer
          ?.reuseExistingServer,
      ).toBe(false);
    },
  );

  it("allows explicit local server reuse", () => {
    expect(
      resolveE2EServerConfig({ E2E_REUSE_EXISTING_SERVER: "true" }).webServer
        ?.reuseExistingServer,
    ).toBe(true);
  });

  it.each([
    "http://127.0.0.1:5199",
    "https://example.test",
    "https://localhost:5173",
  ])("rejects a custom target without explicit reuse: %s", (baseURL) => {
    expect(() => resolveE2EServerConfig({ E2E_BASE_URL: baseURL })).toThrow(
      "E2E_REUSE_EXISTING_SERVER=true",
    );
  });

  it("uses an explicitly chosen existing custom target without starting a local server", () => {
    expect(
      resolveE2EServerConfig({
        E2E_BASE_URL: "https://example.test/app/",
        E2E_REUSE_EXISTING_SERVER: "true",
      }),
    ).toEqual({ baseURL: "https://example.test/app/", webServer: undefined });
  });

  it.each(["", "example.test", "file:///tmp/app"])(
    "rejects an invalid E2E_BASE_URL: %j",
    (baseURL) => {
      expect(() => resolveE2EServerConfig({ E2E_BASE_URL: baseURL })).toThrow(
        "E2E_BASE_URL must be an absolute HTTP(S) URL",
      );
    },
  );

  it("wires the exported environment into the actual Playwright config", async () => {
    vi.stubEnv("CI", "true");
    vi.stubEnv("E2E_PORT", "24003");
    vi.stubEnv("E2E_BASE_URL", undefined);
    vi.stubEnv("E2E_REUSE_EXISTING_SERVER", undefined);
    const { default: config } = await import("../playwright.config");
    expect(config.use?.baseURL).toBe("http://127.0.0.1:24003");
    expect(config.webServer).toMatchObject({
      command: "bun run serve --host 127.0.0.1 --port 24003 --strictPort",
      url: config.use?.baseURL,
      reuseExistingServer: false,
    });
  }, 30000);
});
