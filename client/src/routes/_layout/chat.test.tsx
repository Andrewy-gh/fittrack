import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  mockSearch,
  mockNavigate,
  mockCreateConversation,
  mockGetConversation,
  mockPollConversation,
  mockStreamMessage,
  mockShowErrorToast,
} = vi.hoisted(() => ({
  mockSearch: { conversationId: '41' as string | undefined },
  mockNavigate: vi.fn(),
  mockCreateConversation: vi.fn(),
  mockGetConversation: vi.fn(),
  mockPollConversation: vi.fn(),
  mockStreamMessage: vi.fn(),
  mockShowErrorToast: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  createFileRoute: () => () => ({
    useRouteContext: () => ({ user: { id: 'user-123' } }),
    useSearch: () => mockSearch,
    fullPath: '/chat',
  }),
  useNavigate: () => mockNavigate,
}));

vi.mock('@/lib/api/ai-chat', () => ({
  createAIChatConversation: mockCreateConversation,
  getAIChatConversation: mockGetConversation,
  pollAIChatConversationUntilSettled: mockPollConversation,
  streamAIChatMessage: mockStreamMessage,
}));

vi.mock('@/lib/errors', () => ({
  getErrorMessage: (error: unknown, fallback = 'An unexpected error occurred') =>
    error instanceof Error ? error.message : fallback,
  showErrorToast: mockShowErrorToast,
}));

import { ChatRouteComponent } from './chat';

function conversationDetail(messages: Array<Record<string, unknown>>) {
  return {
    conversation: {
      id: 41,
      created_at: '2026-03-26T17:00:00Z',
      updated_at: '2026-03-26T17:00:00Z',
    },
    messages,
  };
}

function deferredPromise<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;

  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });

  return { promise, resolve, reject };
}

describe('ChatRouteComponent', () => {
  beforeEach(() => {
    mockSearch.conversationId = '41';
    mockNavigate.mockReset();
    mockCreateConversation.mockReset();
    mockGetConversation.mockReset();
    mockPollConversation.mockReset();
    mockStreamMessage.mockReset();
    mockShowErrorToast.mockReset();
  });

  it('recovers a completed reply when the stream dies before the start event reaches the client', async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(
      conversationDetail([])
    );
    mockStreamMessage.mockRejectedValue(
      new Error('AI chat stream ended before a terminal event')
    );
    mockPollConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 71,
          conversation_id: 41,
          role: 'user',
          content: 'hello',
          status: 'completed',
          created_at: '2026-03-26T17:00:01Z',
          updated_at: '2026-03-26T17:00:01Z',
          completed_at: '2026-03-26T17:00:01Z',
        },
        {
          id: 72,
          conversation_id: 41,
          role: 'assistant',
          content: 'Recovered answer',
          status: 'completed',
          created_at: '2026-03-26T17:00:01Z',
          updated_at: '2026-03-26T17:00:02Z',
          completed_at: '2026-03-26T17:00:02Z',
        },
      ])
    );

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        'Ask about training, recovery, exercise choices, or FitTrack usage...'
      ),
      'hello'
    );
    await user.click(screen.getByRole('button', { name: 'Send' }));

    expect(await screen.findByText('Recovered answer')).toBeInTheDocument();
    expect(screen.getByText('hello')).toBeInTheDocument();
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    );
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });

  it('ignores late initial load results after the conversation is cleared', async () => {
    const initialLoad = deferredPromise<ReturnType<typeof conversationDetail>>();
    mockGetConversation.mockReturnValue(initialLoad.promise);

    const view = render(<ChatRouteComponent />);

    await waitFor(() => {
      expect(mockGetConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByText('No messages yet. Start a new chat or send the first prompt.')
      ).toBeInTheDocument();
    });

    initialLoad.resolve(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: 'assistant',
          content: 'stale reply',
          status: 'streaming',
          created_at: '2026-03-26T17:00:00Z',
          updated_at: '2026-03-26T17:00:01Z',
        },
      ])
    );

    await waitFor(() => {
      expect(screen.queryByText('stale reply')).not.toBeInTheDocument();
    });
    expect(mockPollConversation).not.toHaveBeenCalled();
  });

  it('ignores late recovery results after the conversation is cleared', async () => {
    const recovery = deferredPromise<ReturnType<typeof conversationDetail>>();
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: 'assistant',
          content: 'partial',
          status: 'streaming',
          created_at: '2026-03-26T17:00:00Z',
          updated_at: '2026-03-26T17:00:01Z',
        },
      ])
    );
    mockPollConversation.mockReturnValue(recovery.promise);

    const view = render(<ChatRouteComponent />);

    await waitFor(() => {
      expect(mockPollConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByText('No messages yet. Start a new chat or send the first prompt.')
      ).toBeInTheDocument();
    });

    recovery.resolve(
      conversationDetail([
        {
          id: 71,
          conversation_id: 41,
          role: 'assistant',
          content: 'Recovered answer',
          status: 'completed',
          created_at: '2026-03-26T17:00:01Z',
          updated_at: '2026-03-26T17:00:02Z',
          completed_at: '2026-03-26T17:00:02Z',
        },
      ])
    );

    await waitFor(() => {
      expect(screen.queryByText('Recovered answer')).not.toBeInTheDocument();
    });
  });

  it('does not toast when submit recovery is aborted by clearing the conversation', async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue(
      new Error('AI chat stream ended before a terminal event')
    );
    mockPollConversation.mockImplementation(
      (_conversationId: number, options?: { signal?: AbortSignal }) =>
        new Promise((_resolve, reject) => {
          options?.signal?.addEventListener(
            'abort',
            () => reject(new DOMException('Aborted', 'AbortError')),
            { once: true }
          );
        })
    );

    const view = render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        'Ask about training, recovery, exercise choices, or FitTrack usage...'
      ),
      'hello'
    );
    await user.click(screen.getByRole('button', { name: 'Send' }));

    await waitFor(() => {
      expect(mockPollConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText(
          'Ask about training, recovery, exercise choices, or FitTrack usage...'
        )
      ).toBeEnabled();
    });
    expect(
      screen.getByText('No messages yet. Start a new chat or send the first prompt.')
    ).toBeInTheDocument();
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });
});
