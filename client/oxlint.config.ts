import { defineConfig } from "oxlint";

export default defineConfig({
  ignorePatterns: [
    ".agent/**",
    ".agents/**",
    ".claude/**",
    ".codex/**",
    ".continue/**",
    ".cursor/**",
    ".gemini/**",
    ".opencode/**",
    ".pi/**",
    ".roo/**",
    ".windsurf/**",
    "src/routeTree.gen.ts",
    "tools/oxlint/anti-slop/**",
  ],
  jsPlugins: [
    { name: "anti-slop", specifier: "./tools/oxlint/anti-slop/index.ts" },
  ],
  rules: {
    "anti-slop/no-chained-type-assertions": "error",
    "anti-slop/no-conditional-empty-object-spread": "error",
    "anti-slop/no-known-value-widening": "error",
    "anti-slop/no-object-parameters": "error",
    "anti-slop/no-reflect-apply": "error",
    "anti-slop/no-reflect-get": "error",
    "anti-slop/no-unknown-returns": "error",
    "anti-slop/no-unknown-type-aliases": "error",
    "anti-slop/no-widen-then-assert": "error",
  },
  overrides: [
    {
      files: ["src/**/*.{ts,tsx}"],
      excludeFiles: [
        "src/client/**",
        "**/test/**",
        "**/tests/**",
        "**/__tests__/**",
        "**/*.test.ts",
        "**/*.test.tsx",
        "**/*.spec.ts",
        "**/*.spec.tsx",
        "**/*-test.ts",
        "**/*-test.tsx",
      ],
      rules: {
        "anti-slop/no-unsafe-dictionary-type": "error",
        "anti-slop/require-safety-comment-for-type-assertion": "error",
      },
    },
    {
      files: ["src/features/chat/api/ai-chat.ts"],
      rules: {
        // This boundary deliberately parses an unknown-valued JSON object field by field.
        // Its generic property builders also preserve exact computed keys via assertions.
        "anti-slop/no-known-value-widening": "off",
        "anti-slop/no-unsafe-dictionary-type": "off",
      },
    },
    {
      files: ["src/components/reload-prompt.test.tsx"],
      rules: {
        // The browser callback requires the full platform interface, while this test
        // intentionally supplies only the update seam it exercises.
        "anti-slop/no-chained-type-assertions": "off",
      },
    },
    {
      files: ["src/client/**/*.{ts,tsx}"],
      rules: {
        "anti-slop/no-chained-type-assertions": "off",
        "anti-slop/no-conditional-empty-object-spread": "off",
        "anti-slop/no-known-value-widening": "off",
        "anti-slop/no-object-parameters": "off",
        "anti-slop/no-reflect-apply": "off",
        "anti-slop/no-reflect-get": "off",
        "anti-slop/no-unknown-returns": "off",
        "anti-slop/no-unknown-type-aliases": "off",
        "anti-slop/no-unsafe-dictionary-type": "off",
        "anti-slop/no-widen-then-assert": "off",
      },
    },
  ],
});
