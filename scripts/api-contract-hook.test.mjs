import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  mkdtempSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

const repository = new URL("..", import.meta.url).pathname.replace(
  /^\/(\w:)/,
  "$1",
);
const hook = readFileSync(join(repository, ".husky", "pre-commit"), "utf8");
const pathInspector = readFileSync(
  join(repository, "scripts", "api-contract-hook.mjs"),
  "utf8",
);
const shell =
  process.platform === "win32" ? "C:/Program Files/Git/bin/sh.exe" : "sh";

function command(args, cwd, options = {}) {
  const result = spawnSync(args[0], args.slice(1), {
    cwd,
    encoding: "utf8",
    ...options,
  });
  if (result.error) throw result.error;
  return result;
}

function fixture(t, generator = "noop") {
  const root = mkdtempSync(join(tmpdir(), "fittrack-api-hook-test-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  command(["git", "init", "-q"], root);
  command(["git", "config", "user.email", "test@example.invalid"], root);
  command(["git", "config", "user.name", "Hook Fixture"], root);
  mkdirSync(join(root, "scripts"), { recursive: true });
  writeFileSync(join(root, "scripts", "api-contract-hook.mjs"), pathInspector);
  command(["git", "add", "scripts/api-contract-hook.mjs"], root);
  mkdirSync(join(root, "server"), { recursive: true });
  writeFileSync(
    join(root, "server", "Makefile"),
    "test-fast:\n\t@echo make test-fast >> ../calls.log\n",
  );
  command(["git", "add", "server/Makefile"], root);
  writeFileSync(join(root, ".husky-pre-commit"), hook);
  const env = {
    ...process.env,
    CALL_LOG: "calls.log",
    GENERATOR_MODE: generator,
  };
  const stage = (path, body = "fixture\n") => {
    const full = join(root, path);
    mkdirSync(join(full, ".."), { recursive: true });
    writeFileSync(full, body);
    command(["git", "add", "--", path], root);
  };
  const commit = () => command(["git", "commit", "-qm", "fixture"], root);
  const runHook = () =>
    command(
      [
        shell,
        "-c",
        `bun() { printf '%s\\n' "$*" >> "$CALL_LOG"; case "$*" in *api:generate*) [ "$GENERATOR_MODE" != fail ] || return 23; if [ "$GENERATOR_MODE" = change ]; then mkdir -p client/src/client; printf generated > client/src/client/new.gen.ts; fi ;; esac; }; . ./.husky-pre-commit`,
      ],
      root,
      { env },
    );
  const calls = () => {
    try {
      return readFileSync(join(root, env.CALL_LOG), "utf8");
    } catch {
      return "";
    }
  };
  return { root, stage, commit, runHook, calls, env };
}

test("skips generation for frontend, markdown, migrations, and backend tests", (t) => {
  for (const path of [
    "README.md",
    "docs/notes.md",
    "server/migrations/00099_example.sql",
    "server/internal/sample/handler_test.go",
    "client/src/app.tsx",
  ]) {
    const f = fixture(t);
    f.stage("seed.txt");
    f.commit();
    f.stage(path);
    const result = f.runHook();
    assert.equal(result.status, 0, result.stderr);
    assert.doesNotMatch(f.calls(), /api:generate/);
  }
});

test("generates for staged backend source, leaves output unstaged, and accepts reviewed output", (t) => {
  const f = fixture(t, "change");
  f.stage("seed.txt");
  f.commit();
  f.stage("server/internal/workout/assessment/policy/model.go");
  const blocked = f.runHook();
  assert.equal(blocked.status, 1);
  assert.match(blocked.stderr, /Review the output and stage it explicitly/);
  assert.match(f.calls(), /api:generate/);
  assert.match(
    readFileSync(join(f.root, "client/src/client/new.gen.ts"), "utf8"),
    /generated/,
  );
  assert.equal(
    command(["git", "diff", "--cached", "--name-only"], f.root).stdout.trim(),
    "server/internal/workout/assessment/policy/model.go",
  );

  f.stage("client/src/client/new.gen.ts");
  f.env.GENERATOR_MODE = "noop";
  const accepted = f.runHook();
  assert.equal(accepted.status, 0, accepted.stderr);
  assert.match(f.calls(), /fmt:check/);
  assert.match(f.calls(), /test-fast/);
});

test("deletions and both paths of a rename still trigger generation", (t) => {
  for (const mode of ["delete", "rename"]) {
    const f = fixture(t);
    f.stage("server/internal/sample/old.go");
    f.commit();
    if (mode === "delete") {
      command(["git", "rm", "-q", "server/internal/sample/old.go"], f.root);
    } else {
      command(
        [
          "git",
          "mv",
          "server/internal/sample/old.go",
          "server/internal/sample/new.go",
        ],
        f.root,
      );
    }
    const result = f.runHook();
    assert.equal(result.status, 0, result.stderr);
    assert.match(f.calls(), /api:generate/);
  }
});

test("formatter and custom lint inputs trigger generation and refuse partial staging", (t) => {
  for (const path of [
    "client/.oxfmtrc.json",
    "client/tools/oxlint/anti-slop/index.ts",
  ]) {
    const f = fixture(t);
    f.stage("seed.txt");
    f.commit();
    f.stage(path);
    const generated = f.runHook();
    assert.equal(generated.status, 0, generated.stderr);
    assert.match(f.calls(), /api:generate/);
    const callsBefore = f.calls();
    writeFileSync(join(f.root, path), "keep unstaged formatter input\n");
    const refused = f.runHook();
    assert.equal(refused.status, 1, refused.stderr);
    assert.match(refused.stderr, /relevant worktree changes are not staged/);
    assert.equal(f.calls(), callsBefore);
    assert.equal(
      readFileSync(join(f.root, path), "utf8"),
      "keep unstaged formatter input\n",
    );
  }
});

test("refuses relevant unstaged input and generated output without invoking generation", (t) => {
  const f = fixture(t);
  f.stage("seed.txt");
  f.commit();
  f.stage("server/internal/sample/model.go");
  writeFileSync(
    join(f.root, "server/internal/sample/model.go"),
    "partially staged\n",
  );
  const before = readFileSync(
    join(f.root, "server/internal/sample/model.go"),
    "utf8",
  );
  const result = f.runHook();
  assert.equal(result.status, 1);
  assert.match(result.stderr, /relevant worktree changes are not staged/);
  assert.equal(f.calls(), "");
  assert.equal(
    readFileSync(join(f.root, "server/internal/sample/model.go"), "utf8"),
    before,
  );
});

test("refuses unstaged generated files and untracked Go inputs before generation", (t) => {
  const f = fixture(t);
  f.stage("client/src/client/current.gen.ts");
  f.commit();
  f.stage("server/internal/sample/model.go");
  writeFileSync(
    join(f.root, "client/src/client/current.gen.ts"),
    "unstaged output\n",
  );
  const result = f.runHook();
  assert.equal(result.status, 1);
  assert.equal(f.calls(), "");
  // Resolve the output change so only the deep untracked input can block this run.
  command(["git", "add", "client/src/client/current.gen.ts"], f.root);
  mkdirSync(join(f.root, "server/internal/workout/assessment/policy"), {
    recursive: true,
  });
  writeFileSync(
    join(f.root, "server/internal/workout/assessment/policy/untracked.go"),
    "untracked input\n",
  );
  const untracked = f.runHook();
  assert.equal(untracked.status, 1);
  assert.match(
    untracked.stderr,
    /server\/internal\/workout\/assessment\/policy\/untracked\.go/,
  );
  assert.equal(f.calls(), "");
  assert.equal(
    readFileSync(
      join(f.root, "server/internal/workout/assessment/policy/untracked.go"),
      "utf8",
    ),
    "untracked input\n",
  );

  const ignored = fixture(t);
  ignored.stage(".gitignore", "client/src/client/ignored.gen.ts\n");
  ignored.commit();
  ignored.stage("server/internal/sample/model.go");
  mkdirSync(join(ignored.root, "client/src/client"), { recursive: true });
  writeFileSync(
    join(ignored.root, "client/src/client/ignored.gen.ts"),
    "keep ignored output\n",
  );
  const ignoredResult = ignored.runHook();
  assert.equal(ignoredResult.status, 1);
  assert.match(ignoredResult.stderr, /ignored\.gen\.ts/);
  assert.equal(ignored.calls(), "");
});

test("finds quoted non-ASCII untracked generated files without removing them", (t) => {
  const f = fixture(t);
  f.stage("seed.txt");
  f.commit();
  f.stage("server/internal/sample/model.go");
  const untrackedPath = "client/src/client/draft-é.ts";
  const fullPath = join(f.root, untrackedPath);
  mkdirSync(join(fullPath, ".."), { recursive: true });
  writeFileSync(fullPath, "keep this untracked file\n");

  const listed = command(
    ["git", "ls-files", "--others", "--", "client/src/client"],
    f.root,
  );
  assert.equal(listed.status, 0, listed.stderr);
  assert.match(listed.stdout, /draft-\\303\\251\.ts/);

  const result = f.runHook();
  assert.equal(result.status, 1);
  assert.match(result.stderr, /relevant worktree changes are not staged/);
  assert.equal(f.calls(), "");
  assert.equal(readFileSync(fullPath, "utf8"), "keep this untracked file\n");
});

test("unrelated unstaged edits survive while a staged API change is checked", (t) => {
  const f = fixture(t);
  f.stage("seed.txt");
  f.commit();
  f.stage("server/internal/sample/model.go");
  writeFileSync(join(f.root, "README.md"), "unrelated unstaged edit\n");
  const result = f.runHook();
  assert.equal(result.status, 0, result.stderr);
  assert.equal(
    readFileSync(join(f.root, "README.md"), "utf8"),
    "unrelated unstaged edit\n",
  );
  assert.match(f.calls(), /api:generate/);
});

test("generation failures stop the hook and propagate", (t) => {
  const f = fixture(t, "fail");
  f.stage("seed.txt");
  f.commit();
  f.stage("server/internal/sample/model.go");
  const result = f.runHook();
  assert.notEqual(result.status, 0);
  assert.match(f.calls(), /api:generate/);
  assert.equal(
    command(["git", "diff", "--cached", "--name-only"], f.root).stdout.trim(),
    "server/internal/sample/model.go",
  );
});
