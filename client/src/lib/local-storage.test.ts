import { describe, expect, it } from "vitest";
import { createWorkoutDraftStorage, type StorageLike } from "./local-storage";

function createMemoryStorage(
  initialEntries: Record<string, string> = {},
): StorageLike {
  const store = new Map(Object.entries(initialEntries));

  return {
    getItem(key) {
      return store.get(key) ?? null;
    },
    setItem(key, value) {
      store.set(key, value);
    },
    removeItem(key) {
      store.delete(key);
    },
  };
}

describe("workoutDraftStorage", () => {
  it("round-trips draft data through an injected storage backend", () => {
    const storage = createMemoryStorage();
    const draftStorage = createWorkoutDraftStorage(storage);
    const date = new Date("2026-03-24T10:30:00.000Z");

    draftStorage.save(
      {
        date,
        notes: "Heavy day",
        workoutFocus: "Upper",
        exercises: [],
      },
      "user-123",
    );

    const loaded = draftStorage.load("user-123");

    expect(loaded).toEqual({
      date,
      notes: "Heavy day",
      workoutFocus: "Upper",
      exercises: [],
    });
  });

  it("keeps drafts scoped by user id", () => {
    const storage = createMemoryStorage();
    const draftStorage = createWorkoutDraftStorage(storage);

    draftStorage.save(
      {
        date: new Date("2026-03-24T10:30:00.000Z"),
        notes: "User A",
        workoutFocus: "",
        exercises: [],
      },
      "user-a",
    );
    draftStorage.save(
      {
        date: new Date("2026-03-24T11:30:00.000Z"),
        notes: "User B",
        workoutFocus: "",
        exercises: [],
      },
      "user-b",
    );

    expect(draftStorage.load("user-a")?.notes).toBe("User A");
    expect(draftStorage.load("user-b")?.notes).toBe("User B");
  });

  it("clears drafts from the injected storage backend", () => {
    const storage = createMemoryStorage();
    const draftStorage = createWorkoutDraftStorage(storage);

    draftStorage.save(
      {
        date: new Date("2026-03-24T10:30:00.000Z"),
        notes: "To delete",
        workoutFocus: "",
        exercises: [],
      },
      "user-123",
    );
    draftStorage.clear("user-123");

    expect(draftStorage.load("user-123")).toBeNull();
  });

  it("returns null for malformed stored JSON", () => {
    const storage = createMemoryStorage({
      "workout-entry-form-data-user-123": "{bad json",
    });
    const draftStorage = createWorkoutDraftStorage(storage);

    expect(draftStorage.load("user-123")).toBeNull();
  });

  it("returns saved draft data from valid stored JSON", () => {
    const storage = createMemoryStorage({
      "workout-entry-form-data-user-123": JSON.stringify({
        date: "2026-03-24T10:30:00.000Z",
        notes: "Saved draft",
        workoutFocus: "Upper",
        exercises: [
          {
            name: "Bench Press",
            sets: [{ reps: 5, setType: "working", weight: 185 }],
          },
        ],
      }),
    });
    const draftStorage = createWorkoutDraftStorage(storage);

    expect(draftStorage.load("user-123")).toEqual({
      date: new Date("2026-03-24T10:30:00.000Z"),
      notes: "Saved draft",
      workoutFocus: "Upper",
      exercises: [
        {
          name: "Bench Press",
          sets: [{ reps: 5, setType: "working", weight: 185 }],
        },
      ],
    });
  });

  it("returns null for structurally invalid stored drafts", () => {
    const storage = createMemoryStorage({
      "workout-entry-form-data-user-123": JSON.stringify({
        date: "2026-03-24T10:30:00.000Z",
        notes: "Bad draft",
        workoutFocus: "Upper",
        exercises: [
          {
            name: "Bench Press",
            sets: [{ reps: "5", setType: "working", weight: 185 }],
          },
        ],
      }),
    });
    const draftStorage = createWorkoutDraftStorage(storage);

    expect(draftStorage.load("user-123")).toBeNull();
  });
});
