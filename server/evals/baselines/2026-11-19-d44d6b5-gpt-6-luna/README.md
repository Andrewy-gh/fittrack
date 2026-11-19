# Luna chat baseline, November 19, 2026

The model under evaluation was `openai/gpt-6-luna` at FitTrack commit `d44d6b5c68d634f14a15cc0bf1fd102fc69807e8` (PR #372). These are the first complete runs of each pack in this baseline. The fixture pack overlaps the default pack. The prior Gemini baseline is retained.

| Two-turn pack | Passed | Failed | Operational errors | Unscored |
| --- | ---: | ---: | ---: | ---: |
| [Default workout scenarios](baseline-two-turn.json) | 20 | 0 | 0 | 0 |
| [Workout history and profile fixtures](fixtures-two-turn.json) | 20 | 2 | 0 | 0 |

In the fixture pack, Luna generated a workout immediately for the combined workout and meal-plan request (`prompt-18`). The expected first reply asks the user to narrow the request. For `profile-02`, Luna called the required profile-update tool. The text check failed because its reply omitted the required term “remember”.

## Reproduction

Build `./cmd/ai-chat-scenario-sweep` from the evaluated commit’s `server/` directory. Set `AI_CHAT_PRIMARY_MODEL=openai/gpt-6-luna`, disable `AI_CHAT_FALLBACK`, and supply the OpenAI credential for generation and the Gemini credential for the existing narrow-scope judge. Choose a distinct `FITTRACK_AI_CHAT_SWEEP_OUT` path for each run.

```text
ai-chat-scenario-sweep -mode two_turn -scenario-delay 0 -timeout 45m
ai-chat-scenario-sweep -mode two_turn -with-data-fixtures -scenario-delay 0 -timeout 45m
```

The two packs ran concurrently. Recorded timing is observational and may include contention. The scores measure one run per pack.
