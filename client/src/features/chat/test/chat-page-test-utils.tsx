import { vi } from "vitest";
import type {
  AIWorkoutDraft,
  AIWorkoutDraftStatus,
} from "@/features/chat/api/ai-chat";

const mocks = vi.hoisted(() => ({
  mockSearch: {
    conversationId: "41" as string | undefined,
    checkout: undefined as "success" | "cancelled" | undefined,
  },
  mockNavigate: vi.fn(),
  mockCreateConversation: vi.fn(),
  mockGetConversation: vi.fn(),
  mockPollConversation: vi.fn(),
  mockResumeStream: vi.fn(),
  mockReportTelemetry: vi.fn(),
  mockRequestRecovery: vi.fn(),
  mockSaveLatestWorkoutDraft: vi.fn(),
  mockStreamMessage: vi.fn(),
  mockShowErrorToast: vi.fn(),
  mockToastSuccess: vi.fn(),
  mockBillingStatusQueryOptions: vi.fn(),
  mockFeatureAccessQueryOptions: vi.fn(),
  mockCreateBillingCheckoutSession: vi.fn(),
  mockCreateBillingCustomerPortalSession: vi.fn(),
  mockGetBillingStatus: vi.fn(),
  mockRedirectToBillingCheckout: vi.fn(),
  mockRedirectToBillingPortal: vi.fn(),
  mockGetFeatureAccess: vi.fn(),
  mockUseMutation: vi.fn(),
  mockUseQuery: vi.fn(),
  mockRefetchBillingStatus: vi.fn(),
  mockRefetchFeatureAccess: vi.fn(),
  mockBillingQueryResult: { value: undefined as unknown },
  mockFeatureAccessQueryResult: { value: undefined as unknown },
  mockCheckoutAccessQueryResult: { value: undefined as unknown },
}));

export const {
  mockSearch,
  mockNavigate,
  mockCreateConversation,
  mockGetConversation,
  mockPollConversation,
  mockResumeStream,
  mockReportTelemetry,
  mockRequestRecovery,
  mockSaveLatestWorkoutDraft,
  mockStreamMessage,
  mockShowErrorToast,
  mockToastSuccess,
  mockBillingStatusQueryOptions,
  mockFeatureAccessQueryOptions,
  mockCreateBillingCheckoutSession,
  mockCreateBillingCustomerPortalSession,
  mockGetBillingStatus,
  mockRedirectToBillingCheckout,
  mockRedirectToBillingPortal,
  mockGetFeatureAccess,
  mockUseMutation,
  mockUseQuery,
  mockRefetchBillingStatus,
  mockRefetchFeatureAccess,
  mockBillingQueryResult,
  mockFeatureAccessQueryResult,
  mockCheckoutAccessQueryResult,
} = mocks;

vi.mock("@tanstack/react-query", () => ({
  useMutation: mockUseMutation,
  useQuery: mockUseQuery,
}));

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock("@/features/chat/api/ai-chat", () => ({
  createAIChatConversation: mockCreateConversation,
  getAIChatConversation: mockGetConversation,
  pollAIChatConversationUntilSettled: mockPollConversation,
  resumeAIChatMessageStream: mockResumeStream,
  reportAIChatTelemetry: mockReportTelemetry,
  requestAIChatMessageRecovery: mockRequestRecovery,
  saveAIChatLatestWorkoutDraft: mockSaveLatestWorkoutDraft,
  streamAIChatMessage: mockStreamMessage,
}));

vi.mock("@/features/chat/api/billing", () => ({
  billingStatusQueryOptions: mockBillingStatusQueryOptions,
  createBillingCustomerPortalSession: mockCreateBillingCustomerPortalSession,
  createBillingCheckoutSession: mockCreateBillingCheckoutSession,
  getBillingStatus: mockGetBillingStatus,
  redirectToBillingCheckout: mockRedirectToBillingCheckout,
  redirectToBillingPortal: mockRedirectToBillingPortal,
}));

vi.mock("@/features/chat/api/feature-access", () => ({
  featureAccessQueryOptions: mockFeatureAccessQueryOptions,
  getFeatureAccess: mockGetFeatureAccess,
  hasAIChatFeatureAccess: (
    grants?: Array<{
      feature_key: string;
    }>,
  ) => grants?.some((grant) => grant.feature_key === "ai_chatbot") ?? false,
}));

vi.mock("@/lib/errors", () => ({
  getErrorMessage: (
    error: unknown,
    fallback = "An unexpected error occurred",
  ) => (error instanceof Error ? error.message : fallback),
  showErrorToast: mockShowErrorToast,
}));

vi.mock("sonner", () => ({
  toast: {
    success: mockToastSuccess,
  },
}));

const { ChatPage } = await import("@/features/chat/pages/chat-page");

export function ChatRouteComponent() {
  return (
    <ChatPage
      userId="user-123"
      conversationId={parseConversationId(mockSearch.conversationId)}
      conversationIdSearch={mockSearch.conversationId}
      checkout={mockSearch.checkout}
    />
  );
}

export function conversationDetail(
  messages: Array<Record<string, unknown>>,
  activeRun?: Record<string, unknown>,
  latestWorkoutDraft?: AIWorkoutDraft,
  latestWorkoutDraftStatus?: AIWorkoutDraftStatus,
) {
  return {
    conversation: {
      id: 41,
      created_at: "2026-03-26T17:00:00Z",
      updated_at: "2026-03-26T17:00:00Z",
      latest_workout_draft: latestWorkoutDraft,
      latest_workout_draft_status: latestWorkoutDraftStatus,
    },
    messages,
    active_run: activeRun,
  };
}

export function deferredPromise<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;

  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });

  return { promise, resolve, reject };
}

export function resetChatRouteMocks() {
  window.sessionStorage.clear();
  window.localStorage.clear();
  mockSearch.conversationId = "41";
  mockSearch.checkout = undefined;
  mockNavigate.mockReset();
  mockCreateConversation.mockReset();
  mockGetConversation.mockReset();
  mockPollConversation.mockReset();
  mockResumeStream.mockReset();
  mockReportTelemetry.mockReset();
  mockRequestRecovery.mockReset();
  mockSaveLatestWorkoutDraft.mockReset();
  mockStreamMessage.mockReset();
  mockShowErrorToast.mockReset();
  mockToastSuccess.mockReset();
  mockBillingStatusQueryOptions.mockReset();
  mockFeatureAccessQueryOptions.mockReset();
  mockCreateBillingCheckoutSession.mockReset();
  mockCreateBillingCustomerPortalSession.mockReset();
  mockGetBillingStatus.mockReset();
  mockRedirectToBillingCheckout.mockReset();
  mockRedirectToBillingPortal.mockReset();
  mockGetFeatureAccess.mockReset();
  mockUseMutation.mockReset();
  mockUseQuery.mockReset();
  mockRefetchBillingStatus.mockReset();
  mockRefetchFeatureAccess.mockReset();
  mockBillingStatusQueryOptions.mockReturnValue({
    queryKey: ["billing", "ai-chatbot", "status", "user-123"],
    queryFn: vi.fn(),
  });
  mockFeatureAccessQueryOptions.mockReturnValue({
    queryKey: ["feature-access", "user-123"],
    queryFn: vi.fn(),
  });
  mockBillingQueryResult.value = {
    data: {
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_123",
        status: "active",
        cancel_at_period_end: false,
      },
    },
    isLoading: false,
    isPending: false,
    refetch: mockRefetchBillingStatus,
  };
  mockFeatureAccessQueryResult.value = {
    data: [
      {
        feature_key: "ai_chatbot",
      },
    ],
    isLoading: false,
    isPending: false,
    refetch: mockRefetchFeatureAccess,
  };
  mockCheckoutAccessQueryResult.value = {
    data: undefined,
    error: null,
    isFetching: false,
    isError: false,
    isSuccess: false,
  };
  mockUseQuery.mockImplementation((options: { queryKey?: unknown[] }) => {
    if (options.queryKey?.[0] === "feature-access") {
      return mockFeatureAccessQueryResult.value;
    }
    if (options.queryKey?.[2] === "checkout-access") {
      return mockCheckoutAccessQueryResult.value;
    }
    return mockBillingQueryResult.value;
  });
  mockCreateBillingCheckoutSession.mockResolvedValue({
    url: "https://checkout.stripe.test/session",
  });
  mockCreateBillingCustomerPortalSession.mockResolvedValue({
    url: "https://billing.stripe.test/session",
  });
  mockRefetchBillingStatus.mockResolvedValue(mockBillingQueryResult.value);
  mockRefetchFeatureAccess.mockResolvedValue(
    mockFeatureAccessQueryResult.value,
  );
  mockGetBillingStatus.mockResolvedValue(
    (
      mockBillingQueryResult.value as {
        data: unknown;
      }
    ).data,
  );
  mockGetFeatureAccess.mockResolvedValue(
    (
      mockFeatureAccessQueryResult.value as {
        data: unknown;
      }
    ).data,
  );
  mockUseMutation.mockImplementation(
    (options: {
      mutationFn: () => Promise<{ url: string }>;
      onSuccess: (session: { url: string }) => void;
      onError: (error: unknown) => void;
    }) => ({
      isPending: false,
      mutate: () => {
        void options
          .mutationFn()
          .then(options.onSuccess)
          .catch(options.onError);
      },
    }),
  );
  mockReportTelemetry.mockResolvedValue(undefined);
  mockResumeStream.mockResolvedValue({
    doneEvent: {
      type: "done",
      message_id: 61,
      text: "resumed",
    },
    endedWithError: false,
  });
  mockRequestRecovery.mockResolvedValue({
    conversation_id: 41,
    run_id: 61,
    status: "queued",
  });
}

function parseConversationId(value?: string): number | null {
  if (!value) {
    return null;
  }

  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return null;
  }

  return parsed;
}
