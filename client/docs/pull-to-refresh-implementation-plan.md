# Pull-to-Refresh Implementation Plan

## Context
- **App**: React PWA (makeshift) using TanStack Query/Router
- **Issue**: No browser controls in full-screen PWA mode, need refresh capability
- **Current status**: App works fine without refresh, but user wants refresh option for PWA usage

## Technical Stack Analysis
- **Framework**: React 19 with Vite
- **State Management**: TanStack Query v5.84.2
- **UI**: Radix UI components + Tailwind CSS
- **Icons**: Lucide React
- **Testing**: Vitest + Testing Library

## Verified Information from Research
The existing research in `pull-down-to-refresh.md` is **accurate**:
- iOS PWAs disable native pull-to-refresh gesture
- Custom implementation requires touch event handling
- Threshold-based approach with visual feedback is standard
- Medium complexity to implement from scratch

## Recommended Solutions (Priority Order)

### 1. Simple Refresh Button (Recommended Start)
**Effort**: 5 minutes
**Implementation**:
```javascript
// Add RefreshCw icon from lucide-react to header
const refreshData = () => {
  queryClient.invalidateQueries(); // TanStack Query refresh
  // or window.location.reload() for full page refresh
};
```

**Pros**: Accessible, reliable, immediate solution
**Cons**: Less "native" feeling

### 2. Pull-to-Refresh Library
**Effort**: 30 minutes
**Library**: `react-pull-to-refresh`
**Command**: `bunx add react-pull-to-refresh`

**Integration**: Works well with TanStack Query's `invalidateQueries()`
**Pros**: Native-like experience
**Cons**: Platform inconsistencies, accessibility concerns, touch event complexity

### 3. Alternative Gesture Options
- Double-tap header
- Long-press logo/title
- Custom swipe-down implementation

## Implementation Considerations

### Drawbacks to Consider
1. **Platform inconsistency**: Behaves differently across iOS/Android/desktop
2. **Accessibility**: Not keyboard/screen-reader friendly
3. **Performance**: Touch listeners can impact scroll performance
4. **Testing complexity**: Gesture testing harder in Vitest
5. **User expectations**: May not feel exactly like native apps

### Technical Questions for Implementation
1. **Placement**: Where should refresh control be located?
2. **Scope**: Refresh all TanStack queries or specific data?
3. **Feedback**: What visual indicators during refresh?
4. **Fallbacks**: Behavior on non-touch devices?

## Next Steps
1. **Immediate**: Implement simple refresh button as interim solution
2. **Future**: Consider pull-to-refresh if native feel is priority
3. **Alternative**: Explore auto-refresh + manual button combination

## Files to Reference
- `client/package.json` - Current dependencies
- `client/docs/pull-down-to-refresh.md` - Research and technical details

---
*Created for handoff to implementation agent*