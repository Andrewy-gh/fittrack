#!/usr/bin/env node
// Local-only worktree coordination. See docs/worktrees.md for the supported entry points.
import { createHash, randomUUID } from "node:crypto";
import { execFile, execFileSync, spawn } from "node:child_process";
import {
  access,
  mkdir,
  readFile,
  readdir,
  realpath,
  rename,
  rm,
  writeFile,
} from "node:fs/promises";
import type { SpawnOptions } from "node:child_process";
import type { Server } from "node:net";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import { promisify } from "node:util";
import { basename, dirname, isAbsolute, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const VERSION = 1;
const FIRST_PORT = 21000;
const LAST_PORT = 29999;
const BLOCK_SIZE = 10;
const MODES = ["dev", "test", "rls", "e2e"] as const;
type Mode = (typeof MODES)[number];
type Allocation = { readonly root: string; readonly id: string; readonly basePort: number };
type Context = { readonly root: string; readonly stateDir: string; readonly configPath: string };
type Registry = { readonly version: number; readonly entries: Allocation[] };
type LocalConfig = ReturnType<typeof localConfig>;
type Release = () => Promise<void>;
type WindowsRequest =
  | { readonly action: "reserve"; readonly port: number }
  | {
      readonly action: "run";
      readonly receipt: string;
      readonly command: string;
      readonly args: string[];
    };
type WindowsOptions = Pick<SpawnOptions, "cwd" | "env"> & { readonly stdout?: "pipe" | "inherit" };
type ExecutionOptions = {
  readonly cwd: string;
  readonly env: NodeJS.ProcessEnv;
  readonly cancelPath?: string;
  readonly receiptPath: string;
  readonly onUncertain: () => void;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
function errorCode(error: unknown): unknown {
  return isRecord(error) ? error.code : undefined;
}
function isMode(value: string | undefined): value is Mode {
  return MODES.some((mode) => mode === value);
}
const LOCAL_DIRECTORY = ".worktree";
const windowsScript = fileURLToPath(new URL("./worktree-windows.ps1", import.meta.url));

async function windowsExecutable() {
  const source = await readFile(new URL("./worktree-windows.cs", import.meta.url));
  const hash = createHash("sha256").update(source).update(process.arch).digest("hex");
  const directory = join(tmpdir(), "fittrack-worktree-native");
  const executable = join(directory, `${hash}.exe`);
  const exists = () =>
    access(executable).then(
      () => true,
      () => false,
    );
  if (await exists()) return executable;
  await withLock(
    `${executable}.lock`,
    async () => {
      if (await exists()) return;
      const temporary = join(directory, `${randomUUID()}.exe`);
      try {
        await promisify(execFile)(
          join(
            process.env.SystemRoot ?? "C:\\Windows",
            "System32",
            "WindowsPowerShell",
            "v1.0",
            "powershell.exe",
          ),
          [
            "-NoLogo",
            "-NoProfile",
            "-NonInteractive",
            "-ExecutionPolicy",
            "Bypass",
            "-File",
            windowsScript,
            "-OutputPath",
            temporary,
          ],
          { windowsHide: true },
        );
        await rename(temporary, executable);
      } finally {
        await rm(temporary, { force: true });
      }
    },
    () => true,
    30000,
  );
  return executable;
}

function windowsSupervisor(
  executable: string,
  request: WindowsRequest,
  options: WindowsOptions = {},
) {
  const args =
    request.action === "reserve"
      ? ["reserve", String(request.port)]
      : ["run", request.receipt, request.command, ...request.args];
  const child = spawn(executable, args, {
    windowsHide: true,
    ...options,
    stdio: ["pipe", options.stdout ?? "inherit", "inherit"],
  });
  child.stdin?.on("error", () => {});
  return child;
}

async function reserveWindows(basePort: number) {
  const child = windowsSupervisor(
    await windowsExecutable(),
    { action: "reserve", port: basePort },
    { stdout: "pipe" },
  );
  const exited = new Promise<number | null>((done, reject) => {
    child.once("error", reject);
    child.once("exit", done);
  });
  const release = async () => {
    child.stdin?.end();
    await exited;
  };
  try {
    const status = await new Promise<string>((done, reject) => {
      let output = "";
      child.stdout?.on("data", (chunk) => {
        output += chunk;
        if (output.includes("\n")) done(output.trim());
      });
      exited.then(
        () => reject(new Error("Windows port supervisor exited before reserving sockets.")),
        reject,
      );
    });
    if (status !== "READY")
      throw new Error(`Port block ${basePort} cannot be reserved (${status}).`, {
        cause: { code: status },
      });
    return release;
  } catch (error) {
    await release();
    throw error;
  }
}

/** Derive a stable database-safe identifier from the canonical checkout path. */
export function worktreeId(root: string) {
  const label =
    basename(root)
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "_")
      .replace(/^_+|_+$/g, "")
      .slice(0, 20) || "worktree";
  return `${label}_${createHash("sha256").update(root).digest("hex").slice(0, 12)}`;
}

function allocation(root: string, basePort: number): Allocation {
  const id = worktreeId(root);
  return { root, id, basePort };
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

/** Discover the canonical checkout and shared Git coordination directory. */
export async function discover(cwd = process.cwd()): Promise<Context> {
  const git = (...args: string[]) => execFileSync("git", args, { cwd, encoding: "utf8" }).trim();
  const root = await realpath(git("rev-parse", "--show-toplevel"));
  const common = await realpath(resolve(cwd, git("rev-parse", "--git-common-dir")));
  return {
    root,
    stateDir: join(common, "fittrack-worktrees"),
    configPath: join(root, LOCAL_DIRECTORY, "config.json"),
  };
}

async function readJSON(path: string, fallback?: unknown): Promise<unknown> {
  try {
    return JSON.parse(await readFile(path, "utf8"));
  } catch (error) {
    if (errorCode(error) === "ENOENT" && fallback !== undefined) return fallback;
    throw new Error(
      `Cannot read ${path}: ${error instanceof Error ? error.message : "unknown read error"}`,
    );
  }
}

async function atomicJSON(path: string, value: unknown) {
  await mkdir(dirname(path), { recursive: true });
  const temporary = `${path}.${randomUUID()}.tmp`;
  try {
    await writeFile(temporary, `${JSON.stringify(value, null, 2)}\n`, { flag: "wx" });
    await rename(temporary, path);
  } finally {
    await rm(temporary, { force: true });
  }
}

/** Run an operation under an exclusive directory lock without stealing stale locks. */
export async function withLock<T>(
  path: string,
  action: () => Promise<T>,
  shouldRelease = () => true,
  waitMs = 0,
): Promise<T> {
  await mkdir(dirname(path), { recursive: true });
  const deadline = Date.now() + waitMs;
  for (;;) {
    try {
      await mkdir(path);
      break;
    } catch (error) {
      if (errorCode(error) !== "EEXIST") throw error;
      if (Date.now() >= deadline)
        throw new Error(
          `Lock already held: ${path}. Wait for the other command. After a crash, verify its process has stopped before removing this lock directory.`,
        );
      await delay(100);
    }
  }
  try {
    await writeFile(
      join(path, "owner.json"),
      JSON.stringify({ pid: process.pid, startedAt: new Date().toISOString() }),
    );
    return await action();
  } finally {
    if (shouldRelease()) await rm(path, { recursive: true, force: true });
  }
}

function parseRegistry(value: unknown): Registry {
  if (!isRecord(value) || value.version !== VERSION || !Array.isArray(value.entries))
    throw new Error("Unsupported worktree registry. Refusing to overwrite it.");
  const roots = new Set<string>();
  const ids = new Set<string>();
  const ports = new Set<number>();
  const entries: Allocation[] = [];
  for (const candidate of value.entries) {
    const entry: unknown = candidate;
    if (
      !isRecord(entry) ||
      typeof entry.root !== "string" ||
      typeof entry.id !== "string" ||
      entry.id !== worktreeId(entry.root) ||
      typeof entry.basePort !== "number" ||
      !Number.isInteger(entry.basePort) ||
      entry.basePort < FIRST_PORT ||
      entry.basePort + BLOCK_SIZE - 1 > LAST_PORT ||
      (entry.basePort - FIRST_PORT) % BLOCK_SIZE !== 0 ||
      roots.has(entry.root) ||
      ids.has(entry.id) ||
      ports.has(entry.basePort)
    ) {
      throw new Error(
        "Invalid or conflicting worktree registry entry. Refusing to change allocations.",
      );
    }
    roots.add(entry.root);
    ids.add(entry.id);
    ports.add(entry.basePort);
    entries.push({ root: entry.root, id: entry.id, basePort: entry.basePort });
  }
  return { version: VERSION, entries };
}

async function readRegistry(stateDir: string) {
  const registry = await readJSON(join(stateDir, "registry.json"), {
    version: VERSION,
    entries: [],
  });
  return parseRegistry(registry);
}

async function listen(port: number, host: string) {
  const server = createServer();
  await new Promise<void>((resolveListen, reject) => {
    server.once("error", reject);
    server.listen({ port, host, exclusive: true, ipv6Only: host === "::" }, resolveListen);
  });
  return server;
}

// Hold both address-family reservations until every port has been checked.
/** Reserve all ten ports until the returned release function is called. */
export async function reserveBlock(basePort: number): Promise<Release> {
  if (process.platform === "win32") return reserveWindows(basePort);
  const servers: Server[] = [];
  const release = async () => {
    await Promise.all(
      servers.map((server) => new Promise<void>((done) => server.close(() => done()))),
    );
  };
  try {
    for (let port = basePort; port < basePort + BLOCK_SIZE; port++) {
      for (const host of ["0.0.0.0", "::"]) {
        try {
          servers.push(await listen(port, host));
        } catch (error) {
          if (
            host === "::" &&
            ["EAFNOSUPPORT", "EADDRNOTAVAIL"].some((code) => code === errorCode(error))
          )
            continue;
          throw new Error(
            `Port ${port} cannot be reserved (${String(errorCode(error))}). Stop the conflicting listener; assignments never silently change.`,
            { cause: error },
          );
        }
      }
    }
    return release;
  } catch (error) {
    await release();
    throw error;
  }
}

/** Allocate or recover a checkout configuration while holding its port reservations. */
export async function initialize(
  context: Context,
  reserve: (port: number) => Promise<Release> = reserveBlock,
) {
  // Allocation is short-lived; concurrent initializers wait, but never steal locks.
  const lock = join(context.stateDir, "registry.lock");
  return withLock(
    lock,
    async () => {
      const registry = await readRegistry(context.stateDir);
      let entry = registry.entries.find((candidate) => candidate.root === context.root);
      let release: Release | undefined;
      if (entry) {
        release = await reserve(entry.basePort);
      } else {
        for (let port = FIRST_PORT; port + BLOCK_SIZE - 1 <= LAST_PORT; port += BLOCK_SIZE) {
          if (registry.entries.some((candidate) => candidate.basePort === port)) continue;
          try {
            release = await reserve(port);
            entry = allocation(context.root, port);
            break;
          } catch (error) {
            if (!(error instanceof Error && errorCode(error.cause) === "EADDRINUSE")) throw error;
          }
        }
        if (!entry) throw new Error("No free ten-port worktree block remains in 21000â€“29999.");
        const allocated = entry;
        if (registry.entries.some((candidate) => candidate.id === allocated.id))
          throw new Error("Worktree ID collision; refusing allocation.");
        registry.entries.push(entry);
      }
      if (!entry || !release) throw new Error("Port allocation did not complete.");
      try {
        // Registry first: an interrupted local-file write can be retried without reallocating.
        await atomicJSON(join(context.stateDir, "registry.json"), registry);
        const config = localConfig(entry);
        await atomicJSON(context.configPath, config);
        return config;
      } finally {
        await release();
      }
    },
    () => true,
    5000,
  );
}

/** Load an allocation only when the local configuration matches the shared registry. */
export async function loadConfig(context: Context) {
  const registry = await readRegistry(context.stateDir);
  const entry = registry.entries.find((candidate) => candidate.root === context.root);
  if (!entry)
    throw new Error("This worktree has no allocation. Run: node scripts/worktree.ts init");
  const expected = localConfig(entry);
  const actual = await readJSON(context.configPath);
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(
      "Worktree config differs from the registry. Stop this worktree's services and rerun init; never copy another worktree's .worktree directory.",
    );
  }
  return expected;
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

function signalGroup(pid: number, signal: NodeJS.Signals | 0) {
  try {
    process.kill(-pid, signal);
    return true;
  } catch (error) {
    if (errorCode(error) === "ESRCH") return false;
    throw error;
  }
}

const delay = (milliseconds: number) => new Promise<void>((done) => setTimeout(done, milliseconds));

async function executeOwned(
  command: string,
  args: string[],
  { cwd, env, cancelPath, receiptPath, onUncertain }: ExecutionOptions,
): Promise<number> {
  const executable = process.platform === "win32" ? await windowsExecutable() : undefined;
  return new Promise<number>((resolveExit, reject) => {
    // A separate group lets cancellation reach go/bun/shell grandchildren too.
    const windows = process.platform === "win32";
    if (windows) {
      try {
        if (!isAbsolute(command))
          command =
            execFileSync("where.exe", [command], {
              encoding: "utf8",
              windowsHide: true,
              stdio: ["ignore", "pipe", "ignore"],
            })
              .trim()
              .split(/\r?\n/)
              .find((path) => path.toLowerCase().endsWith(".exe")) ?? "";
      } catch {
        reject(
          new Error(
            "ENOENT: executable not found. Use a native executable (node, bun, go) or an explicit shell.",
          ),
        );
        return;
      }
      if (!command?.toLowerCase().endsWith(".exe")) {
        reject(
          new Error(
            "Use a native .exe or an explicit shell; .cmd/.bat files are not implicitly interpreted.",
          ),
        );
        return;
      }
    }
    const child = executable
      ? windowsSupervisor(
          executable,
          { action: "run", command, args, receipt: receiptPath },
          { cwd, env },
        )
      : spawn(command, args, { cwd, env, stdio: "inherit", detached: true });
    let escalation: ReturnType<typeof setTimeout> | undefined;
    const cancel = (signal: NodeJS.Signals) => {
      if (!child.pid) return;
      if (windows) {
        child.stdin?.end();
        return;
      }
      const pid = child.pid;
      signalGroup(pid, signal);
      escalation ??= setTimeout(() => signalGroup(pid, "SIGKILL"), 2000);
    };
    const onInterrupt = () => cancel("SIGINT");
    const onTerminate = () => cancel("SIGTERM");
    const cancellation = cancelPath
      ? setInterval(async () => {
          try {
            await readFile(cancelPath);
            cancel("SIGTERM");
          } catch {}
        }, 100)
      : undefined;
    process.on("SIGINT", onInterrupt);
    process.on("SIGTERM", onTerminate);
    const cleanup = () => {
      clearTimeout(escalation);
      clearInterval(cancellation);
      process.off("SIGINT", onInterrupt);
      process.off("SIGTERM", onTerminate);
    };
    child.once("error", (error) => {
      cleanup();
      reject(error);
    });
    child.once("exit", async (code, signal) => {
      try {
        // Even a normally-exiting shell can leave a background database writer.
        if (windows) {
          const drained = await readFile(receiptPath, "utf8").catch(() => "");
          if (drained !== "drained") {
            onUncertain();
            console.error(
              "Windows job cleanup was not confirmed; coordination records are retained for manual recovery.",
            );
            if (code === 0) code = 1;
          }
        }
        if (!windows && child.pid !== undefined && signalGroup(child.pid, 0)) {
          signalGroup(child.pid, "SIGTERM");
          await delay(250);
          if (signalGroup(child.pid, 0)) {
            signalGroup(child.pid, "SIGKILL");
            await delay(250);
          }
          if (signalGroup(child.pid, 0)) {
            onUncertain();
            console.error(
              `Process group ${child.pid} has not fully disappeared. Its coordination records are retained; verify ownership and termination before recovery.`,
            );
            if (code === 0) code = 1;
          }
        }
        resolveExit(code ?? (signal === "SIGINT" ? 130 : 143));
      } catch (error) {
        onUncertain();
        reject(error);
      } finally {
        cleanup();
      }
    });
  });
}

/** Run and supervise a command, retaining coordination records if cleanup is uncertain. */
export async function runCommand(
  context: Context,
  mode: string,
  command: string | undefined,
  args: string[],
) {
  if (!command) throw new Error("Expected: run <dev|test|rls|e2e> -- <command> [arguments]");
  const config = await loadConfig(context);
  const env = commandEnvironment(config, mode);
  let releaseLock = true;
  const lockPath = join(context.stateDir, `${config.id}-${mode}.lock`);
  const runPath = join(context.stateDir, `${config.id}-runs`, randomUUID());
  await withLock(
    join(context.stateDir, `${config.id}-database.lock`),
    async () => {
      await mkdir(runPath, { recursive: true });
      await atomicJSON(join(runPath, "owner.json"), { pid: process.pid, mode });
    },
    () => true,
    5000,
  );
  const execute = () =>
    executeOwned(command, args, {
      cwd: process.cwd(),
      env,
      cancelPath: join(runPath, "cancel"),
      receiptPath: join(runPath, "drained"),
      onUncertain: () => {
        releaseLock = false;
      },
    });
  try {
    return await (["test", "rls"].includes(mode)
      ? withLock(lockPath, execute, () => releaseLock)
      : execute());
  } finally {
    if (releaseLock) await rm(runPath, { recursive: true, force: true });
  }
}

async function stopCommands(context: Context, mode: string | undefined) {
  if (!isMode(mode)) throw new Error("Expected stop <dev|test|rls|e2e>");
  const config = await loadConfig(context);
  const directory = join(context.stateDir, `${config.id}-runs`);
  await withLock(
    join(context.stateDir, `${config.id}-database.lock`),
    async () => {
      for (const id of await readdir(directory).catch((error) => {
        if (errorCode(error) === "ENOENT") return [];
        throw error;
      })) {
        const owner = await readJSON(join(directory, id, "owner.json"), null);
        if (isRecord(owner) && owner.mode === mode)
          await writeFile(join(directory, id, "cancel"), "cancel\n").catch((error) => {
            if (errorCode(error) !== "ENOENT") throw error;
          });
      }
    },
    () => true,
    5000,
  );
  console.log(
    `Cancellation requested for this worktree's ${mode} commands. Wait for their exit before restarting.`,
  );
}

async function setupDatabases(context: Context, major: string | undefined) {
  if (!major || !/^[1-9][0-9]$/.test(major))
    throw new Error(
      "Expected db setup --major <verified-production-major>. No default version is assumed.",
    );
  const config = await loadConfig(context);
  const env = commandEnvironment(config, "dev");
  env.FITTRACK_DATABASE_CONFIG = JSON.stringify(config);
  env.FITTRACK_POSTGRES_MAJOR = major;
  let releaseLock = true;
  await withLock(
    join(context.stateDir, `${config.id}-database.lock`),
    async () => {
      const runs = await readdir(join(context.stateDir, `${config.id}-runs`)).catch((error) => {
        if (errorCode(error) === "ENOENT") return [];
        throw error;
      });
      if (runs.length)
        throw new Error(
          "Stop this worktree's commands before database setup. Stale run records require manual process verification.",
        );
      const args = ["run", "-p=1", "./cmd/worktree-db"];
      const code = await executeOwned("go", args, {
        cwd: join(context.root, "server"),
        env,
        receiptPath: join(context.stateDir, `${config.id}-database.lock`, "drained"),
        onUncertain: () => {
          releaseLock = false;
        },
      });
      if (code !== 0)
        throw new Error(
          `Database setup exited ${code}; completed databases are preserved. Correct the prerequisite and retry.`,
        );
    },
    () => releaseLock,
  );
}

async function main(args: string[]) {
  if (!args.length || args[0] === "help" || args[0] === "--help") {
    console.log(
      "Usage: node scripts/worktree.ts init | status | check | db setup --major <verified-major> | stop <mode> | run <dev|test|rls|e2e> -- <command> [arguments]\nCreating a Git worktree is not setup. Run init, then db setup before running commands. See docs/worktrees.md.",
    );
    return;
  }
  const context = await discover();
  switch (args[0]) {
    case "init": {
      const config = await initialize(context);
      console.log(
        `Allocated ${config.id}. Config: ${context.configPath}\nNo databases created. See docs/worktrees.md for the staged setup.`,
      );
      break;
    }
    case "status":
      console.log(JSON.stringify(await loadConfig(context), null, 2));
      break;
    case "check": {
      const config = await loadConfig(context);
      const release = await reserveBlock(config.ports.frontend);
      await release();
      console.log(`All ten assigned ports for ${config.id} are currently free.`);
      break;
    }
    case "run":
      if (args[2] !== "--")
        throw new Error("Expected: run <dev|test|rls|e2e> -- <command> [arguments]");
      process.exitCode = await runCommand(context, args[1] ?? "", args[3], args.slice(4));
      break;
    case "stop":
      await stopCommands(context, args[1]);
      break;
    case "db":
      if (args[1] !== "setup" || args[2] !== "--major" || args.length !== 4)
        throw new Error("Expected db setup --major <verified-production-major>");
      await setupDatabases(context, args[3]);
      break;
    default:
      throw new Error(`Unknown command: ${args[0]}`);
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  main(process.argv.slice(2)).catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
}
