import {
  StackClientApp as StackClientAppCtor,
  type StackClientApp,
} from "@stackframe/react";
import { useNavigate } from "@tanstack/react-router";

function parseEnvironmentString(value: unknown): string | undefined {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

const projectId = parseEnvironmentString(import.meta.env.VITE_PROJECT_ID);
const publishableClientKey = parseEnvironmentString(
  import.meta.env.VITE_PUBLISHABLE_CLIENT_KEY,
);

export const isStackConfigured = Boolean(projectId && publishableClientKey);

export const stackClientApp: StackClientApp<true, string> | null = (() => {
  if (!projectId || !publishableClientKey) return null;
  try {
    return new StackClientAppCtor({
      projectId,
      publishableClientKey,
      tokenStore: "cookie",
      redirectMethod: {
        useNavigate: () => {
          const navigate = useNavigate();
          return (to: string) => {
            navigate({ to });
          };
        },
      },
    });
  } catch (err) {
    // Missing/invalid Stack Auth config should not take down demo mode.
    console.error(
      "Failed to initialize Stack Auth client; falling back to demo mode.",
      err,
    );
    return null;
  }
})();
