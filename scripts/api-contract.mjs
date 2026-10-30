#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  renameSync,
  rmSync,
  statSync,
} from "node:fs";
import { dirname, join, relative, resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const checkOnly = process.argv.includes("--check");
const temporaryRoot = mkdtempSync(
  join(root, "client", "src", "client-api-contract-tmp-"),
);
const swaggerOut = join(temporaryRoot, "swagger");
const clientOut = join(temporaryRoot, "client");
const swaggerJson = join(swaggerOut, "swagger.json");
const clientConfig = join(root, "client", "openapi-ts.config.ts");

function run(command, args, cwd, env = {}) {
  const result = spawnSync(command, args, {
    cwd,
    env: { ...process.env, ...env },
    stdio: "inherit",
    shell: process.platform === "win32" && command.endsWith(".cmd"),
  });
  if (result.error) throw result.error;
  if (result.status !== 0)
    throw new Error(
      `${command} exited with status ${result.status ?? "unknown"}`,
    );
}

function filesUnder(path) {
  if (!existsSync(path)) return [];
  const info = statSync(path);
  if (info.isFile()) return [""];
  return readdirSync(path, { withFileTypes: true }).flatMap((entry) => {
    const child = join(path, entry.name);
    return entry.isDirectory()
      ? filesUnder(child).map((file) => join(entry.name, file))
      : [entry.name];
  });
}

try {
  run(
    "go",
    [
      "run",
      "github.com/swaggo/swag/cmd/swag@v1.16.6",
      "init",
      "-g",
      "cmd/api/main.go",
      "-o",
      swaggerOut,
      "--outputTypes",
      "json,yaml",
    ],
    join(root, "server"),
  );
  run(
    "bun",
    ["run", "openapi-ts", "--file", clientConfig, "--no-log-file"],
    join(root, "client"),
    {
      FITTRACK_OPENAPI_INPUT: swaggerJson,
      FITTRACK_OPENAPI_OUTPUT: clientOut,
    },
  );
  run("bun", ["run", "oxfmt", "--", clientOut], join(root, "client"));
  run("bun", ["run", "oxlint", "--fix", clientOut], join(root, "client"));

  for (const file of ["swagger.json", "swagger.yaml"]) {
    if (!existsSync(join(swaggerOut, file)))
      throw new Error(`Swagger generation did not produce ${file}`);
  }
  if (filesUnder(clientOut).length === 0)
    throw new Error("OpenAPI client generation produced no files");

  const generated = [
    [
      swaggerOut,
      join(root, "server", "docs"),
      ["swagger.json", "swagger.yaml"],
    ],
    [clientOut, join(root, "client", "src", "client"), null],
  ];
  const differences = [];
  const swaggerDirectory = join(root, "server", "docs");
  for (const name of readdirSync(swaggerDirectory)) {
    if (
      name.startsWith("swagger.") &&
      !["swagger.json", "swagger.yaml"].includes(name)
    ) {
      differences.push(relative(root, join(swaggerDirectory, name)));
    }
  }
  for (const [fresh, current, only] of generated) {
    const names =
      only ??
      [...new Set([...filesUnder(fresh), ...filesUnder(current)])].sort();
    for (const name of names) {
      const expected = join(fresh, name);
      const actual = join(current, name);
      if (
        !existsSync(expected) ||
        !existsSync(actual) ||
        !readFileSync(expected).equals(readFileSync(actual))
      ) {
        differences.push(relative(root, actual));
      }
    }
  }

  if (checkOnly) {
    if (differences.length) {
      console.error("Generated API files are stale or missing:");
      for (const file of differences) console.error(`  ${file}`);
      console.error(
        "Run `bun run --cwd client api:generate` from the repository root and commit the generated files.",
      );
      process.exitCode = 1;
    } else {
      console.log("Generated API files are fresh.");
    }
  } else {
    const publications = generated.flatMap(([fresh, current, only]) =>
      only
        ? only.map((name) => ({
            source: join(fresh, name),
            destination: join(current, name),
          }))
        : [{ source: fresh, destination: current }],
    );
    const backedUp = [];
    const published = [];
    try {
      for (const item of publications) {
        mkdirSync(dirname(item.destination), { recursive: true });
        item.backup = `${item.destination}.api-contract-backup-${process.pid}`;
        item.hadOriginal = existsSync(item.destination);
        if (item.hadOriginal) renameSync(item.destination, item.backup);
        backedUp.push(item);
      }
      for (const item of publications) {
        renameSync(item.source, item.destination);
        published.push(item);
      }
    } catch (error) {
      for (const item of published.reverse())
        rmSync(item.destination, { recursive: true, force: true });
      for (const item of backedUp.reverse()) {
        if (item.hadOriginal && existsSync(item.backup))
          renameSync(item.backup, item.destination);
      }
      throw error;
    }
    for (const item of backedUp) {
      if (item.hadOriginal)
        rmSync(item.backup, { recursive: true, force: true });
    }
    console.log(
      differences.length
        ? `Updated ${differences.length} generated API file(s).`
        : "Generated API files are already current.",
    );
  }
} catch (error) {
  console.error(`API generation failed: ${error.message}`);
  process.exitCode = 1;
} finally {
  rmSync(temporaryRoot, { recursive: true, force: true });
}
