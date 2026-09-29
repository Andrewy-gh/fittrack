import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { AIWorkoutDraft } from "@/features/chat/api/ai-chat";
import { ChatWorkoutDraftCard } from "./chat-workout-draft-card";

const draft: AIWorkoutDraft = {
  date: "2026-04-20T12:00:00Z",
  workoutFocus: "pull",
  exercises: [
    {
      name: "Chest Supported Row",
      sets: [{ reps: 10, setType: "working" }],
    },
  ],
};

describe("ChatWorkoutDraftCard", () => {
  it("shows a saved draft without an open-workout action when no saved workout is available", () => {
    render(
      <ChatWorkoutDraftCard
        draft={draft}
        saveState={{ status: "saved" }}
        onSave={vi.fn()}
        onEdit={vi.fn()}
      />,
    );

    expect(screen.getByRole("button", { name: "Saved" })).toBeDisabled();
    expect(
      screen.queryByRole("button", { name: "Open saved workout" }),
    ).not.toBeInTheDocument();
  });
});
