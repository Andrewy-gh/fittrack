import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ChatComposer } from "./chat-composer";

describe("ChatComposer", () => {
  it("keeps the mobile textarea font at 16px to avoid iOS focus zoom", () => {
    render(
      <ChatComposer
        value=""
        onChange={vi.fn()}
        onSubmit={vi.fn()}
        placeholder="Ask FitTrack"
      />,
    );

    const textarea = screen.getByPlaceholderText("Ask FitTrack");

    expect(textarea).toHaveClass("text-base", "md:text-sm");
    expect(textarea).not.toHaveClass("text-sm");
  });
});
