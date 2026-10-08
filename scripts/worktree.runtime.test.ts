// Opt-in transport proof against the separately built native Go CLI.
import assert from "node:assert/strict";
import { execFile, spawn } from "node:child_process";
import { cp, mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { promisify } from "node:util";
import test from "node:test";
import type { TestContext } from "node:test";
import { loadConfig } from "./worktree.ts";

const execute = promisify(execFile);
const source = fileURLToPath(new URL("../", import.meta.url));
const script = join(source, "scripts", "worktree.ts");
const runtime = process.env.WORKTREE_RUNTIME;
const options = { skip: !runtime && "Set WORKTREE_RUNTIME to the external Go executable", timeout: 30000 };

function isRecord(value: unknown): value is Readonly<Record<string, unknown>> {
  return typeof value === "object" && value !== null;
}

function outputRecord(output: string) {
  const value: unknown = JSON.parse(output);
  assert.ok(isRecord(value));
  return value;
}

async function fixture(t: TestContext) {
  const temp = await mkdtemp(join(tmpdir(), "fittrack-external-runtime-"));
  t.after(() => rm(temp, { recursive: true, force: true }));
  const root = join(temp, "checkout with spaces");
  await mkdir(root);
  for (const path of ["worktree-runtime.json", "scripts/worktree.ts", "scripts/package.json"]) {
    await cp(join(source, path), join(root, path), { recursive: true });
  }
  await mkdir(join(root, "server"));
  await execute("git", ["init", "--quiet"], { cwd: root, windowsHide: true });
  const env = {
    ...process.env,
    WORKTREE_RUNTIME: resolve(runtime ?? ""),
    WORKTREE_RUNTIME_STATE_DIR: join(temp, "state"),
    FITTRACK_LOCAL_DATABASE_URL: "postgresql://local_owner@127.0.0.1:55432/postgres?sslmode=disable",
    FITTRACK_LOCAL_RUNTIME_DATABASE_URL: "postgresql://fittrack_app@127.0.0.1:55432/postgres?sslmode=disable",
    PGHOSTADDR: "192.0.2.1",
    PGSERVICE: "production",
  };
  const cli = (args: string[], cwd = root, overrides: NodeJS.ProcessEnv = {}) =>
    execute(process.execPath, [script, ...args], { cwd, env: { ...env, ...overrides }, windowsHide: true });
  await cli(["init"]);
  return { root, env, cli, config: await loadConfig(root, env) };
}

async function waitFor(description: string, predicate: () => Promise<boolean>) {
  const deadline = Date.now() + 10000;
  while (Date.now() < deadline) {
    if (await predicate()) return;
    await new Promise((done) => setTimeout(done, 25));
  }
  assert.fail(description);
}

test("external runtime transports FitTrack mode/env and original working directory", options, async (t) => {
  const f = await fixture(t);
  const inspect = `console.log(JSON.stringify({
    cwd:process.cwd(), database:new URL(process.env.DATABASE_URL).pathname,
    runtimeRole:new URL(process.env.DATABASE_URL).username,
    ownerRole:process.env.ADMIN_DATABASE_URL ? new URL(process.env.ADMIN_DATABASE_URL).username : '',
    api:process.env.PORT, metrics:process.env.METRICS_PORT, frontend:process.env.FRONTEND_PORT,
    app:process.env.APP_BASE_URL, e2e:process.env.E2E_BASE_URL,
    proxy:process.env.API_PROXY_TARGET, auth:process.env.E2E_LOCAL_AUTH_API_BASE_URL,
    redirected:!!(process.env.PGHOSTADDR||process.env.PGSERVICE),
    parentPipe:process.env.WORKTREE_RUNTIME_PARENT_PIPE
  }))`;
  const cwd = join(f.root, "server");
  const e2e = outputRecord((await f.cli(["run", "e2e", "--", process.execPath, "-e", inspect], cwd)).stdout);
  assert.equal(e2e.cwd, cwd);
  assert.equal(e2e.database, `/${f.config.databases.e2e}`);
  assert.equal(e2e.api, String(f.config.ports.e2eApi));
  assert.equal(e2e.metrics, String(f.config.ports.e2eMetrics));
  assert.equal(e2e.frontend, String(f.config.ports.e2eFrontend));
  assert.equal(e2e.app, e2e.e2e);
  assert.equal(e2e.proxy, e2e.auth);
  assert.equal(e2e.redirected, false);
  assert.equal(e2e.parentPipe, undefined);
  const rls = outputRecord((await f.cli(["run", "rls", "--", process.execPath, "-e", inspect], cwd)).stdout);
  assert.equal(rls.runtimeRole, "fittrack_app");
  assert.equal(rls.ownerRole, "local_owner");
  assert.equal(rls.database, `/${f.config.databases.rls}`);
  await assert.rejects(f.cli(["run", "rls", "--", process.execPath, "-e", "process.exit(0)"], cwd,
    { FITTRACK_LOCAL_RUNTIME_DATABASE_URL: f.env.FITTRACK_LOCAL_DATABASE_URL }), /existing fittrack_app role/);
  const log = join(f.root, ".worktree", "command.log");
  await f.cli(["run", "dev", "--log-file", log, "--", process.execPath, "-e",
    "console.log('agent stdout');console.error('agent stderr')"], cwd);
  assert.match(await readFile(log, "utf8"), /agent stdout/);
  assert.match(await readFile(log, "utf8"), /agent stderr/);
});

test("killing the Node bridge cancels detached descendants and permits the next test", options, async (t) => {
  const f = await fixture(t);
  const ready = join(f.root, "grandchild-ready");
  const grandchild = `const s=require('node:net').createServer(c=>c.end());
    s.listen(Number(process.env.METRICS_PORT),'127.0.0.1',()=>require('node:fs').writeFileSync(${JSON.stringify(ready)},'ready'));
    setTimeout(()=>process.exit(0),20000);`;
  const parent = `require('node:child_process').spawn(process.execPath,['-e',${JSON.stringify(grandchild)}],
    {stdio:'inherit',detached:process.platform==='win32',windowsHide:true}).unref();setTimeout(()=>process.exit(0),20000);`;
  const bridge = spawn(process.execPath, [script, "run", "test", "--", process.execPath, "-e", parent], {
    cwd: f.root, env: f.env, windowsHide: true, stdio: "ignore",
  });
  const exited = new Promise<void>((done, reject) => {
    bridge.once("error", reject);
    bridge.once("exit", () => done());
  });
  t.after(async () => {
    await f.cli(["stop", "test"]).catch(() => {});
    if (bridge.exitCode === null && bridge.signalCode === null) bridge.kill("SIGKILL");
    await exited;
  });
  await waitFor("managed descendant did not bind", async () => readFile(ready).then(() => true, () => false));
  await assert.rejects(f.cli(["run", "test", "--", process.execPath, "-e", "process.exit(0)"]), /Lock already held/);
  bridge.kill("SIGKILL");
  await exited;
  await waitFor("Node bridge death left an owned listener", async () => f.cli(["check"]).then(() => true, () => false));
  await waitFor("Node bridge death retained a confirmed mode lock", async () =>
    f.cli(["run", "test", "--", process.execPath, "-e", "process.exit(0)"]).then(() => true, () => false));
});

test("application root failure drains an orphan with inherited output", options, async (t) => {
  const f = await fixture(t);
  const ready = join(f.root, "orphan-ready");
  const grandchild = `const s=require('node:net').createServer(c=>c.end());
    s.listen(Number(process.env.METRICS_PORT),'127.0.0.1',()=>require('node:fs').writeFileSync(${JSON.stringify(ready)},'ready'));
    setTimeout(()=>process.exit(0),20000);`;
  const parent = `const child=require('node:child_process').spawn(process.execPath,['-e',${JSON.stringify(grandchild)}],
    {stdio:'inherit',detached:process.platform==='win32',windowsHide:true});child.unref();
    const timer=setInterval(()=>{if(require('node:fs').existsSync(${JSON.stringify(ready)}))process.exit(17)},25);`;
  await assert.rejects(f.cli(["run", "test", "--", process.execPath, "-e", parent]),
    (error: unknown) => typeof error === "object" && error !== null && "code" in error && error.code === 17);
  await f.cli(["check"]);
  await f.cli(["run", "test", "--", process.execPath, "-e", "process.exit(0)"]);
});
