import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ReloadPrompt } from "./reload-prompt";

const swMock = vi.hoisted(() => ({
  needRefresh: false,
  options: undefined as
    | {
        onRegisteredSW?: (
          swUrl: string,
          registration?: ServiceWorkerRegistration,
        ) => void;
      }
    | undefined,
  updateServiceWorker: vi.fn(),
}));

vi.mock("@/lib/pwa-register", async () => {
  const React = await import("react");

  return {
    useRegisterSW(options: typeof swMock.options) {
      swMock.options = options;
      const [needRefresh, setNeedRefresh] = React.useState(swMock.needRefresh);

      return {
        needRefresh: [needRefresh, setNeedRefresh],
        offlineReady: [false, vi.fn()],
        updateServiceWorker: swMock.updateServiceWorker,
      };
    },
  };
});

describe("ReloadPrompt", () => {
  beforeEach(() => {
    swMock.needRefresh = false;
    swMock.options = undefined;
    swMock.updateServiceWorker.mockReset();
  });

  it("shows the update prompt and reloads through the service worker", async () => {
    swMock.needRefresh = true;

    render(<ReloadPrompt />);

    expect(screen.getByText("New update available")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Reload" }));

    expect(swMock.updateServiceWorker).toHaveBeenCalledWith(true);
  });

  it("checks for a service worker update when the installed app resumes", async () => {
    const registration = {
      update: vi.fn().mockResolvedValue(undefined),
    } as unknown as ServiceWorkerRegistration;

    render(<ReloadPrompt />);

    swMock.options?.onRegisteredSW?.("/sw.js", registration);
    expect(registration.update).toHaveBeenCalledTimes(1);

    fireEvent.focus(window);
    expect(registration.update).toHaveBeenCalledTimes(2);

    Object.defineProperty(document, "visibilityState", {
      configurable: true,
      value: "hidden",
    });
    fireEvent(document, new Event("visibilitychange"));
    expect(registration.update).toHaveBeenCalledTimes(2);

    Object.defineProperty(document, "visibilityState", {
      configurable: true,
      value: "visible",
    });
    fireEvent(document, new Event("visibilitychange"));
    expect(registration.update).toHaveBeenCalledTimes(3);
  });
});
