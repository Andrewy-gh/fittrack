import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import {
  mkdtempSync,
  mkdirSync,
  readFileSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "node:test";
import lintConfig from "../oxlint.config.ts";

const client = fileURLToPath(new URL("../", import.meta.url));
const read = (path) => readFileSync(join(client, path), "utf8");
const knipConfig = JSON.parse(read("knip.json"));

function fixture(t) {
  const dir = mkdtempSync(join(tmpdir(), "fittrack-guardrails-"));
  symlinkSync(
    join(client, "node_modules"),
    join(dir, "node_modules"),
    "junction",
  );
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  return dir;
}

function put(dir, path, content) {
  const target = join(dir, path);
  mkdirSync(dirname(target), { recursive: true });
  writeFileSync(target, content);
}

function cli(dir, tool, args) {
  const result = spawnSync(
    process.execPath,
    [join(client, "node_modules", tool), ...args],
    {
      cwd: dir,
      encoding: "utf8",
      timeout: 120_000,
      env: { ...process.env, NO_COLOR: "1" },
    },
  );
  assert.ifError(result.error);
  return { status: result.status, output: result.stdout + result.stderr };
}

function lint(dir, path, code, extra = []) {
  put(dir, path, code);
  return cli(dir, "oxlint/bin/oxlint", [
    "--config",
    "oxlint.json",
    path,
    ...extra,
  ]);
}

test("normal Oxlint config rejects static mutable training queries, not immutable data or text", (t) => {
  const dir = fixture(t);
  put(dir, "package.json", '{"type":"module"}');
  const config = {
    ...lintConfig,
    jsPlugins: lintConfig.jsPlugins.map((plugin) => ({
      ...plugin,
      specifier: resolve(client, plugin.specifier),
    })),
  };
  put(dir, "oxlint.json", JSON.stringify(config));
  const path = "src/features/workouts/pages/workout-route-composition.tsx";
  const invalid =
    "export const options = { ...getWorkoutsQueryOptions(), staleTime: 'static' };";
  const rejected = lint(dir, path, invalid);
  assert.equal(rejected.status, 1, rejected.output);
  assert.match(
    rejected.output,
    /training-cache\(no-static-training-data\)|training-cache\/no-static-training-data/,
  );
  // Negative control: the identical input is accepted without this rule.
  put(
    dir,
    "oxlint.json",
    JSON.stringify({
      ...config,
      overrides: config.overrides.map((override) => ({
        ...override,
        rules: {
          ...override.rules,
          "training-cache/no-static-training-data": "off",
        },
      })),
    }),
  );
  const disabled = lint(dir, path, invalid);
  assert.equal(disabled.status, 0, disabled.output);
  put(dir, "oxlint.json", JSON.stringify(config));

  for (const code of [
    "export const options = { ...getWorkoutsQueryOptions(), staleTime: Infinity };",
    "export const options = { ...getWorkoutsQueryOptions(), staleTime: 0 };",
    "export const options = { ...getWorkoutsQueryOptions() };",
    "export const immutable = { ...getExerciseCatalogQueryOptions(), staleTime: 'static' };",
    "export const unrelated = { staleTime: 'static', label: 'static' }; // staleTime: 'static'",
    "export const options = { ...getWorkoutsQueryOptions(), meta: { staleTime: 'static' } };",
  ]) {
    const accepted = lint(dir, path, code);
    assert.equal(accepted.status, 0, accepted.output);
  }
  for (const code of [
    "export const options = { ...getExerciseListQueryOptions(user), ['staleTime']: 'static' };",
    "export const options = { queryKey: ['training-assessment', user], queryFn: load, staleTime: 'static' };",
    "export const options = { queryKey: ['muscle-contributions', user], queryFn: load, staleTime: 'static' };",
    "export const options = { queryKey: getDemoExercisesQueryKey(), queryFn: load, 'staleTime': 'static' };",
  ]) {
    const result = lint(dir, path, code);
    assert.equal(result.status, 1, result.output);
    assert.match(result.output, /Mutable training data/);
  }
  // Scope is deliberate: immutable callers elsewhere and test fixtures stay valid.
  for (const outside of [
    "src/lib/immutable.ts",
    "src/features/workouts/api/workouts.test.ts",
  ]) {
    const result = lint(dir, outside, invalid);
    assert.equal(result.status, 0, result.output);
  }
  for (const boundary of [
    "src/features/workouts/api/workouts.ts",
    "src/features/workouts/api/workout-query-options.ts",
    "src/features/exercises/api/exercises.ts",
    "src/features/exercises/api/exercise-query-options.ts",
    "src/features/analytics/api/training-assessment.ts",
    "src/features/analytics/api/muscle-contributions.ts",
    "src/features/workouts/hooks/use-new-workout-form-workflow.ts",
    "src/features/workouts/components/form/exercise-recommendation-panel.tsx",
    "src/lib/demo-data/query-options.ts",
  ]) {
    const result = lint(dir, boundary, invalid);
    assert.equal(result.status, 1, result.output);
    assert.match(result.output, /Mutable training data/);
  }
});

test("Vite-only typecheck rejects an inert Vitest test block in the actual config", (t) => {
  const dir = fixture(t);
  put(dir, "package.json", '{"type":"module"}');
  put(dir, "tsconfig.vite.json", read("tsconfig.vite.json"));
  put(
    dir,
    "scripts/local-server-config.ts",
    read("scripts/local-server-config.ts"),
  );
  const original = read("vite.config.js");
  put(dir, "vite.config.js", original);
  const valid = cli(dir, "typescript/bin/tsc", ["-p", "tsconfig.vite.json"]);
  assert.equal(valid.status, 0, valid.output);
  put(
    dir,
    "vite.config.js",
    original.replace(
      "defineConfig({",
      "defineConfig({ test: { setupFiles: ['./never-loaded.ts'] },",
    ),
  );
  const invalid = cli(dir, "typescript/bin/tsc", ["-p", "tsconfig.vite.json"]);
  assert.equal(invalid.status, 2, invalid.output);
  assert.match(invalid.output, /'test' does not exist in type/);
  // Before this guard the JS config was not checked at all.
  const unchecked = cli(dir, "typescript/bin/tsc", [
    "-p",
    "tsconfig.vite.json",
    "--checkJs",
    "false",
  ]);
  assert.equal(unchecked.status, 0, unchecked.output);
});

test("Knip follows active setup and exposes orphan tooling, setup files and dependencies", (t) => {
  const dir = fixture(t);
  const pkg = {
    type: "module",
    scripts: { test: "vitest run" },
    devDependencies: { vitest: "5.0.1" },
  };
  put(dir, "package.json", JSON.stringify(pkg));
  put(dir, "knip.json", JSON.stringify(knipConfig));
  put(
    dir,
    "vitest.config.ts",
    "export default { test: { setupFiles: ['./tests/setup.ts'] } };",
  );
  put(dir, "tests/setup.ts", "console.log('active setup');");
  put(
    dir,
    "tests/example.test.ts",
    "import { test } from 'vitest'; test('example', () => {});",
  );
  const run = (args = []) =>
    cli(dir, "knip/bin/knip.js", ["--config", "knip.json", ...args]);
  const valid = run();
  assert.equal(valid.status, 0, valid.output);
  const orphans = [
    "src/test/setup.ts",
    "scripts/retired-config.ts",
    "tools/retired-tool.ts",
  ];
  for (const path of orphans) put(dir, path, "console.log('orphan');");
  const invalid = run();
  assert.equal(invalid.status, 1, invalid.output);
  for (const path of orphans)
    assert.ok(invalid.output.includes(path), invalid.output);
  for (const path of orphans) rmSync(join(dir, path));
  put(dir, "src/dead.ts", "console.log('dead production file');");
  const production = run(["--production"]);
  assert.equal(production.status, 1, production.output);
  assert.match(production.output, /src\/dead.ts/);
  rmSync(join(dir, "src/dead.ts"));
  put(
    dir,
    "package.json",
    JSON.stringify({ ...pkg, dependencies: { valibot: "1.5.0" } }),
  );
  for (const args of [[], ["--production"]]) {
    const unused = run(args);
    assert.equal(unused.status, 1, unused.output);
    assert.match(unused.output, /Unused dependencies/);
    assert.match(unused.output, /valibot/);
  }
  put(dir, "package.json", JSON.stringify(pkg));
  const restored = run(["--production"]);
  assert.equal(restored.status, 0, restored.output);
  put(
    dir,
    "package.json",
    JSON.stringify({
      ...pkg,
      devDependencies: { ...pkg.devDependencies, valibot: "1.5.0" },
    }),
  );
  const unusedDev = run();
  assert.equal(unusedDev.status, 1, unusedDev.output);
  assert.match(unusedDev.output, /Unused devDependencies/);
  assert.match(unusedDev.output, /valibot/);
  put(
    dir,
    "package.json",
    JSON.stringify({
      ...pkg,
      scripts: { ...pkg.scripts, retired: "fittrack-missing-tool" },
    }),
  );
  put(
    dir,
    "vitest.config.ts",
    "export default { test: { setupFiles: ['./tests/setup.ts', './tests/missing-setup.ts'] } };",
  );
  const brokenConfig = run();
  assert.equal(brokenConfig.status, 1, brokenConfig.output);
  assert.match(brokenConfig.output, /Unresolved imports/);
  assert.match(brokenConfig.output, /missing-setup/);
  assert.match(brokenConfig.output, /Unlisted binaries/);
  assert.match(brokenConfig.output, /fittrack-missing-tool/);
  // The previous include list silently omitted missing config imports and binaries.
  put(
    dir,
    "knip-old.json",
    JSON.stringify({
      ...knipConfig,
      include: knipConfig.include.filter(
        (issue) => issue !== "unresolved" && issue !== "binaries",
      ),
    }),
  );
  const old = cli(dir, "knip/bin/knip.js", ["--config", "knip-old.json"]);
  assert.equal(old.status, 0, old.output);
});
