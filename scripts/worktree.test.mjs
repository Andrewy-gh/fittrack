import assert from "node:assert/strict";
import { execFile, execFileSync, spawn } from "node:child_process";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { promisify } from "node:util";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { commandEnvironment, databaseURL, discover, initialize, loadConfig, localConfig, reserveBlock, withLock, worktreeId } from "./worktree.mjs";

const run = promisify(execFile);
const script = fileURLToPath(new URL("./worktree.mjs", import.meta.url));
const base = "postgresql://local_owner@127.0.0.1:55432/postgres?sslmode=disable";
const runtime = "postgresql://fittrack_app@127.0.0.1:55432/postgres?sslmode=disable";

async function fixture(t) {
  const dir = await mkdtemp(join(tmpdir(), "fittrack-worktree-test-"));
  t.after(() => rm(dir, { recursive: true, force: true }));
  const context = { root: join(dir, "feature"), stateDir: join(dir, "state"), configPath: join(dir, "feature", ".worktree", "config.json") };
  await mkdir(context.root);
  return context;
}

const fakeReserve = async () => async () => {};

function config(root = "/tmp/fittrack") {
  return localConfig({ root, id: worktreeId(root), basePort: 21000 });
}

test("ID is stable across branch changes and safe for unquoted PostgreSQL names", () => {
  const id = worktreeId("/tmp/Feature With 'Odd' Characters ✨");
  assert.match(id, /^[a-z0-9_]+_[a-f0-9]{12}$/);
  assert.equal(id, worktreeId("/tmp/Feature With 'Odd' Characters ✨"));
  assert.notEqual(id, worktreeId("/other/Feature With 'Odd' Characters ✨"));
  for (const name of Object.values(localConfig({ id, root: "/tmp", basePort: 21000 }).databases)) assert.ok(name.length <= 63);
});

test("init persists allocation and reuses it without changing other worktrees", async (t) => {
  const context = await fixture(t);
  const first = await initialize(context, fakeReserve);
  const second = { ...context, root: `${context.root}-other`, configPath: `${context.root}-other/.worktree/config.json` };
  const next = await initialize(second, fakeReserve);
  assert.equal(first.ports.frontend, 21000);
  assert.equal(next.ports.frontend, 21010);
  assert.deepEqual(await initialize(context, fakeReserve), first);
  assert.deepEqual(await loadConfig(second), next);
  assert.equal(new Set(Object.values(first.databases)).size, 4);
  const serialized = await readFile(context.configPath, "utf8");
  assert.ok(!serialized.includes("password"));
  assert.ok(!serialized.includes("DATABASE_URL"));
});

test("new allocations skip busy blocks, existing allocations fail without reassignment", async (t) => {
  const context = await fixture(t);
  const occupied = () => new Error("busy", { cause: { code: "EADDRINUSE" } });
  const chosen = await initialize(context, async (port) => {
    if (port === 21000) throw occupied();
    return async () => {};
  });
  assert.equal(chosen.ports.frontend, 21010);
  const before = await readFile(join(context.stateDir, "registry.json"), "utf8");
  await assert.rejects(initialize(context, async () => { throw occupied(); }), /busy/);
  assert.equal(await readFile(join(context.stateDir, "registry.json"), "utf8"), before);
});

test("unexpected socket errors fail promptly instead of searching every block", async (t) => {
  const context = await fixture(t);
  let attempts = 0;
  await assert.rejects(initialize(context, async () => {
    attempts++;
    throw new Error("denied", { cause: { code: "EACCES" } });
  }), /denied/);
  assert.equal(attempts, 1);
});

test("a held registry lock prevents concurrent writes and is never stolen", async (t) => {
  const context = await fixture(t);
  const lock = join(context.stateDir, "registry.lock");
  await withLock(lock, async () => {
    await assert.rejects(initialize(context, fakeReserve), /Lock already held/);
    assert.ok(JSON.parse(await readFile(join(lock, "owner.json"), "utf8")).pid);
  });
  await initialize(context, fakeReserve);
});

test("a failed initialization releases its registry lock", async (t) => {
  const context = await fixture(t);
  await assert.rejects(initialize(context, async () => { throw new Error("failed"); }), /failed/);
  await initialize(context, fakeReserve);
});

test("invalid/conflicting registry entries and copied local configs are rejected", async (t) => {
  const context = await fixture(t);
  const first = await initialize(context, fakeReserve);
  await writeFile(context.configPath, JSON.stringify({ ...first, id: "copied" }));
  await assert.rejects(loadConfig(context), /differs from the registry/);
  await initialize(context, fakeReserve);
  const path = join(context.stateDir, "registry.json");
  const registry = JSON.parse(await readFile(path, "utf8"));
  registry.entries.push(registry.entries[0]);
  await writeFile(path, JSON.stringify(registry));
  await assert.rejects(initialize(context, fakeReserve), /Invalid or conflicting/);
});

test("interrupted local config writes can recover from the persisted allocation", async (t) => {
  const context = await fixture(t);
  const first = await initialize(context, fakeReserve);
  await rm(context.configPath);
  const repaired = await initialize(context, fakeReserve);
  assert.deepEqual(repaired, first);
});

test("socket reservation detects actual occupied ports and releases prior probes", async (t) => {
  const server = createServer();
  await new Promise((resolveListen) => server.listen(0, "127.0.0.1", resolveListen));
  t.after(() => new Promise((done) => server.close(done)));
  const port = server.address().port;
  await assert.rejects(reserveBlock(port), /cannot be reserved/);
});

test("database URLs always select a unique local database and keep credentials only in memory", () => {
  const current = config();
  const urls = ["dev", "test", "rls", "e2e"].map((mode) => databaseURL(base, current, mode));
  assert.equal(new Set(urls).size, 4);
  for (const [index, mode] of ["dev", "test", "rls", "e2e"].entries()) {
    const url = new URL(urls[index]);
    assert.equal(url.pathname, `/${current.databases[mode]}`);
    assert.equal(url.hostname, "127.0.0.1");
    assert.equal(url.port, "55432");
    assert.equal(url.username, "local_owner");
  }
});

test("missing, remote, alternate-port and query-redirected database templates fail closed", () => {
  for (const template of [undefined, "invalid", "https://127.0.0.1:55432/postgres", "postgres://db.example.com:55432/postgres", "postgres://127.0.0.1:5432/postgres", "postgres://127.0.0.1:55432/postgres?host=remote", "postgres://127.0.0.1:55432/postgres?dbname=production", "postgres://127.0.0.1:55432/postgres?hostaddr=192.0.2.1", "postgres://127.0.0.1:55432/postgres#fragment"]) {
    assert.throws(() => databaseURL(template, config(), "test"));
  }
});

test("dev and E2E commands receive coherent disjoint service endpoints", () => {
  const current = config();
  const inherited = { FITTRACK_LOCAL_DATABASE_URL: base, DATABASE_URL: "wrong", VITE_API_BASE_URL: "https://remote.example/api", PGHOSTADDR: "192.0.2.1", PGSERVICE: "production" };
  const dev = commandEnvironment(current, "dev", inherited);
  const e2e = commandEnvironment(current, "e2e", inherited);
  assert.equal(dev.FRONTEND_PORT, "21000");
  assert.equal(dev.PORT, "21001");
  assert.equal(dev.METRICS_PORT, "21002");
  assert.equal(dev.PREVIEW_PORT, "21006");
  assert.equal(dev.VITE_API_BASE_URL, "/api");
  assert.equal(dev.PGHOSTADDR, "");
  assert.equal(dev.PGSERVICE, "");
  assert.equal(dev.RLS_ENFORCEMENT_REQUIRED, "false");
  assert.equal(e2e.FRONTEND_PORT, e2e.E2E_PORT);
  assert.equal(e2e.PREVIEW_PORT, e2e.E2E_PORT);
  assert.equal(e2e.APP_BASE_URL, e2e.E2E_BASE_URL);
  assert.equal(e2e.API_PROXY_TARGET, e2e.E2E_LOCAL_AUTH_API_BASE_URL);
  assert.equal(e2e.E2E_REUSE_EXISTING_SERVER, "false");
  assert.notEqual(e2e.DATABASE_URL, dev.DATABASE_URL);
  assert.equal(inherited.DATABASE_URL, "wrong");
});

test("test mode serializes Go packages and cannot inherit recommendation's external URL", () => {
  const env = commandEnvironment(config(), "test", { FITTRACK_LOCAL_DATABASE_URL: base, GOFLAGS: "-tags=example", RECOMMENDATION_TEST_DATABASE_URL: "wrong" });
  assert.equal(env.GOFLAGS, "-tags=example -p=1");
  assert.equal(env.RECOMMENDATION_TEST_DATABASE_URL, "");
  assert.equal(env.ADMIN_DATABASE_URL, "");
  assert.match(env.DATABASE_URL, /_test\?/);
});

test("RLS mode requires existing restricted role and keeps owner URL distinct", () => {
  const inherited = { FITTRACK_LOCAL_DATABASE_URL: base };
  assert.throws(() => commandEnvironment(config(), "rls", inherited), /FITTRACK_LOCAL_RUNTIME_DATABASE_URL/);
  assert.throws(() => commandEnvironment(config(), "rls", { ...inherited, FITTRACK_LOCAL_RUNTIME_DATABASE_URL: base }), /fittrack_app/);
  const env = commandEnvironment(config(), "rls", { ...inherited, FITTRACK_LOCAL_RUNTIME_DATABASE_URL: runtime });
  assert.equal(new URL(env.DATABASE_URL).username, "fittrack_app");
  assert.equal(new URL(env.ADMIN_DATABASE_URL).username, "local_owner");
  assert.equal(env.RECOMMENDATION_TEST_DATABASE_URL, env.DATABASE_URL);
  assert.equal(new URL(env.DATABASE_URL).pathname, new URL(env.ADMIN_DATABASE_URL).pathname);
  assert.equal(env.RLS_ENFORCEMENT_REQUIRED, "true");
  assert.equal(env.GOFLAGS, "-p=1");
});

test("actual linked worktrees share registry, survive branch changes, and command exit codes propagate", async (t) => {
  const context = await fixture(t);
  execFileSync("git", ["init", "--quiet", context.root]);
  execFileSync("git", ["-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "--quiet", "--allow-empty", "-m", "fixture"], { cwd: context.root });
  const otherPath = `${context.root}-linked`;
  execFileSync("git", ["worktree", "add", "--quiet", "-b", "other", otherPath], { cwd: context.root });
  const first = await discover(context.root);
  const other = await discover(otherPath);
  assert.equal(first.stateDir, other.stateDir);
  const a = await initialize(first, fakeReserve);
  const b = await initialize(other, fakeReserve);
  assert.notEqual(a.id, b.id);
  assert.notEqual(a.ports.frontend, b.ports.frontend);
  execFileSync("git", ["checkout", "--quiet", "-b", "renamed"], { cwd: context.root });
  assert.deepEqual(await loadConfig(await discover(context.root)), a);
  const env = { ...process.env, FITTRACK_LOCAL_DATABASE_URL: base };
  const output = await run(process.execPath, [script, "run", "test", "--", process.execPath, "-e", "console.log(new URL(process.env.DATABASE_URL).pathname)"], { cwd: context.root, env });
  assert.equal(output.stdout.trim(), `/${a.databases.test}`);
  await assert.rejects(run(process.execPath, [script, "run", "test", "--", process.execPath, "-e", "process.exit(7)"], { cwd: context.root, env }), (error) => error.code === 7);
  await assert.rejects(run(process.execPath, [script, "run", "test", "--", "fittrack-no-such-command"], { cwd: context.root, env }), /ENOENT/);
  await run(process.execPath, [script, "run", "test", "--", process.execPath, "-e", "process.exit(0)"], { cwd: context.root, env });
});

test("cancellation reaches grandchildren before a test lock can be reused", { timeout: 10000 }, async (t) => {
  const context = await fixture(t);
  execFileSync("git", ["init", "--quiet", context.root]);
  const actual = await discover(context.root);
  const current = await initialize(actual, fakeReserve);
  const ready = join(context.root, "ready");
  const stopped = join(context.root, "stopped");
  const grandchildSource = `
    const fs = require('node:fs');
    process.on('SIGTERM', () => { fs.writeFileSync(${JSON.stringify(stopped)}, 'stopped'); process.exit(0); });
    fs.writeFileSync(${JSON.stringify(ready)}, String(process.pid));
    setInterval(() => {}, 100);
  `;
  const childSource = `
    const { spawn } = require('node:child_process');
    const child = spawn(process.execPath, ['-e', ${JSON.stringify(grandchildSource)}], { stdio: 'inherit' });
    let stopped = false, cancelled = false;
    child.on('exit', () => { stopped = true; if (cancelled) process.exit(0); });
    process.on('SIGTERM', () => { cancelled = true; if (stopped) process.exit(0); });
    setInterval(() => {}, 100);
  `;
  const wrapper = spawn(process.execPath, [script, "run", "test", "--", process.execPath, "-e", childSource], {
    cwd: context.root,
    env: { ...process.env, FITTRACK_LOCAL_DATABASE_URL: base },
    stdio: ["ignore", "pipe", "pipe"],
  });
  t.after(() => { if (wrapper.exitCode === null) wrapper.kill("SIGKILL"); });
  const exited = new Promise((done, reject) => { wrapper.once("error", reject); wrapper.once("exit", done); });
  let pid;
  for (let i = 0; i < 100; i++) {
    try { pid = Number(await readFile(ready, "utf8")); break; } catch { await new Promise((done) => setTimeout(done, 20)); }
  }
  assert.ok(pid, "grandchild did not become ready");
  t.after(() => { try { process.kill(pid, "SIGKILL"); } catch {} });
  wrapper.kill("SIGTERM");
  await exited;
  assert.equal(await readFile(stopped, "utf8"), "stopped");
  // The parent reaps its child, so the entire owned group has disappeared.
  assert.throws(() => process.kill(pid, 0), { code: "ESRCH" });
  const output = await run(process.execPath, [script, "run", "test", "--", process.execPath, "-e", "process.exit(0)"], { cwd: context.root, env: { ...process.env, FITTRACK_LOCAL_DATABASE_URL: base } });
  assert.equal(output.stderr, "");
  assert.equal(current.databases.test.endsWith("_test"), true);
});
