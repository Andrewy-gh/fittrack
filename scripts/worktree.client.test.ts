import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";
import { commandEnvironment, localConfig } from "./worktree.ts";

test(
  "managed unit environment preserves the client billing request contract",
  { timeout: 60000 },
  () => {
    const root = fileURLToPath(new URL("../", import.meta.url));
    const config = localConfig({ root, id: "fittrack_0123456789ab", basePort: 21000 });
    // Exercise the existing caller-facing assertions with the adapter's real
    // environment, so direct CI's fallback URL cannot hide a managed-mode defect.
    const result = spawnSync(
      process.platform === "win32" ? "bun.exe" : "bun",
      [
        "--no-env-file",
        "run",
        "test",
        "--",
        "src/features/chat/api/billing.test.ts",
        "--maxWorkers=1",
      ],
      {
        cwd: fileURLToPath(new URL("../client/", import.meta.url)),
        env: commandEnvironment(config, "unit"),
        encoding: "utf8",
        timeout: 55000,
        windowsHide: true,
      },
    );
    assert.ifError(result.error);
    assert.equal(result.status, 0, `${result.stdout}\n${result.stderr}`);
  },
);
