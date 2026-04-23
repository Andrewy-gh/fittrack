type EnvSource = Record<string, string | undefined>;

export type StackAuthBootstrapConfig = {
  email?: string;
  password?: string;
  projectId?: string;
  secretServerKey?: string;
  publishableClientKey?: string;
  apiBaseUrl: string;
  localAuthEnabled: boolean;
  localAuthUserId: string;
  localAuthEmail: string;
  localAuthDisplayName: string;
  localAuthApiBaseUrl: string;
};

function pickFirst(...values: Array<string | undefined>) {
  return values.find((value) => typeof value === "string" && value.length > 0);
}

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
      processEnv.VITE_API_BASE_URL,
      clientEnv.VITE_API_BASE_URL,
      "http://localhost:8080",
    ) ?? "http://localhost:8080";

  return {
    email: pickFirst(processEnv.E2E_STACK_EMAIL),
    password: pickFirst(processEnv.E2E_STACK_PASSWORD),
    projectId: pickFirst(
      processEnv.VITE_PROJECT_ID,
      clientEnv.VITE_PROJECT_ID,
      processEnv.STACK_PROJECT_ID,
      processEnv.PROJECT_ID,
      serverEnv.VITE_PROJECT_ID,
      serverEnv.PROJECT_ID,
    ),
    secretServerKey: pickFirst(
      processEnv.SECRET_SERVER_KEY,
      serverEnv.SECRET_SERVER_KEY,
    ),
    publishableClientKey: pickFirst(
      processEnv.VITE_PUBLISHABLE_CLIENT_KEY,
      clientEnv.VITE_PUBLISHABLE_CLIENT_KEY,
    ),
    apiBaseUrl: pickFirst(
      processEnv.STACK_API_BASE_URL,
      "https://api.stack-auth.com",
    )!,
    localAuthEnabled: isTruthy(
      pickFirst(
        processEnv.E2E_LOCAL_AUTH_ENABLED,
        processEnv.VITE_E2E_LOCAL_AUTH_ENABLED,
        clientEnv.VITE_E2E_LOCAL_AUTH_ENABLED,
        serverEnv.E2E_LOCAL_AUTH_ENABLED,
      ),
    ),
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
