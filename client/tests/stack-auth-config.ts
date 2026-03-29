type EnvSource = Record<string, string | undefined>;

export type StackAuthBootstrapConfig = {
  email?: string;
  password?: string;
  projectId?: string;
  secretServerKey?: string;
  publishableClientKey?: string;
  apiBaseUrl: string;
};

function pickFirst(...values: Array<string | undefined>) {
  return values.find((value) => typeof value === 'string' && value.length > 0);
}

export function resolveStackAuthBootstrapConfig(options: {
  processEnv: EnvSource;
  clientEnv: EnvSource;
  serverEnv: EnvSource;
}): StackAuthBootstrapConfig {
  const { processEnv, clientEnv, serverEnv } = options;

  return {
    email: pickFirst(processEnv.E2E_STACK_EMAIL),
    password: pickFirst(processEnv.E2E_STACK_PASSWORD),
    projectId: pickFirst(
      processEnv.VITE_PROJECT_ID,
      processEnv.STACK_PROJECT_ID,
      processEnv.PROJECT_ID,
      clientEnv.VITE_PROJECT_ID,
      serverEnv.VITE_PROJECT_ID,
      serverEnv.PROJECT_ID
    ),
    secretServerKey: pickFirst(
      processEnv.SECRET_SERVER_KEY,
      serverEnv.SECRET_SERVER_KEY
    ),
    publishableClientKey: pickFirst(
      processEnv.VITE_PUBLISHABLE_CLIENT_KEY,
      clientEnv.VITE_PUBLISHABLE_CLIENT_KEY
    ),
    apiBaseUrl: pickFirst(
      processEnv.STACK_API_BASE_URL,
      'https://api.stack-auth.com'
    )!,
  };
}
