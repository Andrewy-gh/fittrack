# Workout draft weight acceptance baseline

These reports capture the first passing implementation for the starter cases. They are a post-change baseline. A pre-change run was not captured.

| Check | Result |
| --- | ---: |
| [Luna standalone](luna.json) | 5/5 |
| [Gemini standalone](gemini.json) | 5/5 |
| [Gemini after an injected Luna 503](fallback.json) | 1/1 |

The standalone runs used empty workout history. Both included the 35 lb equipment cap case and the user-reported bench weight case. The third starter case passed separately in the Chromium demo test for clearing, saving and reopening a draft weight while preserving an explicit zero.

From the server directory, with the appropriate provider credentials configured:

```text
go run ./cmd/ai-chat-scenario-sweep --draft-weights --model luna
go run ./cmd/ai-chat-scenario-sweep --draft-weights --model gemini
go run ./cmd/ai-chat-scenario-sweep --draft-weights --force-luna-fallback --scenario weight-explicit
```

Set a distinct FITTRACK_AI_CHAT_SWEEP_OUT path for each run. The report records which model answered each scenario.
