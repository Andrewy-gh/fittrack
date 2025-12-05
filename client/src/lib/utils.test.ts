import { formatWeight } from './utils';

describe('formatWeight', () => {
  it('should return "0" for null', () => {
    expect(formatWeight(null)).toBe('0');
  });

  it('should return "0" for undefined', () => {
    expect(formatWeight(undefined)).toBe('0');
  });

  it('should return whole number without decimal for zero', () => {
    expect(formatWeight(0)).toBe('0');
  });

  it('should return whole number without decimal', () => {
    expect(formatWeight(45)).toBe('45');
    expect(formatWeight(135)).toBe('135');
    expect(formatWeight(225)).toBe('225');
  });

  it('should return decimal with one decimal place', () => {
    expect(formatWeight(45.5)).toBe('45.5');
    expect(formatWeight(22.5)).toBe('22.5');
    expect(formatWeight(135.5)).toBe('135.5');
  });

  it('should format decimal with one place even for trailing zeros', () => {
    expect(formatWeight(45.0)).toBe('45'); // Should display as whole number
  });

  it('should handle very small decimal values', () => {
    expect(formatWeight(0.5)).toBe('0.5');
    expect(formatWeight(2.5)).toBe('2.5');
  });

  it('should handle large numbers correctly', () => {
    expect(formatWeight(999999)).toBe('999999');
    expect(formatWeight(999999.5)).toBe('999999.5');
  });

  it('should round to one decimal place for numbers with more precision', () => {
    expect(formatWeight(45.55)).toBe('45.6'); // toFixed rounds
    expect(formatWeight(45.54)).toBe('45.5');
    expect(formatWeight(45.99)).toBe('46.0');
  });

  it('should handle negative numbers if they somehow appear', () => {
    // While validation should prevent this, test defensive behavior
    expect(formatWeight(-45)).toBe('-45');
    expect(formatWeight(-45.5)).toBe('-45.5');
  });
});
