#!/usr/bin/env node

import { spawnSync } from "node:child_process";

function gitPaths(args) {
  const result = spawnSync("git", args, { encoding: null });
  if (result.error) throw result.error;
  if (result.status !== 0)
    throw new Error(
      `git ${args.join(" ")} exited with status ${result.status}`,
    );
  return result.stdout.toString("utf8").split("\0").filter(Boolean);
}

const contractFiles = [
  "server/go.mod",
  "server/go.sum",
  "server/Makefile",
  "server/docs/swagger.json",
  "server/docs/swagger.yaml",
  "client/openapi-ts.config.ts",
  "client/.oxfmtrc.json",
  "client/oxlint.config.ts",
  "client/bun.lock",
  "client/package.json",
  ".husky/pre-commit",
  ".github/workflows/test.yml",
];
const contractDirectories = ["client/src/client/", "client/tools/oxlint/"];

function isApiContractPath(path) {
  if (
    contractFiles.includes(path) ||
    path.startsWith("scripts/api-contract") ||
    contractDirectories.some((directory) => path.startsWith(directory))
  ) {
    return true;
  }

  if (!path.startsWith("server/") || path.endsWith("_test.go")) return false;
  if (path.startsWith("server/migrations/")) return false;
  return /^server\/(?:[^/]+\/)*[^/]+\.go$/.test(path);
}

function run() {
  const staged = gitPaths([
    "diff",
    "--cached",
    "--name-only",
    "--no-renames",
    "--diff-filter=ACDMRTUXB",
    "-z",
  ]);
  const state = {
    apiContractChanged: staged.some(isApiContractPath),
    serverChanged: staged.some((path) => path.startsWith("server/")),
    clientChanged: staged.some((path) => path.startsWith("client/")),
    serverNonDocsChanged: staged.some(
      (path) =>
        path.startsWith("server/") &&
        path !== "server/README.md" &&
        !path.startsWith("server/docs/") &&
        !path.endsWith(".md"),
    ),
  };

  if (state.apiContractChanged) {
    const worktree = [
      ...gitPaths(["diff", "--name-only", "--no-renames", "-z"]),
      // Include ignored contract files, but never enumerate dependency/build trees.
      ...gitPaths([
        "ls-files",
        "--others",
        "-z",
        "--",
        ...contractFiles,
        ...contractDirectories,
        "scripts/api-contract*",
        ":(glob)server/**/*.go",
      ]),
    ];
    const relevant = worktree.find(isApiContractPath);
    if (relevant) {
      console.error(
        `pre-commit: refusing API generation because relevant worktree changes are not staged: ${JSON.stringify(relevant)}`,
      );
      console.error(
        "Stage or otherwise resolve relevant input/output changes, then retry. No API files were generated.",
      );
      process.exitCode = 1;
      return;
    }
  }

  process.stdout.write(
    [
      state.apiContractChanged,
      state.serverChanged,
      state.clientChanged,
      state.serverNonDocsChanged,
    ]
      .map(Number)
      .join(" "),
  );
}

try {
  run();
} catch (error) {
  console.error(`pre-commit: failed to inspect Git paths: ${error.message}`);
  process.exitCode = 1;
}
