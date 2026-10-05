#!/usr/bin/env node
// Local-only worktree coordination. No database, credential, or service lifecycle operations.
import { createHash, randomUUID } from "node:crypto";
import { execFileSync, spawn } from "node:child_process";
import { mkdir, readFile, realpath, rename, rm, writeFile } from "node:fs/promises";
import { createServer } from "node:net";
import { basename, dirname, join, resolve } from "node:path";
import { pathToFileURL } from "node:url";

const VERSION = 1;
const FIRST_PORT = 21000;
const LAST_PORT = 29999;
const BLOCK_SIZE = 10;
const MODES = ["dev", "test", "rls", "e2e"];
const LOCAL_DIRECTORY = ".worktree";

export function worktreeId(root) {
  const label = basename(root).toLowerCase().replace(/[^a-z0-9]+/g, "_").replace(/^_+|_+$/g, "").slice(0, 20) || "worktree";
  return `${label}_${createHash("sha256").update(root).digest("hex").slice(0, 12)}`;
}

function allocation(root, basePort) {
  const id = worktreeId(root);
  return { root, id, basePort };
}

export function localConfig(entry) {
  const { root, id, basePort } = entry;
  return {
    version: VERSION,
    root,
    id,
    postgres: { host: "127.0.0.1", port: 55432 },
    databases: Object.fromEntries(MODES.map((mode) => [mode, `ft_${id}_${mode}`])),
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

export async function discover(cwd = process.cwd()) {
  const git = (...args) => execFileSync("git", args, { cwd, encoding: "utf8" }).trim();
  const root = await realpath(git("rev-parse", "--show-toplevel"));
  const common = await realpath(resolve(cwd, git("rev-parse", "--git-common-dir")));
  return { root, stateDir: join(common, "fittrack-worktrees"), configPath: join(root, LOCAL_DIRECTORY, "config.json") };
}

async function readJSON(path, fallback) {
  try {
    return JSON.parse(await readFile(path, "utf8"));
  } catch (error) {
    if (error.code === "ENOENT" && fallback !== undefined) return fallback;
    throw new Error(`Cannot read ${path}: ${error.message}`);
  }
}

async function atomicJSON(path, value) {
  await mkdir(dirname(path), { recursive: true });
  const temporary = `${path}.${randomUUID()}.tmp`;
  try {
    await writeFile(temporary, `${JSON.stringify(value, null, 2)}\n`, { flag: "wx" });
    await rename(temporary, path);
  } finally {
    await rm(temporary, { force: true });
  }
}

export async function withLock(path, action, shouldRelease = () => true) {
  await mkdir(dirname(path), { recursive: true });
  try {
    await mkdir(path);
  } catch (error) {
    if (error.code === "EEXIST") {
      throw new Error(`Lock already held: ${path}. Wait for the other command. After a crash, verify its process has stopped before removing this lock directory.`);
    }
    throw error;
  }
  try {
    await writeFile(join(path, "owner.json"), JSON.stringify({ pid: process.pid, startedAt: new Date().toISOString() }));
    return await action();
  } finally {
    if (shouldRelease()) await rm(path, { recursive: true, force: true });
  }
}

function validateRegistry(registry) {
  if (registry.version !== VERSION || !Array.isArray(registry.entries)) throw new Error("Unsupported worktree registry. Refusing to overwrite it.");
  const roots = new Set();
  const ids = new Set();
  const ports = new Set();
  for (const entry of registry.entries) {
    if (!entry || typeof entry.root !== "string" || entry.id !== worktreeId(entry.root) || !Number.isInteger(entry.basePort) || entry.basePort < FIRST_PORT || entry.basePort + BLOCK_SIZE - 1 > LAST_PORT || (entry.basePort - FIRST_PORT) % BLOCK_SIZE !== 0 || roots.has(entry.root) || ids.has(entry.id) || ports.has(entry.basePort)) {
      throw new Error("Invalid or conflicting worktree registry entry. Refusing to change allocations.");
    }
    roots.add(entry.root);
    ids.add(entry.id);
    ports.add(entry.basePort);
  }
}

async function readRegistry(stateDir) {
  const registry = await readJSON(join(stateDir, "registry.json"), { version: VERSION, entries: [] });
  validateRegistry(registry);
  return registry;
}

async function listen(port, host) {
  const server = createServer();
  await new Promise((resolveListen, reject) => {
    server.once("error", reject);
    server.listen({ port, host, exclusive: true, ipv6Only: host === "::" }, resolveListen);
  });
  return server;
}

// Hold both address-family reservations until every port has been checked.
export async function reserveBlock(basePort) {
  const servers = [];
  const release = () => Promise.all(servers.map((server) => new Promise((done) => server.close(done))));
  try {
    for (let port = basePort; port < basePort + BLOCK_SIZE; port++) {
      for (const host of ["0.0.0.0", "::"]) {
        try {
          servers.push(await listen(port, host));
        } catch (error) {
          if (host === "::" && ["EAFNOSUPPORT", "EADDRNOTAVAIL"].includes(error.code)) continue;
          throw new Error(`Port ${port} cannot be reserved (${error.code}). Stop the conflicting listener; assignments never silently change.`, { cause: error });
        }
      }
    }
    return release;
  } catch (error) {
    await release();
    throw error;
  }
}

export async function initialize(context, reserve = reserveBlock) {
  return withLock(join(context.stateDir, "registry.lock"), async () => {
    const registry = await readRegistry(context.stateDir);
    let entry = registry.entries.find((candidate) => candidate.root === context.root);
    let release;
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
          if (error.cause?.code !== "EADDRINUSE") throw error;
        }
      }
      if (!entry) throw new Error("No free ten-port worktree block remains in 21000–29999.");
      if (registry.entries.some((candidate) => candidate.id === entry.id)) throw new Error("Worktree ID collision; refusing allocation.");
      registry.entries.push(entry);
    }
    try {
      // Registry first: an interrupted local-file write can be retried without reallocating.
      await atomicJSON(join(context.stateDir, "registry.json"), registry);
      const config = localConfig(entry);
      await atomicJSON(context.configPath, config);
      return config;
    } finally {
      await release();
    }
  });
}

export async function loadConfig(context) {
  const registry = await readRegistry(context.stateDir);
  const entry = registry.entries.find((candidate) => candidate.root === context.root);
  if (!entry) throw new Error("This worktree has no allocation. Run: node scripts/worktree.mjs init");
  const expected = localConfig(entry);
  const actual = await readJSON(context.configPath);
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error("Worktree config differs from the registry. Stop this worktree's services and rerun init; never copy another worktree's .worktree directory.");
  }
  return expected;
}

export function databaseURL(template, config, mode) {
  if (!MODES.includes(mode)) throw new Error("Unknown database mode.");
  if (!template) throw new Error("Set FITTRACK_LOCAL_DATABASE_URL to an existing local owning-role connection URL. Credentials are never created or saved by this helper.");
  let url;
  try {
    url = new URL(template);
  } catch {
    throw new Error("FITTRACK_LOCAL_DATABASE_URL must be a PostgreSQL URL.");
  }
  if (!["postgres:", "postgresql:"].includes(url.protocol) || url.hostname !== config.postgres.host || url.port !== String(config.postgres.port) || url.hash) {
    throw new Error("Database templates must use PostgreSQL at 127.0.0.1:55432. Remote databases and alternate hosts/ports are refused.");
  }
  // libpq/pgx accept query parameters that override the URL host and database.
  for (const key of url.searchParams.keys()) {
    if (!["sslmode", "connect_timeout"].includes(key)) throw new Error(`Unsupported database URL query parameter: ${key}`);
  }
  url.pathname = `/${config.databases[mode]}`;
  return url.toString();
}

export function commandEnvironment(config, mode, inherited = process.env) {
  if (!MODES.includes(mode)) throw new Error(`Mode must be one of: ${MODES.join(", ")}`);
  const isE2E = mode === "e2e";
  const frontend = isE2E ? config.ports.e2eFrontend : config.ports.frontend;
  const api = isE2E ? config.ports.e2eApi : config.ports.api;
  const env = {
    ...inherited,
    FITTRACK_WORKTREE_ID: config.id,
    DATABASE_URL: databaseURL(inherited.FITTRACK_LOCAL_DATABASE_URL, config, mode),
    // Clear inherited alternative database targets before tests read them.
    ADMIN_DATABASE_URL: "",
    RECOMMENDATION_TEST_DATABASE_URL: "",
    PGHOST: config.postgres.host,
    PGPORT: String(config.postgres.port),
    PGDATABASE: config.databases[mode],
    PGHOSTADDR: "",
    PGSERVICE: "",
    PGSERVICEFILE: "",
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
  if (["test", "rls"].includes(mode)) env.GOFLAGS = `${inherited.GOFLAGS ?? ""} -p=1`.trim();
  if (mode === "rls") {
    if (!inherited.FITTRACK_LOCAL_RUNTIME_DATABASE_URL) throw new Error("RLS commands require FITTRACK_LOCAL_RUNTIME_DATABASE_URL for the already-provisioned fittrack_app role.");
    const runtimeURL = databaseURL(inherited.FITTRACK_LOCAL_RUNTIME_DATABASE_URL, config, mode);
    if (new URL(runtimeURL).username !== "fittrack_app") throw new Error("RLS runtime URL must use the existing fittrack_app role.");
    env.ADMIN_DATABASE_URL = env.DATABASE_URL;
    env.DATABASE_URL = runtimeURL;
    env.RECOMMENDATION_TEST_DATABASE_URL = runtimeURL;
    env.RLS_ENFORCEMENT_REQUIRED = "true";
  }
  return env;
}

function signalGroup(pid, signal) {
  try {
    process.kill(-pid, signal);
    return true;
  } catch (error) {
    if (error.code === "ESRCH") return false;
    throw error;
  }
}

const delay = (milliseconds) => new Promise((done) => setTimeout(done, milliseconds));

export async function runCommand(context, mode, command, args) {
  if (!command) throw new Error("Expected: run <dev|test|rls|e2e> -- <command> [arguments]");
  if (process.platform === "win32") throw new Error("The initial command runner requires a POSIX process group; use Linux, macOS, or WSL.");
  const config = await loadConfig(context);
  const env = commandEnvironment(config, mode);
  let releaseLock = true;
  const lockPath = join(context.stateDir, `${config.id}-${mode}.lock`);
  const execute = () => new Promise((resolveExit, reject) => {
    // A separate group lets cancellation reach go/bun/shell grandchildren too.
    const child = spawn(command, args, { cwd: process.cwd(), env, stdio: "inherit", detached: true });
    let escalation;
    const cancel = (signal) => {
      if (!child.pid) return;
      signalGroup(child.pid, signal);
      escalation ??= setTimeout(() => signalGroup(child.pid, "SIGKILL"), 2000);
    };
    const onInterrupt = () => cancel("SIGINT");
    const onTerminate = () => cancel("SIGTERM");
    process.on("SIGINT", onInterrupt);
    process.on("SIGTERM", onTerminate);
    const cleanup = () => {
      clearTimeout(escalation);
      process.off("SIGINT", onInterrupt);
      process.off("SIGTERM", onTerminate);
    };
    child.once("error", (error) => { cleanup(); reject(error); });
    child.once("exit", async (code, signal) => {
      try {
        // Even a normally-exiting shell can leave a background database writer.
        if (signalGroup(child.pid, 0)) {
          signalGroup(child.pid, "SIGTERM");
          await delay(250);
          if (signalGroup(child.pid, 0)) {
            signalGroup(child.pid, "SIGKILL");
            await delay(250);
          }
          if (signalGroup(child.pid, 0)) {
            releaseLock = false;
            console.error(`Process group ${child.pid} has not fully disappeared. Verify that it is stopped before removing ${lockPath}; its lock is retained for test/rls mode.`);
            if (code === 0) code = 1;
          }
        }
        resolveExit(code ?? (signal === "SIGINT" ? 130 : 143));
      } catch (error) {
        releaseLock = false;
        reject(error);
      } finally {
        cleanup();
      }
    });
  });
  return ["test", "rls"].includes(mode)
    ? withLock(lockPath, execute, () => releaseLock)
    : execute();
}

async function main(args) {
  if (!args.length || args[0] === "help" || args[0] === "--help") {
    console.log("Usage: node scripts/worktree.mjs init | status | check | run <dev|test|rls|e2e> -- <command> [arguments]\ninit only reserves names/ports and writes ignored config; it never provisions PostgreSQL. See docs/worktrees.md.");
    return;
  }
  const context = await discover();
  switch (args[0]) {
    case "init": {
      const config = await initialize(context);
      console.log(`Allocated ${config.id}. Config: ${context.configPath}\nNo databases created. See docs/worktrees.md for the staged setup.`);
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
      if (args[2] !== "--") throw new Error("Expected: run <dev|test|rls|e2e> -- <command> [arguments]");
      process.exitCode = await runCommand(context, args[1], args[3], args.slice(4));
      break;
    default:
      throw new Error(`Unknown command: ${args[0]}`);
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  main(process.argv.slice(2)).catch((error) => { console.error(error.message); process.exitCode = 1; });
}
