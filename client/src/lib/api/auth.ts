import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';

export type User = CurrentUser | CurrentInternalUser | null;

/**
 * Validates that a User is not null and has required properties.
 * Throws descriptive errors if validation fails.
 * Returns the user with proper typing (null excluded).
 */
export function checkUser(user: User): asserts user is Exclude<User, null> {
  if (!user) {
    throw new Error('User not found');
  }
  if (!user.id || typeof user.id !== 'string') {
    throw new Error('User ID not found');
  }
}

/**
 * Alternative utility that returns the validated user instead of using type assertion.
 * Useful when you need the return value directly.
 */
export function ensureUser(user: User): Exclude<User, null> {
  checkUser(user);
  return user;
}

export async function getAccessToken(user: User) {
  checkUser(user); // Use the new utility for validation
  const { accessToken } = await user.getAuthJson();
  if (!accessToken) {
    throw new Error('Access token not found');
  }
  return accessToken;
}
