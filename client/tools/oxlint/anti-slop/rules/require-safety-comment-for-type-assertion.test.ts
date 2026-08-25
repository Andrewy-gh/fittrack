import { RuleTester } from "oxlint/plugins-dev";
import { describe, it } from "vitest";

import { requireSafetyCommentForTypeAssertionRule } from "./require-safety-comment-for-type-assertion.ts";

RuleTester.describe = describe;
RuleTester.it = it;
RuleTester.itOnly = it.only;

const tester = new RuleTester({
  languageOptions: { parserOptions: { lang: "ts" } },
});

const error = { messageId: "missingSafetyComment" };

tester.run(
  "anti-slop/require-safety-comment-for-type-assertion",
  requireSafetyCommentForTypeAssertionRule,
  {
    valid: [
      "const exact = { id: 1 } as const;",
      `
        type User = { readonly id: string };
        // SAFETY: parseUser checked this external payload before the branded boundary.
        const user = raw as User;
      `,
      `
        type User = { readonly id: string };
        function readUser(raw: unknown): User {
          // SAFETY: this function accepts only the parsed storage record.
          return raw as User;
        }
      `,
      `
        type User = { readonly id: string };
        // SAFETY: the adapter owns the conversion from its validated protocol record.
        const id = (raw as User).id;
      `,
      `
        type User = { readonly id: string };
        function hasUserId(raw: unknown): boolean {
          // SAFETY: the caller parsed the value before this condition.
          if ((raw as User).id) return true;
          return false;
        }
      `,
      `
        type User = { readonly id: string };
        function hasUserId(raw: unknown): boolean {
          // SAFETY: the caller parsed the value before this loop condition.
          while ((raw as User).id) return true;
          return false;
        }
      `,
      `
        type User = { readonly id: string };
        const user = raw /* SAFETY: parseUser checked the boundary value. */ as User;
      `,
      `
        type User = { readonly id: string };
        // SAFETY: parseUser checked the boundary value before this angle assertion.
        const user = <User>raw;
      `,
    ],
    invalid: [
      {
        code: "type User = { readonly id: string }; const user = raw as User;",
        errors: [error],
      },
      {
        code: "type User = { readonly id: string }; const user = <User>raw;",
        errors: [error],
      },
      {
        code: `
          type User = { readonly id: string };
          const user = raw as User;
          // SAFETY: comments after an assertion do not establish its invariant.
        `,
        errors: [error],
      },
      {
        code: `
          type User = { readonly id: string };
          // This is expected to be a user.
          const user = raw as User;
        `,
        errors: [error],
      },
      {
        code: `
          type User = { readonly id: string };
          // SAFETY: a function-level comment must not cover its control-flow body.
          function hasUserId(raw: unknown): boolean {
            if ((raw as User).id) return true;
            return false;
          }
        `,
        errors: [error],
      },
      {
        code: `
          type User = { readonly id: string };
          const user = raw as /* SAFETY: too late. */ User;
        `,
        errors: [error],
      },
      {
        code: `
          type User = { readonly id: string };
          const user = <User> /* SAFETY: too late. */ raw;
        `,
        errors: [error],
      },
    ],
  },
);
