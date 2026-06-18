import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { BillingStatusResponse } from "@/features/chat/api/billing";
import {
  AIChatBillingActions,
  AIChatBillingCard,
  type AIChatBillingCardAccessState,
} from "./ai-chat-billing-card";

function renderCard(
  status?: BillingStatusResponse,
  options: { accessState?: AIChatBillingCardAccessState } = {},
) {
  const onStartCheckout = vi.fn();
  const onManageBilling = vi.fn();
  const onRefreshAccess = vi.fn();
  const accessState =
    options.accessState ?? (status?.has_access ? "ready" : "blocked");
  render(
    <>
      <AIChatBillingCard
        status={status}
        accessState={accessState}
      />
      <AIChatBillingActions
        status={status}
        accessState={accessState}
        onStartCheckout={onStartCheckout}
        onManageBilling={onManageBilling}
        onRefreshAccess={onRefreshAccess}
      />
    </>,
  );

  return { onStartCheckout, onManageBilling, onRefreshAccess };
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
      <>
        <AIChatBillingCard
          isError
          accessState="billing-error"
        />
        <AIChatBillingActions
          isError
          accessState="billing-error"
          onStartCheckout={vi.fn()}
          onManageBilling={vi.fn()}
          onRefreshAccess={vi.fn()}
        />
      </>,
    );

    expect(screen.getByText("Unavailable")).toBeInTheDocument();
    expect(screen.getByText("Could not confirm billing.")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
  });

  it("offers refresh instead of Checkout when checkout activation verification fails", async () => {
    const user = userEvent.setup();
    const onStartCheckout = vi.fn();
    const onRefreshAccess = vi.fn();
    render(
      <>
        <AIChatBillingCard
          isError
          accessState="checkout-activation-error"
          status={{
            feature_key: "ai_chatbot",
            has_access: false,
          }}
        />
        <AIChatBillingActions
          isError
          accessState="checkout-activation-error"
          status={{
            feature_key: "ai_chatbot",
            has_access: false,
          }}
          onStartCheckout={onStartCheckout}
          onManageBilling={vi.fn()}
          onRefreshAccess={onRefreshAccess}
        />
      </>,
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
      screen.getByText("Checking access after payment."),
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
    const { onManageBilling, onStartCheckout } = renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_trial",
        status: "trialing",
        cancellation_scheduled: false,
      },
      trial_usage: {
        used: 12,
        limit: 30,
      },
    });

    expect(screen.getByText("Trial")).toBeInTheDocument();
    expect(screen.getByText("12 of 30 trial prompts used")).toBeInTheDocument();
    expect(
      screen.queryByText("Your 7-day trial is active."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Manage plan" }));

    expect(onManageBilling).toHaveBeenCalledTimes(1);
    expect(onStartCheckout).not.toHaveBeenCalled();
  });

  it("opens plan management when premium is active without extra success copy", async () => {
    const user = userEvent.setup();
    const { onManageBilling, onStartCheckout } = renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_active",
        status: "active",
        cancellation_scheduled: false,
      },
    });

    expect(screen.getByText("Premium")).toBeInTheDocument();
    expect(
      screen.queryByText("Premium is active. AI chat is available."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Manage plan" }));

    expect(onManageBilling).toHaveBeenCalledTimes(1);
    expect(onStartCheckout).not.toHaveBeenCalled();
  });

  it("disables plan management while the portal is opening", () => {
    const activeStatus = {
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_active",
        status: "active",
        cancellation_scheduled: false,
      },
    } satisfies BillingStatusResponse;
    const props = {
      status: activeStatus,
      accessState: "ready" as const,
      onStartCheckout: vi.fn(),
      onManageBilling: vi.fn(),
      onRefreshAccess: vi.fn(),
    };
    render(
      <AIChatBillingActions
        {...props}
        isBillingPortalLoading
      />,
    );

    expect(
      screen.getByRole("button", { name: "Opening billing..." }),
    ).toBeDisabled();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();
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
      screen.queryByText("AI chat access is active for this account."),
    ).not.toBeInTheDocument();
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
          cancellation_scheduled: false,
        },
      },
      { accessState: "ready" },
    );

    expect(screen.getByText("Access active")).toBeInTheDocument();
    expect(
      screen.getByText("Update billing to keep chat available."),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(
        "AI chat is paused until the payment issue is resolved.",
      ),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Manage plan" }));

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
          cancellation_scheduled: false,
        },
      },
      { accessState: "activating" },
    );

    expect(screen.getByText("Activating")).toBeInTheDocument();
    expect(screen.getByText("Finishing activation.")).toBeInTheDocument();
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
        cancellation_scheduled: true,
        access_ends_at: "2026-05-30T12:00:00Z",
      },
    });

    expect(
      screen.getByText("Access continues until May 30, 2026."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();
  });

  it("shows scheduled cancellation access messaging from access end", () => {
    renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_cancel_at",
        status: "active",
        cancellation_scheduled: true,
        access_ends_at: "2026-07-10T12:00:00Z",
      },
    });

    expect(
      screen.getByText("Access continues until Jul 10, 2026."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();
  });

  it("uses the earlier access end when scheduled cancellation dates differ", () => {
    renderCard({
      feature_key: "ai_chatbot",
      has_access: true,
      subscription: {
        stripe_subscription_id: "sub_cancel_at_after_period",
        status: "active",
        cancellation_scheduled: true,
        access_ends_at: "2026-06-30T12:00:00Z",
      },
    });

    expect(
      screen.getByText("Access continues until Jun 30, 2026."),
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
          cancellation_scheduled: false,
        },
      });

      expect(screen.getByText("Action needed")).toBeInTheDocument();
      expect(screen.getByText(message)).toBeInTheDocument();
      await user.click(screen.getByRole("button", { name: "Manage plan" }));

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
          cancellation_scheduled: false,
        },
      });

      await user.click(screen.getByRole("button", { name: buttonLabel }));

      expect(onStartCheckout).toHaveBeenCalledTimes(1);
      expect(onManageBilling).not.toHaveBeenCalled();
    },
  );
});
