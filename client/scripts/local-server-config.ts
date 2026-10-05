type Environment = Record<string, string | undefined>;

export function readPort(
  name: string,
  fallback: number,
  env: Environment = process.env,
): number {
  const value = env[name];
  if (value === undefined) return fallback;

  if (!/^\d+$/.test(value) || Number(value) < 1 || Number(value) > 65535) {
    throw new Error(`${name} must be an integer port between 1 and 65535`);
  }
  return Number(value);
}

function readHttpUrl(name: string, value: string): URL {
  try {
    const url = new URL(value);
    if (
      (url.protocol === "http:" || url.protocol === "https:") &&
      !url.username &&
      !url.password
    ) {
      return url;
    }
  } catch {
    // Report the setting name without echoing a potentially sensitive value.
  }
  throw new Error(`${name} must be an absolute HTTP(S) URL without credentials`);
}

export function resolveViteServerConfig(env: Environment = process.env) {
  const target = env.API_PROXY_TARGET ?? "http://127.0.0.1:8080";
  readHttpUrl("API_PROXY_TARGET", target);
  const proxy = { "/api": { target, changeOrigin: true } };

  return {
    server: {
      host: "127.0.0.1",
      port: readPort("FRONTEND_PORT", 5173, env),
      strictPort: true,
      proxy,
    },
    preview: {
      host: "127.0.0.1",
      port: readPort("PREVIEW_PORT", 4173, env),
      strictPort: true,
      proxy,
    },
  };
}

export function resolveE2EServerConfig(env: Environment = process.env) {
  const port = readPort("E2E_PORT", 5173, env);
  const baseURL = env.E2E_BASE_URL ?? `http://127.0.0.1:${port}`;
  const url = readHttpUrl("E2E_BASE_URL", baseURL);
  const reuseExistingServer = env.E2E_REUSE_EXISTING_SERVER === "true";
  const isManagedOrigin =
    url.protocol === "http:" &&
    ["127.0.0.1", "localhost"].includes(url.hostname) &&
    Number(url.port || "80") === port;

  if (!isManagedOrigin) {
    if (!reuseExistingServer) {
      throw new Error(
        "E2E_BASE_URL must use HTTP localhost or 127.0.0.1 on E2E_PORT. " +
          "To test an already-running custom target, explicitly set E2E_REUSE_EXISTING_SERVER=true.",
      );
    }
    return { baseURL, webServer: undefined };
  }

  return {
    baseURL,
    webServer: {
      command:
        `bun run ${env.CI ? "serve" : "dev"} --host ${url.hostname} ` +
        `--port ${port} --strictPort`,
      url: baseURL,
      reuseExistingServer,
      timeout: 120000,
      stdout: "pipe" as const,
      stderr: "pipe" as const,
    },
  };
}
