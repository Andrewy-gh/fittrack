/* Mock workout volume data for chart testing */

export interface VolumeData {
  date: string; // ISO date format "YYYY-MM-DD"
  volume: number; // Total volume in kg/lbs
}

/**
 * Generates mock volume data for the past N days
 * Creates realistic-looking workout data with some variation and rest days
 */
function generateMockData(days: number): VolumeData[] {
  const data: VolumeData[] = [];
  const today = new Date();

  for (let i = days - 1; i >= 0; i--) {
    const date = new Date(today);
    date.setDate(date.getDate() - i);

    // Format date as YYYY-MM-DD
    const dateStr = date.toISOString().split('T')[0];

    // Skip some days to simulate rest days (roughly 2-3 rest days per week)
    const isRestDay = Math.random() < 0.3;

    if (!isRestDay) {
      // Generate volume between 5000-15000 with some variation
      // Add some weekly patterns (higher volume mid-week)
      const dayOfWeek = date.getDay();
      const baseMidWeekBonus = [0, 1, 2, 3, 2, 1, 0][dayOfWeek];
      const baseVolume = 8000 + baseMidWeekBonus * 1000;
      const variation = (Math.random() - 0.5) * 4000;
      const volume = Math.round(baseVolume + variation);

      data.push({
        date: dateStr,
        volume: Math.max(3000, volume), // Ensure minimum volume
      });
    }
  }

  return data;
}

// Generate 1 year of data (365 days)
export const mockVolumeData: VolumeData[] = generateMockData(365);

// Pre-filtered datasets for quick testing
export const mockData1Day = mockVolumeData.slice(-1);
export const mockData1Week = mockVolumeData.slice(-7);
export const mockData1Month = mockVolumeData.slice(-30);
export const mockData6Months = mockVolumeData.slice(-180);
export const mockData1Year = mockVolumeData;

// Sample data with specific dates for controlled testing
export const sampleVolumeData: VolumeData[] = [
  { date: '2024-12-01', volume: 8500 },
  { date: '2024-12-02', volume: 9200 },
  { date: '2024-12-04', volume: 7800 },
  { date: '2024-12-05', volume: 10500 },
  { date: '2024-12-07', volume: 8900 },
  { date: '2024-12-08', volume: 11200 },
  { date: '2024-12-09', volume: 9800 },
  { date: '2024-12-11', volume: 8200 },
  { date: '2024-12-12', volume: 10100 },
  { date: '2024-12-14', volume: 9500 },
  { date: '2024-12-15', volume: 12000 },
  { date: '2024-12-16', volume: 8700 },
  { date: '2024-12-18', volume: 9900 },
  { date: '2024-12-19', volume: 10800 },
  { date: '2024-12-21', volume: 8400 },
  { date: '2024-12-22', volume: 11500 },
  { date: '2024-12-23', volume: 9300 },
  { date: '2024-12-26', volume: 10700 },
  { date: '2024-12-27', volume: 9100 },
  { date: '2024-12-28', volume: 11800 },
  { date: '2024-12-29', volume: 10200 },
];

/**
 * Range types matching the Apple Health-style segmented control
 */
export type RangeType = 'D' | 'W' | 'M' | '6M' | 'Y';

export const ranges: Record<RangeType, number> = {
  D: 1,     // 1 day
  W: 7,     // 7 days
  M: 30,    // 30 days
  '6M': 180, // 180 days
  Y: 365    // 365 days
};

/**
 * Filters volume data to the last N days
 */
export function filterDataByDays(data: VolumeData[], days: number): VolumeData[] {
  if (days >= data.length) return data;
  return data.slice(-days);
}

/**
 * Filters volume data by range type with appropriate aggregation
 * - W: Last 7 daily bars
 * - M: Last 30 daily bars
 * - 6M: Last ~26 weekly bars (aggregated)
 * - Y: Last ~12 monthly bars (aggregated)
 */
export function filterDataByRange(data: VolumeData[], range: RangeType): VolumeData[] {
  switch (range) {
    case 'D':
      // Single day - last day only
      return filterDataByDays(data, 1);

    case 'W':
      // Week - last 7 daily bars
      return filterDataByDays(data, 7);

    case 'M':
      // Month - last 30 daily bars
      return filterDataByDays(data, 30);

    case '6M':
      // 6 months - aggregate to weekly, then take last 26 weeks
      const sixMonthData = filterDataByDays(data, 180);
      const weeklyData = aggregateToWeekly(sixMonthData);
      return weeklyData.slice(-26);

    case 'Y':
      // Year - aggregate to monthly, return ~12 bars
      return aggregateToMonthly(data);

    default:
      return data;
  }
}

/**
 * Aggregates daily data into weekly averages (Monday-Sunday)
 * Returns ~26 weeks for 6-month period
 */
export function aggregateToWeekly(data: VolumeData[]): VolumeData[] {
  if (data.length === 0) return [];

  const weekMap = new Map<string, { volumes: number[]; weekStart: Date }>();

  data.forEach((item) => {
    const date = new Date(item.date);

    // Find Monday of this week (ISO week starts on Monday)
    const dayOfWeek = date.getDay();
    const diff = dayOfWeek === 0 ? -6 : 1 - dayOfWeek; // Sunday is 0, adjust to -6
    const monday = new Date(date);
    monday.setDate(date.getDate() + diff);

    // Use Monday's date as the key (YYYY-MM-DD)
    const weekKey = monday.toISOString().split('T')[0];

    if (!weekMap.has(weekKey)) {
      weekMap.set(weekKey, { volumes: [], weekStart: monday });
    }

    weekMap.get(weekKey)!.volumes.push(item.volume);
  });

  // Convert to array and calculate averages
  const weeklyData: VolumeData[] = Array.from(weekMap.entries())
    .map(([weekKey, { volumes }]) => ({
      date: weekKey,
      volume: Math.round(volumes.reduce((sum, v) => sum + v, 0) / volumes.length),
    }))
    .sort((a, b) => a.date.localeCompare(b.date));

  return weeklyData;
}

/**
 * Aggregates daily data into monthly averages (30-day rolling windows)
 * Returns ~12 monthly bars for a year of data
 */
export function aggregateToMonthly(data: VolumeData[]): VolumeData[] {
  if (data.length === 0) return [];

  const monthlyData: VolumeData[] = [];
  const sortedData = [...data].sort((a, b) => a.date.localeCompare(b.date));

  // Get the most recent date
  const endDate = new Date(sortedData[sortedData.length - 1].date);

  // Create 12 monthly buckets going backwards from the end date
  for (let i = 0; i < 12; i++) {
    // Calculate the end of this 30-day window
    const windowEnd = new Date(endDate);
    windowEnd.setDate(endDate.getDate() - (i * 30));

    // Calculate the start of this 30-day window
    const windowStart = new Date(windowEnd);
    windowStart.setDate(windowEnd.getDate() - 30);

    // Filter data within this window
    const windowData = sortedData.filter((item) => {
      const itemDate = new Date(item.date);
      return itemDate >= windowStart && itemDate <= windowEnd;
    });

    if (windowData.length > 0) {
      const avgVolume = Math.round(
        windowData.reduce((sum, d) => sum + d.volume, 0) / windowData.length
      );

      // Use the window end date as the label
      monthlyData.unshift({
        date: windowEnd.toISOString().split('T')[0],
        volume: avgVolume,
      });
    }
  }

  return monthlyData;
}

/**
 * Get the appropriate label for the stats section based on range type
 */
export function getRangeLabel(range: RangeType, count: number): string {
  switch (range) {
    case 'D':
      return `Showing ${count} day${count !== 1 ? 's' : ''}`;
    case 'W':
      return `Showing ${count} day${count !== 1 ? 's' : ''}`;
    case 'M':
      return `Showing ${count} day${count !== 1 ? 's' : ''}`;
    case '6M':
      return `Showing ${count} week${count !== 1 ? 's' : ''}`;
    case 'Y':
      return `Showing ${count} month${count !== 1 ? 's' : ''}`;
    default:
      return `Showing ${count} points`;
  }
}

/**
 * Get the appropriate date format pattern based on range type
 */
export function getDateFormat(range: RangeType): string {
  switch (range) {
    case 'D':
      return 'PPP'; // Full date: January 1, 2024
    case 'W':
    case 'M':
      return 'MMM d'; // Short date: Jan 1
    case '6M':
      return 'MMM d'; // Week start: Jan 1
    case 'Y':
      return 'MMM yyyy'; // Month: Jan 2024
    default:
      return 'MMM d';
  }
}
