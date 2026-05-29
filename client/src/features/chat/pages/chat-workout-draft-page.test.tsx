import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AIWorkoutDraft } from "@/features/chat/api/ai-chat";
import {
  ChatRouteComponent,
  conversationDetail,
  mockGetConversation,
  mockNavigate,
  mockSaveLatestWorkoutDraft,
  mockStreamMessage,
  mockToastSuccess,
  resetChatRouteMocks,
} from "../test/chat-page-test-utils";

describe("ChatRouteComponent", () => {
  beforeEach(resetChatRouteMocks);

  it("shows the latest workout draft on reopen and imports it into the workout form flow", async () => {
    const user = userEvent.setup();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);
    const latestWorkoutDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Keep rest short",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };

    mockGetConversation.mockResolvedValue(
      conversationDetail([], undefined, latestWorkoutDraft),
    );

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("Latest structured workout draft"),
    ).toBeInTheDocument();
    expect(screen.getByText(/Chest Supported Row/)).toBeInTheDocument();
    expect(screen.getByText("Keep rest short")).toBeInTheDocument();

    await user.click(
      screen.getByRole("button", { name: "Edit in workout form" }),
    );

    expect(confirmSpy).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(
        window.localStorage.getItem("workout-entry-form-data-user-123"),
      ).toBe(
        JSON.stringify({
          date: "2026-04-21T12:00:00Z",
          notes: "Keep rest short",
          workoutFocus: "pull",
          exercises: [
            {
              name: "Chest Supported Row",
              sets: [{ reps: 10, setType: "working" }],
            },
          ],
        }),
      );
    });
    expect(mockNavigate).toHaveBeenCalledWith({ to: "/workouts/new" });
    expect(mockSaveLatestWorkoutDraft).not.toHaveBeenCalled();
    expect(mockToastSuccess).toHaveBeenCalledWith(
      "Workout draft loaded into the form",
    );
  });

  it("saves the latest workout draft directly without overwriting unrelated form draft state", async () => {
    const user = userEvent.setup();
    const latestWorkoutDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "  Keep rest short  ",
      workoutFocus: "  pull  ",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };

    window.localStorage.setItem(
      "workout-entry-form-data-user-123",
      JSON.stringify({
        date: "2026-04-20T12:00:00Z",
        notes: "Old imported draft",
        workoutFocus: "push",
        exercises: [
          {
            name: "Bench Press",
            sets: [{ reps: 8, setType: "working", weight: 185 }],
          },
        ],
      }),
    );
    mockGetConversation.mockResolvedValue(
      conversationDetail([], undefined, latestWorkoutDraft),
    );
    mockSaveLatestWorkoutDraft.mockResolvedValue({
      conversation: {
        id: 41,
        created_at: "2026-03-26T17:00:00Z",
        updated_at: "2026-03-26T17:05:00Z",
        latest_workout_draft: latestWorkoutDraft,
        latest_workout_draft_status: {
          is_saved: true,
          saved_workout_id: 88,
          saved_at: "2026-04-21T12:05:00Z",
        },
      },
      workout_id: 88,
    });

    render(<ChatRouteComponent />);

    await user.click(await screen.findByRole("button", { name: "Save now" }));

    expect(mockSaveLatestWorkoutDraft).toHaveBeenCalledWith(41);
    expect(
      window.localStorage.getItem("workout-entry-form-data-user-123"),
    ).toBe(
      JSON.stringify({
        date: "2026-04-20T12:00:00Z",
        notes: "Old imported draft",
        workoutFocus: "push",
        exercises: [
          {
            name: "Bench Press",
            sets: [{ reps: 8, setType: "working", weight: 185 }],
          },
        ],
      }),
    );
    expect(mockToastSuccess).toHaveBeenCalledWith("Workout saved successfully");
    expect(mockNavigate).not.toHaveBeenCalledWith({ to: "/workouts/new" });
    expect(await screen.findByRole("button", { name: "Saved" })).toBeDisabled();

    await user.click(
      await screen.findByRole("button", { name: "Open saved workout" }),
    );

    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/workouts/$workoutId",
      params: { workoutId: 88 },
    });
  });

  it("shows a disabled Saved button and saved workout link when reopening an already-saved draft", async () => {
    const user = userEvent.setup();
    const latestWorkoutDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Keep rest short",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };

    mockGetConversation.mockResolvedValue(
      conversationDetail([], undefined, latestWorkoutDraft, {
        is_saved: true,
        saved_workout_id: 88,
        saved_at: "2026-04-21T12:05:00Z",
      }),
    );

    render(<ChatRouteComponent />);

    const savedButton = await screen.findByRole("button", { name: "Saved" });
    expect(savedButton).toBeDisabled();
    expect(
      screen.queryByRole("button", { name: "Save now" }),
    ).not.toBeInTheDocument();
    expect(mockSaveLatestWorkoutDraft).not.toHaveBeenCalled();

    await user.click(
      screen.getByRole("button", { name: "Open saved workout" }),
    );

    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/workouts/$workoutId",
      params: { workoutId: 88 },
    });
  });

  it("overwrites the latest workout draft CTA after a regenerated structured workout", async () => {
    const user = userEvent.setup();
    const originalDraft: AIWorkoutDraft = {
      date: "2026-04-20T12:00:00Z",
      notes: "Original draft",
      workoutFocus: "push",
      exercises: [
        {
          name: "Bench Press",
          sets: [{ reps: 8, setType: "working", weight: 185 }],
        },
      ],
    };
    const regeneratedDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Regenerated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };

    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([], undefined, originalDraft))
      .mockResolvedValueOnce(
        conversationDetail(
          [
            {
              id: 71,
              conversation_id: 41,
              role: "user",
              content: "regenerate it",
              status: "completed",
              created_at: "2026-03-26T17:00:01Z",
              updated_at: "2026-03-26T17:00:01Z",
              completed_at: "2026-03-26T17:00:01Z",
            },
            {
              id: 72,
              conversation_id: 41,
              role: "assistant",
              content: "I put together a structured workout draft for you.",
              status: "completed",
              created_at: "2026-03-26T17:00:01Z",
              updated_at: "2026-03-26T17:00:02Z",
              completed_at: "2026-03-26T17:00:02Z",
            },
          ],
          undefined,
          regeneratedDraft,
        ),
      );
    mockStreamMessage.mockImplementation(
      async (
        _conversationId: number,
        _prompt: string,
        options?: {
          onStart?: (event: Record<string, unknown>) => void;
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onStart?.({
          type: "start",
          message_id: 72,
        });
        options?.onDone?.({
          type: "done",
          message_id: 72,
          text: "I put together a structured workout draft for you.",
          workout_draft: regeneratedDraft,
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 72,
            text: "I put together a structured workout draft for you.",
            workout_draft: regeneratedDraft,
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText(/Bench Press/)).toBeInTheDocument();

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "regenerate it",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await user.click(
      await screen.findByRole("button", { name: "Edit in workout form" }),
    );

    expect(screen.getByText(/Chest Supported Row/)).toBeInTheDocument();
    expect(screen.queryByText(/Bench Press/)).not.toBeInTheDocument();
    expect(
      screen.queryAllByRole("button", { name: "Edit in workout form" }),
    ).toHaveLength(1);

    await waitFor(() => {
      expect(
        window.localStorage.getItem("workout-entry-form-data-user-123"),
      ).toBe(
        JSON.stringify({
          date: "2026-04-21T12:00:00Z",
          notes: "Regenerated draft",
          workoutFocus: "pull",
          exercises: [
            {
              name: "Chest Supported Row",
              sets: [{ reps: 10, setType: "working" }],
            },
          ],
        }),
      );
    });
  });

  it("keeps the workout draft card with the draft-producing reply after a non-draft follow-up", async () => {
    const user = userEvent.setup();
    const generatedDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Generated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };
    const generatedMessages = [
      {
        id: 71,
        conversation_id: 41,
        role: "user",
        content: "build a pull workout",
        status: "completed",
        created_at: "2026-03-26T17:00:01Z",
        updated_at: "2026-03-26T17:00:01Z",
        completed_at: "2026-03-26T17:00:01Z",
      },
      {
        id: 72,
        conversation_id: 41,
        role: "assistant",
        content: "I put together a structured workout draft for you.",
        status: "completed",
        created_at: "2026-03-26T17:00:01Z",
        updated_at: "2026-03-26T17:00:02Z",
        completed_at: "2026-03-26T17:00:02Z",
      },
    ];
    const followUpMessages = [
      ...generatedMessages,
      {
        id: 73,
        conversation_id: 41,
        role: "user",
        content: "how long should I rest?",
        status: "completed",
        created_at: "2026-03-26T17:01:01Z",
        updated_at: "2026-03-26T17:01:01Z",
        completed_at: "2026-03-26T17:01:01Z",
      },
      {
        id: 74,
        conversation_id: 41,
        role: "assistant",
        content: "Rest 90 seconds between these working sets.",
        status: "completed",
        created_at: "2026-03-26T17:01:01Z",
        updated_at: "2026-03-26T17:01:02Z",
        completed_at: "2026-03-26T17:01:02Z",
      },
    ];

    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockResolvedValueOnce(
        conversationDetail(generatedMessages, undefined, generatedDraft),
      )
      .mockResolvedValueOnce(
        conversationDetail(followUpMessages, undefined, generatedDraft),
      );
    mockStreamMessage
      .mockImplementationOnce(
        async (
          _conversationId: number,
          _prompt: string,
          options?: {
            onStart?: (event: Record<string, unknown>) => void;
            onDone?: (event: Record<string, unknown>) => void;
          },
        ) => {
          options?.onStart?.({ type: "start", message_id: 72 });
          options?.onDone?.({
            type: "done",
            message_id: 72,
            text: "I put together a structured workout draft for you.",
            workout_draft: generatedDraft,
          });

          return {
            doneEvent: {
              type: "done",
              message_id: 72,
              text: "I put together a structured workout draft for you.",
              workout_draft: generatedDraft,
            },
            endedWithError: false,
          };
        },
      )
      .mockImplementationOnce(
        async (
          _conversationId: number,
          _prompt: string,
          options?: {
            onStart?: (event: Record<string, unknown>) => void;
            onDone?: (event: Record<string, unknown>) => void;
          },
        ) => {
          options?.onStart?.({ type: "start", message_id: 74 });
          options?.onDone?.({
            type: "done",
            message_id: 74,
            text: "Rest 90 seconds between these working sets.",
          });

          return {
            doneEvent: {
              type: "done",
              message_id: 74,
              text: "Rest 90 seconds between these working sets.",
            },
            endedWithError: false,
          };
        },
      );

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "build a pull workout",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(screen.getByTestId("chat-message-72")).toBeInTheDocument();
    });
    expect(
      within(screen.getByTestId("chat-message-72")).getByText(
        "Latest structured workout draft",
      ),
    ).toBeInTheDocument();

    await user.type(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "how long should I rest?",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(
      await screen.findByText("Rest 90 seconds between these working sets."),
    ).toBeInTheDocument();
    expect(
      within(screen.getByTestId("chat-message-72")).getByText(
        "Latest structured workout draft",
      ),
    ).toBeInTheDocument();
    expect(
      within(screen.getByTestId("chat-message-74")).queryByText(
        "Latest structured workout draft",
      ),
    ).not.toBeInTheDocument();
  });
});
