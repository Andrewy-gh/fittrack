import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { CurrentUser } from "@stackframe/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockNavigate = vi.fn();
const { mockCreateBillingCustomerPortalSession, mockRedirectToBillingPortal } =
  vi.hoisted(() => ({
    mockCreateBillingCustomerPortalSession: vi.fn(),
    mockRedirectToBillingPortal: vi.fn(),
  }));
const { mockDeleteAccount } = vi.hoisted(() => ({
  mockDeleteAccount: vi.fn(),
}));
const { mockClearCurrentDeviceAccountState } = vi.hoisted(() => ({
  mockClearCurrentDeviceAccountState: vi.fn(),
}));

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock("@/features/chat/api/billing", () => ({
  createBillingCustomerPortalSession: mockCreateBillingCustomerPortalSession,
  redirectToBillingPortal: mockRedirectToBillingPortal,
}));

vi.mock("@/features/account/api/account", () => ({
  deleteAccount: mockDeleteAccount,
}));

vi.mock("@/features/account/utils/current-device-state", () => ({
  clearCurrentDeviceAccountState: mockClearCurrentDeviceAccountState,
}));

import { AccountSettingsPage } from "@/features/account/pages/account-settings-page";

const testUserSignOut = vi.fn();
const testUser = {
  id: "user-123",
  displayName: "Andy",
  primaryEmail: "andy@example.com",
  signOut: testUserSignOut,
} as unknown as CurrentUser;

describe("AccountSettingsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateBillingCustomerPortalSession.mockResolvedValue({
      url: "https://billing.stripe.test/session",
    });
    mockDeleteAccount.mockResolvedValue(undefined);
  });

  it("separates billing management from checkbox-confirmed account deletion", async () => {
    const user = userEvent.setup();

    render(<AccountSettingsPage user={testUser} />);

    expect(
      screen.getByRole("heading", { name: "Account settings" }),
    ).toBeInTheDocument();

    const billingSection = screen
      .getByRole("heading", { name: "Billing" })
      .closest("section");
    expect(billingSection).not.toBeNull();
    expect(
      within(billingSection as HTMLElement).getByRole("button", {
        name: "Manage billing",
      }),
    ).toBeEnabled();

    const deletionSection = screen
      .getByRole("heading", { name: "Delete account" })
      .closest("section");
    expect(deletionSection).not.toBeNull();

    const deletion = within(deletionSection as HTMLElement);
    expect(
      deletion.getByText(
        "Deleting your account cancels your AI chat subscription so it will not renew. Your current billing period may already have been charged. If you recently paid and want FitTrack to review a refund, contact support@fittrack.andrewy.me.",
      ),
    ).toBeInTheDocument();
    expect(
      deletion.getByText(
        /contact privacy@fittrack\.andrewy\.me before deleting/i,
      ),
    ).toBeInTheDocument();

    const deleteButton = deletion.getByRole("button", {
      name: "Delete account",
    });
    expect(deleteButton).toBeDisabled();

    await user.click(
      deletion.getByRole("checkbox", {
        name: /I understand this deletes my FitTrack app data/i,
      }),
    );

    expect(deleteButton).toBeEnabled();
  });

  it("opens the Stripe billing portal from account settings", async () => {
    const user = userEvent.setup();

    render(<AccountSettingsPage user={testUser} />);

    await user.click(screen.getByRole("button", { name: "Manage billing" }));

    expect(mockCreateBillingCustomerPortalSession).toHaveBeenCalledOnce();
    expect(mockRedirectToBillingPortal).toHaveBeenCalledWith(
      "https://billing.stripe.test/session",
    );
  });

  it("clears local app state, signs out, and navigates home after account deletion succeeds", async () => {
    const user = userEvent.setup();

    render(<AccountSettingsPage user={testUser} />);

    const deletionSection = screen
      .getByRole("heading", { name: "Delete account" })
      .closest("section");
    const deletion = within(deletionSection as HTMLElement);

    await user.click(
      deletion.getByRole("checkbox", {
        name: /I understand this deletes my FitTrack app data/i,
      }),
    );
    await user.click(deletion.getByRole("button", { name: "Delete account" }));

    expect(mockDeleteAccount).toHaveBeenCalledOnce();
    expect(mockClearCurrentDeviceAccountState).toHaveBeenCalledWith("user-123");
    expect(testUserSignOut).toHaveBeenCalledOnce();
    expect(mockNavigate).toHaveBeenCalledWith({ to: "/" });
  });

  it("keeps local app state and the signed-in session when account deletion fails", async () => {
    const user = userEvent.setup();
    mockDeleteAccount.mockRejectedValueOnce(new Error("delete failed"));

    render(<AccountSettingsPage user={testUser} />);

    const deletionSection = screen
      .getByRole("heading", { name: "Delete account" })
      .closest("section");
    const deletion = within(deletionSection as HTMLElement);

    await user.click(
      deletion.getByRole("checkbox", {
        name: /I understand this deletes my FitTrack app data/i,
      }),
    );
    await user.click(deletion.getByRole("button", { name: "Delete account" }));

    expect(
      await deletion.findByText(
        "Could not delete your account. Please try again.",
      ),
    ).toBeInTheDocument();
    expect(mockClearCurrentDeviceAccountState).not.toHaveBeenCalled();
    expect(testUserSignOut).not.toHaveBeenCalled();
    expect(mockNavigate).not.toHaveBeenCalled();
  });
});
