import { useRegisterSW } from 'virtual:pwa-register/react';
import { Button } from './ui/button';

export function ReloadPrompt() {
  const {
    offlineReady: [offlineReady, setOfflineReady],
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW({
    onRegistered(r: ServiceWorkerRegistration | undefined) {
      console.log('SW Registered: ' + r);
    },
    onRegisterError(error: Error) {
      console.log('SW registration error', error);
    },
  });

  const close = () => {
    setOfflineReady(false);
    setNeedRefresh(false);
  };

  return (
    <div>
      {(offlineReady || needRefresh) && (
        <div className="fixed bottom-4 right-4 z-50 flex items-center gap-2 rounded-lg border bg-background p-4 shadow-lg">
          <div className="flex-1">
            {offlineReady ? (
              <span>App ready to work offline</span>
            ) : (
              <span>New update available</span>
            )}
          </div>
          <div className="flex gap-2">
            {needRefresh && (
              <Button
                onClick={() => updateServiceWorker(true)}
                size="sm"
              >
                Reload
              </Button>
            )}
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
