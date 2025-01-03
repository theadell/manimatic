package llm

import (
	"context"

	"github.com/alecthomas/jsonschema"
	"github.com/openai/openai-go"
)

type Provider interface {
	Generate(ctx context.Context, prompt string) (Response, error)
	ModelID() string
}

type Response struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Warnings    string `json:"warnings"`
	SceneName   string `json:"scene_name"`
	ValidInput  bool   `json:"valid_input"`
}

type ModelsResponse struct {
	Models       []string `json:"models"`
	DefaultModel string   `json:"default_model"`
}

func GenerateSchema[T any]() any {
	reflector := jsonschema.Reflector{AllowAdditionalProperties: false, DoNotReference: true}
	var v T
	return reflector.Reflect(v)
}

var ManimSchema = GenerateSchema[Response]()

var ResponseFormat = openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
	openai.ResponseFormatJSONSchemaParam{
		Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
		JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        openai.F("manim_script_response"),
			Description: openai.F("A response containing code, description, warnings, scene_name, and valid_input fields."),
			Schema:      openai.F(ManimSchema),
			Strict:      openai.Bool(true),
		}),
	},
)
