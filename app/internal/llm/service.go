package llm

import (
	"context"
	"fmt"
	"sort"
)

type Service struct {
	providers    map[string]Provider
	defaultModel string
	modelCache   []string
}

func NewService(defaultModel string) *Service {
	return &Service{
		providers:    make(map[string]Provider),
		defaultModel: defaultModel,
	}
}

func (s *Service) RegisterProvider(provider Provider) {
	s.providers[provider.ModelID()] = provider
	s.updateModelCache()
}
func (s *Service) updateModelCache() {
	s.modelCache = make([]string, 0, len(s.providers))
	for modelID := range s.providers {
		s.modelCache = append(s.modelCache, modelID)
	}
	// Sort for consistent ordering
	sort.Strings(s.modelCache)
}

func (s *Service) Generate(ctx context.Context, prompt string, model string) (Response, error) {
	if model == "" {
		model = s.defaultModel
	}

	provider, exists := s.providers[model]
	if !exists {
		return Response{}, fmt.Errorf("unsupported model: %s", model)
	}

	return provider.Generate(ctx, prompt)
}

func (s *Service) AvailableModels() []string {
	return s.modelCache
}
func (s *Service) DefaultModel() string {
	return s.defaultModel
}
