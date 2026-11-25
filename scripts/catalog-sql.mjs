#!/usr/bin/env node
// Generate the SQLC membership snapshot or a reviewable forward-migration snippet.
import { readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const catalogPath = resolve(root, "server/internal/exercisecatalog/catalog.json");
const schemaPath = resolve(root, "server/schema.sql");
const begin = "    -- BEGIN GENERATED CATALOG MEMBERSHIP";
const end = "    -- END GENERATED CATALOG MEMBERSHIP";
const [mode, ...rest] = process.argv.slice(2);
if (rest.length || !["--check", "--write", "--snippet"].includes(mode)) {
  console.error("Usage: node scripts/catalog-sql.mjs --check|--write|--snippet");
  process.exit(2);
}

try {
  const entries = JSON.parse(readFileSync(catalogPath, "utf8"));
  if (!Array.isArray(entries) || entries.length === 0) throw new Error("catalog must be a nonempty array");
  const ids = entries.map((entry) => entry?.id);
  if (ids.some((id) => typeof id !== "string" || !id || /[\x00-\x1f\x7f]/.test(id)) || new Set(ids).size !== ids.length) {
    throw new Error("catalog IDs must be unique nonempty strings without control characters");
  }
  ids.sort();
  const literal = (id) => id.includes("\\")
    ? `E'${id.replaceAll("\\", "\\\\").replaceAll("'", "''")}'`
    : `'${id.replaceAll("'", "''")}'`;
  const check = `CHECK (catalog_id IN (\n${ids.map((id) => `    ${literal(id)}`).join(",\n")}\n))`;
  if (mode === "--snippet") {
    console.log(`-- Frozen catalog membership snapshot; review locking, existing data and rollback before use.\nALTER TABLE exercise DROP CONSTRAINT exercise_catalog_id_check;\nALTER TABLE exercise ADD CONSTRAINT exercise_catalog_id_check ${check};`);
  } else {
    const original = readFileSync(schemaPath, "utf8");
    const newline = original.includes("\r\n") ? "\r\n" : "\n";
    const normalized = original.replaceAll("\r\n", "\n");
    const first = normalized.indexOf(begin);
    const last = normalized.indexOf(end);
    if (first < 0 || last < first || normalized.indexOf(begin, first + 1) !== -1 || normalized.indexOf(end, last + 1) !== -1) {
      throw new Error("schema must contain exactly one ordered pair of catalog generation markers");
    }
    const generated = `${begin}\n    catalog_id TEXT ${check},\n${end}`;
    const updated = (normalized.slice(0, first) + generated + normalized.slice(last + end.length)).replaceAll("\n", newline);
    if (mode === "--check") {
      if (original !== updated) throw new Error("schema catalog membership is stale; run node scripts/catalog-sql.mjs --write");
      console.log("Schema catalog membership matches embedded JSON.");
    } else {
      writeFileSync(schemaPath, updated);
      console.log("Updated schema catalog membership from JSON.");
    }
  }
} catch (error) {
  console.error(`catalog-sql: ${error.message}`);
  process.exit(1);
}
