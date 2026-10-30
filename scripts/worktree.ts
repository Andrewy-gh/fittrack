#!/usr/bin/env node
// FitTrack's app adapter for the separately installed worktree-runtime Go CLI.
import { execFileSync, spawn } from "node:child_process";
import { basename, isAbsolute, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { existsSync } from "node:fs";
import { createConnection } from "node:net";

const VERSION = 1;
const MODES = ["dev", "test", "rls", "e2e", "unit"] as const;
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
  if (!isMode(mode) || mode === "unit") throw new Error("Unknown database mode.");
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
      throw new Error("Unsupported database URL query parameter; connection details withheld.");
  }
  url.pathname = `/${config.databases[mode]}`;
  return url.toString();
}

/** Derive a command environment with isolated service ports and database targets. */
export function commandEnvironment(config: LocalConfig, mode: "unit", inherited?: NodeJS.ProcessEnv): NodeJS.ProcessEnv;
export function commandEnvironment(config: LocalConfig, mode: "dev" | "test" | "rls" | "e2e", inherited?: NodeJS.ProcessEnv): NodeJS.ProcessEnv & { DATABASE_URL: string; ADMIN_DATABASE_URL: string };
export function commandEnvironment(config: LocalConfig, mode: string, inherited?: NodeJS.ProcessEnv): NodeJS.ProcessEnv;
export function commandEnvironment(
  config: LocalConfig,
  mode: string,
  inherited: NodeJS.ProcessEnv = process.env,
) {
  if (!isMode(mode)) throw new Error(`Mode must be one of: ${MODES.join(", ")}`);
  if (mode === "unit") {
    const env = { ...inherited };
    // Unit checks have no database target and must not inherit integration credentials.
    for (const key of Object.keys(env)) {
      if (/^(PG|DATABASE_URL$|ADMIN_DATABASE_URL$|RECOMMENDATION_TEST_DATABASE_URL$|FITTRACK_LOCAL_.*DATABASE_URL$|STRIPE_)/i.test(key)) delete env[key];
    }
    // Node's Request requires an absolute URL; this mode never starts a proxy.
    env.VITE_API_BASE_URL = "http://127.0.0.1:8080";
    return env;
  }
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

type Check = { readonly id: string; readonly status: "passed" | "missing" | "inaccessible" | "failed" | "not_required"; readonly repair: string };
type Readiness = { readonly ready: boolean; readonly checks: readonly Check[] };
type ReadinessOptions = { readonly mode: Mode };
function probe(command: string, args: string[], cwd: string, env: NodeJS.ProcessEnv, timeout = 15000) {
  try {
    return { ok: true as const, output: execFileSync(command, args, { cwd, env, windowsHide: true, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"], timeout, maxBuffer: 1024 * 1024 }) };
  } catch { return { ok: false as const }; }
}
function reportReadiness(report: Readiness) {
  // The adapter emits the shared protocol; native runtime owns user formatting.
  console.log(JSON.stringify({ version: 1, ...report }));
  process.exitCode = report.ready ? 0 : 1;
}

/** Diagnose only; database checks use the same endpoint/ownership guard as setup. */
async function readiness(root: string, env: NodeJS.ProcessEnv, options: ReadinessOptions, stage: "bootstrap" | "ready"): Promise<Readiness> {
  const checks: Check[] = [];
  const add = (id: string, passed: boolean, repair: string, failure: Check["status"] = "failed") => checks.push({ id, status: passed ? "passed" : failure, repair: passed ? "" : repair });
  let config: LocalConfig | undefined;
  if (env.WORKTREE_ALLOCATION) {
    try { config = localConfig(parseAllocation(JSON.parse(env.WORKTREE_ALLOCATION))); }
    catch { add("allocation", false, "Native runtime supplied an invalid allocation."); }
  }
  if (["dev", "e2e", "unit"].includes(options.mode)) {
    add("client_dependencies", existsSync(join(root, "client", "node_modules", "vitest", "package.json")), "Run bun install --frozen-lockfile in client/.", "missing");
    add("bun", probe("bun", ["--version"], root, env).ok, "Install the repository-pinned Bun runtime.", "inaccessible");
  }
  if (options.mode === "unit") {
    checks.push({ id: "database_auth_billing", status: "not_required", repair: "" });
    return { ready: checks.every(c => ["passed", "not_required"].includes(c.status)), checks };
  }
  const go = probe("go", ["version"], root, env);
  add("go", go.ok, "Install the repository-pinned Go toolchain.", "inaccessible");
  const goose = probe(env.FITTRACK_GOOSE ?? "goose", ["-version"], root, env);
  add("goose", goose.ok && goose.output.trim() === "goose version: v3.24.3", "Install Goose v3.24.3 or set FITTRACK_GOOSE in the local profile.", "inaccessible");
  const target = config ?? localConfig({ root, id: "doctor_000000000000", basePort: 21000 });
  let credentialsValid = true;
  for (const key of ["FITTRACK_LOCAL_DATABASE_URL", "FITTRACK_LOCAL_RUNTIME_DATABASE_URL"] as const) {
    if (!env[key]) { add(key, false, "Supply this local credential through the session or selected development Doppler config.", "missing"); credentialsValid = false; continue; }
    try {
      const parsed = new URL(databaseURL(env[key], target, "dev"));
      if (!parsed.username || (key === "FITTRACK_LOCAL_RUNTIME_DATABASE_URL" ? parsed.username !== "fittrack_app" : parsed.username === "fittrack_app")) throw new Error();
      add(key, true, "");
    } catch { add(key, false, "Use an explicit local owning role and fittrack_app role at 127.0.0.1:55432; redirected URLs are refused."); credentialsValid = false; }
  }
  const service = await new Promise<boolean>(done => {
    const socket = createConnection({ host: "127.0.0.1", port: 55432 });
    const finish = (ready: boolean) => { socket.destroy(); done(ready); };
    socket.setTimeout(2000, () => finish(false));
    socket.once("error", () => finish(false));
    socket.once("connect", () => finish(true));
  });
  add("postgres_service", service, "Start the documented local PostgreSQL 15 cluster at 127.0.0.1:55432. Doctor never starts or replaces services.", "inaccessible");
  if (credentialsValid && go.ok && service) {
    const checkEnv: NodeJS.ProcessEnv = { ...env, FITTRACK_DATABASE_CONFIG: JSON.stringify(target), FITTRACK_POSTGRES_MAJOR: "15" };
    for (const key of Object.keys(checkEnv)) if (/^(PGHOSTADDR|PGSERVICE|PGSERVICEFILE)$/i.test(key)) delete checkEnv[key];
    const result = probe("go", ["run", "-p=1", "./cmd/worktree-db", "--check", stage === "ready" && config ? options.mode : "bootstrap"], join(root, "server"), {
      ...checkEnv,
    }, 120000);
    // Only parse the deliberately bounded public report; never print subprocess diagnostics.
    if (result.ok) {
      try {
        const data: unknown = JSON.parse(result.output);
        if (isRecord(data) && typeof data.ready === "boolean" && typeof data.code === "string" && ["ready", "cluster_unavailable", "owner_permissions", "runtime_auth", "unsafe_runtime_role", "database_unprepared", "database_ownership", "migrations_pending", "invalid_configuration"].includes(data.code)) {
          add(data.code === "ready" ? "database" : data.code, data.ready, "Check the local cluster, role permissions, and owning credentials; run prepare for missing worktree databases/migrations.");
        } else add("database_probe", false, "Database checker returned an invalid report.");
      } catch { add("database_probe", false, "Database checker returned an invalid report."); }
    } else add("database_probe", false, "Cannot run the read-only database checker; inspect the Go toolchain and local service.", "inaccessible");
  }
  if (["dev", "e2e"].includes(options.mode)) {
    const project = env.PROJECT_ID;
    add("auth_project", Boolean(project && (!env.VITE_PROJECT_ID || env.VITE_PROJECT_ID === project)), "Supply PROJECT_ID (and matching VITE_PROJECT_ID) in the selected development configuration.", "missing");
    if (options.mode === "e2e") {
      const browser = probe(process.execPath, ["-e", "const {createRequire}=require('node:module');const {existsSync}=require('node:fs');const p=createRequire(process.cwd()+'/package.json')('playwright');process.exit(existsSync(p.chromium.executablePath())?0:1)"], join(root, "client"), env);
      add("chromium", browser.ok, "Run bunx playwright install chromium in client/.", "missing");
      const enabled = (value: string | undefined) => ["1", "true", "yes", "on"].includes((value ?? "").trim().toLowerCase());
      const serverLocal = enabled(env.E2E_LOCAL_AUTH_ENABLED), clientLocal = enabled(env.VITE_E2E_LOCAL_AUTH_ENABLED);
      const local = serverLocal && clientLocal;
      const provider = Boolean(env.VITE_PROJECT_ID && env.VITE_PUBLISHABLE_CLIENT_KEY && (env.SECRET_SERVER_KEY || (env.E2E_STACK_EMAIL && env.E2E_STACK_PASSWORD)));
      add("e2e_auth_configuration", (serverLocal || clientLocal) ? local : provider, "Set both E2E_LOCAL_AUTH_ENABLED and VITE_E2E_LOCAL_AUTH_ENABLED for local auth, or configure Stack test login. This checks configuration, not a completed login.", "missing");
    }
  } else checks.push({ id: "auth", status: "not_required", repair: "" });
  checks.push({ id: "billing", status: "not_required", repair: "Billing provider access, webhook forwarding and authenticated checkout require separate scenario verification." });
  return { ready: checks.every(c => ["passed", "not_required"].includes(c.status)), checks };
}

async function main(args: string[]) {
  if (!args.length || args[0] === "help" || args[0] === "--help") {
    console.log("FitTrack adapter: [--profile <absolute-path>] init | status | check | doctor --mode <mode> [--json] | prepare --mode <mode> [--major 15] [--json] | db setup --major <verified-major> | stop <mode> | run <dev|test|rls|e2e|unit> [--log-file path] -- <command> [arguments]\nRequires the external worktree-runtime Go CLI. See docs/worktrees.md.");
    return;
  }
  if (args[0] === "--exec") {
    const mode = args[1], command = args[3];
    if (!isMode(mode) || args[2] !== "--" || !command || process.env.WORKTREE_MODE !== mode)
      throw new Error("Invalid managed FitTrack command.");
    const cwd = process.env.WORKTREE_COMMAND_CWD;
    if (!cwd || !isAbsolute(cwd)) throw new Error("Missing managed command directory.");
    // Bun can restore deleted variables from .env when it executes a file/eval.
    // Disable its automatic reload for unit launches; other modes retain it.
    const commandArgs = mode === "unit" && /^bun(?:\.exe)?$/i.test(basename(command))
      ? ["--no-env-file", ...args.slice(4)] : args.slice(4);
    await execute(command, commandArgs, cwd, commandEnvironment(assignedConfig(), mode));
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
  if (args[0] === "--doctor") {
    const mode = process.env.WORKTREE_MODE, stage = process.env.WORKTREE_DOCTOR_STAGE;
    if (!isMode(mode) || (stage !== "ready" && stage !== "bootstrap")) throw new Error("Invoke doctor through the native worktree-runtime CLI.");
    reportReadiness(await readiness(process.cwd(), process.env, { mode }, stage));
    return;
  }
  // The bridge only translates FitTrack syntax and presents FitTrack endpoints.
  // Profiles, readiness protocol and prepare orchestration belong to the native CLI.
  const forwarded = [...args];
  let commandIndex = 0;
  while (forwarded[commandIndex] === "--profile") commandIndex += 2;
  const command = forwarded[commandIndex];
  if (command === "doctor" || command === "prepare") {
    const major = forwarded.indexOf("--major");
    if (major >= 0) {
      if (forwarded[major + 1] !== "15") throw new Error("FitTrack preparation requires its pinned PostgreSQL major 15.");
      forwarded.splice(major, 2);
    }
  }
  if (command === "db") {
    if (forwarded[commandIndex + 1] !== "setup") throw new Error("Expected db setup --major <verified-production-major>.");
    forwarded.splice(commandIndex, 2, "setup");
  }
  if (command === "status") {
    const result = await runtimeCommand(forwarded, process.cwd(), process.env, true);
    if (result.code) { process.exitCode = result.code; return; }
    console.log(JSON.stringify(localConfig(parseAllocation(JSON.parse(result.output))), null, 2));
    return;
  }
  process.exitCode = (await runtimeCommand(forwarded)).code;
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  main(process.argv.slice(2)).catch((error: unknown) => {
    console.error(error instanceof Error ? error.message : "FitTrack adapter failed.");
    process.exitCode = 1;
  });
}
