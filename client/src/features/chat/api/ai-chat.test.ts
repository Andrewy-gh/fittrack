import { beforeEach, describe, expect, it, vi } from "vitest";

const { getUser } = vi.hoisted(() => ({
  getUser: vi.fn(),
}));

const { applyLocalDevAuthHeader } = vi.hoisted(() => ({
  applyLocalDevAuthHeader: vi.fn((headers: Headers) => {
    headers.set("x-fittrack-dev-e2e-user", "local-e2e-user");
    return headers;
  }),
}));

vi.mock("@/stack", () => ({
  stackClientApp: {
    getUser,
  },
}));

vi.mock("@/lib/local-dev-auth", () => ({
  applyLocalDevAuthHeader,
}));

import "@/lib/api/client-config";
import {
  createAIChatConversation,
  listAIChatConversations,
  pollAIChatConversationUntilSettled,
  reportAIChatTelemetry,
  resumeAIChatMessageStream,
  requestAIChatMessageRecovery,
  saveAIChatLatestWorkoutDraft,
  streamAIChatMessage,
} from "./ai-chat";

function latestRequest(): Request {
  return vi.mocked(fetch).mock.calls.at(-1)?.[0] as Request;
}

describe("ai chat api wrapper", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    getUser.mockReset();
    applyLocalDevAuthHeader.mockClear();
    getUser.mockResolvedValue({
      getAuthJson: vi.fn().mockResolvedValue({ accessToken: "token-123" }),
    });
  });

  it("falls back to the local dev auth header when Stack has no user", async () => {
    getUser.mockResolvedValue(null);

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          conversation: {
            id: 41,
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:00Z",
          },
          messages: [],
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    await requestAIChatMessageRecovery(41);

    expect(applyLocalDevAuthHeader).toHaveBeenCalledTimes(1);
    expect(fetch).toHaveBeenCalledWith(expect.any(Request));

    const request = latestRequest();
    expect(request.url).toContain("/api/ai/conversations/41/messages/recover");
    expect(request.headers.get("x-fittrack-dev-e2e-user")).toBe(
      "local-e2e-user",
    );
  });

  it("parses JSON preflight errors before SSE parsing", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ message: "runtime unavailable" }), {
        status: 503,
        headers: {
          "Content-Type": "application/json",
        },
      }),
    );

    await expect(streamAIChatMessage(41, "hello")).rejects.toEqual({
      message: "runtime unavailable",
    });
  });

  it("streams start delta done events in order", async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              "event: start",
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              "",
              "event: delta",
              'data: {"type":"delta","delta":"hello "}',
              "",
              "event: done",
              'data: {"type":"done","conversation_id":41,"run_id":51,"message_id":61,"text":"hello world"}',
              "",
              "",
            ].join("\n"),
          ),
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          "Content-Type": "text/event-stream",
        },
      }),
    );

    const seen: string[] = [];
    const result = await streamAIChatMessage(41, "hello", {
      onStart: () => seen.push("start"),
      onDelta: (event) => seen.push(event.delta ?? ""),
      onDone: () => seen.push("done"),
    });

    expect(seen).toEqual(["start", "hello ", "done"]);
    expect(result.endedWithError).toBe(false);
    expect(result.doneEvent?.text).toBe("hello world");
  });

  it("parses workout draft payloads from done events", async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              "event: start",
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              "",
              "event: done",
              'data: {"type":"done","conversation_id":41,"run_id":51,"message_id":61,"text":"I put together a structured workout draft for you.","workout_draft":{"date":"2026-04-20T12:00:00Z","workoutFocus":"pull","exercises":[{"name":"Chest Supported Row","sets":[{"reps":10,"setType":"working"}]}]}}',
              "",
              "",
            ].join("\n"),
          ),
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          "Content-Type": "text/event-stream",
        },
      }),
    );

    const result = await streamAIChatMessage(41, "build me a pull workout");

    expect(result.doneEvent?.workout_draft).toEqual({
      date: "2026-04-20T12:00:00Z",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    });
  });

  it("resumes a chat stream after a sequence cursor", async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              "event: start",
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61,"sequence":3}',
              "",
              "event: delta",
              'data: {"type":"delta","delta":"world","sequence":4}',
              "",
              "event: done",
              'data: {"type":"done","conversation_id":41,"run_id":51,"message_id":61,"text":"hello world","sequence":4,"workout_draft":{"date":"2026-04-20T12:00:00Z","exercises":[{"name":"Goblet Squat","sets":[{"reps":10,"setType":"working"}]}]}}',
              "",
              "",
            ].join("\n"),
          ),
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          "Content-Type": "text/event-stream",
        },
      }),
    );

    const seen: Array<string | number> = [];
    const result = await resumeAIChatMessageStream(41, 51, 3, {
      onStart: (event) => seen.push(event.sequence ?? -1),
      onDelta: (event) => seen.push(event.sequence ?? -1),
      onDone: () => seen.push("done"),
    });

    expect(fetch).toHaveBeenCalledWith(
      "/api/ai/conversations/41/messages/stream/resume?runId=51&afterSequence=3",
      expect.objectContaining({
        method: "GET",
      }),
    );
    expect(seen).toEqual([3, 4, "done"]);
    expect(result.doneEvent?.sequence).toBe(4);
    expect(result.doneEvent?.workout_draft?.date).toBe("2026-04-20T12:00:00Z");
  });

  it("suppresses duplicate SSE chunks by event id", async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              "id: 1",
              "event: start",
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              "",
              "id: 1",
              "event: start",
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              "",
              "id: 2",
              "event: done",
              'data: {"type":"done","conversation_id":41,"run_id":51,"message_id":61,"text":"hello world"}',
              "",
              "",
            ].join("\n"),
          ),
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          "Content-Type": "text/event-stream",
        },
      }),
    );

    const seen: string[] = [];
    await streamAIChatMessage(41, "hello", {
      onStart: () => seen.push("start"),
      onDone: () => seen.push("done"),
    });

    expect(seen).toEqual(["start", "done"]);
  });

  it("fails when the SSE stream ends before a terminal event", async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              "id: 1",
              "event: start",
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              "",
              "",
            ].join("\n"),
          ),
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          "Content-Type": "text/event-stream",
        },
      }),
    );

    await expect(streamAIChatMessage(41, "hello")).rejects.toThrow(
      "AI chat stream ended before a terminal event",
    );
  });

  it("polls persisted conversation state until streaming settles", async () => {
    const fetchSpy = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            conversation: {
              id: 41,
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:00Z",
            },
            messages: [
              {
                id: 61,
                conversation_id: 41,
                role: "assistant",
                content: "partial",
                status: "streaming",
                created_at: "2026-03-26T17:00:00Z",
                updated_at: "2026-03-26T17:00:01Z",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            conversation: {
              id: 41,
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:02Z",
            },
            messages: [
              {
                id: 61,
                conversation_id: 41,
                role: "assistant",
                content: "complete",
                status: "completed",
                created_at: "2026-03-26T17:00:00Z",
                updated_at: "2026-03-26T17:00:02Z",
                completed_at: "2026-03-26T17:00:02Z",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      );

    const detail = await pollAIChatConversationUntilSettled(41, {
      intervalMs: 0,
      timeoutMs: 1000,
    });

    expect(detail.messages[0]?.status).toBe("completed");
    expect(fetchSpy).toHaveBeenCalledTimes(2);
  });

  it("runs the streaming callback before retrying persisted conversation polling", async () => {
    const onStreaming = vi.fn().mockResolvedValue(undefined);
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            conversation: {
              id: 41,
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:00Z",
            },
            messages: [
              {
                id: 61,
                conversation_id: 41,
                role: "assistant",
                content: "partial",
                status: "streaming",
                created_at: "2026-03-26T17:00:00Z",
                updated_at: "2026-03-26T17:00:01Z",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            conversation: {
              id: 41,
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:02Z",
            },
            messages: [
              {
                id: 61,
                conversation_id: 41,
                role: "assistant",
                content: "complete",
                status: "completed",
                created_at: "2026-03-26T17:00:00Z",
                updated_at: "2026-03-26T17:00:02Z",
                completed_at: "2026-03-26T17:00:02Z",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      );

    await pollAIChatConversationUntilSettled(41, {
      intervalMs: 0,
      timeoutMs: 1000,
      onStreaming,
    });

    expect(onStreaming).toHaveBeenCalledTimes(1);
    expect(onStreaming).toHaveBeenCalledWith(
      expect.objectContaining({
        messages: [
          expect.objectContaining({
            status: "streaming",
          }),
        ],
      }),
    );
  });

  it("fetches persisted conversation state through the generated client", async () => {
    const fetchSpy = vi
      .spyOn(globalThis, "fetch")
      .mockImplementation(async (input) => {
        expect((input as Request).url).toContain("/api/ai/conversations/41");

        return new Response(
          JSON.stringify({
            conversation: {
              id: 41,
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:02Z",
            },
            messages: [
              {
                id: 61,
                conversation_id: 41,
                role: "assistant",
                content: "complete",
                status: "completed",
                created_at: "2026-03-26T17:00:00Z",
                updated_at: "2026-03-26T17:00:02Z",
                completed_at: "2026-03-26T17:00:02Z",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      });

    await pollAIChatConversationUntilSettled(41, {
      intervalMs: 0,
      timeoutMs: 10,
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
  });

  it("requests chat recovery for an active conversation", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({ conversation_id: 41, run_id: 61, status: "queued" }),
        {
          status: 202,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const response = await requestAIChatMessageRecovery(41);

    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain(
      "/api/ai/conversations/41/messages/recover",
    );
    expect(latestRequest().method).toBe("POST");
    expect(response).toEqual({
      conversation_id: 41,
      run_id: 61,
      status: "queued",
    });
  });

  it("saves the latest workout draft through the ai chat endpoint", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          workout_id: 901,
          conversation: {
            id: 41,
            latest_workout_draft: {
              date: "2026-04-21T12:00:00Z",
              exercises: [
                {
                  name: "Chest Supported Row",
                  sets: [{ reps: 10, setType: "working" }],
                },
              ],
            },
            latest_workout_draft_status: {
              is_saved: true,
              saved_workout_id: 901,
              saved_at: "2026-04-21T12:30:00Z",
            },
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-04-21T12:30:00Z",
          },
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const conversation = await saveAIChatLatestWorkoutDraft(41);

    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain(
      "/api/ai/conversations/41/latest-workout-draft/save",
    );
    expect(latestRequest().method).toBe("POST");
    expect(conversation.id).toBe(41);
    expect(conversation.latest_workout_draft_status?.is_saved).toBe(true);
    expect(conversation.latest_workout_draft_status?.saved_workout_id).toBe(
      901,
    );
  });

  it("posts ai chat telemetry events to the Go API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, {
        status: 202,
      }),
    );

    await reportAIChatTelemetry({
      category: "stream",
      outcome: "transport_ended_pre_terminal",
      stage: "pre_start",
    });

    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    const request = latestRequest();
    expect(request.url).toContain("/api/ai/chat/telemetry");
    expect(request.method).toBe("POST");
    expect(request.keepalive).toBe(true);
    await expect(request.json()).resolves.toEqual({
      category: "stream",
      outcome: "transport_ended_pre_terminal",
      stage: "pre_start",
    });
  });

  it("returns created conversation JSON", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          id: 41,
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:00Z",
        }),
        {
          status: 201,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const conversation = await createAIChatConversation();

    expect(conversation.id).toBe(41);
    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain("/api/ai/conversations");
    expect(latestRequest().method).toBe("POST");
  });

  it("lists recent conversations through the generated client", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            id: 72,
            title: "Leg day plan",
            created_at: "2026-06-25T17:00:00Z",
            updated_at: "2026-06-25T17:05:00Z",
            last_message_at: "2026-06-25T17:05:00Z",
          },
        ]),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const conversations = await listAIChatConversations();

    expect(conversations).toHaveLength(1);
    expect(conversations[0]).toMatchObject({
      id: 72,
      title: "Leg day plan",
    });
    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain("/api/ai/conversations");
    expect(latestRequest().method).toBe("GET");
  });
});
