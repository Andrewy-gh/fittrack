import type { ComponentPropsWithoutRef } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

const { mockUseUser } = vi.hoisted(() => ({
  mockUseUser: vi.fn(),
}));

vi.mock("@stackframe/react", () => ({
  useUser: mockUseUser,
}));

vi.mock("@/components/theme-provider", () => ({
  useTheme: () => ({
    theme: "light",
    setTheme: vi.fn(),
  }),
}));

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    children,
    to,
    ...props
  }: ComponentPropsWithoutRef<"a"> & { to: string }) => (
    <a
      href={to}
      {...props}
    >
      {children}
    </a>
  ),
}));

vi.mock("@/lib/local-dev-auth", () => ({
  getLocalDevRouteUser: () => null,
}));

import { CustomUserButton } from "@/components/custom-user-button";

describe("CustomUserButton", () => {
  it("links signed-in users to account settings from the user menu", async () => {
    const user = userEvent.setup();
    mockUseUser.mockReturnValue({
      id: "user-123",
      displayName: "Andy",
      primaryEmail: "andy@example.com",
      profileImageUrl: null,
      signOut: vi.fn(),
    });

    render(<CustomUserButton />);

    await user.click(screen.getByRole("button"));

    expect(
      await screen.findByRole("menuitem", { name: /settings/i }),
    ).toHaveAttribute("href", "/settings");
  });
});
