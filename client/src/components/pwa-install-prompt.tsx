import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"

const DISMISS_KEY = "fittrack:pwa-install-prompt:v1"

type DevicePlatform = "ios" | "android" | "other"

function getPlatform(): DevicePlatform {
  const ua = navigator.userAgent.toLowerCase()
  if (/iphone|ipad|ipod/.test(ua)) {
    return "ios"
  }
  if (/android/.test(ua)) {
    return "android"
  }
  return "other"
}

function isStandalone(): boolean {
  if (window.matchMedia("(display-mode: standalone)").matches) {
    return true
  }
  const iosStandalone = (navigator as Navigator & { standalone?: boolean })
    .standalone
  return iosStandalone === true
}

function isMobileTouch(platform: DevicePlatform): boolean {
  const coarsePointer = window.matchMedia("(pointer: coarse)").matches
  const touchPoints = navigator.maxTouchPoints > 0
  const isTouch = coarsePointer || touchPoints
  const mobileViewport = window.matchMedia("(max-width: 1024px)").matches
  const platformKnown = platform !== "other"
  return isTouch && (mobileViewport || platformKnown)
}

function getCopy(platform: DevicePlatform): { title: string; body: string } {
  if (platform === "ios") {
    return {
      title: "Install FitTrack",
      body: "Tap Share, then Add to Home Screen for the full app experience.",
    }
  }
  if (platform === "android") {
    return {
      title: "Install FitTrack",
      body: "Tap the menu (⋮), then Add to Home screen or Install app.",
    }
  }
  return {
    title: "Install FitTrack",
    body: "Use your browser menu to add FitTrack to your home screen.",
  }
}

export function PwaInstallPrompt() {
  const [visible, setVisible] = useState(false)
  const [platform, setPlatform] = useState<DevicePlatform>("other")

  useEffect(() => {
    if (typeof window === "undefined") {
      return
    }

    const header = document.querySelector<HTMLElement>("[data-app-header]")
    const updateHeaderHeight = () => {
      if (!header) {
        return
      }
      const height = Math.ceil(header.getBoundingClientRect().height)
      document.documentElement.style.setProperty(
        "--app-header-height",
        `${height}px`
      )
    }

    updateHeaderHeight()

    let observer: ResizeObserver | null = null
    if ("ResizeObserver" in window && header) {
      observer = new ResizeObserver(updateHeaderHeight)
      observer.observe(header)
    }
    window.addEventListener("resize", updateHeaderHeight)

    const nextPlatform = getPlatform()
    if (isStandalone() || !isMobileTouch(nextPlatform)) {
      if (observer) {
        observer.disconnect()
      }
      window.removeEventListener("resize", updateHeaderHeight)
      return
    }

    try {
      if (localStorage.getItem(DISMISS_KEY) === "1") {
        return
      }
    } catch {
      // Ignore storage errors; still show prompt once per load.
    }

    setPlatform(nextPlatform)
    setVisible(true)

    return () => {
      if (observer) {
        observer.disconnect()
      }
      window.removeEventListener("resize", updateHeaderHeight)
    }
  }, [])

  const dismiss = () => {
    setVisible(false)
    try {
      localStorage.setItem(DISMISS_KEY, "1")
    } catch {
      // Ignore storage errors; dismissal will be session-only.
    }
  }

  if (!visible) {
    return null
  }

  const { title, body } = getCopy(platform)

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
  )
}
