import { queryOptions } from "@tanstack/react-query";
import {
  getFeaturesAccess,
  type FeatureaccessFeatureAccessResponse,
} from "@/client";
import "./client-config";

export const AI_CHAT_FEATURE_KEY = "ai_chatbot";

export type FeatureAccessGrant = FeatureaccessFeatureAccessResponse;

type FeatureAccessResponses = {
  200: FeatureAccessGrant[];
};

export function featureAccessQueryOptions() {
  return queryOptions({
    queryKey: ["feature-access"],
    queryFn: ({ signal }) => getFeatureAccess({ signal }),
  });
}

export async function getFeatureAccess(
  options: {
    signal?: AbortSignal;
  } = {},
): Promise<FeatureAccessGrant[]> {
  const response = await getFeaturesAccess<true>({
    signal: options.signal,
    throwOnError: true,
  });

  return response.data as FeatureAccessResponses[200];
}

export function hasAIChatFeatureAccess(grants?: FeatureAccessGrant[]): boolean {
  return (
    grants?.some((grant) => grant.feature_key === AI_CHAT_FEATURE_KEY) ?? false
  );
}
