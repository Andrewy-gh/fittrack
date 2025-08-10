import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';

export type User = CurrentUser | CurrentInternalUser | null;

export async function getAccessToken(user: User) {
  if (!user) {
    throw new Error('User not found');
  }
  if (!user.id || typeof user.id !== 'string') {
    throw new Error('User ID not found');
  }
  const { accessToken } = await user.getAuthJson();
  if (!accessToken) {
    throw new Error('Access token not found');
  }
  return accessToken;
}
