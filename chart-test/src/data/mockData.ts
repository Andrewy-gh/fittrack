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
 * Filters volume data by range type
 */
export function filterDataByRange(data: VolumeData[], range: RangeType): VolumeData[] {
  return filterDataByDays(data, ranges[range]);
}
