/**
 * The authenticated-user capability the application passes through routes and UI.
 *
 * Stack sessions satisfy this contract, while local development provides the same
 * visible identity and sign-out capability without imitating Stack's full API.
 */
export type ApplicationUser = {
  readonly id: string;
  readonly displayName: string | null;
  readonly primaryEmail: string | null;
  readonly profileImageUrl: string | null;
  signOut(): Promise<void>;
};
