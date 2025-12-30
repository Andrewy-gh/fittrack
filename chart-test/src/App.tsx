import { useState } from 'react';
import { mockVolumeData, type RangeType } from './data/mockData';

function App() {
  const [isDarkMode, setIsDarkMode] = useState(true);

  // Apply dark mode class to document
  if (isDarkMode) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }

  return (
    <div className="app-container">
      {/* Header */}
      <header style={{ marginBottom: '2rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h1 style={{ fontSize: '2rem', fontWeight: 'bold', marginBottom: '0.5rem' }}>
              Bar Chart Library Comparison
            </h1>
            <p style={{ color: 'var(--color-muted-foreground)' }}>
              Testing 5 different charting libraries for FitTrack volume visualization
            </p>
          </div>
          <button
            onClick={() => setIsDarkMode(!isDarkMode)}
            style={{
              padding: '0.5rem 1rem',
              borderRadius: 'var(--radius-md)',
              border: '1px solid var(--color-border)',
              backgroundColor: 'var(--color-card)',
              color: 'var(--color-foreground)',
              cursor: 'pointer',
              fontSize: '0.875rem',
            }}
          >
            {isDarkMode ? '☀️ Light Mode' : '🌙 Dark Mode'}
          </button>
        </div>
      </header>

      {/* Info Section */}
      <section className="section">
        <div className="card" style={{ marginBottom: '2rem' }}>
          <h2 className="card-title">Project Overview</h2>
          <p className="card-description" style={{ marginBottom: '1rem' }}>
            This demo compares 5 charting libraries as alternatives to the current Recharts Brush component.
            The goal is to find a better mobile UX for range selection, similar to Apple Health's segmented control.
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
            <div>
              <strong>Current Issue:</strong>
              <p style={{ fontSize: '0.875rem', color: 'var(--color-muted-foreground)' }}>
                Brush slider difficult to control on mobile
              </p>
            </div>
            <div>
              <strong>Goal:</strong>
              <p style={{ fontSize: '0.875rem', color: 'var(--color-muted-foreground)' }}>
                Button-based range selection (D/W/M/6M/Y)
              </p>
            </div>
            <div>
              <strong>Data Points:</strong>
              <p style={{ fontSize: '0.875rem', color: 'var(--color-muted-foreground)' }}>
                {mockVolumeData.length} days of mock volume data
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Chart Implementations Section */}
      <section className="section">
        <h2 className="section-title">Chart Implementations</h2>
        <p className="section-description">
          Each implementation below uses the same mock data and CSS variables for theming.
          Click the range buttons to filter the data.
        </p>

        <div className="chart-grid">
          {/* Placeholder for chart demos */}
          <div className="card">
            <h3 className="card-title">1. Recharts + Custom Buttons</h3>
            <p className="card-description">Implementation coming soon...</p>
            <div style={{
              height: '300px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'var(--color-muted)',
              borderRadius: 'var(--radius-md)',
              marginTop: '1rem'
            }}>
              <span style={{ color: 'var(--color-muted-foreground)' }}>Chart placeholder</span>
            </div>
          </div>

          <div className="card">
            <h3 className="card-title">2. Tremor BarChart</h3>
            <p className="card-description">Implementation coming soon...</p>
            <div style={{
              height: '300px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'var(--color-muted)',
              borderRadius: 'var(--radius-md)',
              marginTop: '1rem'
            }}>
              <span style={{ color: 'var(--color-muted-foreground)' }}>Chart placeholder</span>
            </div>
          </div>

          <div className="card">
            <h3 className="card-title">3. Nivo ResponsiveBar</h3>
            <p className="card-description">Implementation coming soon...</p>
            <div style={{
              height: '300px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'var(--color-muted)',
              borderRadius: 'var(--radius-md)',
              marginTop: '1rem'
            }}>
              <span style={{ color: 'var(--color-muted-foreground)' }}>Chart placeholder</span>
            </div>
          </div>

          <div className="card">
            <h3 className="card-title">4. Chart.js</h3>
            <p className="card-description">Implementation coming soon...</p>
            <div style={{
              height: '300px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'var(--color-muted)',
              borderRadius: 'var(--radius-md)',
              marginTop: '1rem'
            }}>
              <span style={{ color: 'var(--color-muted-foreground)' }}>Chart placeholder</span>
            </div>
          </div>

          <div className="card">
            <h3 className="card-title">5. ApexCharts</h3>
            <p className="card-description">Implementation coming soon...</p>
            <div style={{
              height: '300px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'var(--color-muted)',
              borderRadius: 'var(--radius-md)',
              marginTop: '1rem'
            }}>
              <span style={{ color: 'var(--color-muted-foreground)' }}>Chart placeholder</span>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer style={{ marginTop: '3rem', padding: '2rem 0', borderTop: '1px solid var(--color-border)' }}>
        <p style={{ textAlign: 'center', color: 'var(--color-muted-foreground)', fontSize: '0.875rem' }}>
          Bar Chart Library Research for FitTrack | {new Date().getFullYear()}
        </p>
      </footer>
    </div>
  );
}

export default App;
