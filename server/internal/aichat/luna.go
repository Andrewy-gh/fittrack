package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

const lunaModel = config.LunaModel

func defineLuna(g *genkit.Genkit) {
	client := openai.NewClient(option.WithAPIKey(strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))), option.WithHTTPClient(http.DefaultClient), option.WithMaxRetries(0), option.WithBaseURL("https://api.openai.com/v1"))
	genkit.DefineModel(g, lunaModel, &ai.ModelOptions{Supports: &ai.ModelSupports{Multiturn: true, SystemRole: true, Tools: true, Constrained: "all", Output: []string{"text", "json"}}}, func(ctx context.Context, req *ai.ModelRequest, cb ai.ModelStreamCallback) (*ai.ModelResponse, error) {
		params, err := lunaRequest(req)
		if err != nil {
			return nil, err
		}
		var httpResponse *http.Response
		stream := client.Responses.NewStreaming(ctx, params, option.WithResponseInto(&httpResponse))
		defer stream.Close()
		for stream.Next() {
			event := stream.Current()
			switch event.Type {
			case "response.output_text.delta":
				if cb != nil && event.Delta != "" {
					if strings.TrimSpace(event.Delta) != "" && !commitProvider(ctx) {
						return nil, errPrimarySlow
					}
					if err := cb(ctx, &ai.ModelResponseChunk{Content: []*ai.Part{ai.NewTextPart(event.Delta)}}); err != nil {
						return nil, err
					}
				}
			case "response.completed":
				result, err := lunaResponse(event.Response)
				if errors.Is(err, errEmptyLunaResponse) {
					markTransientProviderFailure(ctx)
				}
				if err == nil && len(result.Message.Content) > 0 {
					if !commitProvider(ctx) {
						return nil, errPrimarySlow
					}
				}
				return result, err
			case "response.failed", "error":
				code := string(event.Response.Error.Code)
				if event.Type == "error" {
					code = event.Code
				}
				if code == "server_error" || code == "rate_limit_exceeded" {
					markTransientProviderFailure(ctx)
				}
				return nil, fmt.Errorf("OpenAI provider stream failed (%s)", code)
			case "response.incomplete":
				return nil, fmt.Errorf("OpenAI provider stream ended with %s", event.Type)
			}
		}
		if err := stream.Err(); err != nil {
			markProviderError(ctx, httpResponse, err)
			return nil, fmt.Errorf("OpenAI provider request: %w", err)
		}
		markProviderError(ctx, httpResponse, io.ErrUnexpectedEOF)
		return nil, errors.New("OpenAI provider stream ended before completion")
	})
}

func lunaRequest(req *ai.ModelRequest) (responses.ResponseNewParams, error) {
	p := responses.ResponseNewParams{Model: strings.TrimPrefix(lunaModel, "openai/"), Store: openai.Bool(false), Include: []responses.ResponseIncludable{responses.ResponseIncludableReasoningEncryptedContent}, Reasoning: shared.ReasoningParam{Effort: "low"}}
	for _, msg := range req.Messages {
		role := responses.EasyInputMessageRole(msg.Role)
		if msg.Role == ai.RoleModel {
			role = "assistant"
		}
		for _, part := range msg.Content {
			// Genkit retains part metadata between tool calls within this turn.
			if raw, ok := part.Metadata[lunaReasoningKey].(string); ok {
				var items []responses.ResponseReasoningItemParam
				if err := json.Unmarshal([]byte(raw), &items); err != nil {
					return p, fmt.Errorf("decode OpenAI reasoning continuity: %w", err)
				}
				for i := range items {
					p.Input.OfInputItemList = append(p.Input.OfInputItemList, responses.ResponseInputItemUnionParam{OfReasoning: &items[i]})
				}
			}
			var item responses.ResponseInputItemUnionParam
			switch {
			case part.ToolRequest != nil:
				call := part.ToolRequest
				args, err := json.Marshal(call.Input)
				if err != nil {
					return p, err
				}
				item = responses.ResponseInputItemParamOfFunctionCall(string(args), call.Ref, call.Name)
			case part.ToolResponse != nil:
				result := part.ToolResponse
				data, err := json.Marshal(result.Output)
				if err != nil {
					return p, err
				}
				item = responses.ResponseInputItemParamOfFunctionCallOutput(string(data))
				item.OfFunctionCallOutput.CallID = openai.String(result.Ref)
			case part.IsText():
				item = responses.ResponseInputItemParamOfMessage(part.Text, role)
			default:
				return p, fmt.Errorf("unsupported FitTrack chat message part")
			}
			p.Input.OfInputItemList = append(p.Input.OfInputItemList, item)
		}
	}
	for _, tool := range req.Tools {
		t := responses.ToolParamOfFunction(tool.Name, tool.InputSchema, false)
		t.OfFunction.Description = openai.String(tool.Description)
		p.Tools = append(p.Tools, t)
	}
	if req.Output != nil && len(req.Output.Schema) > 0 {
		p.Text.Format.OfJSONSchema = &responses.ResponseFormatTextJSONSchemaConfigParam{Name: "workout_draft", Schema: req.Output.Schema, Strict: openai.Bool(false)}
	}
	return p, nil
}

var errEmptyLunaResponse = errors.New("OpenAI provider returned no text or tool calls")

const lunaReasoningKey = "fittrack_openai_reasoning"

func lunaResponse(resp responses.Response) (*ai.ModelResponse, error) {
	msg := &ai.Message{Role: ai.RoleModel}
	var reasoning []json.RawMessage
	for _, out := range resp.Output {
		start := len(msg.Content)
		switch out.Type {
		case "reasoning":
			reasoning = append(reasoning, json.RawMessage(out.RawJSON()))
		case "message":
			for _, content := range out.Content {
				switch content.Type {
				case "output_text":
					if strings.TrimSpace(content.Text) != "" {
						msg.Content = append(msg.Content, ai.NewTextPart(content.Text))
					}
				case "refusal":
					if strings.TrimSpace(content.Refusal) != "" {
						msg.Content = append(msg.Content, ai.NewTextPart(content.Refusal))
					}
				}
			}
		case "function_call":
			var args map[string]any
			if err := json.Unmarshal([]byte(out.AsFunctionCall().Arguments), &args); err != nil {
				return nil, fmt.Errorf("invalid arguments for %s: %w", out.Name, err)
			}
			msg.Content = append(msg.Content, ai.NewToolRequestPart(&ai.ToolRequest{Ref: out.CallID, Name: out.Name, Input: args}))
		}
		if len(reasoning) > 0 && len(msg.Content) > start {
			encoded, err := json.Marshal(reasoning)
			if err != nil {
				return nil, err
			}
			msg.Content[start].Metadata = map[string]any{lunaReasoningKey: string(encoded)}
			reasoning = nil
		}
	}
	slog.Info("OpenAI chat request completed", "model", resp.Model, "input_tokens", resp.Usage.InputTokens, "output_tokens", resp.Usage.OutputTokens)
	if len(msg.Content) == 0 {
		return nil, errEmptyLunaResponse
	}
	return &ai.ModelResponse{Message: msg, FinishReason: ai.FinishReasonStop}, nil
}
