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
});
