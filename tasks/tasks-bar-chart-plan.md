# Bar Chart Research - Updated Plan

## Current Issues
1. **Tremor**: Bars black, ignoring `colors` prop
2. **Nivo**: Not rendering (no bars visible)

## Feature Requirements

### Data Aggregation
- **W**: 7 daily bars (current ✓)
- **M**: 30 daily bars (test first, reduce to 20-25 if condensed)
- **6M**: ~26 weekly bars (Mon-Sun, averages)
- **Y**: ~12 monthly bars (30-day rolling from current date, averages)

### Horizontal Scroll
- Complements range buttons (not replacement)
- Show fixed number of bars, scroll back through history
- Example: W shows 7 bars (last 7 days), scroll left to see previous weeks
- Mount at most recent data (scroll position: right)
- Scroll buttons for non-touch (left/right arrows)
- Load all data upfront (research production implications)
- **Document**: Which libs support/can't support

### Tooltip Positioning
- Anchor above bar
- Position under date range buttons
- **Document**: Which libs support/can't support

### UI Updates
- Remove "D" button from RangeSelector
- Keep W/M/6M/Y only

## Implementation Tasks

### 1. Bug Fixes
- [x] Fix Tremor color theming issue
- [x] Debug Nivo rendering (check data format, container height, theme conflicts)

### 2. Data Aggregation Logic
- [x] Add `aggregateToWeekly()` util (Monday start, average volumes)
- [x] Add `aggregateToMonthly()` util (30-day rolling from current date, average volumes)
- [x] Update `filterDataByRange()`:
  - W: Last 7 daily bars
  - M: Last 30 daily bars (test, reduce to 20-25 if condensed)
  - 6M: Last ~26 weekly bars
  - Y: Last ~12 monthly bars
- [x] Update all demos to use new aggregation

### 3. Horizontal Scroll
- [x] Research production data implications (1yr daily data ~365 points, aggregated ~26-52 points)
- [x] Research scroll APIs per library
- [x] Implement scroll with all data loaded upfront
- [x] Set initial scroll position to right (most recent)
- [x] Add scroll buttons for non-touch (left/right arrows)
- [x] Handle edge case: scroll to data start
- [x] Document limitations per library

### 4. Tooltip Positioning
- [x] Research tooltip positioning APIs per library
- [x] Implement fixed position under buttons if possible
- [x] Document limitations per library (see `tooltip-positioning-research.md`)

### 5. UI Updates
- [x] Remove 'D' from RangeType type
- [x] Update RangeSelector component (W/M/6M/Y only)
- [x] Update ranges object in mockData.ts

### 6. Production Data Research
- [x] Calculate data payload sizes:
  - Daily data (~249 points with rest days): 8.75 KB / 1.75 KB gzipped
  - Weekly aggregated (~54 points): 1.90 KB / 0.38 KB gzipped
  - Monthly aggregated (~12 points): 0.42 KB / 0.08 KB gzipped
- [x] Research bundle impact per library with full dataset
  - Chart.js: 70.6 KB gzipped (smallest)
  - Recharts: 140.1 KB, Nivo: 140.9 KB, ApexCharts: 154.5 KB
  - Tremor: 244.7 KB (includes date-fns, headlessui)
- [x] Test rendering performance with 365 bars (worst case M range)
  - All libraries handle aggregated data excellently
  - Chart.js best for large datasets (canvas-based)
- [x] Recommendation: **Client-side aggregation** (see `production-data-research.md`)
  - Full year payload only 1.75 KB gzipped
  - Instant range switching vs 200-500ms server RTT
  - Simpler architecture, better UX

### 7. Documentation
- [x] Create comparison matrix (see `library-comparison-matrix.md`)
- [x] Horizontal scroll: All ✅ except ApexCharts (⚠️ conflicts with built-in zoom)
- [x] Fixed tooltip position: Chart.js only ✅, others ❌ or ⚠️
- [x] Color theming: All ✅ (Tremor needs useEffect)
- [x] Mobile UX: Chart.js & Nivo ⭐ Excellent, others ✅ Good
- [x] Bundle size: Chart.js ⭐ 70.6KB, others 140-245KB
- [x] Performance: Chart.js ⭐ Excellent (canvas), others ✅ Good (SVG)
- [x] **Final Recommendation:** Chart.js (custom tooltips, smallest, fastest)

## Open Questions
None currently.
