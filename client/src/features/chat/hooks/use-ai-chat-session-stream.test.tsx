import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { client } from "@/client/client.gen";
import type {
  AIChatConversationDetail,
  AIChatStreamEvent,
} from "@/features/chat/api/ai-chat";
import { useAIChatSession } from "./use-ai-chat-session";

const timestamp = "2026-04-21T12:00:00Z";
const draft = {
  date: timestamp,
  workoutFocus: "Pull",
  exercises: [
    { name: "Row", sets: [{ reps: 10, setType: "working" as const }] },
  ],
};

describe("chat stream state before authoritative reload", () => {
  beforeEach(() => {
    window.sessionStorage.clear();
    client.setConfig({ baseUrl: "http://localhost/api" });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    client.setConfig({ baseUrl: "/api" });
  });

  // Existing page tests supply the completed conversation on reload. Keep that
  // request pending here so it cannot hide a lost stream update or draft marker.
  it.each(["submit", "resume"] as const)(
    "%s applies parsed deltas, cursors and a completed workout draft before reload",
    async (mode) => {
      const detail: AIChatConversationDetail = {
        conversation: {
          id: 41,
          created_at: timestamp,
          updated_at: timestamp,
          latest_workout_draft_status: { is_saved: true, saved_workout_id: 12 },
        },
        messages:
          mode === "resume"
            ? [
                {
                  id: 72,
                  conversation_id: 41,
                  role: "assistant",
                  content: "Prior ",
                  status: "streaming",
                  created_at: timestamp,
                  updated_at: timestamp,
                },
              ]
            : [],
      };
      if (mode === "resume") {
        detail.active_run = {
          id: 91,
          assistant_message_id: 72,
          status: "streaming",
          latest_sequence: 1,
        };
      }
      let writer: ReadableStreamDefaultController<Uint8Array> | undefined;
      const body = new ReadableStream<Uint8Array>({
        start(controller) {
          writer = controller;
        },
      });
      let detailRequests = 0;
      let releaseReload: ((response: Response) => void) | undefined;
      const reload = new Promise<Response>((resolve) => {
        releaseReload = resolve;
      });
      const requests: string[] = [];
      vi.stubGlobal("fetch", async (input: RequestInfo | URL) => {
        const url = new URL(
          input instanceof Request ? input.url : String(input),
          "http://localhost",
        );
        requests.push(url.href);
        if (url.pathname.includes("/messages/stream")) {
          return new Response(body, {
            headers: { "Content-Type": "text/event-stream" },
          });
        }
        if (url.pathname === "/api/ai/conversations/41") {
          detailRequests += 1;
          return detailRequests === 1 ? Response.json(detail) : reload;
        }
        if (url.pathname === "/api/ai/chat/telemetry")
          return new Response(null, { status: 204 });
        throw new Error(`Unexpected request: ${url.pathname}`);
      });
      const started: number[] = [];
      const { result } = renderHook(() =>
        useAIChatSession({
          conversationId: 41,
          initialPrompt: "",
          onPromptChange() {},
          onPromptStarted(id) {
            started.push(id);
          },
          onNewConversationCreated() {},
          async onConversationCreated() {},
        }),
      );
      await waitFor(() =>
        expect(result.current.isLoadingConversation).toBe(false),
      );
      let submitted: Promise<void> | undefined;
      if (mode === "submit") {
        act(() => {
          submitted = result.current.submitPromptValue("Build a workout");
        });
      }
      await waitFor(() =>
        expect(requests.some((url) => url.includes("/messages/stream"))).toBe(
          true,
        ),
      );
      const send = (event: AIChatStreamEvent) => {
        if (!writer) throw new Error("Stream writer not initialized");
        writer.enqueue(
          new TextEncoder().encode(`data: ${JSON.stringify(event)}\n\n`),
        );
      };
      await act(async () => {
        send({ type: "start", message_id: 72, run_id: 91, sequence: 1 });
        // The real parser omits delta identity. Submit retains its zero
        // fallback while resume keeps the loaded run identity.
        send({ type: "delta", delta: "Partial", sequence: 2 });
      });
      await waitFor(() =>
        expect(result.current.messages.at(-1)?.content).toBe(
          mode === "resume" ? "Prior Partial" : "Partial",
        ),
      );
      expect(
        JSON.parse(
          window.sessionStorage.getItem("fittrack.ai-chat.resume:41") ?? "null",
        ),
      ).toEqual({
        runId: mode === "resume" ? 91 : 0,
        sequence: 2,
        assistantMessageId: 72,
      });
      expect(started).toEqual(mode === "submit" ? [41] : []);
      await act(async () => {
        send({
          type: "done",
          message_id: 73,
          text: "Final workout",
          workout_draft: draft,
        });
      });
      await waitFor(() => expect(detailRequests).toBe(2));
      expect(result.current.messages.at(-1)).toMatchObject({
        id: 73,
        content: "Final workout",
        status: "completed",
      });
      expect(result.current.conversation?.latest_workout_draft).toEqual(draft);
      expect(result.current.conversation?.latest_workout_draft_status).toEqual({
        is_saved: false,
      });
      expect(result.current.latestWorkoutDraftMessageId).toBe(73);
      expect(
        window.sessionStorage.getItem("fittrack.ai-chat.resume:41"),
      ).toBeNull();
      await act(async () => {
        releaseReload?.(
          Response.json({ ...detail, messages: [], active_run: undefined }),
        );
        await submitted;
      });
      await waitFor(() => expect(result.current.isSubmitting).toBe(false));
    },
  );
});
