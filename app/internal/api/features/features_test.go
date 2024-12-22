package features

import (
	"encoding/json"
	"testing"
)

func TestNewFeatures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[FeatureKey]bool
	}{
		{
			name:  "No features enabled",
			input: "",
			expected: map[FeatureKey]bool{
				UserCompile: false,
				HighQuality: false,
			},
		},
		{
			name:  "UserCompile enabled",
			input: "user-compile",
			expected: map[FeatureKey]bool{
				UserCompile: true,
				HighQuality: false,
			},
		},
		{
			name:  "All features enabled",
			input: "user-compile, high-quality",
			expected: map[FeatureKey]bool{
				UserCompile: true,
				HighQuality: true,
			},
		},
		{
			name:  "Input with spaces",
			input: "  user-compile ,   high-quality ",
			expected: map[FeatureKey]bool{
				UserCompile: true,
				HighQuality: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feat := New(tt.input)
			for key, expectedEnabled := range tt.expected {
				if feat.IsEnabled(key) != expectedEnabled {
					t.Errorf("Feature %s enabled state mismatch. Expected %v, got %v", key, expectedEnabled, feat.IsEnabled(key))
				}
			}
		})
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		key      FeatureKey
		expected bool
	}{
		{
			name:     "UserCompile enabled",
			input:    "user-compile",
			key:      UserCompile,
			expected: true,
		},
		{
			name:     "HighQuality disabled",
			input:    "user-compile",
			key:      HighQuality,
			expected: false,
		},
		{
			name:     "All features enabled",
			input:    "user-compile, high-quality",
			key:      HighQuality,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feat := New(tt.input)
			if feat.IsEnabled(tt.key) != tt.expected {
				t.Errorf("Expected %v for feature %s, got %v", tt.expected, tt.key, feat.IsEnabled(tt.key))
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	feat := New("user-compile")
	expectedJSON := `{"version":"0.1.0","features":[{"key":"user-compile","description":"Allows users to edit scripts and compile them with arbitrary input.","enabled":true},{"key":"high-quality","description":"Enables high-quality (4K) rendering of animations.","enabled":false}]}`

	data, err := json.Marshal(feat)
	if err != nil {
		t.Fatalf("Failed to marshal Features to JSON: %v", err)
	}

	actualJSON := string(data)
	if actualJSON != expectedJSON {
		t.Errorf("JSON output mismatch. Expected: %s, Got: %s", expectedJSON, actualJSON)
	}
}
