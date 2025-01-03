package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"manimatic/internal/llm"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Model string

const (
	ChatModelGPT4o     = Model(openai.ChatModelGPT4o)
	ChatModelGPT4oMini = Model(openai.ChatModelGPT4oMini)
)

var defaultModels = []Model{
	ChatModelGPT4o,
	ChatModelGPT4oMini,
}

type provider struct {
	client  *openai.Client
	modelID string
}

func RegisterWith(service *llm.Service, apiKey string) {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	for _, model := range defaultModels {
		p := &provider{
			client:  client,
			modelID: string(model),
		}
		service.RegisterProvider(p)
	}
}

func (p *provider) ModelID() string {
	return p.modelID
}

func (p *provider) Generate(ctx context.Context, prompt string) (llm.Response, error) {
	resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(llm.DefaultSystemPrompt),
			openai.UserMessage(prompt),
		}),
		ResponseFormat: llm.ResponseFormat,
		Model:          openai.F(p.modelID),
	})
	if err != nil {
		return llm.Response{}, fmt.Errorf("openai api call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return llm.Response{}, fmt.Errorf("no response choices returned")
	}

	var result llm.Response
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return llm.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}
