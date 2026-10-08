#!/usr/bin/env node
// FitTrack's app adapter for the separately installed worktree-runtime Go CLI.
import { execFileSync, spawn } from "node:child_process";
import { isAbsolute, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const VERSION = 1;
const MODES = ["dev", "test", "rls", "e2e"] as const;
type Mode = (typeof MODES)[number];
type Allocation = { readonly root: string; readonly id: string; readonly basePort: number };
type LocalConfig = ReturnType<typeof localConfig>;

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
function isMode(value: string | undefined): value is Mode {
  return MODES.some((mode) => mode === value);
}
function parseAllocation(value: unknown): Allocation {
  if (!isRecord(value) || value.version !== VERSION || typeof value.root !== "string" ||
      !isAbsolute(value.root) || typeof value.id !== "string" ||
      !/^[a-z0-9_]+_[a-f0-9]{12}$/.test(value.id) || typeof value.basePort !== "number" ||
      !Number.isInteger(value.basePort) || value.basePort < 21000 || value.basePort > 29990 ||
      (value.basePort - 21000) % 10 !== 0) {
    throw new Error("Unsupported external runtime allocation; use worktree-runtime protocol 1.");
  }
  return { root: value.root, id: value.id, basePort: value.basePort };
}
function assignedConfig() {
  try {
    return localConfig(parseAllocation(JSON.parse(process.env.WORKTREE_ALLOCATION ?? "")));
  } catch {
    throw new Error("Invoke this app adapter through the external worktree-runtime CLI.");
  }
}

/** Build the credential-free database names and service ports for an allocation. */
export function localConfig(entry: Allocation) {
  const { root, id, basePort } = entry;
  return {
    version: VERSION,
    root,
    id,
    postgres: { host: "127.0.0.1", port: 55432 },
    databases: {
      dev: `ft_${id}_dev`,
      test: `ft_${id}_test`,
      rls: `ft_${id}_rls`,
      e2e: `ft_${id}_e2e`,
    },
    ports: {
      frontend: basePort,
      api: basePort + 1,
      metrics: basePort + 2,
      e2eFrontend: basePort + 3,
      e2eApi: basePort + 4,
      e2eMetrics: basePort + 5,
      preview: basePort + 6,
    },
  };
}

/** Select a worktree database while refusing remote or redirected connection templates. */
export function databaseURL(template: string | undefined, config: LocalConfig, mode: string) {
  if (!isMode(mode)) throw new Error("Unknown database mode.");
  if (!template)
    throw new Error(
      "Set FITTRACK_LOCAL_DATABASE_URL to an existing local owning-role connection URL. Credentials are never created or saved by this helper.",
    );
  let url;
  try {
    url = new URL(template);
  } catch {
    throw new Error("FITTRACK_LOCAL_DATABASE_URL must be a PostgreSQL URL.");
  }
  if (
    !["postgres:", "postgresql:"].includes(url.protocol) ||
    url.hostname !== config.postgres.host ||
    url.port !== String(config.postgres.port) ||
    url.hash
  ) {
    throw new Error(
      "Database templates must use PostgreSQL at 127.0.0.1:55432. Remote databases and alternate hosts/ports are refused.",
    );
  }
  // libpq/pgx accept query parameters that override the URL host and database.
  for (const key of url.searchParams.keys()) {
    if (!["sslmode", "connect_timeout"].includes(key))
      throw new Error(`Unsupported database URL query parameter: ${key}`);
  }
  url.pathname = `/${config.databases[mode]}`;
  return url.toString();
}

/** Derive a command environment with isolated service ports and database targets. */
export function commandEnvironment(
  config: LocalConfig,
  mode: string,
  inherited: NodeJS.ProcessEnv = process.env,
) {
  if (!isMode(mode)) throw new Error(`Mode must be one of: ${MODES.join(", ")}`);
  const isE2E = mode === "e2e";
  const frontend = isE2E ? config.ports.e2eFrontend : config.ports.frontend;
  const api = isE2E ? config.ports.e2eApi : config.ports.api;
  const env: NodeJS.ProcessEnv & { DATABASE_URL: string; ADMIN_DATABASE_URL: string } = {
    ...inherited,
    FITTRACK_WORKTREE_ID: config.id,
    DATABASE_URL: databaseURL(inherited.FITTRACK_LOCAL_DATABASE_URL, config, mode),
    // Clear inherited alternative database targets before tests read them.
    ADMIN_DATABASE_URL: "",
    RECOMMENDATION_TEST_DATABASE_URL: "",
    PGHOST: config.postgres.host,
    PGPORT: String(config.postgres.port),
    PGDATABASE: config.databases[mode],
    PORT: String(api),
    METRICS_PORT: String(isE2E ? config.ports.e2eMetrics : config.ports.metrics),
    FRONTEND_PORT: String(frontend),
    PREVIEW_PORT: String(isE2E ? frontend : config.ports.preview),
    API_PROXY_TARGET: `http://127.0.0.1:${api}`,
    VITE_API_BASE_URL: "/api",
    APP_BASE_URL: `http://127.0.0.1:${frontend}`,
    E2E_PORT: String(config.ports.e2eFrontend),
    E2E_BASE_URL: `http://127.0.0.1:${config.ports.e2eFrontend}`,
    E2E_LOCAL_AUTH_API_BASE_URL: `http://127.0.0.1:${config.ports.e2eApi}`,
    E2E_REUSE_EXISTING_SERVER: "false",
    ENVIRONMENT: "development",
    RLS_ENFORCEMENT_REQUIRED: "false",
  };
  // libpq treats PGSERVICE="" as a request for a service named "", not unset.
  for (const key of Object.keys(env)) {
    if (/^(PGHOSTADDR|PGSERVICE|PGSERVICEFILE)$/i.test(key)) delete env[key];
  }
  if (["test", "rls"].includes(mode)) env.GOFLAGS = `${inherited.GOFLAGS ?? ""} -p=1`.trim();
  if (mode === "rls") {
    if (!inherited.FITTRACK_LOCAL_RUNTIME_DATABASE_URL)
      throw new Error(
        "RLS commands require FITTRACK_LOCAL_RUNTIME_DATABASE_URL for the already-provisioned fittrack_app role.",
      );
    const runtimeURL = databaseURL(inherited.FITTRACK_LOCAL_RUNTIME_DATABASE_URL, config, mode);
    if (new URL(runtimeURL).username !== "fittrack_app")
      throw new Error("RLS runtime URL must use the existing fittrack_app role.");
    env.ADMIN_DATABASE_URL = env.DATABASE_URL;
    env.DATABASE_URL = runtimeURL;
    env.RECOMMENDATION_TEST_DATABASE_URL = runtimeURL;
    env.RLS_ENFORCEMENT_REQUIRED = "true";
  }
  return env;
}

// This bridge contains no allocation, lock, process-tree or recovery policy.
function runtimeCommand(args: string[], cwd = process.cwd(), inherited = process.env, capture = false) {
  return new Promise<{ readonly output: string; readonly code: number }>((done, reject) => {
    const child = spawn(inherited.WORKTREE_RUNTIME ?? "worktree-runtime", args, {
      cwd,
      env: { ...inherited, WORKTREE_RUNTIME_PARENT_PIPE: "1" },
      windowsHide: true,
      // libuv's implicit Windows job would kill the runtime itself with this
      // wrapper, preventing its confirmed cleanup receipt and lock release.
      detached: process.platform === "win32",
      stdio: ["pipe", capture ? "pipe" : "inherit", "inherit"],
    });
    child.stdin?.on("error", () => {});
    let output = "";
    child.stdout?.on("data", (chunk) => { output += chunk; });
    const cancel = () => child.stdin?.end();
    process.on("SIGINT", cancel);
    process.on("SIGTERM", cancel);
    const cleanup = () => {
      process.off("SIGINT", cancel);
      process.off("SIGTERM", cancel);
      child.stdin?.end();
    };
    child.once("error", () => {
      cleanup();
      reject(new Error("Install the external worktree-runtime CLI on PATH or set WORKTREE_RUNTIME to its native executable."));
    });
    child.once("exit", (code) => {
      cleanup();
      done({ output, code: code ?? 130 });
    });
  });
}

/** Read the external allocation and project it into FitTrack's local endpoints. */
export async function loadConfig(cwd = process.cwd(), inherited: NodeJS.ProcessEnv = process.env) {
  const result = await runtimeCommand(["status", "--json"], cwd, inherited, true);
  if (result.code) throw new Error("External allocation could not be loaded.");
  const value: unknown = JSON.parse(result.output);
  return localConfig(parseAllocation(value));
}

function nativeCommand(command: string) {
  if (process.platform !== "win32") return command;
  try {
    const path = isAbsolute(command) ? command :
      execFileSync("where.exe", [command], { encoding: "utf8", windowsHide: true,
        stdio: ["ignore", "pipe", "ignore"] }).trim().split(/\r?\n/)
        .find((candidate) => candidate.toLowerCase().endsWith(".exe"));
    if (path?.toLowerCase().endsWith(".exe")) return path;
  } catch {}
  throw new Error("ENOENT: use a native .exe or an explicit shell; .cmd/.bat files are not implicitly interpreted.");
}

function execute(command: string, args: string[], cwd: string, env: NodeJS.ProcessEnv) {
  return new Promise<void>((done, reject) => {
    const child = spawn(nativeCommand(command), args, { cwd, env, windowsHide: true, stdio: "inherit" });
    child.once("error", () => reject(new Error("App command could not start; arguments and environment withheld.")));
    // Observe exit, not pipe EOF: grandchildren are contained by the Go runtime.
    child.once("exit", (code) => { process.exitCode = code ?? 130; done(); });
  });
}

async function main(args: string[]) {
  if (!args.length || args[0] === "help" || args[0] === "--help") {
    console.log("FitTrack adapter: init | status | check | db setup --major <verified-major> | stop <mode> | run <dev|test|rls|e2e> [--log-file path] -- <command> [arguments]\nRequires the external worktree-runtime Go CLI. See docs/worktrees.md.");
    return;
  }
  if (args[0] === "--exec") {
    const mode = args[1], command = args[3];
    if (!isMode(mode) || args[2] !== "--" || !command || process.env.WORKTREE_MODE !== mode)
      throw new Error("Invalid managed FitTrack command.");
    const cwd = process.env.WORKTREE_COMMAND_CWD;
    if (!cwd || !isAbsolute(cwd)) throw new Error("Missing managed command directory.");
    await execute(command, args.slice(4), cwd, commandEnvironment(assignedConfig(), mode));
    return;
  }
  if (args[0] === "--setup") {
    if (args.length !== 3 || args[1] !== "--major" || !/^[1-9][0-9]$/.test(args[2] ?? ""))
      throw new Error("Expected db setup --major <verified-production-major>.");
    const config = assignedConfig();
    const env = commandEnvironment(config, "dev");
    env.FITTRACK_DATABASE_CONFIG = JSON.stringify(config);
    env.FITTRACK_POSTGRES_MAJOR = args[2];
    await execute("go", ["run", "-p=1", "./cmd/worktree-db"], join(config.root, "server"), env);
    return;
  }
  if (args[0] === "status") {
    if (args.length !== 1) throw new Error("Expected status.");
    console.log(JSON.stringify(await loadConfig(), null, 2));
    return;
  }
  if (args[0] === "db") {
    if (args[1] !== "setup") throw new Error("Expected db setup --major <verified-production-major>.");
    process.exitCode = (await runtimeCommand(["setup", ...args.slice(2)])).code;
    return;
  }
  if (!["init", "check", "stop", "run"].includes(args[0] ?? ""))
    throw new Error("Unknown FitTrack adapter command.");
  process.exitCode = (await runtimeCommand(args)).code;
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  main(process.argv.slice(2)).catch((error: unknown) => {
    console.error(error instanceof Error ? error.message : "FitTrack adapter failed.");
    process.exitCode = 1;
  });
}
