// Opt-in real PostgreSQL proof. Starts ONLY a fresh, task-owned temporary cluster.
import assert from "node:assert/strict";
import type { ChildProcess, ExecFileOptionsWithStringEncoding } from "node:child_process";
import { discover, loadConfig } from "./worktree.ts";
import { execFile, spawn } from "node:child_process";
import { randomBytes } from "node:crypto";
import { cp, mkdir, mkdtemp, readFile, rm, symlink, writeFile } from "node:fs/promises";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { promisify } from "node:util";
import test from "node:test";

const dockerHost =
  process.platform === "win32" ? "npipe:////./pipe/docker_engine" : "unix:///var/run/docker.sock";
const localDockerEnv = (env: NodeJS.ProcessEnv) =>
  Object.fromEntries(
    Object.entries(env).filter(([key]) => !key.toUpperCase().startsWith("DOCKER_")),
  );
const execute = (
  command: string,
  args: string[],
  options: ExecFileOptionsWithStringEncoding = {},
) =>
  promisify(execFile)(
    command,
    command === "docker" ? ["--host", dockerHost, ...args] : args,
    command === "docker"
      ? {
          ...options,
          env: localDockerEnv(options.env ?? process.env),
          windowsHide: true,
          encoding: "utf8",
        }
      : { ...options, encoding: "utf8" },
  );
const source = dirname(dirname(fileURLToPath(import.meta.url)));
const script = join(source, "scripts", "worktree.ts");
const binaries = process.env.FITTRACK_TEST_POSTGRES_BIN ?? "";
const docker = process.env.FITTRACK_TEST_DOCKER === "1";
const suffix = process.platform === "win32" ? ".exe" : "";

test(
  "real linked worktrees migrate, isolate writes, preserve dev data and reject unsafe setup",
  {
    skip:
      !docker &&
      !binaries &&
      "Set FITTRACK_TEST_DOCKER=1 or FITTRACK_TEST_POSTGRES_BIN, and FITTRACK_GOOSE to opt in to a disposable local cluster",
    timeout: 1200000,
  },
  async (t) => {
    // Fail before creating anything if the designated endpoint is occupied.
    const probe = createServer();
    await new Promise<void>((done, reject) => {
      probe.once("error", reject);
      probe.listen(55432, "127.0.0.1", done);
    });
    await new Promise((done) => probe.close(done));
    const temp = await mkdtemp(join(tmpdir(), "fittrack-db-integration-"));
    const password = randomBytes(24).toString("hex");
    let major: string;
    const container = "fittrack-pg15-test-" + randomBytes(6).toString("hex");
    if (docker) {
      major = "15";
      let started = false;
      t.after(async () => {
        if (started) {
          const label = await execute("docker", [
            "inspect",
            "--format",
            '{{index .Config.Labels "fittrack.test-owner"}}',
            container,
          ]);
          assert.equal(label.stdout.trim(), container);
          await execute("docker", ["rm", "-f", container]);
        }
        await rm(temp, { recursive: true, force: true });
      });
      await execute(
        "docker",
        [
          "run",
          "-d",
          "--name",
          container,
          "--label",
          "fittrack.test-owner=" + container,
          "--label",
          "fittrack.local-postgres=15.19",
          "--tmpfs",
          "/var/lib/postgresql/data:rw",
          "-p",
          "127.0.0.1:55432:5432",
          "-e",
          "POSTGRES_USER=ft_test_owner",
          "-e",
          "POSTGRES_PASSWORD",
          "-e",
          "POSTGRES_DB=postgres",
          "postgres:15.19-bookworm@sha256:539ceaaae49b3a7c8a04467cf00cc6788d8e3f1675df41860d86eebc4c40524f",
        ],
        { env: { ...process.env, POSTGRES_PASSWORD: password }, windowsHide: true },
      );
      started = true;
      let version;
      for (let i = 0; i < 60; i++) {
        try {
          version = (
            await execute("docker", [
              "exec",
              container,
              "psql",
              "-U",
              "ft_test_owner",
              "-d",
              "postgres",
              "-X",
              "-Atqc",
              "SHOW server_version_num",
            ])
          ).stdout.trim();
          if (version) break;
        } catch {}
        await new Promise((done) => setTimeout(done, 1000));
      }
      assert.equal(version, "150019");
      const inspections: unknown = JSON.parse(
        (await execute("docker", ["inspect", container])).stdout,
      );
      assert.ok(Array.isArray(inspections));
      const inspection: unknown = inspections[0];
      assert.ok(
        typeof inspection === "object" &&
          inspection !== null &&
          "HostConfig" in inspection &&
          "Mounts" in inspection,
      );
      const host = inspection.HostConfig;
      assert.ok(
        typeof host === "object" &&
          host !== null &&
          "Tmpfs" in host &&
          typeof host.Tmpfs === "object" &&
          host.Tmpfs !== null,
      );
      assert.ok(Object.hasOwn(host.Tmpfs, "/var/lib/postgresql/data"));
      assert.ok(Array.isArray(inspection.Mounts));
      assert.ok(
        inspection.Mounts.every(
          (mount: unknown) =>
            typeof mount === "object" && mount !== null && "Type" in mount && mount.Type !== "bind",
        ),
      );
      t.diagnostic(
        "Actual PostgreSQL " + version + "; Docker host 127.0.0.1:55432 maps to bridge port 5432.",
      );
    } else {
      const data = join(temp, "cluster");
      const pgctl = join(binaries, `pg_ctl${suffix}`);
      let started = false;
      t.after(async () => {
        if (started)
          await execute(pgctl, ["-D", data, "-m", "fast", "-w", "stop"], { windowsHide: true });
        await rm(temp, { recursive: true, force: true });
      });
      const pgVersion = await execute(join(binaries, `postgres${suffix}`), ["--version"]);
      const detectedMajor = pgVersion.stdout.match(/PostgreSQL\) (\d+)/)?.[1];
      assert.ok(detectedMajor, "PostgreSQL must report its major version");
      major = detectedMajor;
      const passwordFile = join(temp, "password");
      await writeFile(passwordFile, password, { mode: 0o600 });
      await execute(
        join(binaries, `initdb${suffix}`),
        [
          "-D",
          data,
          "-U",
          "ft_test_owner",
          "-A",
          "scram-sha-256",
          "--pwfile",
          passwordFile,
          "--encoding=UTF8",
          "--no-locale",
        ],
        { windowsHide: true },
      );
      await rm(passwordFile);
      // Windows pg_ctl's detached server can inherit pipes: wait for exit, not pipe EOF.
      await new Promise<void>((done, reject) => {
        const child = spawn(
          pgctl,
          [
            "-D",
            data,
            "-l",
            join(temp, "postgres.log"),
            "-o",
            "-h 127.0.0.1 -p 55432",
            "-w",
            "start",
          ],
          { windowsHide: true, stdio: "ignore" },
        );
        child.once("error", reject);
        child.once("exit", (code) => {
          started = code === 0;
          code === 0 ? done() : reject(new Error(`pg_ctl start exited ${code}`));
        });
      });
    }
    const env = {
      ...process.env,
      FITTRACK_LOCAL_DATABASE_URL: `postgresql://ft_test_owner:${password}@127.0.0.1:55432/postgres?sslmode=disable`,
      FITTRACK_LOCAL_RUNTIME_DATABASE_URL: `postgresql://fittrack_app:${randomBytes(24).toString("hex")}@127.0.0.1:55432/postgres?sslmode=disable`,
      FITTRACK_LOCAL_POSTGRES_CONTAINER: docker ? container : "",
      GOMAXPROCS: "2",
    };
    async function sql(
      database: string,
      statement: string,
      roleURL = env.FITTRACK_LOCAL_DATABASE_URL,
    ) {
      const url = new URL(roleURL);
      url.pathname = `/${database}`;
      const sqlEnv: NodeJS.ProcessEnv = {
        ...env,
        PGDATABASE: database,
        PGHOST: "127.0.0.1",
        PGPORT: "55432",
        PGUSER: decodeURIComponent(url.username),
        PGPASSWORD: decodeURIComponent(url.password),
        PGOPTIONS: "",
        PGSSLMODE: "disable",
      };
      for (const key of Object.keys(sqlEnv))
        if (/^(PGHOSTADDR|PGSERVICE|PGSERVICEFILE)$/i.test(key)) delete sqlEnv[key];
      if (docker) sqlEnv.PGPORT = "5432";
      const command = docker ? "docker" : join(binaries, `psql${suffix}`);
      const args = docker
        ? [
            "--host",
            dockerHost,
            "exec",
            "-i",
            ...[
              "PGDATABASE",
              "PGHOST",
              "PGPORT",
              "PGUSER",
              "PGPASSWORD",
              "PGOPTIONS",
              "PGSSLMODE",
            ].flatMap((key) => ["-e", key]),
            container,
            "psql",
          ]
        : [];
      const child = spawn(command, [...args, "-X", "-At", "-v", "ON_ERROR_STOP=1"], {
        env: docker ? localDockerEnv(sqlEnv) : sqlEnv,
        windowsHide: true,
        stdio: ["pipe", "pipe", "pipe"],
      });
      let output = "",
        errorOutput = "";
      child.stdout.on("data", (chunk) => {
        output += chunk;
      });
      child.stderr.on("data", (chunk) => {
        errorOutput += chunk;
      });
      const result = new Promise<string>((done, reject) => {
        child.once("error", () => reject(new Error("psql could not start")));
        child.once("exit", (code) =>
          code === 0
            ? done(output.trim())
            : reject(
                new Error(
                  `psql failed with exit ${code}: ${errorOutput.replaceAll(password, "[redacted]").replaceAll(decodeURIComponent(url.password), "[redacted]")}`,
                ),
              ),
        );
      });
      child.stdin.end(statement);
      return result;
    }
    const repo = join(temp, "repository");
    await mkdir(repo);
    // Commit only synthetic fixture inputs. No real checkout or user files are staged.
    const inputs = [
      "server/go.mod",
      "server/go.sum",
      "server/cmd/worktree-db",
      "server/scripts/provision-runtime-role.sql",
      "server/migrations",
    ];
    if (process.env.FITTRACK_TEST_API)
      inputs.push(
        "client/vite.config.js",
        "client/playwright.config.ts",
        "client/scripts",
        "client/package.json",
        "client/tsconfig.json",
        "client/index.html",
        "client/src",
      );
    for (const path of inputs) {
      await cp(join(source, path), join(repo, path), { recursive: true });
    }
    const git = (...args: string[]) => execute("git", args, { cwd: repo });
    await git("init", "--quiet");
    await git("add", ".");
    await git(
      "-c",
      "user.name=Fixture",
      "-c",
      "user.email=fixture@example.test",
      "commit",
      "--quiet",
      "-m",
      "fixture",
    );
    await writeFile(join(repo, "unrelated.txt"), "preserve me");
    const roots = [join(temp, "agent one"), join(temp, "agent two")] as const;
    for (const [index, root] of roots.entries())
      await git("worktree", "add", "--quiet", "-b", `agent-${index}`, root);
    if (process.env.FITTRACK_TEST_API) {
      for (const root of roots)
        await symlink(
          join(source, "client/node_modules"),
          join(root, "client/node_modules"),
          process.platform === "win32" ? "junction" : "dir",
        );
    }
    const cli = (root: string, args: string[], overrides: NodeJS.ProcessEnv = {}) =>
      execute(process.execPath, [script, ...args], {
        cwd: root,
        env: { ...env, ...overrides },
        windowsHide: true,
      });
    await Promise.all(roots.map((root) => cli(root, ["init"])));
    const readConfig = async (root: string) => {
      const config = await loadConfig(await discover(root));
      assert.deepEqual(JSON.parse((await cli(root, ["status"])).stdout), config);
      return config;
    };
    const configs = await Promise.all([readConfig(roots[0]), readConfig(roots[1])]);
    assert.notEqual(configs[0].ports.api, configs[1].ports.api);
    assert.equal(new Set(configs.flatMap((config) => Object.values(config.databases))).size, 8);
    await assert.rejects(
      cli(roots[0], ["db", "setup", "--major", String(Number(major) + 1)]),
      /major does not match/,
    );
    await assert.rejects(cli(roots[0], ["db", "setup"]), /Expected db setup/);
    assert.equal(
      await sql("postgres", "SELECT count(*) FROM pg_database WHERE datname LIKE 'ft_%';"),
      "0",
    );
    // Concurrent setup safely serializes at the database boundary: retry only lock contention.
    async function setup(root: string) {
      for (let attempt = 0; attempt < 60; attempt++) {
        try {
          return await cli(root, ["db", "setup", "--major", major]);
        } catch (error) {
          if (
            !(
              error instanceof Error &&
              "stderr" in error &&
              typeof error.stderr === "string" &&
              error.stderr.includes("cluster lock")
            ) ||
            attempt === 59
          )
            throw error;
          await new Promise((done) => setTimeout(done, 1000));
        }
      }
    }
    async function setupBoth() {
      const results = await Promise.allSettled(roots.map(setup));
      for (const result of results) if (result.status === "rejected") throw result.reason;
    }
    if (docker) {
      // The former endpoint guard must reject this real NAT mapping without opt-in.
      await assert.rejects(
        cli(roots[0], ["db", "setup", "--major", major], { FITTRACK_LOCAL_POSTGRES_CONTAINER: "" }),
        /server endpoint/,
      );
      // A stopped, unlabelled task-owned container cannot authorize the live endpoint.
      const unowned = container + "-unowned";
      await execute("docker", [
        "create",
        "--name",
        unowned,
        "--label",
        "fittrack.test-owner=" + container,
        "postgres:15.19-bookworm@sha256:539ceaaae49b3a7c8a04467cf00cc6788d8e3f1675df41860d86eebc4c40524f",
      ]);
      try {
        await assert.rejects(
          cli(roots[0], ["db", "setup", "--major", major], {
            FITTRACK_LOCAL_POSTGRES_CONTAINER: unowned,
          }),
          /refusing unowned/,
        );
      } finally {
        await execute("docker", ["rm", "-v", unowned]);
      }
      await assert.rejects(
        cli(roots[0], ["db", "setup", "--major", major], {
          FITTRACK_LOCAL_POSTGRES_CONTAINER: "missing-fittrack-test-container",
        }),
        /cannot inspect/,
      );
    }
    for (const value of [
      "postgresql://owner:unused@remote.example:55432/postgres",
      "postgresql://owner:unused@127.0.0.1:5432/postgres",
      "postgresql://owner:unused@127.0.0.1:55432/postgres?host=remote.example",
    ]) {
      await assert.rejects(
        cli(roots[0], ["db", "setup", "--major", major], { FITTRACK_LOCAL_DATABASE_URL: value }),
        /Remote databases|query parameter/,
      );
    }
    await setupBoth();
    for (const config of configs) {
      await sql(
        config.databases.rls,
        await readFile(join(source, "server/scripts/verify-runtime-role.sql"), "utf8"),
        env.FITTRACK_LOCAL_RUNTIME_DATABASE_URL,
      );
      await sql(
        config.databases.rls,
        await readFile(join(source, "server/scripts/test-runtime-role.sql"), "utf8"),
        env.FITTRACK_LOCAL_RUNTIME_DATABASE_URL,
      );
    }
    if (docker)
      await cli(roots[0], ["db", "setup", "--major", major], {
        DOCKER_HOST: "tcp://192.0.2.1:2375",
        DOCKER_CONTEXT: "unrelated-remote",
      });
    const roleBefore = await sql(
      "postgres",
      "SELECT rolpassword FROM pg_authid WHERE rolname='fittrack_app';",
    );
    for (const [index, config] of configs.entries()) {
      for (const database of Object.values(config.databases)) {
        assert.equal(
          await sql(database, "SELECT max(version_id) FROM goose_db_version WHERE is_applied;"),
          "33",
        );
        await sql(
          database,
          `CREATE TABLE worktree_probe(value text); INSERT INTO worktree_probe VALUES ('agent-${index}');`,
        );
      }
    }
    await setupBoth();
    for (const [index, config] of configs.entries()) {
      for (const database of Object.values(config.databases))
        assert.equal(await sql(database, "SELECT value FROM worktree_probe;"), `agent-${index}`);
      assert.equal(
        await sql(
          config.databases.rls,
          "SELECT current_user;",
          env.FITTRACK_LOCAL_RUNTIME_DATABASE_URL,
        ),
        "fittrack_app",
      );
    }
    assert.ok(
      (await sql("postgres", "SELECT rolpassword FROM pg_authid WHERE rolname='fittrack_app';")) ===
        roleBefore,
      "existing runtime credential must not change",
    );
    if (process.env.FITTRACK_TEST_API) {
      const servers: {
        root: string;
        child: ChildProcess;
        output: string;
        exited: Promise<number | null>;
      }[] = [];
      const endpointsEnv = {
        ...env,
        PROJECT_ID: process.env.FITTRACK_TEST_PROJECT_ID ?? "local-worktree-integration",
        E2E_LOCAL_AUTH_ENABLED: "true",
        VITE_E2E_LOCAL_AUTH_ENABLED: "true",
        INNGEST_EVENT_KEY: "",
        INNGEST_SIGNING_KEY: "",
        STRIPE_SECRET_KEY: "",
        STRIPE_WEBHOOK_SECRET: "",
        GEMINI_API_KEY: "",
        GOOGLE_API_KEY: "",
        NODE_OPTIONS: "--max-old-space-size=384",
        DB_MAX_CONNS: "3",
      };
      const redact = (value: string) =>
        value
          .replaceAll(password, "[redacted]")
          .replaceAll(new URL(env.FITTRACK_LOCAL_RUNTIME_DATABASE_URL).password, "[redacted]")
          .replace(/postgres(?:ql)?:\/\/\S+/g, "[connection withheld]");
      const launch = (
        root: string,
        command: string,
        args: string[],
        extra: NodeJS.ProcessEnv = {},
      ) => {
        const child = spawn(process.execPath, [script, "run", "e2e", "--", command, ...args], {
          cwd: root,
          env: { ...endpointsEnv, ...extra },
          windowsHide: true,
          stdio: ["ignore", "pipe", "pipe"],
        });
        const server = { root, child, output: "" };
        const capture = (chunk: Buffer) => {
          server.output = (server.output + chunk).slice(-4000);
        };
        child.stdout.on("data", capture);
        child.stderr.on("data", capture);
        const exited = new Promise<number | null>((done, reject) => {
          child.once("error", reject);
          child.once("exit", done);
        });
        servers.push({
          ...server,
          exited,
          get output() {
            return server.output;
          },
        });
      };
      const waitHTTP = async (url: string, wanted = 200) => {
        for (let attempt = 0; attempt < 300; attempt++) {
          try {
            const response = await fetch(url, { signal: AbortSignal.timeout(1000) });
            if (response.status === wanted) return response;
          } catch {}
          const failed = servers.find((server) => server.child.exitCode !== null);
          if (failed)
            throw new Error(
              `Endpoint process exited ${failed.child.exitCode}: ${redact(failed.output)}`,
            );
          await new Promise((done) => setTimeout(done, 100));
        }
        throw new Error(
          `Endpoint did not become ready: ${url}\n${redact(servers.map((server) => server.output).join("\n"))}`,
        );
      };
      try {
        for (const [index, root] of roots.entries()) {
          launch(root, process.env.FITTRACK_TEST_API, [], {
            E2E_LOCAL_AUTH_USER_ID: `agent-${index}`,
          });
        }
        for (const [index, config] of configs.entries()) {
          const root = roots[index];
          assert.ok(root);
          const api = `http://127.0.0.1:${config.ports.e2eApi}`;
          const frontend = `http://127.0.0.1:${config.ports.e2eFrontend}`;
          await waitHTTP(`${api}/ready`);
          await waitHTTP(`http://127.0.0.1:${config.ports.e2eMetrics}/metrics`);
          launch(root, process.execPath, [
            join(source, "client/node_modules/vite/bin/vite.js"),
            join(root, "client"),
            "--config",
            join(root, "client/vite.config.js"),
          ]);
          const bootstrap = await fetch(`${api}/dev/e2e/auth/bootstrap`, { method: "POST" });
          assert.equal(bootstrap.status, 200);
          const identity: unknown = await bootstrap.json();
          assert.ok(typeof identity === "object" && identity !== null && "user_id" in identity);
          assert.equal(identity.user_id, `agent-${index}`);
          await waitHTTP(`${frontend}/api/workouts`, 401);
          const result = await fetch(`${frontend}/api/workouts`, {
            headers: { "x-fittrack-dev-e2e-user": `agent-${index}` },
          });
          assert.equal(result.status, 200, "real Vite proxy must reach its own authenticated API");
          const crossUser = await fetch(`${frontend}/api/workouts`, {
            headers: { "x-fittrack-dev-e2e-user": `agent-${1 - index}` },
          });
          assert.equal(crossUser.status, 401);
          assert.equal(
            await sql(config.databases.e2e, "SELECT string_agg(user_id, ',') FROM users;"),
            `agent-${index}`,
          );
          const configURL = pathToFileURL(join(root, "client/playwright.config.ts")).href;
          const output = await cli(root, [
            "run",
            "e2e",
            "--",
            "bun",
            "-e",
            `import config from ${JSON.stringify(configURL)}; console.log(JSON.stringify({base:config.use.baseURL, server:config.webServer}));`,
          ]);
          const playwright = JSON.parse(output.stdout.trim());
          assert.equal(playwright.base, frontend);
          assert.equal(playwright.server.url, frontend);
          assert.equal(playwright.server.reuseExistingServer, false);
          assert.ok(
            playwright.server.command.includes(`--port ${config.ports.e2eFrontend} --strictPort`),
          );
        }
        await assert.rejects(setup(roots[0]), /Stop this worktree's commands/);
        await cli(roots[0], ["stop", "e2e"]);
        await Promise.all(
          servers.filter((server) => server.root === roots[0]).map((server) => server.exited),
        );
        await cli(roots[0], ["check"]);
        await waitHTTP(`http://127.0.0.1:${configs[1].ports.e2eApi}/ready`);
        t.diagnostic(
          "Two real APIs/Vite proxies: distinct auth identities, cross-user refusal, Playwright origins, and stopping one preserves the other.",
        );
      } finally {
        for (const root of roots) await cli(root, ["stop", "e2e"]);
        await Promise.all(servers.map((server) => server.exited));
        for (const root of roots) await cli(root, ["check"]);
      }
    } else {
      t.diagnostic(
        "API/Vite/Playwright runtime proof skipped; set FITTRACK_TEST_API to a binary built from this checkout and install client dependencies.",
      );
    }
    // Real lock contention with another database client, no mock of the setup lock.
    const lockTag = "ft_lock_" + randomBytes(8).toString("hex");
    const lock = sql(
      "postgres",
      `SET application_name='${lockTag}'; SELECT pg_advisory_lock(714624891036); SELECT pg_sleep(120);`,
    ).catch((error) => error);
    try {
      let held = false;
      for (let attempt = 0; attempt < 40; attempt++) {
        held =
          (await sql(
            "postgres",
            `SELECT EXISTS(SELECT 1 FROM pg_locks l JOIN pg_stat_activity a USING(pid) WHERE a.application_name='${lockTag}' AND l.locktype='advisory' AND l.granted);`,
          )) === "t";
        if (held) break;
        await new Promise((done) => setTimeout(done, 100));
      }
      assert.ok(held, "independent client must hold the setup lock before competing");
      await assert.rejects(cli(roots[0], ["db", "setup", "--major", major]), /cluster lock/);
    } finally {
      await sql(
        "postgres",
        `SELECT pg_cancel_backend(pid) FROM pg_stat_activity WHERE application_name='${lockTag}';`,
      );
      await lock;
    }
    await sql("postgres", `COMMENT ON DATABASE "${configs[0].databases.dev}" IS 'unrelated';`);
    await assert.rejects(setup(roots[0]), /owner\/marker mismatch/);
    assert.equal(
      await sql(configs[0].databases.dev, "SELECT value FROM worktree_probe;"),
      "agent-0",
    );
    await sql(
      "postgres",
      `COMMENT ON DATABASE "${configs[0].databases.dev}" IS 'fittrack-worktree:${configs[0].id}'; ALTER ROLE fittrack_app BYPASSRLS;`,
    );
    await assert.rejects(setup(roots[0]), /existing fittrack_app is unsafe/);
    assert.equal(
      await sql("postgres", "SELECT rolbypassrls FROM pg_roles WHERE rolname='fittrack_app';"),
      "t",
    );
    await sql("postgres", "ALTER ROLE fittrack_app NOBYPASSRLS;");
    await setup(roots[0]);
    assert.equal(await readFile(join(repo, "unrelated.txt"), "utf8"), "preserve me");
    assert.equal((await git("status", "--porcelain")).stdout.trim(), "?? unrelated.txt");
    if (docker) {
      const logs = await execute("docker", ["logs", container]);
      await writeFile(join(temp, "postgres.log"), logs.stdout + logs.stderr);
    }
    assert.ok(!(await readFile(join(temp, "postgres.log"), "utf8")).includes(password));
    assert.ok(
      !(await readFile(join(temp, "postgres.log"), "utf8")).includes(
        new URL(env.FITTRACK_LOCAL_RUNTIME_DATABASE_URL).password,
      ),
    );
    t.diagnostic(
      `PostgreSQL ${major}: eight migrated databases, two real linked worktrees; production version selection is independent of this disposable test.`,
    );
  },
);
