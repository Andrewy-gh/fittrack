import { queryOptions, useMutation } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type { ApiError } from "@/lib/errors";
import { queryClient } from "@/lib/api/api";
import "@/lib/api/client-config";

export type TrainingProfile = {
  primary_goal: string | null;
  experience_level: string | null;
  preferred_session_duration_minutes: number | null;
  usual_training_location: string | null;
  available_equipment: string[];
  avoided_exercises: string[];
  movement_limitations: string[] | null;
};

export type UpdateTrainingProfileRequest = TrainingProfile;

type TrainingProfileResponses = {
  200: TrainingProfile;
};

const trainingProfileQueryKey = ["training-profile"] as const;

export function trainingProfileQueryOptions() {
  return queryOptions({
    queryKey: trainingProfileQueryKey,
    queryFn: ({ signal }) => getTrainingProfile({ signal }),
  });
}

export async function getTrainingProfile(
  options: {
    signal?: AbortSignal;
  } = {},
): Promise<TrainingProfile> {
  const response = await client.get<TrainingProfileResponses, ApiError, true>({
    url: "/training-profile",
    signal: options.signal,
    throwOnError: true,
  });

  return response.data;
}

export async function updateTrainingProfile(
  body: UpdateTrainingProfileRequest,
): Promise<TrainingProfile> {
  const response = await client.put<TrainingProfileResponses, ApiError, true>({
    url: "/training-profile",
    body,
    throwOnError: true,
  });

  return response.data;
}

export function useUpdateTrainingProfileMutation() {
  return useMutation({
    mutationFn: updateTrainingProfile,
    meta: { skipGlobalErrorHandler: true },
    onSuccess: (profile) => {
      queryClient.setQueryData(trainingProfileQueryKey, profile);
      queryClient.invalidateQueries({ queryKey: trainingProfileQueryKey });
    },
  });
}
