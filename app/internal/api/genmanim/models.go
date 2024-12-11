package genmanim

import "github.com/alecthomas/jsonschema"

type ManimScriptResponse struct {
	Code        string `json:"code" jsonschema_description:"The complete Manim script, empty if the input is unrelated."`
	Description string `json:"description" jsonschema_description:"A brief explanation of the script's functionality."`
	Warnings    string `json:"warnings" jsonschema_description:"Any warnings or assumptions made."`
	SceneName   string `json:"scene_name" jsonschema_description:"The name of the primary scene class."`
	ValidInput  bool   `json:"valid_input" jsonschema_description:"Indicates if the input is valid for creating Manim scripts."`
}

func GenerateSchema[T any]() any {
	reflector := jsonschema.Reflector{AllowAdditionalProperties: false, DoNotReference: true}
	var v T
	return reflector.Reflect(v)
}

// Precompute the schema for ManimScriptResponse.
var ManimSchema = GenerateSchema[ManimScriptResponse]()
