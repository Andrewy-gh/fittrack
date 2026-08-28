import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AnalyticsPage } from "./analytics-page";

const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }));

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => navigate,
}));

vi.mock("@tanstack/react-query", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-query")>();

  return {
    ...actual,
    useSuspenseQuery: () => ({
      data: [
        {
          id: 1,
          name: "Bench press",
          created_at: "2026-08-21T00:00:00Z",
          updated_at: "2026-08-21T00:00:00Z",
          user_id: "user-1",
        },
        {
          id: 2,
          name: "Squat",
          created_at: "2026-08-21T00:00:00Z",
          updated_at: "2026-08-21T00:00:00Z",
          user_id: "user-1",
        },
      ],
    }),
    useQuery: () => ({ data: undefined, isLoading: false }),
  };
});

vi.mock("@/features/analytics/components/analytics-dashboard", () => ({
  AnalyticsDashboard: ({
    onSelectExercise,
  }: {
    onSelectExercise: (id: number) => void;
  }) => (
    <button
      type="button"
      onClick={() => onSelectExercise(2)}
    >
      Select Squat
    </button>
  ),
}));

describe("AnalyticsPage", () => {
  it("preserves page scroll when exercise selection updates the route", () => {
    render(
      <AnalyticsPage
        exerciseId={1}
        user={null}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Select Squat" }));

    expect(navigate).toHaveBeenCalledWith({
      search: expect.any(Function),
      resetScroll: false,
    });

    const [{ search }] = navigate.mock.calls[0];
    expect(search({ tab: "overview", exerciseId: 1 })).toEqual({
      tab: "overview",
      exerciseId: 2,
    });
  });
});
