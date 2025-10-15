# PWA Update Implementation Plan

## Steps
1. `cd client && bun add -D vite-plugin-pwa`
2. Update `vite.config.js`:
   - Import VitePWA
   - registerType: 'prompt'
   - workbox strategies for assets/api
3. Fix `manifest.json`: FitTrack branding, proper names
4. Create `src/components/ReloadPrompt.tsx`:
   - useRegisterSW hook
   - Toast/dialog when update available
   - User clicks → reload with new version
5. Add `<ReloadPrompt />` to `App.tsx`
6. Test: `bun run build && bun run serve`
   - Chrome DevTools → Application → Service Workers
   - Rebuild to simulate update, verify prompt appears

## Testing PWA Updates
- **Local**: Preview server + DevTools SW panel
- **Staging**: Deploy, install PWA, redeploy, check prompt
- **Manual testing standard** - SW behavior hard to automate
- **Skip TDD for SW** - use manual/e2e instead

## TDD Best Practice
- **Yes**: Business logic, utils, components
- **No**: SW updates, browser APIs, PWA lifecycle
- **Production**: Mix both - TDD for logic, manual for platform features

## Key Config
```js
VitePWA({
  registerType: 'prompt',
  workbox: {
    globPatterns: ['**/*.{js,css,html,ico,png,svg}']
  }
})
```

## Files to modify
- `client/vite.config.js`
- `client/public/manifest.json`
- `client/src/components/ReloadPrompt.tsx` (new)
- `client/src/App.tsx`
