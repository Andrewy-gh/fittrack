import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { getPolicySections } from "@/features/privacy/privacy-content";

describe("privacy policy account deletion copy", () => {
  it("describes in-app deletion and the pre-deletion data request path", () => {
    const choicesSection = getPolicySections().find(
      (section) => section.id === "your-choices-and-rights",
    );

    expect(choicesSection).toBeDefined();

    render(<>{choicesSection?.content}</>);

    expect(
      screen.getByText(
        /You can delete your FitTrack account in Account settings/i,
      ),
    ).toBeInTheDocument();
    expect(
      screen.getByText((_, element) =>
        Boolean(
          element?.tagName === "P" &&
          element?.textContent?.includes(
            "If you want a copy of your data, request a copy at privacy@fittrack.andrewy.me before deleting",
          ),
        ),
      ),
    ).toBeInTheDocument();
  });
});
