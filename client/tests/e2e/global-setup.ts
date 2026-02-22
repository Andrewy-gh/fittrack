import { chromium, type FullConfig } from '@playwright/test';
import { mkdir, readFile } from 'node:fs/promises';
import { randomUUID } from 'node:crypto';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { signInWithStack } from './helpers/stack-auth';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const authFile = path.join(__dirname, '.auth', 'stack.json');
const serverEnvPath = path.resolve(__dirname, '..', '..', '..', 'server', '.env');
const clientEnvPath = path.resolve(__dirname, '..', '..', '..', 'client', '.env');

type StackTokens = { refreshToken: string; accessToken: string };

function parseEnvFile(contents: string): Record<string, string> {
  const result: Record<string, string> = {};
  for (const line of contents.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const eqIndex = trimmed.indexOf('=');
    if (eqIndex === -1) continue;
    const key = trimmed.slice(0, eqIndex).trim();
    let value = trimmed.slice(eqIndex + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    result[key] = value;
  }
  return result;
}

async function loadEnvFile(filePath: string) {
  try {
    const contents = await readFile(filePath, 'utf8');
    return parseEnvFile(contents);
  } catch {
    return {};
  }
}

function buildStackHeaders(options: {
  accessType: 'client' | 'server' | 'admin';
  projectId: string;
  publishableClientKey?: string;
  secretServerKey?: string;
}) {
  return {
    'content-type': 'application/json',
    'x-stack-access-type': options.accessType,
    'x-stack-project-id': options.projectId,
    ...(options.publishableClientKey
      ? { 'x-stack-publishable-client-key': options.publishableClientKey }
      : {}),
    ...(options.secretServerKey
      ? { 'x-stack-secret-server-key': options.secretServerKey }
      : {}),
  } as Record<string, string>;
}

async function stackRequest<T>(
  baseUrl: string,
  pathName: string,
  options: {
    method?: 'GET' | 'POST';
    headers: Record<string, string>;
    body?: unknown;
  }
): Promise<{ status: number; body: T }> {
  const response = await fetch(`${baseUrl}${pathName}`, {
    method: options.method ?? 'GET',
    headers: options.headers,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });
  const text = await response.text();
  let data: T;
  try {
    data = JSON.parse(text) as T;
  } catch {
    data = text as T;
  }
  return { status: response.status, body: data };
}

async function ensureUserId(options: {
  baseUrl: string;
  projectId: string;
  secretServerKey: string;
  publishableClientKey?: string;
  email: string;
}) {
  const headers = buildStackHeaders({
    accessType: 'server',
    projectId: options.projectId,
    publishableClientKey: options.publishableClientKey,
    secretServerKey: options.secretServerKey,
  });

  const createResponse = await stackRequest<{ id?: string; code?: string }>(
    options.baseUrl,
    '/users',
    {
      method: 'POST',
      headers,
      body: {
        primary_email: options.email,
      },
    }
  );

  if (createResponse.status === 201 && createResponse.body.id) {
    return createResponse.body.id;
  }

  if (
    createResponse.status !== 409 ||
    typeof createResponse.body !== 'object' ||
    createResponse.body === null ||
    (createResponse.body as { code?: string }).code !== 'USER_EMAIL_ALREADY_EXISTS'
  ) {
    const details =
      typeof createResponse.body === 'string'
        ? createResponse.body
        : JSON.stringify(createResponse.body);
    throw new Error(`Failed to create Stack Auth user: ${createResponse.status} ${details}`);
  }

  const listResponse = await stackRequest<{ items?: Array<{ id: string; primary_email?: string }> }>(
    options.baseUrl,
    `/users?query=${encodeURIComponent(options.email)}`,
    { headers }
  );

  const existingUser = listResponse.body.items?.find(
    (user) => user.primary_email?.toLowerCase() === options.email.toLowerCase()
  );

  if (!existingUser) {
    throw new Error('Stack Auth user exists but could not be found in list results.');
  }

  return existingUser.id;
}

async function createSession(options: {
  baseUrl: string;
  projectId: string;
  secretServerKey: string;
  publishableClientKey?: string;
  userId: string;
}): Promise<StackTokens> {
  const headers = buildStackHeaders({
    accessType: 'server',
    projectId: options.projectId,
    publishableClientKey: options.publishableClientKey,
    secretServerKey: options.secretServerKey,
  });

  const response = await stackRequest<{ refresh_token: string; access_token: string }>(
    options.baseUrl,
    '/auth/sessions',
    {
      method: 'POST',
      headers,
      body: {
        user_id: options.userId,
        expires_in_millis: 1000 * 60 * 60 * 2,
      },
    }
  );

  if (response.status !== 200) {
    throw new Error(`Failed to create Stack Auth session: ${response.status}`);
  }

  return {
    refreshToken: response.body.refresh_token,
    accessToken: response.body.access_token,
  };
}

function buildAuthCookies(options: {
  baseUrl: string;
  projectId: string;
  refreshToken: string;
  accessToken: string;
}) {
  const appUrl = new URL(options.baseUrl);
  const isSecure = appUrl.protocol === 'https:';
  const cookieUrl = new URL('/', appUrl).toString();
  const refreshName = `${isSecure ? '__Host-' : ''}stack-refresh-${options.projectId}--default`;
  const legacyRefreshName = `stack-refresh-${options.projectId}`;
  const refreshPayload = JSON.stringify({
    refresh_token: options.refreshToken,
    updated_at_millis: Date.now(),
  });
  const accessPayload = JSON.stringify([options.refreshToken, options.accessToken]);

  const nowSeconds = Math.floor(Date.now() / 1000);
    return [
      {
        name: legacyRefreshName,
        value: options.refreshToken,
        url: cookieUrl,
        expires: nowSeconds + 60 * 60 * 24 * 365,
      },
      {
        name: refreshName,
        value: refreshPayload,
        url: cookieUrl,
        expires: nowSeconds + 60 * 60 * 24 * 365,
      },
      {
        name: 'stack-access',
        value: accessPayload,
        url: cookieUrl,
        expires: nowSeconds + 60 * 60 * 24,
      },
    ];
}

async function globalSetup(config: FullConfig) {
  const email = process.env.E2E_STACK_EMAIL;
  const password = process.env.E2E_STACK_PASSWORD;

  const baseURL = config.projects[0]?.use?.baseURL ?? 'http://localhost:5173';
  const serverEnv = await loadEnvFile(serverEnvPath);
  const clientEnv = await loadEnvFile(clientEnvPath);

  // The browser app uses Vite-exposed env vars, so only consider VITE_* here.
  // If Stack isn't configured for the client, skip auth state generation so auth e2e tests will skip.
  const projectId = process.env.VITE_PROJECT_ID || clientEnv.VITE_PROJECT_ID;
  const secretServerKey =
    process.env.SECRET_SERVER_KEY || serverEnv.SECRET_SERVER_KEY;
  const publishableClientKey =
    process.env.VITE_PUBLISHABLE_CLIENT_KEY || clientEnv.VITE_PUBLISHABLE_CLIENT_KEY;
  const apiBaseUrl = process.env.STACK_API_BASE_URL || 'https://api.stack-auth.com';

  const browser = await chromium.launch();
  const context = await browser.newContext({ baseURL });
  const canUseServerAuth = Boolean(projectId && secretServerKey);
  const canUseUiAuth = Boolean(email && password);

  if (canUseServerAuth) {
    const generatedEmail = `e2e-${randomUUID()}@example.test`;
    const resolvedEmail = email ?? generatedEmail;
    const baseUrl = `${apiBaseUrl}/api/v1`;
    const userId = await ensureUserId({
      baseUrl,
      projectId,
      secretServerKey,
      publishableClientKey,
      email: resolvedEmail,
    });
    const tokens = await createSession({
      baseUrl,
      projectId,
      secretServerKey,
      publishableClientKey,
      userId,
    });

    await context.addCookies(
      buildAuthCookies({
        baseUrl: baseURL,
        projectId,
        refreshToken: tokens.refreshToken,
        accessToken: tokens.accessToken,
      })
    );
  } else if (canUseUiAuth) {
    if (!email || !password) {
      await browser.close();
      return;
    }
    const page = await context.newPage();
    await signInWithStack(page, email, password);
  } else {
    await browser.close();
    return;
  }

  await mkdir(path.dirname(authFile), { recursive: true });
  await context.storageState({ path: authFile });
  await browser.close();
}

export default globalSetup;
