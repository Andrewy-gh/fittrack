type EnvSource = Record<string, string | undefined>;

/** Resolved bootstrap inputs shared by Playwright and the read-only doctor. */
export type StackAuthBootstrapConfig = {
  readonly browserProjectId: string | undefined;
  readonly serverProjectId: string | undefined;
  readonly serverLocalAuthEnabled: boolean;
  readonly clientLocalAuthEnabled: boolean;
  readonly email?: string | undefined;
  readonly password?: string | undefined;
  readonly projectId?: string | undefined;
  readonly secretServerKey?: string | undefined;
  readonly publishableClientKey?: string | undefined;
  readonly apiBaseUrl: string;
  readonly localAuthEnabled: boolean;
  readonly localAuthUserId: string;
  readonly localAuthEmail: string;
  readonly localAuthDisplayName: string;
  readonly localAuthApiBaseUrl: string;
};

function pickFirst(...values: Array<string | undefined>) {
  return values.find((value) => typeof value === "string" && value.length > 0);
}

/** Resolve session inputs before file fallbacks; explicit empty values remain authoritative. */
export function resolveStackAuthBootstrapConfig(options: {
  processEnv: EnvSource;
  clientEnv: EnvSource;
  serverEnv: EnvSource;
}): StackAuthBootstrapConfig {
  const { processEnv, clientEnv, serverEnv } = options;
  const localAuthApiBaseUrl =
    pickFirst(
      processEnv.E2E_LOCAL_AUTH_API_BASE_URL,
      serverEnv.E2E_LOCAL_AUTH_API_BASE_URL,
      processEnv.API_PROXY_TARGET,
      processEnv.VITE_API_BASE_URL,
      clientEnv.VITE_API_BASE_URL,
      "http://localhost:8080",
    ) ?? "http://localhost:8080";

  const browserProjectId =
    processEnv.VITE_PROJECT_ID ?? clientEnv.VITE_PROJECT_ID;
  const serverProjectId = processEnv.PROJECT_ID ?? serverEnv.PROJECT_ID;
  const serverLocalAuthEnabled = isTruthy(
    processEnv.E2E_LOCAL_AUTH_ENABLED ?? serverEnv.E2E_LOCAL_AUTH_ENABLED,
  );
  const clientLocalAuthEnabled = isTruthy(
    processEnv.VITE_E2E_LOCAL_AUTH_ENABLED ??
      clientEnv.VITE_E2E_LOCAL_AUTH_ENABLED,
  );
  return {
    browserProjectId,
    serverProjectId,
    serverLocalAuthEnabled,
    clientLocalAuthEnabled,
    email: processEnv.E2E_STACK_EMAIL,
    password: processEnv.E2E_STACK_PASSWORD,
    projectId:
      browserProjectId ?? processEnv.STACK_PROJECT_ID ?? serverProjectId,
    secretServerKey:
      processEnv.SECRET_SERVER_KEY ?? serverEnv.SECRET_SERVER_KEY,
    publishableClientKey:
      processEnv.VITE_PUBLISHABLE_CLIENT_KEY ??
      clientEnv.VITE_PUBLISHABLE_CLIENT_KEY,
    apiBaseUrl:
      pickFirst(processEnv.STACK_API_BASE_URL, "https://api.stack-auth.com") ??
      "https://api.stack-auth.com",
    localAuthEnabled: serverLocalAuthEnabled || clientLocalAuthEnabled,
    localAuthUserId:
      pickFirst(
        processEnv.E2E_LOCAL_AUTH_USER_ID,
        serverEnv.E2E_LOCAL_AUTH_USER_ID,
        "local-e2e-user",
      ) ?? "local-e2e-user",
    localAuthEmail:
      pickFirst(
        processEnv.E2E_LOCAL_AUTH_EMAIL,
        serverEnv.E2E_LOCAL_AUTH_EMAIL,
        "local-e2e-user@example.test",
      ) ?? "local-e2e-user@example.test",
    localAuthDisplayName:
      pickFirst(
        processEnv.E2E_LOCAL_AUTH_DISPLAY_NAME,
        serverEnv.E2E_LOCAL_AUTH_DISPLAY_NAME,
        "Local E2E User",
      ) ?? "Local E2E User",
    localAuthApiBaseUrl,
  };
}

function isTruthy(value: string | undefined): boolean {
  if (!value) {
    return false;
  }

  switch (value.trim().toLowerCase()) {
    case "1":
    case "true":
    case "yes":
    case "on":
      return true;
    default:
      return false;
  }
}

/** Configuration readiness only; no provider session or API access has been verified. */
export type E2EAuthReadiness =
  | { readonly kind: "invalid"; readonly message: string }
  | { readonly kind: "guest" | "local" | "server" | "ui" };

/** Select the supported bootstrap or return a safe repair without echoing inputs. */
export function e2eAuthReadiness(
  config: StackAuthBootstrapConfig,
  isCI: boolean,
): E2EAuthReadiness {
  const invalid = (message: string): E2EAuthReadiness => ({
    kind: "invalid",
    message,
  });
  if (
    !config.localAuthEnabled &&
    !config.secretServerKey &&
    !config.email &&
    !config.password
  ) {
    return { kind: "guest" };
  }
  if (config.localAuthEnabled && isCI) {
    return invalid(
      "Local E2E auth requires the Vite development server. CI uses production preview, where local auth is disabled. Unset CI for local checks or disable both local-auth flags and configure real Stack auth for preview.",
    );
  }
  if (!config.browserProjectId || !config.publishableClientKey) {
    return invalid(
      "Authenticated E2E requires VITE_PROJECT_ID and VITE_PUBLISHABLE_CLIENT_KEY in the frontend environment; otherwise the app renders as a guest.",
    );
  }
  // StackClientApp rejects malformed project IDs before the frontend can mount.
  // Keep the installed SDK's constructor contract; key/session validity remains live proof.
  if (
    config.browserProjectId !== "internal" &&
    !/^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(
      config.browserProjectId,
    )
  ) {
    return invalid(
      "VITE_PROJECT_ID must be a Stack project UUID (or Stack's internal project identifier). Check PROJECT_ID and VITE_PROJECT_ID; values are withheld.",
    );
  }
  if (
    !config.serverProjectId ||
    config.serverProjectId !== config.browserProjectId
  ) {
    return invalid(
      "Authenticated E2E requires PROJECT_ID and VITE_PROJECT_ID to identify the same Stack project.",
    );
  }
  if (config.localAuthEnabled) {
    if (!config.serverLocalAuthEnabled || !config.clientLocalAuthEnabled) {
      return invalid(
        "Local E2E auth requires both E2E_LOCAL_AUTH_ENABLED and VITE_E2E_LOCAL_AUTH_ENABLED to be enabled (1, true, yes or on). Restart the API and Vite after changing flags.",
      );
    }
    return { kind: "local" };
  }
  if (config.secretServerKey) return { kind: "server" };
  if (config.email && config.password) return { kind: "ui" };
  return invalid(
    "Configure SECRET_SERVER_KEY or both E2E_STACK_EMAIL and E2E_STACK_PASSWORD for Stack test login.",
  );
}
