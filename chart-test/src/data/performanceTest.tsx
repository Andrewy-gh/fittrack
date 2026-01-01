/**
 * Performance test component to measure rendering time with large datasets
 * Test approach: Measure time from component mount to paint completion
 */

import { useState, useEffect, useRef } from 'react';
import { BarChart } from 'recharts';
import { ResponsiveBar } from '@nivo/bar';
import { Bar } from 'react-chartjs-2';
import { format, parseISO } from 'date-fns';
import { VolumeData } from './mockData';

// Generate test data
function generateTestData(days: number): VolumeData[] {
  const data: VolumeData[] = [];
  const today = new Date();

  for (let i = days - 1; i >= 0; i--) {
    const date = new Date(today);
    date.setDate(date.getDate() - i);
    const dateStr = date.toISOString().split('T')[0];

    if (Math.random() > 0.3) {
      data.push({
        date: dateStr,
        volume: Math.round(8000 + (Math.random() - 0.5) * 4000),
      });
    }
  }

  return data;
}

interface PerformanceTestProps {
  library: 'recharts' | 'nivo' | 'chartjs';
  dataSize: number;
  onComplete: (duration: number) => void;
}

export function PerformanceTest({ library, dataSize, onComplete }: PerformanceTestProps) {
  const startTime = useRef(performance.now());
  const [data] = useState(() => generateTestData(dataSize));

  useEffect(() => {
    // Wait for next paint to measure actual render time
    requestAnimationFrame(() => {
      const endTime = performance.now();
      const duration = endTime - startTime.current;
      onComplete(duration);
    });
  }, [onComplete]);

  if (library === 'recharts') {
    return (
      <div style={{ width: '100%', height: '400px' }}>
        <BarChart width={800} height={400} data={data}>
          <bar dataKey="volume" fill="#ea580c" />
        </BarChart>
      </div>
    );
  }

  if (library === 'nivo') {
    const nivoData = data.map((d) => ({
      date: format(parseISO(d.date), 'MMM d'),
      volume: d.volume,
    }));

    return (
      <div style={{ width: '100%', height: '400px' }}>
        <ResponsiveBar
          data={nivoData}
          keys={['volume']}
          indexBy="date"
          colors={['#ea580c']}
        />
      </div>
    );
  }

  if (library === 'chartjs') {
    const chartData = {
      labels: data.map((d) => format(parseISO(d.date), 'MMM d')),
      datasets: [
        {
          label: 'Volume',
          data: data.map((d) => d.volume),
          backgroundColor: '#ea580c',
        },
      ],
    };

    return (
      <div style={{ width: '100%', height: '400px' }}>
        <Bar data={chartData} options={{ responsive: true, maintainAspectRatio: false }} />
      </div>
    );
  }

  return null;
}

// Test runner component
export function PerformanceTestRunner() {
  const [results, setResults] = useState<Record<string, number>>({});
  const [currentTest, setCurrentTest] = useState<{ library: string; size: number } | null>(null);

  const tests = [
    { library: 'recharts', size: 30 },
    { library: 'recharts', size: 365 },
    { library: 'nivo', size: 30 },
    { library: 'nivo', size: 365 },
    { library: 'chartjs', size: 30 },
    { library: 'chartjs', size: 365 },
  ];

  const runTests = () => {
    setResults({});
    setCurrentTest(tests[0] as any);
  };

  const handleComplete = (duration: number) => {
    if (!currentTest) return;

    const key = `${currentTest.library}-${currentTest.size}`;
    setResults((prev) => ({ ...prev, [key]: duration }));

    const currentIndex = tests.findIndex(
      (t) => t.library === currentTest.library && t.size === currentTest.size
    );

    if (currentIndex < tests.length - 1) {
      setTimeout(() => {
        setCurrentTest(tests[currentIndex + 1] as any);
      }, 1000);
    } else {
      setCurrentTest(null);
    }
  };

  return (
    <div style={{ padding: '20px' }}>
      <h2>Chart Performance Test</h2>
      <button onClick={runTests}>Run Tests</button>

      {Object.keys(results).length > 0 && (
        <div style={{ marginTop: '20px' }}>
          <h3>Results:</h3>
          <pre>{JSON.stringify(results, null, 2)}</pre>
        </div>
      )}

      {currentTest && (
        <div style={{ marginTop: '20px' }}>
          <p>Testing: {currentTest.library} with {currentTest.size} data points</p>
          <PerformanceTest
            library={currentTest.library as any}
            dataSize={currentTest.size}
            onComplete={handleComplete}
          />
        </div>
      )}
    </div>
  );
}
