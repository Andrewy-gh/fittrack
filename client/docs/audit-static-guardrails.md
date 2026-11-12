# Audit static-analysis guardrails

These checks prevent specific regressions; they do not close additional audit findings or establish new product policy.

| Command                                    | Contract                                                                                                                                                                                                                                                                   |
| ------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bun run lint`                             | `training-cache/no-static-training-data` rejects literal `staleTime: 'static'` on recognized mutable training query objects in the ten production boundaries listed in `oxlint.config.ts`. Use omitted freshness, a number or `Infinity`, which still allows invalidation. |
| `bun run tsc`                              | Also checks `vite.config.js` using `tsconfig.vite.json`: JavaScript checking is enabled and Vitest's Vite type augmentation is absent. Put test options in `vitest.config.ts`, not an ignored Vite `test` block (A055).                                                    |
| `bun run knip` / `bun run knip:production` | Report unresolved imports and unlisted executable names in addition to existing unused files/dependencies/exports. Repair the importing config or declare the actual tool dependency; do not hide an entire directory.                                                     |
| `bun run test:guardrails`                  | Exercises the real CLIs/configs with disposable fixtures and negative controls. CI runs this separately from Vitest.                                                                                                                                                       |

## Discovery and exclusions (A055/A099/A100)

Knip 6.38 already checks test support and tooling in **ordinary** mode. A pattern such as `!tests/**!` is a **production-only exclusion**, not an ordinary-mode blanket ignore. Likewise `dependencies` already includes unused development dependencies in ordinary mode. Neither needs a speculative replacement.

The installed Knip plugins model the actual tool graph: Vite reads `vite.config.js` and HTML entrypoints; Vitest reads `vitest.config.ts`, its `tests/setup.ts`, and test imports; Oxlint reads `oxlint.config.ts` and its `jsPlugins`; Playwright reads its configs/reporters. Package scripts identify executable entrypoints such as media capture and Husky installation. CLI fixtures verify active setup remains reachable, disconnected `src/test`, `scripts`, and `tools` files are reported, and ordinary/production dependency checks still fail on unused dependencies. Production deliberately excludes development support, not production app code.

New checks expose one existing external executable: `ffmpeg` in the media-capture script. It is an intentionally machine-installed media prerequisite, not a JavaScript package; `ignoreBinaries: ["ffmpeg"]` is the sole new suppression. Existing generated/shared UI export exceptions and parked feature/file exceptions are unchanged. Missing `ffmpeg` installation is still a media-runtime concern.

## Cache rule scope (A005/A006)

The AST rule recognizes direct spreads of the established mutable workout/exercise/demo/analytics factory names, matching query-key factory calls, and the `training-assessment` / `muscle-contributions` key prefixes. It checks only a sibling literal string-valued `staleTime` property (including quoted/computed literal property names). It does not scan source text. Immutable catalog options, unrelated objects/strings/comments, nested metadata, and test files remain valid.

This is intentionally syntactic, not type/data-flow analysis: aliases, namespace/member calls, values hidden in variables/spreads, template expressions, renamed/moved boundaries, and new query families need human review or a deliberate rule update. It does not prove invalidation coverage, successful refetches, account-switch safety, or authorization. Existing behavioral cache tests remain responsible for those observable contracts. No automatic fix changes freshness policy.

## Deferred candidates

- A003 module-evaluation date capture: detecting eager evaluation through helper calls/initializers without also banning legitimate dates requires more analysis. Existing rollover/draft behavior tests remain the guard; no blanket `Date` ban.
- A061 environment centralization, A082 request-time Stripe global mutation, A092 Go test helper placement, and A040/A042 transport boundaries need separate language/runtime/design work, not generic string checks.
- Route/spec completeness, migration immutability, PWA builds, CORS/logging/ownership/nullability require behavioral or dedicated checks.
- A018 numeric/resource limits and A088 service/query removal remain unapproved and untouched.
