import { queryOptions } from "@tanstack/react-query";
import {
  getFeaturesAccess,
  type FeatureaccessFeatureAccessResponse,
} from "@/client";
import "@/lib/api/client-config";

export const AI_CHAT_FEATURE_KEY = "ai_chatbot";

export type FeatureAccessGrant = FeatureaccessFeatureAccessResponse;

export function featureAccessQueryOptions(userId?: string) {
  return queryOptions({
    queryKey: ["feature-access", userId],
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

  if (!response.data) {
    throw new Error("Feature access response did not include a grant list");
  }

  return response.data;
}

export function hasAIChatFeatureAccess(grants?: FeatureAccessGrant[]): boolean {
  return (
    grants?.some((grant) => grant.feature_key === AI_CHAT_FEATURE_KEY) ?? false
  );
}
