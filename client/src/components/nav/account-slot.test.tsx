import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { AccountSlot } from "@/components/nav/account-slot";

vi.mock("@/components/custom-user-button", () => ({
  CustomUserButton: () => <button type="button">User menu</button>,
}));

vi.mock("@/components/guest-user-button", () => ({
  GuestUserButton: () => <button type="button">Guest menu</button>,
}));

describe("AccountSlot", () => {
  it("renders the signed-in account menu for authenticated users", () => {
    render(<AccountSlot user={{} as CurrentUser} />);

    expect(screen.getByRole("button", { name: "User menu" })).toBeVisible();
  });

  it("renders the guest account menu without a user", () => {
    render(<AccountSlot user={null} />);

    expect(screen.getByRole("button", { name: "Guest menu" })).toBeVisible();
  });
});
