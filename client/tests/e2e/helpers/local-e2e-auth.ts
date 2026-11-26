import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { WorkoutCreateWorkoutRequest } from "../../../src/client/types.gen";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export const authStatePath = path.join(__dirname, "..", ".auth", "stack.json");

export function hasAuthState(): boolean {
  return existsSync(authStatePath);
}

export type SeededConversationRequest = {
  title: string;
  user_prompt: string;
  assistant_text: string;
  latest_workout_draft: WorkoutCreateWorkoutRequest;
  expire_ai_chat_access_after_seed?: boolean;
};

export async function seedLocalAIChatConversation(
  request: SeededConversationRequest,
): Promise<{ conversation_id: number }> {
  const apiBaseUrl =
    process.env.E2E_LOCAL_AUTH_API_BASE_URL ?? "http://localhost:8080";
  const response = await fetch(`${apiBaseUrl}/dev/e2e/ai-chat/conversations`, {
    method: "POST",
    headers: {
      "content-type": "application/json",
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    throw new Error(
      `Failed to seed local AI chat conversation: ${response.status} ${await response.text()}`,
    );
  }

  return (await response.json()) as { conversation_id: number };
}

export async function restoreLocalAIChatAccess(): Promise<void> {
  const apiBaseUrl =
    process.env.E2E_LOCAL_AUTH_API_BASE_URL ?? "http://localhost:8080";
  const response = await fetch(`${apiBaseUrl}/dev/e2e/auth/bootstrap`, {
    method: "POST",
    headers: {
      "content-type": "application/json",
    },
  });
  if (!response.ok) {
    throw new Error(
      `Failed to restore local AI chat access: ${response.status} ${await response.text()}`,
    );
  }
}
