import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

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
  latest_workout_draft: {
    date: string;
    notes?: string;
    workoutFocus?: string;
    exercises: Array<{
      name: string;
      sets: Array<{
        reps: number;
        setType: "warmup" | "working";
        weight?: number;
      }>;
    }>;
  };
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
