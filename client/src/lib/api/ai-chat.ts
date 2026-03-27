import type { ApiError } from '@/lib/errors';
import { stackClientApp } from '@/stack';

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

export type AIChatConversation = {
  id: number;
  title?: string;
  created_at: string;
  updated_at: string;
  last_message_at?: string;
};

export type AIChatMessage = {
  id: number;
  conversation_id: number;
  role: 'user' | 'assistant';
  content: string;
  status: 'streaming' | 'completed' | 'failed';
  error_message?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
};

export type AIChatConversationDetail = {
  conversation: AIChatConversation;
  messages: AIChatMessage[];
};

export type AIChatStreamEvent = {
  type: 'start' | 'delta' | 'done' | 'error';
  request_id?: string;
  conversation_id?: number;
  run_id?: number;
  message_id?: number;
  model?: string;
  delta?: string;
  text?: string;
  message?: string;
};

export type AIChatStreamResult = {
  doneEvent?: AIChatStreamEvent;
  endedWithError: boolean;
};

type ParsedSSEChunk = {
  event: AIChatStreamEvent;
  id?: string;
};

type StreamHandlers = {
  onStart?: (event: AIChatStreamEvent) => void;
  onDelta?: (event: AIChatStreamEvent) => void;
  onDone?: (event: AIChatStreamEvent) => void;
  onErrorEvent?: (event: AIChatStreamEvent) => void;
  signal?: AbortSignal;
};

type ConversationPollOptions = {
  signal?: AbortSignal;
  intervalMs?: number;
  timeoutMs?: number;
};

type ConversationRequestOptions = {
  signal?: AbortSignal;
};

const defaultConversationPollIntervalMs = 1000;
const defaultConversationPollTimeoutMs = 55000;

async function getAuthHeaders(contentType = false): Promise<Headers> {
  const headers = new Headers();
  if (contentType) {
    headers.set('Content-Type', 'application/json');
  }

  if (!stackClientApp) {
    return headers;
  }

  const user = await stackClientApp.getUser();
  if (!user) {
    return headers;
  }

  const { accessToken } = await user.getAuthJson();
  if (accessToken) {
    headers.set('x-stack-access-token', accessToken);
  }

  return headers;
}

async function readApiError(response: Response): Promise<ApiError> {
  try {
    return (await response.json()) as ApiError;
  } catch {
    return {
      message: `${response.status} ${response.statusText}`,
    };
  }
}

export async function createAIChatConversation(): Promise<AIChatConversation> {
  const response = await fetch(`${BASE_URL}/ai/conversations`, {
    method: 'POST',
    headers: await getAuthHeaders(),
  });

  if (!response.ok) {
    throw await readApiError(response);
  }

  return (await response.json()) as AIChatConversation;
}

export async function getAIChatConversation(
  conversationId: number,
  options: ConversationRequestOptions = {}
): Promise<AIChatConversationDetail> {
  const response = await fetch(`${BASE_URL}/ai/conversations/${conversationId}`, {
    method: 'GET',
    headers: await getAuthHeaders(),
    signal: options.signal,
  });

  if (!response.ok) {
    throw await readApiError(response);
  }

  return (await response.json()) as AIChatConversationDetail;
}

export async function streamAIChatMessage(
  conversationId: number,
  prompt: string,
  handlers: StreamHandlers = {}
): Promise<AIChatStreamResult> {
  const response = await fetch(
    `${BASE_URL}/ai/conversations/${conversationId}/messages/stream`,
    {
      method: 'POST',
      headers: await getAuthHeaders(true),
      body: JSON.stringify({ prompt }),
      signal: handlers.signal,
    }
  );

  if (!response.ok) {
    throw await readApiError(response);
  }

  if (!response.body) {
    throw new Error('No body in SSE response');
  }

  const reader = response.body.pipeThrough(new TextDecoderStream()).getReader();
  let buffer = '';
  let lastEventId = '';
  let sawTerminalEvent = false;

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }

      buffer += value;
      const chunks = buffer.split('\n\n');
      buffer = chunks.pop() ?? '';

      for (const chunk of chunks) {
        const parsed = parseSSEChunk(chunk);
        if (!parsed) {
          continue;
        }
        if (parsed.id && parsed.id === lastEventId) {
          continue;
        }
        lastEventId = parsed.id ?? lastEventId;

        const event = parsed.event;

        switch (event.type) {
          case 'start':
            handlers.onStart?.(event);
            break;
          case 'delta':
            handlers.onDelta?.(event);
            break;
          case 'done':
            sawTerminalEvent = true;
            handlers.onDone?.(event);
            return { doneEvent: event, endedWithError: false };
          case 'error':
            sawTerminalEvent = true;
            handlers.onErrorEvent?.(event);
            return { doneEvent: event, endedWithError: true };
          default:
            break;
        }
      }
    }
  } finally {
    reader.releaseLock();
  }

  if (!sawTerminalEvent) {
    throw new Error('AI chat stream ended before a terminal event');
  }

  return { endedWithError: false };
}

export async function pollAIChatConversationUntilSettled(
  conversationId: number,
  options: ConversationPollOptions = {}
): Promise<AIChatConversationDetail> {
  const intervalMs = options.intervalMs ?? defaultConversationPollIntervalMs;
  const timeoutMs = options.timeoutMs ?? defaultConversationPollTimeoutMs;
  const deadline = Date.now() + timeoutMs;

  while (true) {
    throwIfAborted(options.signal);
    const detail = await getAIChatConversation(conversationId, {
      signal: options.signal,
    });
    throwIfAborted(options.signal);
    if (!detail.messages.some((message) => message.status === 'streaming')) {
      return detail;
    }
    if (Date.now() >= deadline) {
      throw new Error('AI chat recovery timed out while waiting for persisted conversation state');
    }
    await delay(intervalMs, options.signal);
  }
}

function parseSSEChunk(chunk: string): ParsedSSEChunk | null {
  const lines = chunk.split('\n');
  const dataLines: string[] = [];
  let id = '';

  for (const line of lines) {
    if (line.startsWith('id:')) {
      id = line.replace(/^id:\s*/, '').trim();
    }
    if (line.startsWith('data:')) {
      dataLines.push(line.replace(/^data:\s*/, ''));
    }
  }

  if (dataLines.length === 0) {
    return null;
  }

  return {
    event: JSON.parse(dataLines.join('\n')) as AIChatStreamEvent,
    id,
  };
}

function throwIfAborted(signal?: AbortSignal): void {
  if (signal?.aborted) {
    throw signal.reason instanceof Error
      ? signal.reason
      : new DOMException('Aborted', 'AbortError');
  }
}

function delay(ms: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => {
      cleanup();
      resolve();
    }, ms);

    const onAbort = () => {
      cleanup();
      reject(
        signal?.reason instanceof Error
          ? signal.reason
          : new DOMException('Aborted', 'AbortError')
      );
    };

    const cleanup = () => {
      window.clearTimeout(timer);
      signal?.removeEventListener('abort', onAbort);
    };

    signal?.addEventListener('abort', onAbort, { once: true });
  });
}
