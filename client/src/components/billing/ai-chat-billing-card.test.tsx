import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { BillingStatusResponse } from "@/lib/api/billing";
import { AIChatBillingCard } from "./ai-chat-billing-card";

function renderCard(status?: BillingStatusResponse) {
  const onStartCheckout = vi.fn();
  render(
    <AIChatBillingCard
      status={status}
      onStartCheckout={onStartCheckout}
    />,
  );

  return { onStartCheckout };
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
        onStartCheckout={vi.fn()}
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

  it("shows the trial badge and prompt usage while trialing", () => {
    renderCard({
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
  });

  it("shows premium status when active", () => {
    renderCard({
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
  });

  it.each([
    ["past_due", "AI chat is paused until the payment issue is resolved."],
    ["unpaid", "AI chat is paused until the payment issue is resolved."],
    ["canceled", "AI chat is blocked because the subscription is canceled."],
    ["incomplete", "AI chat is blocked until Checkout is completed."],
    [
      "incomplete_expired",
      "AI chat is blocked because the Checkout session expired.",
    ],
  ] as const)("shows a blocked state for %s", (subscriptionStatus, message) => {
    renderCard({
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
    expect(
      screen.getByRole("button", { name: "Update billing" }),
    ).toBeInTheDocument();
  });
});
