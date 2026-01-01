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
- [ ] Research tooltip positioning APIs per library
- [ ] Implement fixed position under buttons if possible
- [ ] Document limitations per library

### 5. UI Updates
- [ ] Remove 'D' from RangeType type
- [ ] Update RangeSelector component (W/M/6M/Y only)
- [ ] Update ranges object in mockData.ts

### 6. Production Data Research
- [ ] Calculate data payload sizes:
  - Daily data (365 points): ~X KB JSON
  - Weekly aggregated (52 points): ~X KB JSON
  - Monthly aggregated (12 points): ~X KB JSON
- [ ] Research bundle impact per library with full dataset
- [ ] Test rendering performance with 365 bars (worst case M range)
- [ ] Recommendation: Client-side vs server-side aggregation

### 7. Documentation
Create comparison matrix:
- Horizontal scroll: Supported/Partial/No
- Fixed tooltip position: Supported/No
- Color theming: Works/Issues
- Mobile UX: Excellent/Good/Fair
- Data loading performance: Excellent/Good/Fair
- Recommendation: Keep/Consider/Reject

## Open Questions
None currently.
