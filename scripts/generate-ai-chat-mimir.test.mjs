import { test } from "node:test";
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));

test("generates the namespaced source byte-for-byte and rejects drift without rewriting", (t) => {
  const dir = mkdtempSync(join(tmpdir(), "ai-chat-alerts-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  for (const path of ["scripts", "ops/prometheus", "ops/grafana-cloud"]) {
    mkdirSync(join(dir, path), { recursive: true });
  }
  const script = "scripts/generate-ai-chat-mimir.mjs";
  const source = "ops/prometheus/ai-chat-alerts.yml";
  const target = "ops/grafana-cloud/ai-chat-alerts.mimir.yml";
  copyFileSync(join(root, script), join(dir, script));
  copyFileSync(join(root, source), join(dir, source));
  const run = (...args) => spawnSync(process.execPath, [join(dir, script), ...args], { encoding: "utf8" });

  assert.equal(run("--check").status, 1, "missing output must fail check");
  assert.equal(run().status, 0);
  const sourceBytes = readFileSync(join(dir, source));
  const newline = sourceBytes.includes(Buffer.from("\r\n")) ? "\r\n" : "\n";
  const expected = Buffer.concat([Buffer.from(`namespace: ai_chat_rollout${newline}`), sourceBytes]);
  assert.deepEqual(readFileSync(join(dir, target)), expected);
  assert.equal(run("--check").status, 0);
  const drift = Buffer.concat([expected, Buffer.from("# drift\n")]);
  writeFileSync(join(dir, target), drift);
  assert.equal(run("--check").status, 1);
  assert.deepEqual(readFileSync(join(dir, target)), drift, "check must not rewrite drift");
  assert.equal(run().status, 0);
  assert.deepEqual(readFileSync(join(dir, target)), expected);
  assert.equal(run("--unknown").status, 2);
});
