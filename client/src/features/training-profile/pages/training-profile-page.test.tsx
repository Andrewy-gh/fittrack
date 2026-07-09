import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { TrainingProfile } from "@/features/training-profile/api/training-profile";

const { mockGetTrainingProfile, mockUpdateTrainingProfile } = vi.hoisted(
  () => ({
    mockGetTrainingProfile: vi.fn(),
    mockUpdateTrainingProfile: vi.fn(),
  }),
);

const { mockToastSuccess } = vi.hoisted(() => ({
  mockToastSuccess: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: mockToastSuccess,
  },
}));

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    children,
    to,
    className,
  }: {
    children: ReactNode;
    to: string;
    className?: string;
  }) => (
    <a
      href={to}
      className={className}
    >
      {children}
    </a>
  ),
}));

vi.mock("@/features/training-profile/api/training-profile", async () => {
  const reactQuery = await vi.importActual<
    typeof import("@tanstack/react-query")
  >("@tanstack/react-query");

  return {
    trainingProfileQueryOptions: () =>
      reactQuery.queryOptions({
        queryKey: ["training-profile"],
        queryFn: () => mockGetTrainingProfile(),
      }),
    useUpdateTrainingProfileMutation: () =>
      reactQuery.useMutation({
        mutationFn: mockUpdateTrainingProfile,
      }),
  };
});

import { TrainingProfilePage } from "@/features/training-profile/pages/training-profile-page";

const emptyProfile: TrainingProfile = {
  primary_goal: null,
  experience_level: null,
  preferred_session_duration_minutes: null,
  usual_training_location: null,
  available_equipment: [],
  avoided_exercises: [],
  movement_limitations: null,
};

function renderPage(profile: TrainingProfile = emptyProfile) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  mockGetTrainingProfile.mockResolvedValue(profile);
  mockUpdateTrainingProfile.mockImplementation(async (payload) => payload);

  render(
    <QueryClientProvider client={queryClient}>
      <TrainingProfilePage />
    </QueryClientProvider>,
  );
}

function latestUpdatePayload(): TrainingProfile {
  return mockUpdateTrainingProfile.mock.calls.at(-1)?.[0] as TrainingProfile;
}

describe("TrainingProfilePage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the frozen intro copy and enum options", async () => {
    renderPage();

    expect(
      await screen.findByText(
        "FitTrack's AI coach uses this profile to personalize workouts — it fills in your usual setup so you don't repeat yourself in chat. The AI updates it when you tell it something lasting; you can correct anything here.",
      ),
    ).toBeInTheDocument();

    const primaryGoal = screen.getByLabelText("Primary goal");
    expect(within(primaryGoal).getByRole("option", { name: "Not set" }));
    for (const option of [
      "Strength",
      "Hypertrophy",
      "Endurance",
      "General fitness",
      "Weight loss",
      "Mobility",
    ]) {
      expect(within(primaryGoal).getByRole("option", { name: option }));
    }

    const experience = screen.getByLabelText("Experience level");
    for (const option of ["Not set", "Beginner", "Intermediate", "Advanced"]) {
      expect(within(experience).getByRole("option", { name: option }));
    }
  });

  it("keeps save disabled until dirty and shows a success toast", async () => {
    const user = userEvent.setup();
    renderPage();

    const save = await screen.findByRole("button", { name: "Save" });
    expect(save).toBeDisabled();

    await user.selectOptions(screen.getByLabelText("Primary goal"), "strength");
    expect(save).toBeEnabled();
    await user.click(save);

    await waitFor(() => {
      expect(latestUpdatePayload()).toEqual(
        expect.objectContaining({ primary_goal: "strength" }),
      );
    });
    expect(mockToastSuccess).toHaveBeenCalledWith("Training profile saved");
  });

  it("sends null when movement limitations are not specified", async () => {
    const user = userEvent.setup();
    renderPage({
      ...emptyProfile,
      movement_limitations: [],
    });

    await user.click(
      await screen.findByRole("radio", { name: "Not specified" }),
    );
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(latestUpdatePayload()).toEqual(
        expect.objectContaining({ movement_limitations: null }),
      );
    });
  });

  it("sends an empty array when there are no known limitations", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(
      await screen.findByRole("radio", { name: "No known limitations" }),
    );
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(latestUpdatePayload()).toEqual(
        expect.objectContaining({ movement_limitations: [] }),
      );
    });
  });

  it("sends the limitation list when limitations are provided", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(
      await screen.findByRole("radio", { name: "I have limitations:" }),
    );
    await user.type(
      screen.getByLabelText("Movement limitation details"),
      "knee pain{Enter}",
    );
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(latestUpdatePayload()).toEqual(
        expect.objectContaining({ movement_limitations: ["knee pain"] }),
      );
    });
  });
});
