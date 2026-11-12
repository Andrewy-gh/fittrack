import {
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type {
  TrainingprofileProfileResponse,
  TrainingprofileUpdateProfileRequest,
  GetTrainingProfileResponses,
  PutTrainingProfileResponses,
} from "@/client/types.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

/** Profile returned by the server, including explicit nullable fields. */
export type TrainingProfile = TrainingprofileProfileResponse;

/** Server-supported profile update fields. */
export type UpdateTrainingProfileRequest = TrainingprofileUpdateProfileRequest;

/** Keeps authenticated profile reads scoped to the current account. */
export function trainingProfileQueryOptions(userId: string) {
  return queryOptions({
    queryKey: ["training-profile", userId],
    queryFn: ({ signal }) => getTrainingProfile({ signal }),
  });
}

export async function getTrainingProfile(
  options: {
    signal?: AbortSignal;
  } = {},
): Promise<TrainingProfile> {
  const response = await client.get<
    GetTrainingProfileResponses,
    ApiError,
    true
  >({
    url: "/training-profile",
    signal: options.signal,
    throwOnError: true,
  });

  return response.data;
}

export async function updateTrainingProfile(
  body: UpdateTrainingProfileRequest,
): Promise<TrainingProfile> {
  const response = await client.put<
    PutTrainingProfileResponses,
    ApiError,
    true
  >({
    url: "/training-profile",
    body,
    throwOnError: true,
  });

  return response.data;
}

/** Updates the current provider's cache for the account whose form was saved. */
export function useUpdateTrainingProfileMutation(userId: string) {
  const queryClient = useQueryClient();
  const trainingProfileQueryKey = ["training-profile", userId];
  return useMutation({
    mutationFn: updateTrainingProfile,
    meta: { skipGlobalErrorHandler: true },
    onSuccess: (profile) => {
      queryClient.setQueryData(trainingProfileQueryKey, profile);
      queryClient.invalidateQueries({ queryKey: trainingProfileQueryKey });
    },
  });
}
