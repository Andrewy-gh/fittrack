import { defineRule } from "@oxlint/plugins";

import type { ESTree, SourceCode } from "@oxlint/plugins";

type TypeAssertion = ESTree.TSAsExpression | ESTree.TSTypeAssertion;

const commentOwnerKinds = new Set([
  "DoWhileStatement",
  "ExpressionStatement",
  "ForInStatement",
  "ForOfStatement",
  "ForStatement",
  "IfStatement",
  "PropertyDefinition",
  "ReturnStatement",
  "SwitchStatement",
  "ThrowStatement",
  "TryStatement",
  "VariableDeclaration",
  "WhileStatement",
  "WithStatement",
]);

const functionBoundaryKinds = new Set([
  "ArrowFunctionExpression",
  "FunctionDeclaration",
  "FunctionExpression",
]);

function isConstAssertion(node: TypeAssertion): boolean {
  return (
    node.typeAnnotation.type === "TSTypeReference" &&
    node.typeAnnotation.typeName.type === "Identifier" &&
    node.typeAnnotation.typeName.name === "const"
  );
}

function hasSafetyComment(
  sourceCode: SourceCode,
  node: TypeAssertion,
): boolean {
  const assertionOperatorStart =
    node.type === "TSAsExpression"
      ? (sourceCode.getTokenAfter(node.expression)?.start ?? node.start)
      : node.start;
  let current: ESTree.Node = node;

  while (true) {
    if (functionBoundaryKinds.has(current.type)) return false;
    if (
      sourceCode
        .getCommentsBefore(current)
        .some(
          (comment) =>
            comment.end <= assertionOperatorStart &&
            /\bSAFETY\s*:/u.test(comment.value),
        ) ||
      (current === node &&
        node.type === "TSAsExpression" &&
        sourceCode
          .getCommentsInside(node)
          .some(
            (comment) =>
              comment.end <= assertionOperatorStart &&
              /\bSAFETY\s*:/u.test(comment.value),
          ))
    ) {
      return true;
    }

    if (
      commentOwnerKinds.has(current.type) ||
      current.parent.type === "Program"
    )
      return false;
    current = current.parent;
  }
}

/** Require every non-const type assertion to state the invariant TypeScript cannot express. */
export const requireSafetyCommentForTypeAssertionRule = defineRule({
  meta: {
    type: "problem",
    docs: {
      description:
        "Require a nearby SAFETY comment for every TypeScript type assertion except const assertions.",
    },
    messages: {
      missingSafetyComment:
        "This type assertion has no `SAFETY:` justification. State the checked invariant immediately before the assertion or its containing statement.",
    },
  },
  createOnce(context) {
    const checkAssertion = (node: TypeAssertion) => {
      if (isConstAssertion(node) || hasSafetyComment(context.sourceCode, node))
        return;
      context.report({ node, messageId: "missingSafetyComment" });
    };

    return {
      TSAsExpression: checkAssertion,
      TSTypeAssertion: checkAssertion,
    };
  },
});
