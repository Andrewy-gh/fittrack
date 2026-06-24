import { useEffect, useRef } from "react";
import { useRegisterSW } from "@/lib/pwa-register";
import { Button } from "./ui/button";

export function ReloadPrompt() {
  const registrationRef = useRef<ServiceWorkerRegistration | null>(null);

  const checkForUpdate = () => {
    void registrationRef.current?.update().catch((error: unknown) => {
      if (import.meta.env.DEV) {
        console.error("SW update check failed", error);
      }
    });
  };

  const {
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW({
    onRegisteredSW(_swUrl, registration) {
      registrationRef.current = registration ?? null;
      checkForUpdate();
    },
    onRegisterError(error: Error) {
      if (import.meta.env.DEV) {
        console.error("SW registration error", error);
      }
    },
  });

  useEffect(() => {
    const checkWhenVisible = () => {
      if (document.visibilityState === "visible") {
        checkForUpdate();
      }
    };

    document.addEventListener("visibilitychange", checkWhenVisible);
    window.addEventListener("focus", checkForUpdate);

    return () => {
      document.removeEventListener("visibilitychange", checkWhenVisible);
      window.removeEventListener("focus", checkForUpdate);
    };
  }, []);

  const close = () => {
    setNeedRefresh(false);
  };

  return (
    <div>
      {needRefresh && (
        <div className="fixed bottom-4 right-4 z-50 flex items-center gap-2 rounded-lg border bg-background p-4 shadow-lg">
          <div className="flex-1">
            <span>New update available</span>
          </div>
          <div className="flex gap-2">
            <Button
              onClick={() => updateServiceWorker(true)}
              size="sm"
            >
              Reload
            </Button>
            <Button
              onClick={close}
              variant="outline"
              size="sm"
            >
              Close
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
