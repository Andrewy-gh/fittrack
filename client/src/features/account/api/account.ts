import { client } from "@/client/client.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

type DeleteAccountResponses = {
  204: void;
};

export async function deleteAccount(): Promise<void> {
  await client.delete<DeleteAccountResponses, ApiError, true>({
    url: "/account",
    throwOnError: true,
  });
}
