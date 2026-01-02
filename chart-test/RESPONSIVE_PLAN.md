# Mobile-Responsive Bar Charts Implementation

## Overview

Make all 5 demo charts responsive for mobile touch screens by:
- Responsive bar widths (30px mobile, 40px tablet, 50px desktop)
- Responsive fonts/margins for readability
- Migrate ChartJs/Apex to ScrollableChart for consistency
- Mobile-friendly RangeSelector buttons
- Touch scroll optimizations

**User requirements:**
- Keep horizontal scroll ✓
- Keep fixed 320px height ✓
- Mobile-first (main interface on touch screens)

## Implementation Plan

### Commit 1: Add responsive infrastructure

**New: `src/hooks/useBreakpoint.ts`**
- Hook using window.matchMedia
- Returns: 'mobile' (<640px) | 'tablet' (640-1024px) | 'desktop' (>1024px)
- Event listeners for breakpoint changes

**New: `src/utils/responsiveConfig.ts`**
- Centralized config object:
  - `barWidth`: { mobile: 30, tablet: 40, desktop: 50 }
  - `fontSize`: { mobile: 10, tablet: 11, desktop: 12 }
  - `chartMargins`: responsive margins for Recharts
  - `nivoMargins`: responsive margins for Nivo
  - `buttonPadding`: responsive classes for RangeSelector
  - `scrollButton`: responsive size/icon dimensions
- Export `getResponsiveValue()` helper

### Commit 2: Make ScrollableChart responsive

**Modify: `src/components/ScrollableChart.tsx`**
- Import useBreakpoint hook
- Accept optional `responsiveBarWidths` prop override
- Use breakpoint to determine barWidth from config
- Responsive scroll button sizing:
  - Mobile: `p-1.5`, 14px icons, `left-1/right-1`
  - Desktop: `p-2`, 16px icons, `left-2/right-2`
- Update minChartWidth calculation with responsive barWidth

### Commit 3: Update charts using ScrollableChart

**Modify: `src/components/RechartsDemo.tsx`**
- Import useBreakpoint, responsiveConfig
- Get current breakpoint
- Pass responsive barWidth to ScrollableChart
- XAxis/YAxis fontSize from config
- Margins from config
- Line 36: `barWidth={getResponsiveValue(responsiveConfig.barWidth, breakpoint)}`
- Lines 40-44: margins from config
- Lines 55, 65: fontSize from config

**Modify: `src/components/TremorDemo.tsx`**
- Similar to RechartsDemo
- Pass responsive barWidth
- Adjust yAxisWidth responsively (mobile: 40, desktop: 48)

**Modify: `src/components/NivoDemo.tsx`**
- Pass responsive barWidth
- Update theme.text.fontSize from config
- Update margins from nivoMargins config
- Responsive axisBottom/axisLeft tickPadding

### Commit 4: Migrate ChartJs/Apex to ScrollableChart

**Modify: `src/components/ChartJsDemo.tsx`**
- Import ScrollableChart, useBreakpoint, responsiveConfig
- Replace `<div className="h-80 w-full">` with ScrollableChart
- Wrap `<Bar>` component in ScrollableChart
- Add responsive fontSize to scales.x/y.ticks.font.size
- Line 154-156: Replace with ScrollableChart wrapper

**Modify: `src/components/ApexDemo.tsx`**
- Wrap ReactApexChart with ScrollableChart
- Add responsive fontSize to xaxis/yaxis labels
- Update plotOptions.bar.columnWidth based on breakpoint
- Remove h-80 wrapper div

### Commit 5: Make RangeSelector responsive

**Modify: `src/components/RangeSelector.tsx`**
- Import useBreakpoint, responsiveConfig
- Apply responsive padding/text classes:
  - Mobile: `px-3 py-1.5 text-xs`, container gap-0.5, p-0.5
  - Tablet: `px-3.5 py-2 text-sm`
  - Desktop: `px-4 py-2 text-sm`
- Line 17: dynamic padding/gap from config
- Line 22: dynamic button classes

### Commit 6: CSS touch optimizations

**Modify: `src/index.css`**
- Add mobile media query:
```css
@media (max-width: 640px) {
  .chart-grid {
    gap: 1.5rem;
  }

  .overflow-x-auto {
    -webkit-overflow-scrolling: touch;
    touch-action: pan-x;
  }
}
```

## File Summary

**New files:**
- `src/hooks/useBreakpoint.ts`
- `src/utils/responsiveConfig.ts`

**Modified files:**
- `src/components/ScrollableChart.tsx`
- `src/components/RechartsDemo.tsx`
- `src/components/TremorDemo.tsx`
- `src/components/NivoDemo.tsx`
- `src/components/ChartJsDemo.tsx`
- `src/components/ApexDemo.tsx`
- `src/components/RangeSelector.tsx`
- `src/index.css`

## Technical Decisions

**Breakpoints:** 640px/1024px (Tailwind standard)
**Bar widths:** 30/40/50px (narrower on mobile = more data visible)
**Font sizes:** 10/11/12px (modest reduction maintains readability)
**Height:** Fixed 320px all screens (per user requirement)
**Scroll:** Keep horizontal scroll (per user requirement)
**Migration:** ChartJs/Apex use ScrollableChart for consistency
**Touch:** Browser native overflow-x-auto + -webkit-overflow-scrolling

## Commit Messages

```
feat(responsive): add breakpoint hook and config

- Create useBreakpoint hook for mobile/tablet/desktop detection
- Add responsiveConfig utility with bar widths, fonts, margins
```

```
feat(responsive): make ScrollableChart responsive

- Dynamic bar width based on breakpoint (30/40/50px)
- Responsive scroll button sizing and positioning
- Smaller buttons on mobile for better UX
```

```
feat(responsive): make Recharts/Tremor/Nivo responsive

- Apply responsive bar widths from config
- Responsive font sizes for axes (10/11/12px)
- Responsive margins for better mobile layout
```

```
feat(responsive): migrate ChartJs/Apex to ScrollableChart

- Wrap both charts in ScrollableChart component
- Add responsive font sizes to axis config
- Consistent scroll behavior across all demos
```

```
feat(responsive): make RangeSelector mobile-friendly

- Responsive button padding and text sizes
- Tighter spacing on mobile (<640px)
- Better touch targets for mobile users
```

```
style(responsive): add mobile touch optimizations

- Reduce chart-grid gap on mobile (1.5rem)
- Add smooth touch scrolling for iOS
- Set touch-action: pan-x for horizontal scroll
```
