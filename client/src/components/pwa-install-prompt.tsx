import { useEffect, useState } from "react";
import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { Button } from "@/components/ui/button";
import type { DisplayMode } from "@/hooks/use-display-mode";

export const PWA_INSTALL_PROMPT_DISMISS_KEY = "fittrack:pwa-install-prompt:v1";

type DevicePlatform = "ios" | "android" | "other";
type PromptUser = CurrentUser | CurrentInternalUser | null;

interface PwaInstallPromptProps {
  displayMode: DisplayMode;
  pathname: string;
  user: PromptUser;
}

const PROMPT_ROUTES = ["/workouts", "/exercises", "/analytics", "/chat"];

export function isPwaInstallPromptRoute(pathname: string): boolean {
  return PROMPT_ROUTES.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`),
  );
}

function getPlatform(): DevicePlatform {
  const ua = navigator.userAgent.toLowerCase();
  if (/iphone|ipad|ipod/.test(ua)) {
    return "ios";
  }
  if (/android/.test(ua)) {
    return "android";
  }
  return "other";
}

function isMobileTouch(platform: DevicePlatform): boolean {
  const coarsePointer = window.matchMedia("(pointer: coarse)").matches;
  const touchPoints = navigator.maxTouchPoints > 0;
  const isTouch = coarsePointer || touchPoints;
  const mobileViewport = window.matchMedia("(max-width: 1024px)").matches;
  const platformKnown = platform !== "other";
  return isTouch && (mobileViewport || platformKnown);
}

function getCopy(platform: DevicePlatform): { title: string; body: string } {
  if (platform === "ios") {
    return {
      title: "Install FitTrack",
      body: "Tap Share, then Add to Home Screen for the full app experience.",
    };
  }
  if (platform === "android") {
    return {
      title: "Install FitTrack",
      body: "Tap the menu (⋮), then Add to Home screen or Install app.",
    };
  }
  return {
    title: "Install FitTrack",
    body: "Use your browser menu to add FitTrack to your home screen.",
  };
}

function hasDurableDismissal(): boolean {
  try {
    return localStorage.getItem(PWA_INSTALL_PROMPT_DISMISS_KEY) === "1";
  } catch {
    return false;
  }
}

export function PwaInstallPrompt({
  displayMode,
  pathname,
  user,
}: PwaInstallPromptProps) {
  const [visible, setVisible] = useState(false);
  const [platform, setPlatform] = useState<DevicePlatform>("other");
  const [durablyDismissed, setDurablyDismissed] = useState(hasDurableDismissal);
  const [shownForUserId, setShownForUserId] = useState<string | null>(null);
  const userId = user?.id ?? null;

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const header = document.querySelector<HTMLElement>("[data-app-header]");
    const updateHeaderHeight = () => {
      if (!header) {
        return;
      }
      const height = Math.ceil(header.getBoundingClientRect().height);
      document.documentElement.style.setProperty(
        "--app-header-height",
        `${height}px`,
      );
    };

    updateHeaderHeight();

    let observer: ResizeObserver | null = null;
    if ("ResizeObserver" in window && header) {
      observer = new ResizeObserver(updateHeaderHeight);
      observer.observe(header);
    }
    window.addEventListener("resize", updateHeaderHeight);

    return () => {
      if (observer) {
        observer.disconnect();
      }
      window.removeEventListener("resize", updateHeaderHeight);
    };
  }, [displayMode, pathname]);

  useEffect(() => {
    if (!userId) {
      setShownForUserId(null);
      setVisible(false);
      return;
    }

    const dismissed = hasDurableDismissal();
    setDurablyDismissed(dismissed);

    const nextPlatform = getPlatform();
    const eligible =
      displayMode === "web" &&
      isPwaInstallPromptRoute(pathname) &&
      !dismissed &&
      isMobileTouch(nextPlatform);

    if (!eligible) {
      setVisible(false);
      return;
    }

    if (shownForUserId === userId) {
      return;
    }

    setPlatform(nextPlatform);
    setVisible(true);
    setShownForUserId(userId);
  }, [displayMode, durablyDismissed, pathname, shownForUserId, userId]);

  const dismiss = () => {
    setVisible(false);
    try {
      localStorage.setItem(PWA_INSTALL_PROMPT_DISMISS_KEY, "1");
      setDurablyDismissed(true);
    } catch {
      // Ignore storage errors; dismissal will be session-only.
    }
  };

  if (!visible) {
    return null;
  }

  const { title, body } = getCopy(platform);

  return (
    <div
      className="fixed left-1/2 z-50 w-[min(94vw,520px)] -translate-x-1/2 rounded-xl border border-destructive/40 bg-destructive p-3 text-destructive-foreground shadow-lg top-[calc(env(safe-area-inset-top)+var(--app-header-height,0px)+0.75rem)]"
      role="status"
      aria-live="polite"
    >
      <div className="flex items-start gap-3">
        <div className="flex-1">
          <p className="text-sm font-semibold text-destructive-foreground">
            {title}
          </p>
          <p className="text-xs text-destructive-foreground/90">{body}</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={dismiss}
          className="text-black dark:text-white"
        >
          Not now
        </Button>
      </div>
    </div>
  );
}
