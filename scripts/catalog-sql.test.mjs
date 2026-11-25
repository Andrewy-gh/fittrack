import { test } from "node:test";
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, statSync, utimesSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));

test("catalog expansion regenerates schema and emits a frozen forward snippet", (t) => {
  const dir = mkdtempSync(join(tmpdir(), "catalog-sql-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  for (const path of ["scripts", "server/internal/exercisecatalog"]) mkdirSync(join(dir, path), { recursive: true });
  for (const path of ["scripts/catalog-sql.mjs", "server/internal/exercisecatalog/catalog.json", "server/schema.sql"]) copyFileSync(join(root, path), join(dir, path));
  const run = (arg) => spawnSync(process.execPath, [join(dir, "scripts/catalog-sql.mjs"), arg], { encoding: "utf8" });
  const source = join(dir, "server/internal/exercisecatalog/catalog.json");
  const schema = join(dir, "server/schema.sql");
  assert.equal(run("--check").status, 0);
  const entries = JSON.parse(readFileSync(source, "utf8"));
  entries.push({ id: "Synthetic_'Expansion", name: "synthetic" });
  writeFileSync(source, JSON.stringify(entries));
  const oldSchema = readFileSync(schema, "utf8");
  assert.equal(run("--check").status, 1);
  assert.equal(readFileSync(schema, "utf8"), oldSchema, "check must not rewrite");
  const snippet = run("--snippet");
  assert.equal(snippet.status, 0, snippet.stderr);
  assert.match(snippet.stdout, /'Synthetic_''Expansion'/);
  assert.equal(run("--write").status, 0);
  assert.equal(run("--check").status, 0);
  assert.match(readFileSync(schema, "utf8"), /'Synthetic_''Expansion'/);
  const generated = readFileSync(schema, "utf8");
  assert.equal(run("--write").status, 0);
  assert.equal(readFileSync(schema, "utf8"), generated, "generation is deterministic");
  entries.pop();
  writeFileSync(source, JSON.stringify(entries));
  assert.match(snippet.stdout, /'Synthetic_''Expansion'/, "saved snippet is a frozen snapshot");
  assert.equal(run("--check").status, 1);
  entries.push(entries[0]);
  writeFileSync(source, JSON.stringify(entries));
  assert.notEqual(run("--write").status, 0, "duplicate IDs are refused");
  assert.equal(readFileSync(schema, "utf8"), generated);
});

test("write preserves bytes outside its block, is a no-op when current, and rejects invalid Unicode", (t) => {
  const dir = mkdtempSync(join(tmpdir(), "catalog-sql-bytes-"));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  for (const path of ["scripts", "server/internal/exercisecatalog"]) mkdirSync(join(dir, path), { recursive: true });
  for (const path of ["scripts/catalog-sql.mjs", "server/internal/exercisecatalog/catalog.json", "server/schema.sql"]) copyFileSync(join(root, path), join(dir, path));
  const run = (arg) => spawnSync(process.execPath, [join(dir, "scripts/catalog-sql.mjs"), arg], { encoding: "utf8" });
  const source = join(dir, "server/internal/exercisecatalog/catalog.json");
  const schema = join(dir, "server/schema.sql");
  const original = readFileSync(schema);
  const prefix = Buffer.from([0xff, 0x0d, 0x0a, 0x0a]);
  const suffix = Buffer.from([0x0a, 0x0d, 0x0a, 0xfe]);
  const mixed = Buffer.concat([prefix, original, suffix]);
  writeFileSync(schema, mixed);
  const begin = Buffer.from("    -- BEGIN GENERATED CATALOG MEMBERSHIP");
  const end = Buffer.from("    -- END GENERATED CATALOG MEMBERSHIP");
  const before = mixed.subarray(0, mixed.indexOf(begin));
  const after = mixed.subarray(mixed.indexOf(end) + end.length);
  const entries = JSON.parse(readFileSync(source, "utf8"));
  entries.push({ id: "Synthetic_😀" });
  writeFileSync(source, JSON.stringify(entries));
  assert.equal(run("--write").status, 0);
  const result = readFileSync(schema);
  assert.deepEqual(result.subarray(0, result.indexOf(begin)), before);
  assert.deepEqual(result.subarray(result.indexOf(end) + end.length), after);
  assert.equal(run("--check").status, 0);
  utimesSync(schema, new Date(1000000000000), new Date(1000000000000));
  const mtime = statSync(schema).mtimeMs;
  assert.equal(run("--write").status, 0);
  assert.equal(statSync(schema).mtimeMs, mtime, "current generation must not rewrite");
  assert.deepEqual(readFileSync(schema), result);
  for (const id of ["\ud800", "\udc00"]) {
    writeFileSync(source, JSON.stringify([{ id }]));
    for (const mode of ["--write", "--check", "--snippet"]) {
      const failure = run(mode);
      assert.notEqual(failure.status, 0);
      assert.equal(failure.stdout, "");
    }
    assert.deepEqual(readFileSync(schema), result);
    assert.equal(statSync(schema).mtimeMs, mtime);
  }
  // Raw invalid UTF-8 must not be confused with a legitimate U+FFFD ID.
  writeFileSync(source, Buffer.concat([Buffer.from('[{"id":"'), Buffer.from([0xff]), Buffer.from('"}]')]));
  for (const mode of ["--write", "--check", "--snippet"]) {
    const failure = run(mode);
    assert.notEqual(failure.status, 0);
    assert.equal(failure.stdout, "");
    assert.deepEqual(readFileSync(schema), result);
    assert.equal(statSync(schema).mtimeMs, mtime);
  }
  writeFileSync(source, '[{"id":"\ufffd"}]');
  const valid = run("--snippet");
  assert.equal(valid.status, 0, valid.stderr);
  assert.match(valid.stdout, /\ufffd/);
});
