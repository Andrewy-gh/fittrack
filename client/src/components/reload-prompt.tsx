import { useRegisterSW } from "virtual:pwa-register/react";
import { Button } from "./ui/button";

export function ReloadPrompt() {
  const {
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW({
    onRegisterError(error: Error) {
      if (import.meta.env.DEV) {
        console.error("SW registration error", error);
      }
    },
  });

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
