package features

import (
	"encoding/json"
	"fmt"
	"strings"
)

type FeatureKey string

const (
	UserCompile FeatureKey = "user-compile"
	HighQuality FeatureKey = "high-quality"
)

type Feature struct {
	Key         FeatureKey `json:"key"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
}

type Features struct {
	featureList []Feature
	version     string
}

func (f *Features) String() string {
	var enabledFeatures []string
	for _, feature := range f.featureList {
		if feature.Enabled {
			enabledFeatures = append(enabledFeatures, string(feature.Key))
		}
	}
	return fmt.Sprintf("Enabled features: [%s], Version: %s", strings.Join(enabledFeatures, ", "), f.version)
}

var features = []Feature{
	{Key: UserCompile, Description: "Allows users to edit scripts and compile them with arbitrary input."},
	{Key: HighQuality, Description: "Enables high-quality (4K) rendering of animations."},
}

func New(input string) *Features {

	enabledKeys := parseInput(input)

	for i := range features {
		features[i].Enabled = enabledKeys[features[i].Key]
	}

	return &Features{
		featureList: features,
		version:     "0.1.0",
	}
}

func parseInput(input string) map[FeatureKey]bool {
	enabled := make(map[FeatureKey]bool)
	items := strings.Split(input, ",")
	for _, item := range items {
		key := FeatureKey(strings.TrimSpace(item))
		enabled[key] = true
	}
	return enabled
}

func (f *Features) IsEnabled(key FeatureKey) bool {
	for _, feature := range f.featureList {
		if feature.Key == key {
			return feature.Enabled
		}
	}
	return false
}

func (f *Features) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"version":  f.version,
		"features": f.featureList,
	})
}
