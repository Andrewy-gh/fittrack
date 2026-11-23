#!/usr/bin/env node

import { readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const source = resolve(root, "ops/prometheus/ai-chat-alerts.yml");
const target = resolve(root, "ops/grafana-cloud/ai-chat-alerts.mimir.yml");
const args = process.argv.slice(2);
if (args.length > 1 || (args.length === 1 && args[0] !== "--check")) {
  console.error("Usage: node scripts/generate-ai-chat-mimir.mjs [--check]");
  process.exit(2);
}

const rules = readFileSync(source);
// Match the source's line endings while leaving every source byte untouched.
const newline = rules.includes(Buffer.from("\r\n")) ? "\r\n" : "\n";
const expected = Buffer.concat([Buffer.from(`namespace: ai_chat_rollout${newline}`), rules]);
if (args[0] === "--check") {
  let actual;
  try {
    actual = readFileSync(target);
  } catch (error) {
    if (error.code !== "ENOENT") throw error;
  }
  if (!actual?.equals(expected)) {
    console.error("Mimir alerts are missing or out of date; run node scripts/generate-ai-chat-mimir.mjs");
    process.exit(1);
  }
  console.log("Mimir alerts match the Prometheus source.");
} else {
  writeFileSync(target, expected);
  console.log("Generated ops/grafana-cloud/ai-chat-alerts.mimir.yml");
}
