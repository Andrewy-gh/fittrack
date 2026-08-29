import { describe, expect, it } from "vitest";

import { buildHistoricalRangeTimeline } from "./historical-range-timeline";

const today = new Date(2026, 2, 24);

describe("historical range timeline", () => {
  it.each([
    ["W", "day", 7, "2026-03-18", "2026-03-24"],
    ["M", "day", 30, "2026-02-23", "2026-03-24"],
    ["6M", "week", 26, "2025-09-29", "2026-03-23"],
    ["Y", "month", 1, "2026-03-01", "2026-03-01"],
  ] as const)(
    "builds the minimum %s timeline",
    (range, bucket, count, firstDate, lastDate) => {
      const timeline = buildHistoricalRangeTimeline({
        range,
        items: [],
        dateOf: () => undefined,
        today,
      });

      expect(timeline.bucket).toBe(bucket);
      expect(timeline.buckets).toHaveLength(count);
      expect(timeline.buckets[0]?.date).toBe(firstDate);
      expect(timeline.buckets.at(-1)?.date).toBe(lastDate);
    },
  );

  it("extends monthly history to the first item and fills empty buckets", () => {
    const first = { date: "2024-01-15", value: 10 };
    const latest = { date: "2026-03-02T12:00:00.000Z", value: 20 };
    const timeline = buildHistoricalRangeTimeline({
      range: "Y",
      items: [first, latest],
      dateOf: (item) => item.date,
      today,
    });

    expect(timeline.buckets).toHaveLength(27);
    expect(timeline.buckets[0]).toEqual({
      date: "2024-01-01",
      items: [first],
    });
    expect(timeline.buckets[1]).toEqual({
      date: "2024-02-01",
      items: [],
    });
    expect(timeline.buckets.at(-1)).toEqual({
      date: "2026-03-01",
      items: [latest],
    });
  });

  it("groups sessions into Monday-aligned weekly buckets", () => {
    const mondaySession = { date: "2026-03-02", id: 1 };
    const sundaySession = { date: "2026-03-08", id: 2 };
    const timeline = buildHistoricalRangeTimeline({
      range: "6M",
      items: [mondaySession, sundaySession],
      dateOf: (item) => item.date,
      today,
    });

    expect(
      timeline.buckets.find((bucket) => bucket.date === "2026-03-02"),
    ).toEqual({
      date: "2026-03-02",
      items: [mondaySession, sundaySession],
    });
  });

  it("ignores missing, invalid, and future item dates", () => {
    const current = { date: "2026-03-24", id: 1 };
    const timeline = buildHistoricalRangeTimeline({
      range: "W",
      items: [
        current,
        { date: undefined, id: 2 },
        { date: "not-a-date", id: 3 },
        { date: "2026-02-30", id: 4 },
        { date: "2026-03-25", id: 5 },
      ],
      dateOf: (item) => item.date,
      today,
    });

    expect(timeline.buckets).toHaveLength(7);
    expect(timeline.buckets.flatMap((bucket) => bucket.items)).toEqual([
      current,
    ]);
  });
});
