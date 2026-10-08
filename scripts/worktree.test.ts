import assert from "node:assert/strict";
import test from "node:test";
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
