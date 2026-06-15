import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { BillingStatusResponse } from "@/features/chat/api/billing";
import {
  AIChatBillingCard,
  type AIChatBillingCardAccessState,
} from "./ai-chat-billing-card";

function renderCard(
  status?: BillingStatusResponse,
  options: { accessState?: AIChatBillingCardAccessState } = {},
) {
  const onStartCheckout = vi.fn();
  const onManageBilling = vi.fn();
  const onCancelPlan = vi.fn();
  const onRefreshAccess = vi.fn();
  render(
    <AIChatBillingCard
      status={status}
      accessState={
        options.accessState ?? (status?.has_access ? "ready" : "blocked")
      }
      onStartCheckout={onStartCheckout}
      onManageBilling={onManageBilling}
      onCancelPlan={onCancelPlan}
      onRefreshAccess={onRefreshAccess}
    />,
  );

  return { onStartCheckout, onManageBilling, onCancelPlan, onRefreshAccess };
}

describe("AIChatBillingCard", () => {
  it("shows the trial CTA when the user has no active access", async () => {
    const user = userEvent.setup();
    const { onStartCheckout } = renderCard({
      feature_key: "ai_chatbot",
      has_access: false,
    });

    await user.click(screen.getByRole("button", { name: "Start 7-day trial" }));

    expect(onStartCheckout).toHaveBeenCalledTimes(1);
    expect(screen.getByText("Trial available")).toBeInTheDocument();
    expect(screen.getByText(/30 AI prompts/)).toBeInTheDocument();
  });

  it("does not show Checkout when billing status cannot be confirmed", () => {
    render(
      <AIChatBillingCard
        isError
        accessState="billing-error"
        onStartCheckout={vi.fn()}
        onManageBilling={vi.fn()}
        onCancelPlan={vi.fn()}
        onRefreshAccess={vi.fn()}
      />,
    );

    expect(screen.getByText("Unavailable")).toBeInTheDocument();
    expect(
      screen.getByText(
        "We could not confirm billing status. Refresh the page or try again soon.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
  });

  it("offers refresh instead of Checkout when checkout activation verification fails", async () => {
    const user = userEvent.setup();
    const onStartCheckout = vi.fn();
    const onRefreshAccess = vi.fn();
    render(
      <AIChatBillingCard
        isError
        accessState="checkout-activation-error"
        status={{
          feature_key: "ai_chatbot",
          has_access: false,
        }}
        onStartCheckout={onStartCheckout}
        onManageBilling={vi.fn()}
        onCancelPlan={vi.fn()}
        onRefreshAccess={onRefreshAccess}
      />,
    );

    expect(screen.getByText("Unavailable")).toBeInTheDocument();
    expect(
      screen.getByText(
        "Checkout finished, but we could not refresh AI chat access. Try refreshing access.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(onRefreshAccess).toHaveBeenCalledTimes(1);
    expect(onStartCheckout).not.toHaveBeenCalled();
  });

  it("shows payment confirmation after Checkout without offering Checkout again", async () => {
    const user = userEvent.setup();
    const { onRefreshAccess, onStartCheckout } = renderCard(
      {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      { accessState: "payment-confirming" },
    );

    expect(screen.getByText("Confirming")).toBeInTheDocument();
    expect(
      screen.getByText(
        "Payment complete. We are confirming your AI chat access and will keep checking automatically.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(onRefreshAccess).toHaveBeenCalledTimes(1);
    expect(onStartCheckout).not.toHaveBeenCalled();
  });

  it("opens plan management while trialing", async () => {
    const user = userEvent.setup();
    const { onManageBilling, onCancelPlan, onStartCheckout } = renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_trial",
        status: "trialing",
        cancel_at_period_end: false,
      },
      trial_usage: {
        used: 12,
        limit: 30,
      },
    });

    expect(screen.getByText("Trial")).toBeInTheDocument();
    expect(screen.getByText("12 of 30 trial prompts used")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Manage plan" }));

    expect(onManageBilling).toHaveBeenCalledTimes(1);
    expect(onCancelPlan).not.toHaveBeenCalled();
    expect(onStartCheckout).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Cancel plan" }));

    expect(onCancelPlan).toHaveBeenCalledTimes(1);
  });

  it("opens plan management when premium is active", async () => {
    const user = userEvent.setup();
    const { onManageBilling, onCancelPlan, onStartCheckout } = renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_active",
        status: "active",
        cancel_at_period_end: false,
      },
    });

    expect(screen.getByText("Premium")).toBeInTheDocument();
    expect(
      screen.getByText("Premium is active. AI chat is available."),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Manage plan" }));

    expect(onManageBilling).toHaveBeenCalledTimes(1);
    expect(onCancelPlan).not.toHaveBeenCalled();
    expect(onStartCheckout).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Cancel plan" }));

    expect(onCancelPlan).toHaveBeenCalledTimes(1);
  });

  it("disables both portal actions while either active plan action is opening", () => {
    const activeStatus = {
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_active",
        status: "active",
        cancel_at_period_end: false,
      },
    } satisfies BillingStatusResponse;
    const props = {
      status: activeStatus,
      accessState: "ready" as const,
      onStartCheckout: vi.fn(),
      onManageBilling: vi.fn(),
      onCancelPlan: vi.fn(),
      onRefreshAccess: vi.fn(),
    };
    const { rerender } = render(
      <AIChatBillingCard
        {...props}
        isBillingPortalLoading
      />,
    );

    expect(
      screen.getByRole("button", { name: "Opening billing..." }),
    ).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel plan" })).toBeDisabled();

    rerender(
      <AIChatBillingCard
        {...props}
        isCancelPlanLoading
      />,
    );

    expect(screen.getByRole("button", { name: "Manage plan" })).toBeDisabled();
    expect(
      screen.getByRole("button", { name: "Opening cancellation..." }),
    ).toBeDisabled();
  });

  it("shows active access without Checkout when access comes from a non-Stripe grant", () => {
    renderCard(
      {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      { accessState: "ready" },
    );

    expect(screen.getByText("Access active")).toBeInTheDocument();
    expect(
      screen.getByText("AI chat access is active for this account."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
  });

  it("keeps billing management available when a non-Stripe grant overrides blocked Stripe billing", async () => {
    const user = userEvent.setup();
    const { onManageBilling, onStartCheckout } = renderCard(
      {
        feature_key: "ai_chatbot",
        has_access: false,
        subscription: {
          stripe_subscription_id: "sub_past_due",
          status: "past_due",
          cancel_at_period_end: false,
        },
      },
      { accessState: "ready" },
    );

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

    await user.click(screen.getByRole("button", { name: "Update billing" }));

    expect(onManageBilling).toHaveBeenCalledTimes(1);
    expect(onStartCheckout).not.toHaveBeenCalled();
  });

  it("offers refresh instead of Checkout while paid access is activating", async () => {
    const user = userEvent.setup();
    const { onRefreshAccess, onStartCheckout } = renderCard(
      {
        feature_key: "ai_chatbot",
        has_access: true,
        subscription: {
          stripe_subscription_id: "sub_active",
          status: "active",
          cancel_at_period_end: false,
        },
      },
      { accessState: "activating" },
    );

    expect(screen.getByText("Activating")).toBeInTheDocument();
    expect(
      screen.getByText(
        "Premium is active. We are finishing AI chat activation and will keep checking automatically.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(onRefreshAccess).toHaveBeenCalledTimes(1);
    expect(onStartCheckout).not.toHaveBeenCalled();
  });

  it("shows cancellation-at-period-end access messaging", () => {
    renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_cancel_later",
        status: "active",
        cancel_at_period_end: true,
        current_period_end: "2026-05-30T12:00:00Z",
      },
    });

    expect(
      screen.getByText("Access continues until May 30, 2026."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();
  });

  it("shows scheduled cancellation access messaging from cancel_at", () => {
    renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_cancel_at",
        status: "active",
        cancel_at_period_end: false,
        cancel_at: "2026-07-10T03:39:36Z",
        current_period_end: "2026-07-10T03:39:36Z",
      },
    });

    expect(
      screen.getByText("Access continues until Jul 9, 2026."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();
  });

  it.each([
    ["past_due", "AI chat is paused until the payment issue is resolved."],
    ["unpaid", "AI chat is paused until the payment issue is resolved."],
  ] as const)(
    "opens billing management for payment issue state %s",
    async (subscriptionStatus, message) => {
      const user = userEvent.setup();
      const { onManageBilling, onStartCheckout } = renderCard({
        feature_key: "ai_chatbot",
        has_access: false,
        subscription: {
          stripe_subscription_id: `sub_${subscriptionStatus}`,
          status: subscriptionStatus,
          cancel_at_period_end: false,
        },
      });

      expect(screen.getByText("Action needed")).toBeInTheDocument();
      expect(screen.getByText(message)).toBeInTheDocument();
      await user.click(screen.getByRole("button", { name: "Update billing" }));

      expect(onManageBilling).toHaveBeenCalledTimes(1);
      expect(onStartCheckout).not.toHaveBeenCalled();
    },
  );

  it.each([
    ["canceled", "Restart premium"],
    ["incomplete", "Finish Checkout"],
    ["incomplete_expired", "Restart Checkout"],
  ] as const)(
    "uses Checkout for restartable blocked state %s",
    async (subscriptionStatus, buttonLabel) => {
      const user = userEvent.setup();
      const { onStartCheckout, onManageBilling } = renderCard({
        feature_key: "ai_chatbot",
        has_access: false,
        subscription: {
          stripe_subscription_id: `sub_${subscriptionStatus}`,
          status: subscriptionStatus,
          cancel_at_period_end: false,
        },
      });

      await user.click(screen.getByRole("button", { name: buttonLabel }));

      expect(onStartCheckout).toHaveBeenCalledTimes(1);
      expect(onManageBilling).not.toHaveBeenCalled();
    },
  );
});
