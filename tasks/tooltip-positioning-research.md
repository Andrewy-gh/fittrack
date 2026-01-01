# Tooltip Positioning Research

## Requirement
- Anchor tooltip above bar
- Position tooltip under date range buttons
- Fixed Y position, X follows bar position

## Library Comparison

### 1. Recharts (Baseline)
**API:** `<Tooltip position={{ x, y }} />`

**Capabilities:**
- ✅ Supports fixed position prop
- ❌ Position is NOT relative to data points
- ❌ When `position` is set, tooltip stays at fixed coordinates regardless of bar
- ❌ X coordinate doesn't follow bar position

**Implementation:**
```tsx
<Tooltip position={{ y: 0 }} />
```

**Limitations:**
- Position is relative to CartesianGrid, not data points
- Cannot anchor to bar while maintaining fixed Y
- GitHub issues [#222](https://github.com/recharts/recharts/issues/222), [#488](https://github.com/recharts/recharts/issues/488), [#1848](https://github.com/recharts/recharts/issues/1848) confirm this limitation

**Verdict:** ❌ Cannot achieve requirement

---

### 2. Tremor
**API:** `customTooltip` prop (wrapper around Recharts)

**Capabilities:**
- Built on Recharts with same limitations
- Uses Radix UI tooltip primitives
- Inherits Recharts positioning behavior

**Implementation:**
```tsx
<BarChart customTooltip={(props) => <CustomTooltip {...props} />} />
```

**Limitations:**
- Same as Recharts - cannot fix position relative to data points
- `customTooltip` controls content, not positioning

**Verdict:** ❌ Cannot achieve requirement

---

### 3. Nivo ResponsiveBar
**API:** `tooltip` prop with anchor system

**Capabilities:**
- ✅ Custom tooltip component
- ⚠️ Anchor system ('top', 'bottom', 'left', 'right', 'center')
- ❌ No true fixed positioning
- ❌ Tooltips go off-screen at edges

**Implementation:**
```tsx
<ResponsiveBar
  tooltip={({ id, value, indexValue }) => <CustomTooltip />}
/>
```

**Limitations:**
- Must manually set anchor, no auto-calculation
- Anchor applies offsets, doesn't support fixed Y + dynamic X
- GitHub issues [#580](https://github.com/plouc/nivo/issues/580), [#1358](https://github.com/plouc/nivo/issues/1358), [#2403](https://github.com/plouc/nivo/issues/2403) discuss positioning problems

**Verdict:** ❌ Cannot achieve requirement

---

### 4. Chart.js ⭐
**API:** Custom positioner functions via `Tooltip.positioners`

**Capabilities:**
- ✅ Fully custom positioning logic
- ✅ Access to chart area, data point coordinates
- ✅ Can combine fixed Y with dynamic X
- ✅ Best flexibility for requirement

**Implementation:**
```tsx
// Register custom positioner
const customPositioner: TooltipPositionerFunction<'bar'> = function(items) {
  const pos = Tooltip.positioners.average.call(this, items);
  return {
    x: pos.x,              // Follows bar X position
    y: this.chart.chartArea.top,  // Fixed Y at top
    xAlign: 'center',
    yAlign: 'bottom',
  };
};

Tooltip.positioners.fixedTop = customPositioner;

// Use in options
options: {
  plugins: {
    tooltip: {
      position: 'fixedTop'
    }
  }
}
```

**Limitations:**
- None for this use case
- Requires understanding of Chart.js internals

**Verdict:** ✅ **Fully supports requirement**

**References:**
- [Chart.js Tooltip Position Docs](https://www.chartjs.org/docs/latest/samples/tooltip/position.html)
- [Tooltip API](https://www.chartjs.org/docs/latest/configuration/tooltip.html)

---

### 5. ApexCharts
**API:** `tooltip.fixed` configuration

**Capabilities:**
- ✅ Fixed positioning support
- ⚠️ Limited to predefined positions (topRight, topLeft, bottomRight, bottomLeft)
- ✅ Supports offsetX/offsetY
- ❌ Cannot dynamically position X based on bar

**Implementation:**
```tsx
tooltip: {
  fixed: {
    enabled: true,
    position: 'topLeft',
    offsetX: 120,
    offsetY: 0,
  }
}
```

**Limitations:**
- Tooltip stays in corner, doesn't follow bars
- Cannot achieve "X follows bar, Y fixed" requirement
- Custom tooltip HTML doesn't override fixed positioning behavior

**Verdict:** ⚠️ **Partial support** - can fix position but not relative to bars

**References:**
- [ApexCharts Tooltip Docs](https://apexcharts.com/docs/options/tooltip/)
- GitHub issues [#1178](https://github.com/apexcharts/apexcharts.js/issues/1178), [#3257](https://github.com/apexcharts/apexcharts.js/issues/3257)

---

## Summary Matrix

| Library | Fixed Position | Anchor to Bar | X Follows Bar | Implementation | Verdict |
|---------|---------------|---------------|---------------|----------------|---------|
| **Recharts** | ⚠️ Partial | ❌ No | ❌ No | `position={{ y: 0 }}` | ❌ Not viable |
| **Tremor** | ⚠️ Partial | ❌ No | ❌ No | Built on Recharts | ❌ Not viable |
| **Nivo** | ❌ No | ⚠️ Anchor only | ❌ No | `tooltip` prop | ❌ Not viable |
| **Chart.js** | ✅ Yes | ✅ Yes | ✅ Yes | Custom positioner | ✅ **Best option** |
| **ApexCharts** | ✅ Yes | ❌ No | ❌ No | `fixed.position` | ⚠️ Corner only |

## Recommendation

**Chart.js** is the only library that fully supports the tooltip positioning requirement:
- Fixed Y position (under buttons)
- Dynamic X position (follows bar)
- Clean API via custom positioner functions

All other libraries have fundamental limitations that prevent achieving the desired behavior.

## Alternative Approaches

If Chart.js is not viable, alternatives include:

1. **Custom Portal Tooltip**: Render tooltip outside chart using React Portal, calculate position with DOM measurements
2. **Accept Default Behavior**: Use standard hover-following tooltips (all libraries)
3. **Simplified Fixed**: Use ApexCharts fixed corners (acceptable for some use cases)

---

**Research completed:** 2026-01-01
