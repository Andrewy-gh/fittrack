import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { GuestUserButton } from "@/components/guest-user-button";
import { ThemeProvider } from "@/components/theme-provider";

describe("GuestUserButton", () => {
  it("names the icon-only menu trigger", () => {
    render(
      <ThemeProvider>
        <GuestUserButton />
      </ThemeProvider>,
    );

    expect(
      screen.getByRole("button", { name: "Guest user menu" }),
    ).toBeInTheDocument();
  });
});
