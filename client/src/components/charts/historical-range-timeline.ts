import {
  addDays,
  addMonths,
  differenceInCalendarDays,
  differenceInCalendarMonths,
  format,
  startOfMonth,
  startOfWeek,
} from "date-fns";

import type { RangeType } from "./chart-bar-vol.utils";

type HistoricalBucket = "day" | "week" | "month";

type HistoricalRangeTimelineBucket<T> = {
  readonly date: string;
  readonly items: readonly T[];
};

type HistoricalRangeTimeline<T> = {
  readonly bucket: HistoricalBucket;
  readonly buckets: readonly HistoricalRangeTimelineBucket<T>[];
};

type HistoricalRangeTimelineOptions<T> = {
  readonly range: RangeType;
  readonly items: readonly T[];
  readonly dateOf: (item: T) => string | null | undefined;
  readonly today?: Date;
};

const toIsoDate = (date: Date) => format(date, "yyyy-MM-dd");
const monday = (date: Date) => startOfWeek(date, { weekStartsOn: 1 });

function getBucket(range: RangeType): HistoricalBucket {
  return range === "Y" ? "month" : range === "6M" ? "week" : "day";
}

function alignToBucket(date: Date, bucket: HistoricalBucket): Date {
  if (bucket === "month") return startOfMonth(date);
  if (bucket === "week") return monday(date);
  return date;
}

function getDefaultStart(
  end: Date,
  range: RangeType,
  bucket: HistoricalBucket,
): Date {
  if (bucket === "month") return end;
  if (bucket === "week") return addDays(end, -25 * 7);
  return addDays(end, range === "M" ? -29 : -6);
}

function getBucketCount(
  start: Date,
  end: Date,
  bucket: HistoricalBucket,
): number {
  if (bucket === "month") {
    return differenceInCalendarMonths(end, start) + 1;
  }
  if (bucket === "week") {
    return Math.floor(differenceInCalendarDays(end, start) / 7) + 1;
  }
  return differenceInCalendarDays(end, start) + 1;
}

function addBuckets(
  start: Date,
  amount: number,
  bucket: HistoricalBucket,
): Date {
  if (bucket === "month") return addMonths(start, amount);
  return addDays(start, amount * (bucket === "week" ? 7 : 1));
}

function parseCalendarDate(value: string | null | undefined): Date | undefined {
  const calendarDate = value?.split("T")[0];
  if (!calendarDate) return undefined;

  const parsed = new Date(`${calendarDate}T00:00:00`);
  if (Number.isNaN(parsed.getTime()) || toIsoDate(parsed) !== calendarDate) {
    return undefined;
  }

  return parsed;
}

/**
 * Builds a continuous chart timeline that extends from the selected range's
 * minimum span back to the earliest valid historical item.
 *
 * Items are grouped into daily, Monday-aligned weekly, or monthly buckets.
 * Missing, invalid, and future dates do not affect the timeline.
 */
export function buildHistoricalRangeTimeline<T>({
  range,
  items,
  dateOf,
  today = new Date(),
}: HistoricalRangeTimelineOptions<T>): HistoricalRangeTimeline<T> {
  const bucket = getBucket(range);
  const todayStart = new Date(
    today.getFullYear(),
    today.getMonth(),
    today.getDate(),
  );
  const end = alignToBucket(todayStart, bucket);
  const itemsByBucket = new Map<string, T[]>();
  let firstHistoricalBucket: Date | undefined;

  for (const item of items) {
    const date = parseCalendarDate(dateOf(item));
    if (!date || date > todayStart) continue;

    const bucketDate = alignToBucket(date, bucket);
    if (!firstHistoricalBucket || bucketDate < firstHistoricalBucket) {
      firstHistoricalBucket = bucketDate;
    }

    const key = toIsoDate(bucketDate);
    const bucketItems = itemsByBucket.get(key) ?? [];
    bucketItems.push(item);
    itemsByBucket.set(key, bucketItems);
  }

  const defaultStart = getDefaultStart(end, range, bucket);
  const start =
    firstHistoricalBucket && firstHistoricalBucket < defaultStart
      ? firstHistoricalBucket
      : defaultStart;
  const count = getBucketCount(start, end, bucket);

  const buckets = Array.from({ length: count }, (_, index) => {
    const date = toIsoDate(addBuckets(start, index, bucket));
    return {
      date,
      items: itemsByBucket.get(date) ?? [],
    };
  });

  return { bucket, buckets };
}
