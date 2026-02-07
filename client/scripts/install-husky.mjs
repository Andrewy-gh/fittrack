import path from 'node:path';
import { fileURLToPath } from 'node:url';

function repoRootFromClientScripts() {
  const __filename = fileURLToPath(import.meta.url);
  const __dirname = path.dirname(__filename);
  // client/scripts -> client -> repo root
  return path.resolve(__dirname, '..', '..');
}

async function main() {
  // Husky is just git hooks; skip for CI or when explicitly disabled.
  if (process.env.HUSKY === '0' || process.env.CI) return;

  const root = repoRootFromClientScripts();
  process.chdir(root);

  const { default: install } = await import('husky');
  const result = install();

  // Don't fail dependency install if .git isn't present (e.g. CI tarballs).
  if (result) console.log(result);
}

main().catch((err) => {
  console.warn('husky install failed:', err);
});

