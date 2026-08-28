import { describe, expect, it } from "vitest";

import {
  buildSessionMetricChartData,
  getSessionMetricBucket,
  getSessionMetricBucketLabel,
  getSessionMetricPeriodLabel,
} from "./session-metric-buckets";

const today = new Date(2026, 2, 24);
const points = [
  { date: "2026-03-02", workout_id: 1, value: 10 },
  { date: "2026-03-03", workout_id: 2, value: 20 },
];

describe("session metric buckets", () => {
  it("uses monthly and weekly buckets for the long ranges", () => {
    expect(getSessionMetricBucket("Y")).toBe("month");
    expect(getSessionMetricBucket("6M")).toBe("week");
    expect(getSessionMetricPeriodLabel("Y")).toBe("Monthly");
    expect(getSessionMetricPeriodLabel("6M")).toBe("Weekly");
    expect(getSessionMetricPeriodLabel("M")).toBe("Daily");
    expect(getSessionMetricBucketLabel("Y")).toBe(
      "Monthly bars from your first exercise session",
    );
    expect(getSessionMetricBucketLabel("6M")).toBe(
      "Weekly bars from your first exercise session, 26 visible at a time",
    );
    expect(getSessionMetricBucketLabel("M")).toBe(
      "Daily bars from your first exercise session, 26 visible at a time",
    );
    expect(getSessionMetricBucketLabel("W")).toBe(
      "Daily bars from your first exercise session, 7 visible at a time",
    );
  });

  it("builds a continuous yearly timeline from the first exercise session", () => {
    const data = buildSessionMetricChartData(
      [
        { date: "2024-01-15", workout_id: 1, value: 10 },
        { date: "2026-03-02", workout_id: 2, value: 20 },
      ],
      "Y",
      (point) => point.value,
      "sum",
      today,
    );

    expect(data).toHaveLength(27);
    expect(data[0]).toMatchObject({ date: "2024-01-01", value: 10 });
    expect(data.at(-1)).toMatchObject({ date: "2026-03-01", value: 20 });
  });

  it("combines sessions into weekly bars for six months", () => {
    const data = buildSessionMetricChartData(
      points,
      "6M",
      (point) => point.value,
      "average",
      today,
    );

    expect(data).toHaveLength(26);
    expect(data.find((point) => point.date === "2026-03-02")?.value).toBe(15);
  });

  it.each([
    ["W", "2025-01-15"],
    ["M", "2025-01-15"],
    ["6M", "2025-01-13"],
  ] as const)(
    "starts the %s timeline at the first exercise session",
    (range, expectedStart) => {
      const data = buildSessionMetricChartData(
        [{ date: "2025-01-15", workout_id: 1, value: 10 }],
        range,
        (point) => point.value,
        "maximum",
        today,
      );

      expect(data[0]?.date).toBe(expectedStart);
      expect(data[0]?.value).toBe(10);
      expect(data.length).toBeGreaterThan(range === "6M" ? 26 : 30);
    },
  );

  it.each(["6M", "Y"] as const)(
    "excludes future sessions from the current %s bucket",
    (range) => {
      const data = buildSessionMetricChartData(
        [
          { date: "2026-03-24", value: 30 },
          { date: "2026-03-25", value: 999 },
        ],
        range,
        (point) => point.value,
        "sum",
        today,
      );

      expect(data.at(-1)?.value).toBe(30);
    },
  );

  it.each(["W", "M"] as const)(
    "includes zero-value daily buckets for %s",
    (range) => {
      const sessionDate = range === "W" ? "2026-03-23" : "2026-03-02";
      const data = buildSessionMetricChartData(
        [{ date: sessionDate, workout_id: 1, value: 10 }],
        range,
        (point) => point.value,
        "maximum",
        today,
      );

      expect(data).toHaveLength(range === "W" ? 7 : 30);
      expect(data.find((point) => point.date === sessionDate)?.value).toBe(10);
      expect(data.find((point) => point.date === "2026-03-24")?.value).toBe(0);
    },
  );
});
