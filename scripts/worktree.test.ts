import assert from "node:assert/strict";
import test from "node:test";
import { resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { commandEnvironment, databaseURL, localConfig } from "./worktree.ts";

const base = "postgresql://local_owner@127.0.0.1:55432/postgres?sslmode=disable";
const runtime = "postgresql://fittrack_app@127.0.0.1:55432/postgres?sslmode=disable";
function config() {
  return localConfig({ root: "/tmp/fittrack", id: "fittrack_0123456789ab", basePort: 21000 });
}

test("database URLs always select a unique local database and keep credentials only in memory", () => {
  const current = config();
  const urls = ["dev", "test", "rls", "e2e"].map((mode) => databaseURL(base, current, mode));
  assert.equal(new Set(urls).size, 4);
  for (const [index, mode] of (["dev", "test", "rls", "e2e"] as const).entries()) {
    const value = urls[index];
    assert.ok(value);
    const url = new URL(value);
    assert.equal(url.pathname, `/${current.databases[mode]}`);
    assert.equal(url.hostname, "127.0.0.1");
    assert.equal(url.port, "55432");
    assert.equal(url.username, "local_owner");
  }
});

test("missing, remote, alternate-port and query-redirected database templates fail closed", () => {
  for (const template of [
    undefined,
    "invalid",
    "https://127.0.0.1:55432/postgres",
    "postgres://db.example.com:55432/postgres",
    "postgres://127.0.0.1:5432/postgres",
    "postgres://127.0.0.1:55432/postgres?host=remote",
    "postgres://127.0.0.1:55432/postgres?dbname=production",
    "postgres://127.0.0.1:55432/postgres?hostaddr=192.0.2.1",
    "postgres://127.0.0.1:55432/postgres#fragment",
  ]) {
    assert.throws(() => databaseURL(template, config(), "test"));
  }
});

test("dev and E2E commands receive coherent disjoint service endpoints", () => {
  const current = config();
  const inherited = {
    FITTRACK_LOCAL_DATABASE_URL: base,
    DATABASE_URL: "wrong",
    VITE_API_BASE_URL: "https://remote.example/api",
    PGHOSTADDR: "192.0.2.1",
    PGSERVICE: "production",
  };
  const dev = commandEnvironment(current, "dev", inherited);
  const e2e = commandEnvironment(current, "e2e", inherited);
  assert.equal(dev.FRONTEND_PORT, "21000");
  assert.equal(dev.PORT, "21001");
  assert.equal(dev.METRICS_PORT, "21002");
  assert.equal(dev.PREVIEW_PORT, "21006");
  assert.equal(dev.VITE_API_BASE_URL, "/api");
  assert.equal(dev.PGHOSTADDR, undefined);
  assert.equal(dev.PGSERVICE, undefined);
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
  const env = commandEnvironment(config(), "test", {
    FITTRACK_LOCAL_DATABASE_URL: base,
    GOFLAGS: "-tags=example",
    RECOMMENDATION_TEST_DATABASE_URL: "wrong",
  });
  assert.equal(env.GOFLAGS, "-tags=example -p=1");
  assert.equal(env.RECOMMENDATION_TEST_DATABASE_URL, "");
  assert.equal(env.ADMIN_DATABASE_URL, "");
  assert.match(env.DATABASE_URL, /_test\?/);
});

test("RLS mode requires existing restricted role and keeps owner URL distinct", () => {
  const inherited = { FITTRACK_LOCAL_DATABASE_URL: base };
  assert.throws(
    () => commandEnvironment(config(), "rls", inherited),
    /FITTRACK_LOCAL_RUNTIME_DATABASE_URL/,
  );
  assert.throws(
    () =>
      commandEnvironment(config(), "rls", {
        ...inherited,
        FITTRACK_LOCAL_RUNTIME_DATABASE_URL: base,
      }),
    /fittrack_app/,
  );
  const env = commandEnvironment(config(), "rls", {
    ...inherited,
    FITTRACK_LOCAL_RUNTIME_DATABASE_URL: runtime,
  });
  assert.equal(new URL(env.DATABASE_URL).username, "fittrack_app");
  assert.equal(new URL(env.ADMIN_DATABASE_URL).username, "local_owner");
  assert.equal(env.RECOMMENDATION_TEST_DATABASE_URL, env.DATABASE_URL);
  assert.equal(new URL(env.DATABASE_URL).pathname, new URL(env.ADMIN_DATABASE_URL).pathname);
  assert.equal(env.RLS_ENFORCEMENT_REQUIRED, "true");
  assert.equal(env.GOFLAGS, "-p=1");
});

test("unit commands receive no database or Stripe credentials and can construct a Node Request", () => {
  const env = commandEnvironment(config(), "unit", { DATABASE_URL: "remote", ADMIN_DATABASE_URL: "remote", FITTRACK_LOCAL_DATABASE_URL: base, PGHOST: "remote", STRIPE_SECRET_KEY: "synthetic-secret" });
  for (const key of ["DATABASE_URL", "ADMIN_DATABASE_URL", "FITTRACK_LOCAL_DATABASE_URL", "PGHOST", "STRIPE_SECRET_KEY"]) assert.equal(env[key], undefined);
  assert.equal(new Request(`${env.VITE_API_BASE_URL}/workouts`).url, "http://localhost/api/workouts");
});

test("E2E doctor rejects unusable auth and accepts local/provider configuration without hiding other blockers", () => {
  const configured = {
    PROJECT_ID: "11111111-1111-4111-8111-111111111111", VITE_PROJECT_ID: "11111111-1111-4111-8111-111111111111",
    VITE_PUBLISHABLE_CLIENT_KEY: "fixture-public-key",
    E2E_LOCAL_AUTH_ENABLED: "true", VITE_E2E_LOCAL_AUTH_ENABLED: "true",
    SECRET_SERVER_KEY: "", E2E_STACK_EMAIL: "", E2E_STACK_PASSWORD: "", CI: "", E2E_FRONTEND_MODE: undefined,
  };
  const cases: readonly { name: string; env: NodeJS.ProcessEnv; repair?: string }[] = [
    { name: "missing frontend project", env: { VITE_PROJECT_ID: "" }, repair: "VITE_PROJECT_ID" },
    { name: "missing frontend key", env: { VITE_PUBLISHABLE_CLIENT_KEY: "" }, repair: "VITE_PUBLISHABLE_CLIENT_KEY" },
    { name: "missing API project", env: { PROJECT_ID: "" }, repair: "PROJECT_ID" },
    { name: "invalid frontend project", env: { PROJECT_ID: "malformed-project", VITE_PROJECT_ID: "malformed-project" }, repair: "Stack project UUID" },
    { name: "project mismatch", env: { PROJECT_ID: "another-project" }, repair: "same Stack project" },
    { name: "server flag alone", env: { VITE_E2E_LOCAL_AUTH_ENABLED: "off", SECRET_SERVER_KEY: "fixture-secret-never-use" }, repair: "both E2E_LOCAL_AUTH_ENABLED" },
    { name: "client flag alone", env: { E2E_LOCAL_AUTH_ENABLED: "false" }, repair: "both E2E_LOCAL_AUTH_ENABLED" },
    { name: "CI production preview", env: { CI: "true" }, repair: "production preview" },
    { name: "nonempty CI false still selects preview", env: { CI: "false" }, repair: "production preview" },
    { name: "CI explicit development", env: { CI: "true", E2E_FRONTEND_MODE: "development" } },
    { name: "local explicit preview", env: { E2E_FRONTEND_MODE: "preview" }, repair: "production preview" },
    { name: "invalid frontend mode", env: { E2E_FRONTEND_MODE: "typo" }, repair: "E2E_FRONTEND_MODE" },
    { name: "valid local", env: {} },
    { name: "truthy flags", env: { E2E_LOCAL_AUTH_ENABLED: " YES ", VITE_E2E_LOCAL_AUTH_ENABLED: "1" } },
    { name: "provider server preview", env: { CI: "true", E2E_LOCAL_AUTH_ENABLED: "off", VITE_E2E_LOCAL_AUTH_ENABLED: "no", SECRET_SERVER_KEY: "fixture-secret-never-use" } },
    { name: "provider UI", env: { E2E_LOCAL_AUTH_ENABLED: "0", VITE_E2E_LOCAL_AUTH_ENABLED: "false", E2E_STACK_EMAIL: "fixture@example.test", E2E_STACK_PASSWORD: "fixture-password-never-use" } },
    { name: "partial provider", env: { E2E_LOCAL_AUTH_ENABLED: "", VITE_E2E_LOCAL_AUTH_ENABLED: "", E2E_STACK_EMAIL: "fixture@example.test" }, repair: "both E2E_STACK_EMAIL" },
    { name: "guest cannot claim authenticated preparation", env: { E2E_LOCAL_AUTH_ENABLED: "", VITE_E2E_LOCAL_AUTH_ENABLED: "" }, repair: "Configure both local-auth flags" },
  ];
  for (const scenario of cases) {
    const result = spawnSync(process.execPath, [resolve(import.meta.dirname, "worktree.ts"), "--doctor"], {
      cwd: resolve(import.meta.dirname, ".."), encoding: "utf8",
      env: { ...process.env, ...configured, ...scenario.env, WORKTREE_MODE: "e2e", WORKTREE_DOCTOR_STAGE: "bootstrap",
        FITTRACK_LOCAL_DATABASE_URL: "", FITTRACK_LOCAL_RUNTIME_DATABASE_URL: "", WORKTREE_ALLOCATION: "" },
      timeout: 30000, windowsHide: true,
    });
    assert.equal(result.error, undefined, scenario.name);
    const report = JSON.parse(result.stdout);
    const auth = report.checks.find((check: { id: string }) => check.id === "e2e_auth_configuration");
    assert.ok(auth, scenario.name);
    assert.equal(auth.status, scenario.repair ? "missing" : "passed", scenario.name);
    if (scenario.repair) assert.ok(auth.repair.includes(scenario.repair), scenario.name);
    assert.equal(report.ready, false, "auth success must not bypass unavailable DB credentials");
    assert.equal(result.status, 1);
    assert.ok(!result.stdout.includes("fixture-secret-never-use") && !result.stdout.includes("fixture-password-never-use"));
  }
});
