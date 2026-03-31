import { beforeEach, describe, expect, it, vi } from 'vitest';

const { getUser } = vi.hoisted(() => ({
  getUser: vi.fn(),
}));

vi.mock('@/stack', () => ({
  stackClientApp: {
    getUser,
  },
}));

import {
  createAIChatConversation,
  pollAIChatConversationUntilSettled,
  reportAIChatTelemetry,
  streamAIChatMessage,
} from './ai-chat';

describe('ai chat api wrapper', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    getUser.mockReset();
    getUser.mockResolvedValue({
      getAuthJson: vi.fn().mockResolvedValue({ accessToken: 'token-123' }),
    });
  });

  it('parses JSON preflight errors before SSE parsing', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ message: 'runtime unavailable' }), {
        status: 503,
        headers: {
          'Content-Type': 'application/json',
        },
      })
    );

    await expect(streamAIChatMessage(41, 'hello')).rejects.toEqual({
      message: 'runtime unavailable',
    });
  });

  it('streams start delta done events in order', async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              'event: start',
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              '',
              'event: delta',
              'data: {"type":"delta","delta":"hello "}',
              '',
              'event: done',
              'data: {"type":"done","conversation_id":41,"run_id":51,"message_id":61,"text":"hello world"}',
              '',
              '',
            ].join('\n')
          )
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
        },
      })
    );

    const seen: string[] = [];
    const result = await streamAIChatMessage(41, 'hello', {
      onStart: () => seen.push('start'),
      onDelta: (event) => seen.push(event.delta ?? ''),
      onDone: () => seen.push('done'),
    });

    expect(seen).toEqual(['start', 'hello ', 'done']);
    expect(result.endedWithError).toBe(false);
    expect(result.doneEvent?.text).toBe('hello world');
  });

  it('suppresses duplicate SSE chunks by event id', async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              'id: 1',
              'event: start',
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              '',
              'id: 1',
              'event: start',
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              '',
              'id: 2',
              'event: done',
              'data: {"type":"done","conversation_id":41,"run_id":51,"message_id":61,"text":"hello world"}',
              '',
              '',
            ].join('\n')
          )
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
        },
      })
    );

    const seen: string[] = [];
    await streamAIChatMessage(41, 'hello', {
      onStart: () => seen.push('start'),
      onDone: () => seen.push('done'),
    });

    expect(seen).toEqual(['start', 'done']);
  });

  it('fails when the SSE stream ends before a terminal event', async () => {
    const stream = new ReadableStream({
      start(controller) {
        controller.enqueue(
          new TextEncoder().encode(
            [
              'id: 1',
              'event: start',
              'data: {"type":"start","conversation_id":41,"run_id":51,"message_id":61}',
              '',
              '',
            ].join('\n')
          )
        );
        controller.close();
      },
    });

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(stream, {
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
        },
      })
    );

    await expect(streamAIChatMessage(41, 'hello')).rejects.toThrow(
      'AI chat stream ended before a terminal event'
    );
  });

  it('polls persisted conversation state until streaming settles', async () => {
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            conversation: { id: 41, created_at: '2026-03-26T17:00:00Z', updated_at: '2026-03-26T17:00:00Z' },
            messages: [{ id: 61, conversation_id: 41, role: 'assistant', content: 'partial', status: 'streaming', created_at: '2026-03-26T17:00:00Z', updated_at: '2026-03-26T17:00:01Z' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        )
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            conversation: { id: 41, created_at: '2026-03-26T17:00:00Z', updated_at: '2026-03-26T17:00:02Z' },
            messages: [{ id: 61, conversation_id: 41, role: 'assistant', content: 'complete', status: 'completed', created_at: '2026-03-26T17:00:00Z', updated_at: '2026-03-26T17:00:02Z', completed_at: '2026-03-26T17:00:02Z' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        )
      );

    const detail = await pollAIChatConversationUntilSettled(41, {
      intervalMs: 0,
      timeoutMs: 10,
    });

    expect(detail.messages[0]?.status).toBe('completed');
    expect(fetchSpy).toHaveBeenCalledTimes(2);
  });

  it('passes abort signals through persisted conversation polling fetches', async () => {
    const controller = new AbortController();
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementation(async (_input, init) => {
        expect(init?.signal).toBe(controller.signal);

        return new Response(
          JSON.stringify({
            conversation: { id: 41, created_at: '2026-03-26T17:00:00Z', updated_at: '2026-03-26T17:00:02Z' },
            messages: [{ id: 61, conversation_id: 41, role: 'assistant', content: 'complete', status: 'completed', created_at: '2026-03-26T17:00:00Z', updated_at: '2026-03-26T17:00:02Z', completed_at: '2026-03-26T17:00:02Z' }],
          }),
          { status: 200, headers: { 'Content-Type': 'application/json' } }
        );
      });

    await pollAIChatConversationUntilSettled(41, {
      signal: controller.signal,
      intervalMs: 0,
      timeoutMs: 10,
    });

    expect(fetchSpy).toHaveBeenCalledTimes(1);
  });

  it('returns created conversation JSON', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          id: 41,
          created_at: '2026-03-26T17:00:00Z',
          updated_at: '2026-03-26T17:00:00Z',
        }),
        {
          status: 201,
          headers: {
            'Content-Type': 'application/json',
          },
        }
      )
    );

    const conversation = await createAIChatConversation();

    expect(conversation.id).toBe(41);
    expect(fetch).toHaveBeenCalledWith(
      '/api/ai/conversations',
      expect.objectContaining({
        method: 'POST',
      })
    );
  });

  it('posts ai chat telemetry events to the Go API', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(null, {
        status: 202,
      })
    );

    await reportAIChatTelemetry({
      category: 'stream',
      outcome: 'transport_ended_pre_terminal',
      stage: 'pre_start',
    });

    expect(fetch).toHaveBeenCalledWith(
      '/api/ai/chat/telemetry',
      expect.objectContaining({
        method: 'POST',
        keepalive: true,
        body: JSON.stringify({
          category: 'stream',
          outcome: 'transport_ended_pre_terminal',
          stage: 'pre_start',
        }),
      })
    );
  });
});
