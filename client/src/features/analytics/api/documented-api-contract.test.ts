import { transferableAbortController } from "node:util";
import { QueryClient } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

// Keep these transport tests local even on machines with provider credentials.
vi.hoisted(() => {
  vi.stubEnv("VITE_PROJECT_ID", "");
  vi.stubEnv("VITE_PUBLISHABLE_CLIENT_KEY", "");
});

import { client } from "@/client/client.gen";
import { deleteAccount } from "@/features/account/api/account";
import {
  createBillingCheckoutSession,
  createBillingCustomerPortalSession,
  getBillingStatus,
} from "@/features/chat/api/billing";
import {
  getTrainingProfile,
  updateTrainingProfile,
  type TrainingProfile,
} from "@/features/training-profile/api/training-profile";
import {
  muscleContributionsQueryOptions,
  type MuscleContributions,
} from "./muscle-contributions";
import {
  trainingAssessmentQueryOptions,
  type StrengthAssessment,
} from "./training-assessment";

const initialConfig = client.getConfig();
beforeEach(() => {
  // Query cancellation must use the same AbortSignal realm as Node's Request.
  vi.stubGlobal(
    "AbortController",
    class {
      constructor() {
        return transferableAbortController();
      }
    },
  );
});
afterEach(() => {
  client.setConfig(initialConfig);
  vi.unstubAllGlobals();
});

const period = {
  startDate: "2026-09-14",
  endDate: "2026-09-21",
  timezone: "America/New_York",
  startAt: "2026-09-14T04:00:00Z",
  endAt: "2026-09-21T04:00:00Z",
  observedAt: "2026-09-21T04:00:00Z",
  partial: false,
};

it("loads generated analytics contracts through account-scoped query callers", async () => {
  const muscleSummary: MuscleContributions = {
    period,
    workingSets: 0,
    invalidWorkingSets: 0,
    unclassifiedSets: 0,
    unreviewedSets: 0,
    mappingVersion: "muscle-contributions-v1",
    catalogVersion: "local-fixture",
    classificationBasis: "current_exercise_catalog_link",
    muscles: [],
    exercises: [],
  };
  const assessment: StrengthAssessment = {
    exerciseId: 42,
    period,
    sessions: [],
    workingSets: 0,
    trainingDays: 0,
    sessionsMeetingReference: 0,
    policyVersion: "strength-session-reference-v1",
    reference: "General reference",
    sourceUrl: "https://example.test/reference",
    applicability: "General adults",
    limitations: ["Logged training only"],
  };
  const requests: Request[] = [];
  client.setConfig({
    baseUrl: "http://localhost/api",
    fetch: async (input, init) => {
      const request = new Request(input, init);
      requests.push(request);
      return Response.json(
        request.url.includes("/assessment") ? assessment : muscleSummary,
      );
    },
  });
  const queries = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  try {
    const muscleOptions = muscleContributionsQueryOptions(
      "account-a",
      period.startDate,
      period.timezone,
    );
    const assessmentOptions = trainingAssessmentQueryOptions(
      "account-a",
      42,
      period.startDate,
      period.timezone,
    );
    expect(await queries.fetchQuery(muscleOptions)).toEqual(muscleSummary);
    expect(await queries.fetchQuery(assessmentOptions)).toEqual(assessment);
    expect(
      queries.getQueryData([
        "muscle-contributions",
        "account-a",
        period.startDate,
        period.timezone,
      ]),
    ).toEqual(muscleSummary);
    expect(
      queries.getQueryData([
        "training-assessment",
        "account-a",
        42,
        period.startDate,
        period.timezone,
      ]),
    ).toEqual(assessment);
    expect(
      queries.getQueryData([
        "muscle-contributions",
        "account-b",
        period.startDate,
        period.timezone,
      ]),
    ).toBeUndefined();
    expect(requests.map((request) => new URL(request.url).pathname)).toEqual([
      "/api/analytics/muscle-contributions",
      "/api/exercises/42/assessment",
    ]);
    for (const request of requests) {
      expect(request.method).toBe("GET");
      expect(new URL(request.url).searchParams.get("startDate")).toBe(
        period.startDate,
      );
      expect(new URL(request.url).searchParams.get("timezone")).toBe(
        period.timezone,
      );
    }
  } finally {
    queries.clear();
  }
});

it("preserves omitted, null and empty profile fields on the wire", async () => {
  const profile: TrainingProfile = {
    primary_goal: null,
    goals: [],
    experience_level: null,
    preferred_session_duration_minutes: null,
    usual_training_location: null,
    available_equipment: [],
    avoided_exercises: [],
    movement_limitations: null,
  };
  const bodies: unknown[] = [];
  client.setConfig({
    baseUrl: "http://localhost/api",
    fetch: async (input, init) => {
      const request = new Request(input, init);
      expect(new URL(request.url).pathname).toBe("/api/training-profile");
      if (request.method === "PUT") bodies.push(await request.json());
      return Response.json(profile);
    },
  });
  expect(await getTrainingProfile()).toEqual(profile);
  await updateTrainingProfile({});
  await updateTrainingProfile({ movement_limitations: null });
  await updateTrainingProfile({ movement_limitations: [] });
  expect(bodies).toEqual([
    {},
    { movement_limitations: null },
    { movement_limitations: [] },
  ]);
});

for (const [name, call, status] of [
  ["billing status", getBillingStatus, 503],
  ["checkout quota", createBillingCheckoutSession, 403],
  ["billing portal", createBillingCustomerPortalSession, 404],
  ["account deletion", deleteAccount, 500],
  ["expired account session", deleteAccount, 401],
  ["profile read", getTrainingProfile, 401],
] as const) {
  it(`retains structured errors for ${name}`, async () => {
    const error = {
      message: `Local ${status} failure`,
      request_id: "local-contract",
    };
    client.setConfig({
      baseUrl: "http://localhost/api",
      fetch: async () => Response.json(error, { status }),
    });
    await expect(call()).rejects.toEqual(error);
  });
}
