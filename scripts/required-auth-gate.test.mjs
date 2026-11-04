// Exercise the real Playwright process: reporter unit mocks cannot prove exit status.
import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, writeFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { promisify } from "node:util";
import test from "node:test";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const playwright = join(root, "client/node_modules/playwright/cli.js");
const testImport = JSON.stringify(pathToFileURL(join(root, "client/node_modules/@playwright/test/index.mjs")).href);
const reporter = join(root, "client/scripts/required-auth-reporter.ts");

test("required auth gate preserves Playwright failures and rejects incomplete coverage", { timeout: 120000 }, async t => {
  const directory = await mkdtemp(join(tmpdir(), "fittrack-required-auth-"));
  t.after(() => rm(directory, { recursive: true, force: true }));
  await writeFile(join(directory, "package.json"), '{"type":"module"}');
  await writeFile(join(directory, "playwright.config.mjs"), `export default {
    testDir: '.', workers: 1, retries: 1, forbidOnly: true, timeout: 1000,
    reporter: [['line'], [${JSON.stringify(reporter)}]]
  };`);
  for (const [name, body, succeeds, args = []] of [
    ["success", 'test("works", () => expect(2 + 2).toBe(4));', true],
    ["missing auth", 'test("needs auth", () => { test.skip(!process.env.FITTRACK_GATE_AUTH, "Missing auth state"); });', false],
    ["static skip", 'test.skip("disabled", () => {});', false],
    ["assertion failure", 'test("broken product", () => expect(1).toBe(2));', false],
    ["failure then retry skip", 'test("retry", ({}, info) => { test.skip(info.retry > 0); expect(1).toBe(2); });', false],
    ["failure then retry pass", 'test("flaky", ({}, info) => expect(info.retry).toBe(1));', false],
    ["expected failure", 'test("known failure", () => { test.fail(); expect(1).toBe(2); });', false],
    ["empty selection", 'test("exists", () => {});', false, ['--grep', 'no-such-test', '--pass-with-no-tests']],
  ]) {
    await t.test(name, async () => {
      await writeFile(join(directory, "gate.test.mjs"), `import { test, expect } from ${testImport};\n${body}`);
      const result = await promisify(execFile)(process.execPath, [playwright, "test", "--config", join(directory, "playwright.config.mjs"), ...args], {
        cwd: directory, env: { ...process.env, FITTRACK_GATE_AUTH: "", FORCE_COLOR: "0" }, timeout: 20000, windowsHide: true,
      }).then(value => ({ code: 0, output: value.stdout + value.stderr }), error => ({ code: error.code, output: error.stdout + error.stderr }));
      assert.equal(result.code, succeeds ? 0 : 1, result.output);
      assert.match(result.output, /Required authenticated suite:/);
      if (name === "missing auth" || name === "static skip" || name === "failure then retry skip") assert.match(result.output, /1 skipped\/unexecuted/);
      if (name === "empty selection") assert.match(result.output, /0 selected/);
    });
  }
});
