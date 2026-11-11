import { afterEach, expect, it, vi } from "vitest";
import { getInitialValues } from "./form-options";
import { createWorkoutDraftStorage } from "@/lib/local-storage";

afterEach(() => vi.useRealTimers());

it("starts new drafts with the current date after rollover while retaining restored dates", () => {
  vi.useFakeTimers({ toFake: ["Date"] });
  const entries = new Map<string, string>();
  const storage = createWorkoutDraftStorage({
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => {
      entries.set(key, value);
    },
    removeItem: (key) => {
      entries.delete(key);
    },
  });
  vi.setSystemTime(new Date("2026-10-08T23:59:00Z"));
  expect(getInitialValues("owner", storage).date).toBe(
    "2026-10-08T23:59:00.000Z",
  );
  vi.setSystemTime(new Date("2026-10-09T00:01:00Z"));
  expect(getInitialValues("owner", storage).date).toBe(
    "2026-10-09T00:01:00.000Z",
  );
  storage.save({ date: "2026-10-01T12:00:00Z", exercises: [] }, "owner");
  expect(getInitialValues("owner", storage).date).toBe("2026-10-01T12:00:00Z");
});
