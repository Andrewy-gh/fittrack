import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
  it("labels historical loads and opens a separately selected progression in the workout form", async () => {
    const user = userEvent.setup();
    const original: AIWorkoutDraft = {
      ...draft,
      exercises: [
        {
          name: "Bench Press",
          draftProgression: { weight: 117.5, reps: 5, sets: 1 },
          sets: [
            { reps: 6, setType: "warmup", weight: 0 },
            {
              reps: 8,
              setType: "working",
              weight: 115,
              weightSourceDate: "2026-04-19T12:00:00Z",
            },
          ],
        },
        ...draft.exercises,
      ],
    };
    const onEdit = vi.fn();
    render(
      <ChatWorkoutDraftCard
        draft={original}
        saveState={{ status: "idle" }}
        onSave={vi.fn()}
        onEdit={onEdit}
      />,
    );

    expect(screen.getByText(/115 lb.*Suggested from/)).toBeVisible();
    expect(screen.getByText(/No added weight/)).toBeVisible();
    expect(screen.getByText(/Choose weight/)).toBeVisible();
    await user.click(screen.getByRole("button", { name: /Try 117.5 lb/ }));
    expect(onEdit.mock.calls[0]?.[0].exercises[0].sets).toEqual([
      { reps: 6, setType: "warmup", weight: 0 },
      { reps: 5, setType: "working", weight: 117.5 },
    ]);
    expect(screen.getByText(/115 lb.*Suggested from/)).toBeVisible();
  });

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
