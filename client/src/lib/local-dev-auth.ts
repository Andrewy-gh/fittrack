const STORAGE_KEY = "fittrack-local-e2e-auth";
const DEV_AUTH_HEADER = "x-fittrack-dev-e2e-user";

export type LocalDevAuthSession = {
  userId: string;
  email: string;
  displayName: string;
};

function parseBooleanFlag(value: unknown): boolean {
  if (typeof value !== "string") {
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

export function isLocalDevAuthEnabled(): boolean {
  return (
    import.meta.env.DEV &&
    parseBooleanFlag(import.meta.env.VITE_E2E_LOCAL_AUTH_ENABLED)
  );
}

export function getLocalDevAuthStorageKey(): string {
  return STORAGE_KEY;
}

export function getLocalDevAuthHeaderName(): string {
  return DEV_AUTH_HEADER;
}

export function getLocalDevAuthSession(): LocalDevAuthSession | null {
  if (!isLocalDevAuthEnabled() || typeof window === "undefined") {
    return null;
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }

    const parsed = JSON.parse(raw) as Partial<LocalDevAuthSession>;
    if (
      typeof parsed.userId !== "string" ||
      typeof parsed.email !== "string" ||
      typeof parsed.displayName !== "string"
    ) {
      return null;
    }

    return {
      userId: parsed.userId,
      email: parsed.email,
      displayName: parsed.displayName,
    };
  } catch {
    return null;
  }
}

export function getLocalDevAuthHeaderValue(): string | null {
  return getLocalDevAuthSession()?.userId ?? null;
}

export function applyLocalDevAuthHeader(headers: Headers): Headers {
  const userId = getLocalDevAuthHeaderValue();
  if (userId) {
    headers.set(DEV_AUTH_HEADER, userId);
  }
  return headers;
}

export function getLocalDevRouteUser():
  | CurrentUser
  | CurrentInternalUser
  | null {
  const session = getLocalDevAuthSession();
  if (!session) {
    return null;
  }

  return {
    id: session.userId,
    displayName: session.displayName,
    primaryEmail: session.email,
    profileImageUrl: null,
    async signOut() {
      clearLocalDevAuthSession();
      window.location.assign("/");
    },
  } as unknown as CurrentUser;
}

export function clearLocalDevAuthSession(): void {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Ignore local-only cleanup failures.
  }
}
import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
