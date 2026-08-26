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

function isConstAssertion(node: TypeAssertion): boolean {
  return (
    node.typeAnnotation.type === "TSTypeReference" &&
    node.typeAnnotation.typeName.type === "Identifier" &&
    node.typeAnnotation.typeName.name === "const"
  );
}

function getAsOperatorToken(sourceCode: SourceCode, node: TypeAssertion) {
  if (node.type === "TSTypeAssertion") return null;
  return (
    sourceCode
      .getTokensBetween(node.expression, node.typeAnnotation)
      .find((token) => token.value === "as") ?? null
  );
}

function hasSafetyComment(
  sourceCode: SourceCode,
  node: TypeAssertion,
): boolean {
  const asOperator = getAsOperatorToken(sourceCode, node);
  const assertionOperatorStart = asOperator?.start ?? node.start;
  const tokenBeforeAs =
    asOperator === null ? null : sourceCode.getTokenBefore(asOperator);
  let current: ESTree.Node = node;

  while (true) {
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
              tokenBeforeAs !== null &&
              comment.start >= tokenBeforeAs.end &&
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
