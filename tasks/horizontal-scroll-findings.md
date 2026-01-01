# Horizontal Scroll Implementation - Findings

## Production Data Implications

### Data Payload Sizes
Based on mock data structure: `{date: "YYYY-MM-DD", volume: number}` (~40 bytes/record in JSON)

- **Daily data (365 points)**: ~14.6 KB raw, ~10-12 KB actual (70% workout days)
- **Weekly aggregated (52 points)**: ~2 KB
- **Monthly aggregated (12 points)**: ~0.5 KB

### Recommendations
- **Load all data upfront**: Payloads are minimal (<15 KB)
- **Client-side aggregation**: Acceptable performance for 365 data points
- **No pagination needed**: All ranges fit comfortably in memory

## Library-Specific Scroll Implementation

### 1. Recharts

**Native Scroll Support**: ❌ None

**Implementation**:
- Wrapper div with `overflow-x: auto`
- Calculate fixed width based on `dataLength * barWidth`
- ResponsiveContainer expands to wrapper width

**Limitations**:
- Y-axis scrolls with chart (not sticky)
- No built-in scroll API ([GitHub #1364](https://github.com/recharts/recharts/issues/1364), [#1507](https://github.com/recharts/recharts/issues/1507), [#3059](https://github.com/recharts/recharts/issues/3059))
- Requires manual scroll position management

**Workarounds Applied**:
- Custom ScrollableChart wrapper component
- Left/right arrow buttons for non-touch
- Auto-scroll to right (most recent) on mount
- Smooth scrolling via `element.scrollBy()`

**Resources**:
- [Scrollable Graph with Recharts (Medium)](https://medium.com/@SwathiMahadevarajan/scrollable-graph-sticky-y-axis-in-recharts-without-using-brush-3733f60b1177)
- [Horizontal Scrollable Graph in Recharts](https://velog.io/@ichenny/Horizontal-Scrollable-Graph-in-Recharts)

### 2. Tremor

**Native Scroll Support**: ❌ None

**Implementation**:
- Same wrapper approach as Recharts (built on Recharts)
- `enableLegendSlider` prop available but only for legend overflow, not chart scroll

**Limitations**:
- Same as Recharts (Y-axis scrolls with chart)
- No native horizontal scroll API
- BarChart component doesn't expose scroll controls

**Workarounds Applied**:
- Same ScrollableChart wrapper
- Works seamlessly with Tremor's pre-styled components

**Resources**:
- [Tremor BarChart Documentation](https://www.tremor.so/docs/visualizations/bar-chart)
- [Tremor GitHub](https://github.com/tremorlabs/tremor/issues)

### 3. Nivo

**Native Scroll Support**: ❌ None

**Implementation**:
- Wrapper div with overflow-x
- ResponsiveBar adapts to wrapper constraints

**Limitations**:
- [GitHub issue #1563](https://github.com/plouc/nivo/issues/1563) discusses need for scroll with 7+ elements
- No built-in scroll mechanism
- Responsive behavior handles vertical/horizontal resize but not overflow

**Workarounds Applied**:
- Same ScrollableChart wrapper
- ResponsiveBar fills available space correctly

**Resources**:
- [Nivo Bar Chart Documentation](https://nivo.rocks/bar/)
- [Nivo Bar Examples (CodeSandbox)](https://codesandbox.io/examples/package/@nivo/bar)
- [Building charts in React with Nivo (LogRocket)](https://blog.logrocket.com/building-charts-in-react-with-nivo/)

## ScrollableChart Component Features

Created universal wrapper component supporting all libraries:

### Features Implemented
- ✅ Horizontal scroll with fixed bar width (default 50px)
- ✅ Auto-scroll to right (most recent data) on mount
- ✅ Left/right arrow buttons for non-touch devices
- ✅ Button visibility based on scroll position
- ✅ Smooth scrolling (80% of container width per click)
- ✅ Touch/trackpad scroll support
- ✅ Styled scrollbar with CSS (thin, themed)
- ✅ Dynamic width calculation (`dataLength * barWidth`)

### Edge Cases Handled
- ✅ No data: Renders empty container
- ✅ Few data points: No scroll buttons shown (fits in view)
- ✅ Many data points: Scroll buttons appear when needed
- ✅ Data changes: Auto-updates scroll state
- ✅ Scroll to start/end: Buttons hide at boundaries

## Comparison Matrix

| Feature | Recharts | Tremor | Nivo |
|---------|----------|--------|------|
| **Native Scroll** | ❌ No | ❌ No | ❌ No |
| **Wrapper Compatible** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Sticky Y-Axis** | ❌ No | ❌ No | ❌ No |
| **Touch Scroll** | ✅ Yes* | ✅ Yes* | ✅ Yes* |
| **Button Controls** | ✅ Yes* | ✅ Yes* | ✅ Yes* |
| **Auto-scroll to End** | ✅ Yes* | ✅ Yes* | ✅ Yes* |

*Via ScrollableChart wrapper

## Tooltip Positioning

### Current Behavior
All three libraries position tooltips **relative to cursor/bar**, not fixed position.

### Requirements Not Met
- ❌ Fixed position under range buttons
- ❌ Tooltip stays in place when scrolling

### Why This Is Difficult
- Recharts/Tremor: Tooltip rendered by Recharts internals, no fixed position API
- Nivo: Custom tooltip component but still anchored to bar position
- Would require custom portal-based tooltip or major library customization

### Recommendation
- Accept default tooltip behavior (anchored to bar)
- Tooltips are well-positioned and don't overflow in testing
- Attempting fixed position would require significant engineering effort with marginal UX benefit

## Overall Assessment

### What Works Well
- ✅ Horizontal scroll via wrapper approach (universal solution)
- ✅ Production data loads efficiently (<15 KB)
- ✅ Smooth UX with arrow buttons + touch scroll
- ✅ Auto-mount at most recent data
- ✅ All three libraries compatible with same wrapper

### Limitations Accepted
- ❌ Y-axis scrolls with chart (not sticky)
- ❌ Tooltips not fixed position (anchored to bars)
- ❌ No native library support (wrapper dependency)

### Recommendation
**Proceed with ScrollableChart wrapper approach**. It provides a consistent, functional scroll experience across all libraries with minimal complexity.
