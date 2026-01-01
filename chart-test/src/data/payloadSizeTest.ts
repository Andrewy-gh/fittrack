/**
 * Test script to calculate JSON payload sizes for production data
 * Run: bun run src/data/payloadSizeTest.ts
 */

import { VolumeData } from './mockData';

// Helper to calculate JSON size in bytes
function getJSONSize(data: any): number {
  const jsonString = JSON.stringify(data);
  return new TextEncoder().encode(jsonString).length;
}

// Helper to format bytes to KB
function formatBytes(bytes: number): string {
  return `${(bytes / 1024).toFixed(2)} KB`;
}

// Generate test data
function generateTestData(days: number): VolumeData[] {
  const data: VolumeData[] = [];
  const today = new Date();

  for (let i = days - 1; i >= 0; i--) {
    const date = new Date(today);
    date.setDate(date.getDate() - i);
    const dateStr = date.toISOString().split('T')[0];

    // Simulate ~70% workout days (30% rest days)
    if (Math.random() > 0.3) {
      data.push({
        date: dateStr,
        volume: Math.round(8000 + (Math.random() - 0.5) * 4000),
      });
    }
  }

  return data;
}

// Aggregate to weekly
function aggregateToWeekly(data: VolumeData[]): VolumeData[] {
  const weekMap = new Map<string, number[]>();

  data.forEach((item) => {
    const date = new Date(item.date);
    const dayOfWeek = date.getDay();
    const diff = dayOfWeek === 0 ? -6 : 1 - dayOfWeek;
    const monday = new Date(date);
    monday.setDate(date.getDate() + diff);
    const weekKey = monday.toISOString().split('T')[0];

    if (!weekMap.has(weekKey)) {
      weekMap.set(weekKey, []);
    }
    weekMap.get(weekKey)!.push(item.volume);
  });

  return Array.from(weekMap.entries())
    .map(([date, volumes]) => ({
      date,
      volume: Math.round(volumes.reduce((sum, v) => sum + v, 0) / volumes.length),
    }))
    .sort((a, b) => a.date.localeCompare(b.date));
}

// Aggregate to monthly
function aggregateToMonthly(data: VolumeData[]): VolumeData[] {
  const monthlyData: VolumeData[] = [];
  const sortedData = [...data].sort((a, b) => a.date.localeCompare(b.date));

  if (sortedData.length === 0) return [];

  const endDate = new Date(sortedData[sortedData.length - 1].date);

  for (let i = 0; i < 12; i++) {
    const windowEnd = new Date(endDate);
    windowEnd.setDate(endDate.getDate() - (i * 30));

    const windowStart = new Date(windowEnd);
    windowStart.setDate(windowEnd.getDate() - 30);

    const windowData = sortedData.filter((item) => {
      const itemDate = new Date(item.date);
      return itemDate >= windowStart && itemDate <= windowEnd;
    });

    if (windowData.length > 0) {
      monthlyData.unshift({
        date: windowEnd.toISOString().split('T')[0],
        volume: Math.round(
          windowData.reduce((sum, d) => sum + d.volume, 0) / windowData.length
        ),
      });
    }
  }

  return monthlyData;
}

// Run tests
console.log('📊 Production Data Payload Size Analysis\n');

// 1. Daily data (365 days)
const dailyData = generateTestData(365);
const dailySize = getJSONSize(dailyData);
console.log(`1️⃣  Daily Data (365 days):`);
console.log(`   Points: ${dailyData.length}`);
console.log(`   Size: ${formatBytes(dailySize)} (${dailySize} bytes)`);
console.log(`   Sample: ${JSON.stringify(dailyData[0])}`);
console.log();

// 2. Weekly aggregated data (~52 weeks)
const weeklyData = aggregateToWeekly(dailyData);
const weeklySize = getJSONSize(weeklyData);
console.log(`2️⃣  Weekly Aggregated (~52 weeks):`);
console.log(`   Points: ${weeklyData.length}`);
console.log(`   Size: ${formatBytes(weeklySize)} (${weeklySize} bytes)`);
console.log(`   Reduction: ${((1 - weeklySize / dailySize) * 100).toFixed(1)}% smaller`);
console.log(`   Sample: ${JSON.stringify(weeklyData[0])}`);
console.log();

// 3. Monthly aggregated data (~12 months)
const monthlyData = aggregateToMonthly(dailyData);
const monthlySize = getJSONSize(monthlyData);
console.log(`3️⃣  Monthly Aggregated (~12 months):`);
console.log(`   Points: ${monthlyData.length}`);
console.log(`   Size: ${formatBytes(monthlySize)} (${monthlySize} bytes)`);
console.log(`   Reduction: ${((1 - monthlySize / dailySize) * 100).toFixed(1)}% smaller`);
console.log(`   Sample: ${JSON.stringify(monthlyData[0])}`);
console.log();

// 4. Range-specific payloads
console.log(`4️⃣  Range-Specific Payloads:`);

const ranges = [
  { name: 'W (7 days)', data: dailyData.slice(-7) },
  { name: 'M (30 days)', data: dailyData.slice(-30) },
  { name: '6M (~26 weeks)', data: aggregateToWeekly(dailyData.slice(-180)).slice(-26) },
  { name: 'Y (~12 months)', data: monthlyData },
];

ranges.forEach(({ name, data }) => {
  const size = getJSONSize(data);
  console.log(`   ${name}: ${data.length} points, ${formatBytes(size)}`);
});

console.log();
console.log(`📦 Summary:`);
console.log(`   Full year daily: ${formatBytes(dailySize)}`);
console.log(`   Full year weekly: ${formatBytes(weeklySize)}`);
console.log(`   Full year monthly: ${formatBytes(monthlySize)}`);
console.log();

// Compression estimates (typical JSON gzip compression ratio: 4-10x)
console.log(`🗜️  Estimated Gzipped Sizes (5x compression ratio):`);
console.log(`   Daily: ~${formatBytes(dailySize / 5)}`);
console.log(`   Weekly: ~${formatBytes(weeklySize / 5)}`);
console.log(`   Monthly: ~${formatBytes(monthlySize / 5)}`);
