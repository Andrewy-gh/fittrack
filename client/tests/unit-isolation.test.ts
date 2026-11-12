import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { spawn } from "node:child_process";
import path from "node:path";
import { describe, expect, it } from "vitest";

const client = process.cwd();
const synthetic = {
  VITE_PROJECT_ID: "00000000-0000-4000-8000-000000000001",
  VITE_PUBLISHABLE_CLIENT_KEY: "synthetic-unit-key",
};
const modulePath = (name: string) =>
  JSON.stringify(path.join(client, name).replaceAll("\\", "/"));

async function fixture(source: string, inherited: boolean) {
  const root = await mkdtemp(path.join(client, "tests", ".unit-isolation-"));
  try {
    // An independent offline fuse is installed BEFORE the boundary under test.
    // Even a regression/removal of that boundary cannot send a provider request.
    await writeFile(
      path.join(root, "offline.ts"),
      `
      import { Socket } from 'node:net';
      const connect = Socket.prototype.connect;
      Socket.prototype.connect = function (...args) {
        const first = Array.isArray(args[0]) ? args[0][0] : args[0];
        const host = typeof first === 'object' ? first.host : args[1];
        const port = typeof first === 'object' ? first.port : first;
        if (Number(port) === 1 || !['127.0.0.1', '::1', 'localhost'].includes(host)) throw new Error('OFFLINE_FUSE_SOCKET');
        return Reflect.apply(connect, this, args);
      };
      export const nativeFetch = globalThis.fetch;
      const realFetch = nativeFetch;
      globalThis.fetch = (input, init) => {
        const url = new URL(input instanceof Request ? input.url : String(input));
        if (!['127.0.0.1', '[::1]', 'localhost'].includes(url.hostname)) throw new Error('OFFLINE_FUSE_FETCH');
        return realFetch(input, init);
      };
    `,
    );
    await writeFile(
      path.join(root, ".env.test.local"),
      Object.entries(synthetic)
        .map(([key, value]) => `${key}=${value}`)
        .join("\n") + "\nVITE_UNIT_ENV_FILE_SENTINEL=synthetic-file-marker",
    );
    await writeFile(path.join(root, "case.fixture.ts"), source);
    await writeFile(
      path.join(root, "vitest.config.ts"),
      `
      import base from ${modulePath("vitest.config.ts")};
      export default { ...base, root: ${JSON.stringify(root)}, test: {
        ...base.test, include: ['case.fixture.ts'], testTimeout: 30000,
        setupFiles: [${JSON.stringify(path.join(root, "offline.ts"))}, ${modulePath("tests/setup.ts")}],
      }};
    `,
    );
    // Do not copy inherited provider settings into a fixture subprocess.
    const env = { ...process.env };
    for (const key of Object.keys(synthetic)) delete env[key];
    if (inherited) Object.assign(env, synthetic);
    const result = await new Promise<{ code: number | null; output: string }>(
      (resolve, reject) => {
        const child = spawn(
          process.execPath,
          [
            path.join(client, "node_modules/vitest/vitest.mjs"),
            "run",
            "--config",
            path.join(root, "vitest.config.ts"),
            "--maxWorkers=1",
          ],
          { cwd: client, env, windowsHide: true, timeout: 90000 },
        );
        let output = "";
        child.stdout.on("data", (chunk) => {
          output += chunk;
        });
        child.stderr.on("data", (chunk) => {
          output += chunk;
        });
        child.on("error", reject);
        child.on("exit", (code) => resolve({ code, output }));
      },
    );
    return result;
  } finally {
    await rm(root, { recursive: true, force: true });
  }
}

describe("ordinary unit isolation entrypoint", () => {
  it.each([true, false])(
    "ignores ambient provider settings (inherited=%s), keeps DOM/storage and explicit mocks",
    async (inherited) => {
      const result = await fixture(
        `
      import { expect, it, vi } from 'vitest';
      import { stackClientApp } from ${modulePath("src/stack.ts")};
      it('ambient provider is absent', async () => {
        expect(stackClientApp === null).toBe(true);
        expect(import.meta.env.VITE_PROJECT_ID).toBe('');
        expect(import.meta.env.VITE_PUBLISHABLE_CLIENT_KEY).toBe('');
        expect(import.meta.env.VITE_UNIT_ENV_FILE_SENTINEL).toBeUndefined();
        localStorage.setItem('real-storage', 'ok');
        expect(localStorage.getItem('real-storage')).toBe('ok');
        expect(document.createElement('button')).toBeInstanceOf(HTMLButtonElement);
        vi.stubGlobal('fetch', async () => new Response('explicit mock'));
        expect(await (await fetch('https://provider.invalid')).text()).toBe('explicit mock');
        vi.unstubAllGlobals();
        vi.stubEnv('VITE_PROJECT_ID', 'synthetic-config-test');
        expect(import.meta.env.VITE_PROJECT_ID).toBe('synthetic-config-test');
        vi.unstubAllEnvs();
      });
    `,
        inherited,
      );
      expect(result.output).not.toContain("OFFLINE_FUSE");
      expect(result.code, result.output).toBe(0);
    },
    120000,
  );

  it("fails the run for caught pre-import fetch, HTTP, HTTPS and Node fetch attempts", async () => {
    const result = await fixture(
      `
      import { expect, it, vi } from 'vitest';
      import http from 'node:http';
      import https from 'node:https';
      import { nativeFetch } from './offline';
      // This executes at module import, before tests/hooks. Simulate an SDK
      // swallowing its rejection; the independent violation ledger must fail.
      await fetch('https://provider.invalid/identity').catch(() => {});
      it('all transports reject before reaching the offline fuse', async () => {
        for (const transport of [http, https]) {
          expect(() => transport.get(transport === http ? 'http://provider.invalid' : 'https://provider.invalid')).toThrow('Unit network isolation');
        }
        expect(() => http.get('http://127.0.0.1:1')).toThrow('Unit network isolation');
        // The offline setup captured Node's fetch before the global wrapper.
        const error = await nativeFetch('https://provider.invalid/analytics').catch(error => error);
        expect(error.cause?.message).toContain('Unit network isolation');
        vi.stubEnv('VITE_PROJECT_ID', '00000000-0000-4000-8000-000000000001');
        vi.stubEnv('VITE_PUBLISHABLE_CLIENT_KEY', 'synthetic-unit-key');
        vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
        try {
          const { stackClientApp } = await import(${modulePath("src/stack.ts")});
          expect(stackClientApp !== null).toBe(true);
          await expect(stackClientApp.getUser({ or: 'anonymous' })).rejects.toThrow('Unit network isolation');
        } finally {
          vi.clearAllTimers();
          vi.useRealTimers();
          vi.unstubAllEnvs();
        }
        console.log('TRANSPORT_ASSERTIONS_PASSED');
      });
    `,
      false,
    );
    expect(result.code, result.output).not.toBe(0);
    expect(result.output).toContain("TRANSPORT_ASSERTIONS_PASSED");
    expect(result.output).toContain("Unit network isolation was violated");
    expect(result.output).toContain("Unhandled Rejection");
    expect(result.output).not.toContain("OFFLINE_FUSE");
  }, 120000);

  it("permits only worker-owned loopback HTTP fixtures", async () => {
    const result = await fixture(
      `
      import { expect, it } from 'vitest';
      import { createServer, get } from 'node:http';
      import { once } from 'node:events';
      it('reads a real local response with fetch and HTTP', async () => {
        const server = createServer((_, res) => res.end('owned fixture'));
        server.listen(0, '127.0.0.1');
        await once(server, 'listening');
        const url = 'http://127.0.0.1:' + server.address().port;
        try {
          expect(await (await fetch(url)).text()).toBe('owned fixture');
          const body = await new Promise((resolve, reject) => {
            get(url, res => { let text = ''; res.on('data', c => text += c); res.on('end', () => resolve(text)); }).on('error', reject);
          });
          expect(body).toBe('owned fixture');
        } finally { await new Promise(resolve => server.close(resolve)); }
      });
    `,
      false,
    );
    expect(result.code, result.output).toBe(0);
  }, 120000);
});
