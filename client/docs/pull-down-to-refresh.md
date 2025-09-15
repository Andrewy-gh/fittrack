**I have a web app which I am currently using as a PWA, although I haven't formally set it up as a PWA, I just added it to home on my iOS device. How do I implement a feature where when you pull down from the screen it refreshes?**

Short answer: In iOS “Add to Home Screen” mode, the native pull‑to‑refresh gesture is disabled, so it won’t automatically reload like in Safari tabs; to get pull‑to‑refresh, implement a custom JavaScript touch/scroll handler that shows a refresh control and triggers a reload or data refetch when pulled past a threshold.[1][2]

### Why it’s disabled
iOS treats standalone/home‑screen web apps differently from Safari tabs and removes the built‑in pull‑to‑refresh UI, so the usual “drag to refresh” won’t appear in a PWA shell even if it does in Safari proper. Developers typically re‑implement this behavior within the page if they want it.[2][1]

### Implementation approach
- Use a top‑level scroll container (not the body) and track touchstart/touchmove when the scroll position is at the very top. On downward drag past a threshold, show a spinner and run either location.reload() or a data refetch function; then stop the spinner when done. This mimics UIRefreshControl behavior used with WKWebView.[3][4]
- Consider overscroll control: on Android/Chrome you can manage edge effects with CSS overscroll-behavior; iOS Safari support is limited, so rely on JS for detection and visuals.[5][6]

### Minimal JS pattern
- Container setup: Place page content in a scrollable wrapper div that fills the viewport. Listen for touch events only when scrollTop == 0 to detect a pull.[4]
- Threshold and spinner: Add a hidden top “refresh header” that slides down as the user pulls; when the pull distance exceeds a threshold (e.g., 60–80px), trigger the refresh and lock the header until complete.[4]
- Action: Choose either a soft refresh (refetch network data and re-render) or a hard reload via location.reload() depending on the app architecture.[3][4]

### Example outline
- HTML: a wrapper with a refresh header and content region.[4]
- JS: track startY, currentY, pulling state, distance; translate the header; on release past threshold call refresh() and then collapse the header when complete. This mirrors native WKWebView + UIRefreshControl patterns used in native shells.[3][4]

### Notes and pitfalls
- Don’t try to “re‑enable” the native gesture in standalone mode—it’s not available; use a custom component instead.[1][2]
- If disabling browser bounce or default pull is needed elsewhere, some developers add touchmove handlers or CSS to prevent default behaviors; do this carefully to avoid breaking scrolling.[7][6][5]
- If the app ever ships as a true native wrapper, iOS provides UIRefreshControl on WKWebView’s scrollView to get native pull‑to‑refresh with few lines of code.[8][3]

[1](https://stackoverflow.com/questions/75972895/ios-pwa-how-to-re-enable-pull-to-refresh)
[2](https://vaadin.com/forum/t/how-to-make-pwa-to-refresh-on-pull-down/164811)
[3](https://spin.atomicobject.com/reload-wkwebview/)
[4](https://dev.to/chicio/implement-a-pull-to-refresh-component-for-you-web-application-1pcg)
[5](https://developer.chrome.com/blog/overscroll-behavior)
[6](https://www.matuzo.at/blog/2022/100daysof-day53)
[7](https://dev.to/khaled17/how-to-prevent-pull-down-to-refresh-of-mobile-browser-mjp)
[8](https://developer.apple.com/documentation/webkit/wkwebview)
[9](https://invisioncommunity.com/forums/topic/473218-ios-pwa-needs-reload-and-back-buttons-on-every-page/)
[10](https://www.reddit.com/r/ios/comments/1hfhmo9/hard_refresh_in_mobile_safari/)
[11](https://news.ycombinator.com/item?id=16850614)
[12](https://forum.bubble.io/t/pull-to-refresh-after-ios-wrapping/184702)
[13](https://stackoverflow.com/questions/53268384/pull-to-refresh-with-wkwebview)
[14](https://stackoverflow.com/questions/69261011/disable-pull-to-refresh-in-ios-15-safari)
[15](https://meta.discourse.org/t/ios-pwa-app-pull-to-refresh/343262?tl=en)
[16](https://www.youtube.com/shorts/rwriYAFY1NU)
[17](https://forum.xojo.com/t/web-disable-pull-to-refresh/81492)
[18](https://heltweg.org/posts/checklist-issues-progressive-web-apps-how-to-fix/)
[19](https://brainhub.eu/library/pwa-on-ios)
[20](https://forum.bubble.io/t/prevent-pull-to-refresh-on-mobile/308800)