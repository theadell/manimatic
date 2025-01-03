package xai

import (
	"context"
	"encoding/json"
	"fmt"
	"manimatic/internal/llm"
	"net/http"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Model string

const (
	Grok2Latest Model  = "grok-2-latest"
	baseURL     string = "https://api.x.ai"
)

var defaultModels = []Model{
	Grok2Latest,
}

func urlMiddleware(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
	if !strings.Contains(req.URL.Path, "/v1") {
		req.URL.Path = "/v1" + req.URL.Path
	}
	return next(req)
}

type provider struct {
	client  *openai.Client // Using OpenAI client since API is compatible
	modelID string
}

func RegisterWith(service *llm.Service, apiKey string) error {

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
		option.WithMiddleware(urlMiddleware),
	)

	for _, model := range defaultModels {
		p := &provider{
			client:  client,
			modelID: string(model),
		}
		service.RegisterProvider(p)
	}

	return nil
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
		return llm.Response{}, fmt.Errorf("xai api call failed: %w", err)
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
