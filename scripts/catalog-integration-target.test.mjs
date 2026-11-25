import { test } from "node:test";
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { resolve, join } from "node:path";
import { tmpdir } from "node:os";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));

test("make test-integration refuses missing native runtime before database setup or tests", () => {
  const result = spawnSync("make", ["-C", join(root, "server"), "test-integration"], {
    encoding: "utf8",
    timeout: 30000,
    env: { ...process.env, WORKTREE_RUNTIME: join(tmpdir(), "fittrack-nonexistent-runtime-for-test") },
  });
  assert.equal(result.error, undefined);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr + result.stdout, /external worktree-runtime CLI/);
  assert.doesNotMatch(result.stderr + result.stdout, /Starting database container|Applying migrations|Running go test/);
});
