package genmanim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type LLMManimService struct {
	client *openai.Client
}

func NewLLMManimService(apiKey string) *LLMManimService {

	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &LLMManimService{client: client}
}

func (s *LLMManimService) GenerateScript(ctx context.Context, userPrompt string, moderationOn bool) (ManimScriptResponse, error) {
	if moderationOn {
		safe, err := s.runModeration(ctx, userPrompt)
		if err != nil {
			return ManimScriptResponse{}, fmt.Errorf("moderation failed: %w", err)
		}
		if !safe {
			return ManimScriptResponse{
				ValidInput:  false,
				Description: "The request cannot be fulfilled due to disallowed content.",
				Warnings:    "Content flagged by moderation.",
			}, nil
		}
	}

	return s.generateManimResponse(ctx, userPrompt)
}

func (s *LLMManimService) runModeration(ctx context.Context, prompt string) (bool, error) {
	modResp, err := s.client.Moderations.New(ctx, openai.ModerationNewParams{
		Input: openai.Raw[openai.ModerationNewParamsInputUnion](prompt),
	})
	if err != nil {
		return false, err
	}

	if len(modResp.Results) > 0 && modResp.Results[0].Flagged {
		return false, nil
	}

	return true, nil
}

const systemPrompt = `
You are an assistant that generates Manim code based on a user prompt.
You MUST return exactly one JSON object that conforms to the given JSON schema:
- code: The full Python Manim script if valid_input is true; empty string if not valid.
- description: A brief explanation of what the script does (or why it's invalid).
- warnings: Any warnings, assumptions, or reasons for invalidity.
- scene_name: The primary scene class name if valid; otherwise empty if invalid.
- valid_input: True if the user's prompt can be turned into a Manim animation; false if unrelated or disallowed.

No additional text outside the JSON. No markdown formatting.
If the user's request is unrelated to Manim or not actionable, set valid_input to false, provide a helpful description and possibly warnings, and leave code empty.
If valid_input is true, the code should have:
- A docstring at the top.
- Necessary imports from manim.
- A Scene class with a construct method implementing the animation.
- Comments explaining key steps in the code.
`

func (s *LLMManimService) generateManimResponse(ctx context.Context, prompt string) (ManimScriptResponse, error) {

	chatCompletion, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(prompt),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("manim_script_response"),
					Description: openai.F("A response containing code, description, warnings, scene_name, and valid_input fields."),
					Schema:      openai.F(ManimSchema),
					Strict:      openai.Bool(true),
				}),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o2024_08_06),
	})
	if err != nil {
		return ManimScriptResponse{}, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return ManimScriptResponse{}, errors.New("no choices returned by the API")
	}

	var resp ManimScriptResponse
	if err := json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &resp); err != nil {
		return ManimScriptResponse{}, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return resp, nil
}
