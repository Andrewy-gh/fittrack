import { useEffect, useState } from "react";

export type DisplayMode = "pwa" | "web";

function isStandaloneDisplayMode() {
  if (typeof window === "undefined") return false;

  return window.matchMedia("(display-mode: standalone)").matches;
}

function isIosStandalone() {
  if (typeof navigator === "undefined") return false;

  return (
    (navigator as Navigator & { standalone?: boolean }).standalone === true
  );
}

function getDisplayMode(): DisplayMode {
  return isStandaloneDisplayMode() || isIosStandalone() ? "pwa" : "web";
}

export function useDisplayMode(): DisplayMode {
  const [displayMode, setDisplayMode] = useState<DisplayMode>(getDisplayMode);

  useEffect(() => {
    const query = window.matchMedia("(display-mode: standalone)");
    const updateDisplayMode = () => setDisplayMode(getDisplayMode());

    updateDisplayMode();
    query.addEventListener("change", updateDisplayMode);

    return () => query.removeEventListener("change", updateDisplayMode);
  }, []);

  return displayMode;
}
