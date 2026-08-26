import type { ApplicationUser } from "@/lib/application-user";

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

function parseLocalDevAuthSession(value: unknown): LocalDevAuthSession | null {
  if (
    typeof value !== "object" ||
    value === null ||
    Array.isArray(value) ||
    !("userId" in value) ||
    !("email" in value) ||
    !("displayName" in value)
  ) {
    return null;
  }

  const { userId, email, displayName } = value;
  if (
    typeof userId !== "string" ||
    typeof email !== "string" ||
    typeof displayName !== "string"
  ) {
    return null;
  }

  return { userId, email, displayName };
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

    const parsed: unknown = JSON.parse(raw);
    return parseLocalDevAuthSession(parsed);
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

export function getLocalDevRouteUser(): ApplicationUser | null {
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
  };
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
