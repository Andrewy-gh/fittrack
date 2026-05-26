import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AIWorkoutDraft, AIWorkoutDraftStatus } from "@/lib/api/ai-chat";

const {
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
} = vi.hoisted(() => ({
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

vi.mock("@tanstack/react-query", () => ({
  useMutation: mockUseMutation,
  useQuery: mockUseQuery,
}));

vi.mock("@tanstack/react-router", () => ({
  createFileRoute: () => () => ({
    useRouteContext: () => ({ user: { id: "user-123" } }),
    useSearch: () => mockSearch,
    fullPath: "/chat",
  }),
  useNavigate: () => mockNavigate,
}));

vi.mock("@/lib/api/ai-chat", () => ({
  createAIChatConversation: mockCreateConversation,
  getAIChatConversation: mockGetConversation,
  pollAIChatConversationUntilSettled: mockPollConversation,
  resumeAIChatMessageStream: mockResumeStream,
  reportAIChatTelemetry: mockReportTelemetry,
  requestAIChatMessageRecovery: mockRequestRecovery,
  saveAIChatLatestWorkoutDraft: mockSaveLatestWorkoutDraft,
  streamAIChatMessage: mockStreamMessage,
}));

vi.mock("@/lib/api/billing", () => ({
  billingStatusQueryOptions: mockBillingStatusQueryOptions,
  createBillingCustomerPortalSession: mockCreateBillingCustomerPortalSession,
  createBillingCheckoutSession: mockCreateBillingCheckoutSession,
  getBillingStatus: mockGetBillingStatus,
  redirectToBillingCheckout: mockRedirectToBillingCheckout,
  redirectToBillingPortal: mockRedirectToBillingPortal,
}));

vi.mock("@/lib/api/feature-access", () => ({
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

import { ChatRouteComponent } from "./chat";

function conversationDetail(
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

function deferredPromise<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;

  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });

  return { promise, resolve, reject };
}

describe("ChatRouteComponent", () => {
  beforeEach(() => {
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
      queryKey: ["billing", "ai-chatbot", "status"],
      queryFn: vi.fn(),
    });
    mockFeatureAccessQueryOptions.mockReturnValue({
      queryKey: ["feature-access"],
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
  });

  it("shows checkout success and cleans the query state", async () => {
    mockSearch.checkout = "success";
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "Checkout complete. We are refreshing your AI chat access.",
      ),
    ).toBeInTheDocument();
    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/chat",
      search: { conversationId: "41" },
      replace: true,
    });
  });

  it("uses checkout access polling results when the webhook grant appears", async () => {
    mockSearch.checkout = "success";
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockCheckoutAccessQueryResult.value = {
      data: {
        billingStatus: {
          feature_key: "ai_chatbot",
          has_access: true,
          subscription: {
            stripe_subscription_id: "sub_123",
            status: "active",
            cancel_at_period_end: false,
          },
        },
        featureAccess: [{ feature_key: "ai_chatbot" }],
      },
      isFetching: false,
      isError: false,
      isSuccess: true,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeEnabled();
    expect(screen.getByText("Premium")).toBeInTheDocument();
    expect(mockRefetchFeatureAccess).not.toHaveBeenCalled();
    expect(mockRefetchBillingStatus).not.toHaveBeenCalled();
  });

  it("shows an activation recovery action when checkout polling exhausts retries after billing becomes active", async () => {
    const user = userEvent.setup();
    mockSearch.checkout = "success";
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockCheckoutAccessQueryResult.value = {
      data: undefined,
      isFetching: false,
      isError: true,
      isSuccess: false,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("AI chat activation is still finishing."),
    ).toBeInTheDocument();
    expect(screen.getByText("Activating")).toBeInTheDocument();
    expect(
      screen.queryByText("Checking your AI chat access..."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(mockRefetchBillingStatus).toHaveBeenCalledTimes(1);
    expect(mockRefetchFeatureAccess).toHaveBeenCalledTimes(1);
  });

  it("does not offer Checkout when billing status is active but the feature grant has not refreshed", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: true,
        subscription: {
          stripe_subscription_id: "sub_active",
          status: "active",
          cancel_at_period_end: false,
        },
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Activating")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeDisabled();
    expect(screen.getByText("AI chat activation is still finishing."));
  });

  it("starts Checkout from the no-access trial CTA", async () => {
    const user = userEvent.setup();
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    await user.click(
      await screen.findByRole("button", { name: "Start 7-day trial" }),
    );

    await waitFor(() => {
      expect(mockCreateBillingCheckoutSession).toHaveBeenCalledTimes(1);
    });
    expect(mockRedirectToBillingCheckout).toHaveBeenCalledWith(
      "https://checkout.stripe.test/session",
    );
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeDisabled();
  });

  it("keeps the trial CTA hidden while feature access is still loading", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: undefined,
      isLoading: true,
      isPending: true,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("Checking your AI chat access..."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeDisabled();
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });

  it("keeps chat enabled when feature access is granted without a Stripe subscription", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
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
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeEnabled();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("Access active")).toBeInTheDocument();
  });

  it("keeps chat enabled when a manual feature grant overrides blocked Stripe billing", async () => {
    const user = userEvent.setup();
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
        subscription: {
          stripe_subscription_id: "sub_past_due",
          status: "past_due",
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
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeEnabled();
    expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
    expect(screen.getByText("Access active")).toBeInTheDocument();
    expect(
      screen.getByText(
        "AI chat access is active for this account. Billing still needs attention.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(
        "AI chat is paused until the payment issue is resolved.",
      ),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText("Start or restore premium access to use AI chat."),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Update billing" }));

    await waitFor(() => {
      expect(mockCreateBillingCustomerPortalSession).toHaveBeenCalledTimes(1);
    });
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });

  it("recovers a completed reply when the stream dies before the start event reaches the client", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockPollConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 71,
          conversation_id: 41,
          role: "user",
          content: "hello",
          status: "completed",
          created_at: "2026-03-26T17:00:01Z",
          updated_at: "2026-03-26T17:00:01Z",
          completed_at: "2026-03-26T17:00:01Z",
        },
        {
          id: 72,
          conversation_id: 41,
          role: "assistant",
          content: "Recovered answer",
          status: "completed",
          created_at: "2026-03-26T17:00:01Z",
          updated_at: "2026-03-26T17:00:02Z",
          completed_at: "2026-03-26T17:00:02Z",
        },
      ]),
    );

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "stream",
      outcome: "transport_ended_pre_terminal",
      stage: "pre_start",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_completed",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_suppressed_due_to_successful_recovery",
    });
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });

  it("does not treat preflight api failures as transport interruptions", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue({
      message: "ai chat runtime is not configured",
      request_id: "req-123",
    });

    render(<ChatRouteComponent />);

    const promptBox = await screen.findByPlaceholderText(
      "Ask about training, recovery, exercise choices, or FitTrack usage...",
    );
    await user.type(promptBox, "hello");
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockShowErrorToast).toHaveBeenCalledWith(
        expect.objectContaining({
          message: "ai chat runtime is not configured",
        }),
        "Failed to stream AI chat response",
      );
    });
    expect(mockRequestRecovery).not.toHaveBeenCalled();
    expect(mockPollConversation).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "stream",
      outcome: "server_error",
      stage: "pre_start",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_shown",
    });
    expect(
      screen.getByText(
        "No messages yet. Start a new chat or send the first prompt.",
      ),
    ).toBeInTheDocument();
    expect(promptBox).toHaveValue("hello");
  });

  it("keeps the prompt visible and shows the recovery failure when submit recovery fails before stream start", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockRequestRecovery.mockRejectedValue(
      new Error("ai chat recovery is not configured"),
    );

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(await screen.findByText("hello")).toBeInTheDocument();
    expect(
      await screen.findByText("ai chat recovery is not configured"),
    ).toBeInTheDocument();
    expect(mockShowErrorToast).toHaveBeenCalledWith(
      expect.objectContaining({
        message: "ai chat recovery is not configured",
      }),
      "Failed to stream AI chat response",
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_failed",
    });
  });

  it("ignores late initial load results after the conversation is cleared", async () => {
    const initialLoad =
      deferredPromise<ReturnType<typeof conversationDetail>>();
    mockGetConversation.mockReturnValue(initialLoad.promise);

    const view = render(<ChatRouteComponent />);

    await waitFor(() => {
      expect(mockGetConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByText(
          "No messages yet. Start a new chat or send the first prompt.",
        ),
      ).toBeInTheDocument();
    });

    initialLoad.resolve(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "stale reply",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
    );

    await waitFor(() => {
      expect(screen.queryByText("stale reply")).not.toBeInTheDocument();
    });
    expect(mockPollConversation).not.toHaveBeenCalled();
  });

  it("ignores late recovery results after the conversation is cleared", async () => {
    const recovery = deferredPromise<ReturnType<typeof conversationDetail>>();
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "partial",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
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
        screen.getByText(
          "No messages yet. Start a new chat or send the first prompt.",
        ),
      ).toBeInTheDocument();
    });

    recovery.resolve(
      conversationDetail([
        {
          id: 71,
          conversation_id: 41,
          role: "assistant",
          content: "Recovered answer",
          status: "completed",
          created_at: "2026-03-26T17:00:01Z",
          updated_at: "2026-03-26T17:00:02Z",
          completed_at: "2026-03-26T17:00:02Z",
        },
      ]),
    );

    await waitFor(() => {
      expect(screen.queryByText("Recovered answer")).not.toBeInTheDocument();
    });
  });

  it("does not toast when submit recovery is aborted by clearing the conversation", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockPollConversation.mockImplementation(
      (_conversationId: number, options?: { signal?: AbortSignal }) =>
        new Promise((_resolve, reject) => {
          options?.signal?.addEventListener(
            "abort",
            () => reject(new DOMException("Aborted", "AbortError")),
            { once: true },
          );
        }),
    );

    const view = render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockPollConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText(
          "Ask about training, recovery, exercise choices, or FitTrack usage...",
        ),
      ).toBeEnabled();
    });
    expect(
      screen.getByText(
        "No messages yet. Start a new chat or send the first prompt.",
      ),
    ).toBeInTheDocument();
    expect(mockShowErrorToast).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovery_aborted",
    });
  });

  it("stops retrying load-triggered recovery after the handoff is queued", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "partial",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
    );
    mockRequestRecovery
      .mockResolvedValueOnce({
        conversation_id: 41,
        status: "not_needed",
      })
      .mockResolvedValueOnce({
        conversation_id: 41,
        run_id: 61,
        status: "queued",
      });
    mockPollConversation.mockImplementation(
      async (
        _conversationId: number,
        options?: {
          onStreaming?: (
            detail: ReturnType<typeof conversationDetail>,
          ) => Promise<void> | void;
        },
      ) => {
        await options?.onStreaming?.(
          conversationDetail([
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: "partial",
              status: "streaming",
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:01Z",
            },
          ]),
        );
        await options?.onStreaming?.(
          conversationDetail([
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: "partial",
              status: "streaming",
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:01Z",
            },
          ]),
        );

        return conversationDetail([
          {
            id: 61,
            conversation_id: 41,
            role: "assistant",
            content: "Recovered answer",
            status: "completed",
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]);
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(mockRequestRecovery).toHaveBeenCalledTimes(2);
    expect(mockRequestRecovery).toHaveBeenNthCalledWith(
      1,
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockRequestRecovery).toHaveBeenNthCalledWith(
      2,
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockShowErrorToast).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_completed",
    });
  });

  it("resumes an active run from the last seen sequence without duplicating replayed text", async () => {
    window.sessionStorage.setItem(
      "fittrack.ai-chat.resume:41",
      JSON.stringify({
        runId: 91,
        sequence: 3,
        assistantMessageId: 61,
      }),
    );
    mockGetConversation
      .mockResolvedValueOnce(
        conversationDetail(
          [
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: "hello",
              status: "streaming",
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:01Z",
            },
          ],
          {
            id: 91,
            assistant_message_id: 61,
            status: "streaming",
            latest_sequence: 3,
          },
        ),
      )
      .mockResolvedValueOnce(
        conversationDetail([
          {
            id: 61,
            conversation_id: 41,
            role: "assistant",
            content: "hello world",
            status: "completed",
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
      );
    mockResumeStream.mockImplementation(
      async (
        conversationId: number,
        runId: number,
        afterSequence: number,
        options?: {
          onDelta?: (event: Record<string, unknown>) => void;
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        expect(conversationId).toBe(41);
        expect(runId).toBe(91);
        expect(afterSequence).toBe(3);
        options?.onDelta?.({
          type: "delta",
          delta: " world",
          sequence: 4,
        });
        options?.onDone?.({
          type: "done",
          message_id: 61,
          text: "hello world",
          sequence: 4,
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 61,
            text: "hello world",
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("hello world")).toBeInTheDocument();
    expect(screen.queryByText("hello world world")).not.toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledWith(
      41,
      91,
      3,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockRequestRecovery).not.toHaveBeenCalled();
    expect(mockPollConversation).not.toHaveBeenCalled();
  });

  it("falls back to recovery polling when resume is unavailable", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail(
        [
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
        {
          id: 91,
          assistant_message_id: 61,
          status: "streaming",
          latest_sequence: 1,
        },
      ),
    );
    mockResumeStream.mockRejectedValue({ message: "resume unavailable" });
    mockPollConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "Recovered answer",
          status: "completed",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:02Z",
          completed_at: "2026-03-26T17:00:02Z",
        },
      ]),
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledTimes(1);
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
  });

  it("falls back to recovery polling when the resume stream exits before a terminal event", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail(
        [
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
        {
          id: 91,
          assistant_message_id: 61,
          status: "streaming",
          latest_sequence: 1,
        },
      ),
    );
    mockResumeStream.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockPollConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "Recovered answer",
          status: "completed",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:02Z",
          completed_at: "2026-03-26T17:00:02Z",
        },
      ]),
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledTimes(1);
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_completed",
    });
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });

  it("finishes reconnect without recovery when the resume stream returns the completed reply", async () => {
    mockGetConversation
      .mockResolvedValueOnce(
        conversationDetail(
          [
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
          {
            id: 91,
            assistant_message_id: 61,
            status: "streaming",
            latest_sequence: 1,
          },
        ),
      )
      .mockResolvedValueOnce(
        conversationDetail([
          {
            id: 61,
            conversation_id: 41,
            role: "assistant",
            content: "Completed answer",
            status: "completed",
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
      );
    mockResumeStream.mockImplementation(
      async (
        _conversationId: number,
        _runId: number,
        _afterSequence: number,
        options?: {
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onDone?.({
          type: "done",
          message_id: 61,
          text: "Completed answer",
          sequence: 2,
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 61,
            text: "Completed answer",
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Completed answer")).toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledTimes(1);
    expect(mockRequestRecovery).not.toHaveBeenCalled();
    expect(mockPollConversation).not.toHaveBeenCalled();
  });

  it("shows a user-visible failure when load-triggered recovery times out", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "partial",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
    );
    mockPollConversation.mockRejectedValue(
      new Error(
        "AI chat recovery timed out while waiting for persisted conversation state",
      ),
    );

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "AI chat recovery timed out while waiting for persisted conversation state",
      ),
    ).toBeInTheDocument();
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovery_timeout",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_shown",
    });
    expect(mockShowErrorToast).toHaveBeenCalledWith(
      expect.objectContaining({
        message:
          "AI chat recovery timed out while waiting for persisted conversation state",
      }),
      "Failed to recover AI chat conversation",
    );
  });

  it("does not reclassify a completed stream when the follow-up refresh fails", async () => {
    const user = userEvent.setup();
    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockRejectedValueOnce(new Error("refresh failed"));
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
          text: "Completed answer",
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 72,
            text: "Completed answer",
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
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(await screen.findByText("refresh failed")).toBeInTheDocument();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "stream",
      outcome: "completed",
      stage: "terminal",
    });
    expect(mockReportTelemetry).not.toHaveBeenCalledWith(
      expect.objectContaining({
        category: "stream",
        outcome: "transport_ended_pre_terminal",
      }),
    );
    expect(mockPollConversation).not.toHaveBeenCalled();
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });

  it("refreshes billing status after a submitted prompt so trial usage updates", async () => {
    const user = userEvent.setup();
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: true,
        subscription: {
          stripe_subscription_id: "sub_trial",
          status: "trialing",
          cancel_at_period_end: false,
        },
        trial_usage: {
          used: 1,
          limit: 30,
        },
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockResolvedValueOnce(
        conversationDetail([
          {
            id: 71,
            conversation_id: 41,
            role: "user",
            content: "hello",
            status: "completed",
            created_at: "2026-03-26T17:00:01Z",
            updated_at: "2026-03-26T17:00:01Z",
            completed_at: "2026-03-26T17:00:01Z",
          },
          {
            id: 72,
            conversation_id: 41,
            role: "assistant",
            content: "Completed answer",
            status: "completed",
            created_at: "2026-03-26T17:00:01Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
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
          text: "Completed answer",
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 72,
            text: "Completed answer",
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("1 of 30 trial prompts used"),
    ).toBeInTheDocument();
    await user.type(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockRefetchBillingStatus).toHaveBeenCalledTimes(1);
    });
    expect(await screen.findByText("Completed answer")).toBeInTheDocument();
  });

  it("keeps the active stream running when new chat creation fails", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    let streamSignal: AbortSignal | undefined;
    mockStreamMessage.mockImplementation(
      (
        _conversationId: number,
        _prompt: string,
        options?: { signal?: AbortSignal },
      ) => {
        streamSignal = options?.signal;
        return new Promise(() => {});
      },
    );
    mockCreateConversation.mockRejectedValue(new Error("create failed"));

    const view = render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockStreamMessage).toHaveBeenCalledTimes(1);
    });

    await user.click(screen.getByRole("button", { name: "New Chat" }));

    await waitFor(() => {
      expect(mockShowErrorToast).toHaveBeenCalledWith(
        expect.objectContaining({ message: "create failed" }),
        "Failed to create chat conversation",
      );
    });
    expect(streamSignal?.aborted).toBe(false);
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(screen.getByText("...")).toBeInTheDocument();

    view.unmount();
  });

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
