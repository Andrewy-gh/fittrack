import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import test from "node:test";

const hook = readFileSync(new URL("../.husky/pre-push", import.meta.url), "utf8");
const shell = process.platform === "win32" ? "C:/Program Files/Git/bin/sh.exe" : "sh";
const zero = "0".repeat(40);

function fixture(t) {
  const root = mkdtempSync(join(tmpdir(), "fittrack-push-hook-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  function cmd(file, args, options = {}) {
    const result = spawnSync(file, args, { cwd: root, encoding: "utf8", ...options });
    if (result.error) throw result.error;
    return result;
  }
  function git(...args) {
    const result = cmd("git", args);
    assert.equal(result.status, 0, result.stderr);
    return result.stdout.trim();
  }
  git("init", "-q");
  git("config", "user.email", "fixture@example.invalid");
  git("config", "user.name", "Push Fixture");
  // Fixture commits must never discover the host's hooks.
  git("config", "core.hooksPath", "/dev/null");
  function commit(path, body = "fixture\n") {
    mkdirSync(dirname(join(root, path)), { recursive: true });
    writeFileSync(join(root, path), body);
    git("add", "--", path);
    git("commit", "-qm", "fixture");
    return git("rev-parse", "HEAD");
  }
  const base = commit("seed.txt");
  mkdirSync(join(root, "bin"));
  mkdirSync(join(root, "server"));
  mkdirSync(join(root, "scripts"));
  writeFileSync(join(root, "hook"), hook);
  const locate = (name) => {
    const r = cmd(shell, ["-c", `command -v ${name}`]);
    assert.equal(r.status, 0, r.stderr);
    return r.stdout.trim();
  };
  const realGit = locate("git");
  const bin = cmd(shell, ["-c", "cd bin && pwd"]).stdout.trim();
  const gitDir = cmd(shell, ["-c", `dirname '${realGit}'`]).stdout.trim();
  const env = { ...process.env, CALL_LOG: join(root, "calls.log"), READY: "0", TEST_EXIT: "0" };
  writeFileSync(env.CALL_LOG, "");
  function executable(name, body) {
    writeFileSync(join(root, "bin", name), `#!/bin/sh\n${body}\n`, { mode: 0o755 });
  }
  executable("git", `if [ "$FAIL_GIT" = "$1" ]; then exit 73; fi\nexec '${realGit}' "$@"`);
  executable("grep", `exec '${locate("grep")}' "$@"`);
  for (const name of ["make", "docker", "go", "goose", "doppler"]) {
    executable(name, `echo 'UNSAFE ${name}' >> "$CALL_LOG"; exit 89`);
  }
  writeFileSync(join(root, "server", "setenv.sh"), 'echo "UNSAFE setenv" >> "$CALL_LOG"; exit 89\n');
  executable("node", `
if read -r unexpected; then echo 'CONSUMED REF STDIN' >> "$CALL_LOG"; exit 88; fi
if [ "\${GIT_DIR+x}\${GIT_WORK_TREE+x}\${GIT_PREFIX+x}" != '' ]; then
  echo 'LEAKED GIT HOOK CONTEXT' >> "$CALL_LOG"; exit 85
fi
case "$*" in
  'scripts/worktree.ts doctor --mode test --json')
    [ "$PWD" = "$FIXTURE_ROOT" ] || exit 86
    echo doctor >> "$CALL_LOG"
    [ "$READY" = 1 ] || exit 1 ;;
  '../scripts/worktree.ts run test -- go test -p 1 -short ./...')
    [ "$PWD" = "$FIXTURE_ROOT/server" ] || exit 86
    echo 'managed test' >> "$CALL_LOG"
    exit "$TEST_EXIT" ;;
  *) echo "UNSAFE node command" >> "$CALL_LOG"; exit 87 ;;
esac`);
  env.FIXTURE_ROOT = cmd(shell, ["-c", "pwd"]).stdout.trim();
  const line = (local, remote = base) => `refs/heads/topic ${local} refs/heads/topic ${remote}\n`;
  function run(input) {
    return cmd(shell, ["-c", `export PATH='${bin}:${gitDir}'; . ./hook origin unused`], { env, input });
  }
  const calls = () => { try { return readFileSync(env.CALL_LOG, "utf8"); } catch { return ""; } };
  return { root, base, commit, git, env, line, run, calls };
}

test("unprepared server push refuses before tests, legacy environment, or service mutations", (t) => {
  const f = fixture(t);
  const sha = f.commit("server/model.go");
  const r = f.run(f.line(sha));
  assert.notEqual(r.status, 0);
  assert.equal(f.calls(), "doctor\n");
  assert.match(r.stderr, /prepare --mode test/);
});

test("prepared first/existing/multiple refs run the full short suite in managed TEST and preserve dirtiness", (t) => {
  for (const kind of ["first", "existing", "multiple", "root"]) {
    const f = fixture(t);
    const sha = f.commit("server/model.go");
    f.env.READY = "1";
    if (kind === "existing") {
      f.env.GIT_DIR = f.git("rev-parse", "--absolute-git-dir");
      f.env.GIT_WORK_TREE = f.env.FIXTURE_ROOT;
      f.env.GIT_PREFIX = "";
    }
    f.commit("staged.txt");
    writeFileSync(join(f.root, "staged.txt"), "keep staged\n");
    f.git("add", "staged.txt");
    writeFileSync(join(f.root, "seed.txt"), "keep unstaged\n");
    const before = f.git("status", "--porcelain=v1");
    const index = f.git("diff", "--cached");
    if (kind === "root") {
      // A first push must also examine a server file in the initial commit.
      f.git("checkout", "--orphan", "root-only");
      f.git("rm", "-rf", "--cached", ".");
      f.commit("server/initial.go");
    }
    const local = kind === "root" ? f.git("rev-parse", "HEAD") : sha;
    const input = kind === "multiple" ? f.line(f.base) + f.line(sha) + f.line(zero) : f.line(local, kind === "existing" ? f.base : zero);
    const r = f.run(input);
    assert.equal(r.status, 0, r.stderr);
    assert.equal(f.calls(), "doctor\nmanaged test\n");
    if (kind !== "root") {
      assert.equal(f.git("status", "--porcelain=v1"), before);
      assert.equal(f.git("diff", "--cached"), index);
      assert.equal(readFileSync(join(f.root, "seed.txt"), "utf8"), "keep unstaged\n");
    }
  }
});

test("missing Node refuses with setup guidance and no commands", (t) => {
  const f = fixture(t);
  const sha = f.commit("server/model.go");
  rmSync(join(f.root, "bin", "node"));
  const r = f.run(f.line(sha));
  assert.notEqual(r.status, 0);
  assert.match(r.stderr, /Node 24\.22\.1/);
  assert.equal(f.calls(), "");
});

test("managed verification failure blocks publication", (t) => {
  const f = fixture(t);
  const sha = f.commit("server/model.go");
  f.env.READY = "1";
  f.env.TEST_EXIT = "27";
  assert.equal(f.run(f.line(sha)).status, 27);
  assert.equal(f.calls(), "doctor\nmanaged test\n");
});

test("docs-only, client-only, unchanged refs and ref deletion skip verification", (t) => {
  for (const path of ["server/README.md", "server/docs/spec.json", "server/nested/notes.md", "client/app.ts"]) {
    const f = fixture(t);
    const sha = f.commit(path);
    const r = f.run(f.line(sha) + f.line(zero) + f.line(f.base));
    assert.equal(r.status, 0, r.stderr);
    assert.equal(f.calls(), "");
  }
});

test("first push excludes already published server history", (t) => {
  const f = fixture(t);
  const published = f.commit("server/published.go");
  f.git("update-ref", "refs/remotes/origin/main", published);
  const docs = f.commit("server/notes.md");
  assert.equal(f.run(f.line(docs, zero)).status, 0);
  assert.equal(f.calls(), "");
});

test("server changes introduced only by merge resolution require verification", (t) => {
  const f = fixture(t);
  const docs = f.commit("notes.md");
  writeFileSync(join(f.root, "server", "resolution.go"), "merge resolution\n");
  f.git("add", "server/resolution.go");
  const tree = f.git("write-tree");
  const merge = f.git("commit-tree", tree, "-p", f.base, "-p", docs, "-m", "resolution");
  f.env.READY = "1";
  const r = f.run(f.line(merge, docs));
  assert.equal(r.status, 0, r.stderr);
  assert.equal(f.calls(), "doctor\nmanaged test\n");
});

test("server file deletion still requires managed verification", (t) => {
  const f = fixture(t);
  const old = f.commit("server/deleted.go");
  f.git("rm", "server/deleted.go");
  f.git("commit", "-qm", "delete");
  f.env.READY = "1";
  assert.equal(f.run(f.line(f.git("rev-parse", "HEAD"), old)).status, 0);
  assert.equal(f.calls(), "doctor\nmanaged test\n");
});

test("Git inspection errors fail closed, including later refs", (t) => {
  for (const failure of ["rev-list", "diff-tree", "rev-parse", "invalid-ref"]) {
    const f = fixture(t);
    const sha = f.commit("server/model.go");
    f.env.FAIL_GIT = failure;
    const input = failure === "invalid-ref" ? f.line(sha) + f.line("bad-object") : f.line(sha);
    assert.notEqual(f.run(input).status, 0);
    assert.equal(f.calls(), "");
  }
});
