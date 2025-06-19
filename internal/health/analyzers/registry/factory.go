// Package registry provides analyzer registration and factory functionality
package registry

import (
	"fmt"
	"sync"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers/common"
)

// AnalyzerFactory creates analyzer instances
type AnalyzerFactory func(walker common.FileWalker, logger core.Logger) common.FullAnalyzer

// FactoryRegistry manages analyzer registration and creation using factory pattern
type FactoryRegistry struct {
	mu        sync.RWMutex
	factories map[string]AnalyzerFactory
}

// NewFactoryRegistry creates a new analyzer factory registry
func NewFactoryRegistry() *FactoryRegistry {
	return &FactoryRegistry{
		factories: make(map[string]AnalyzerFactory),
	}
}

// Register registers an analyzer factory for a language
func (r *FactoryRegistry) Register(language string, factory AnalyzerFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[language] = factory
}

// Unregister removes an analyzer factory for a language
func (r *FactoryRegistry) Unregister(language string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, language)
}

// GetAnalyzer creates an analyzer instance for the specified language
func (r *FactoryRegistry) GetAnalyzer(language string, walker common.FileWalker, logger core.Logger) (common.FullAnalyzer, error) {
	r.mu.RLock()
	factory, exists := r.factories[language]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("analyzer not found for language: %s", language)
	}

	return factory(walker, logger), nil
}

// GetSupportedLanguages returns all supported languages
func (r *FactoryRegistry) GetSupportedLanguages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var languages []string
	for lang := range r.factories {
		languages = append(languages, lang)
	}
	return languages
}

// HasAnalyzer checks if an analyzer is registered for the language
func (r *FactoryRegistry) HasAnalyzer(language string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[language]
	return exists
}

// GlobalFactoryRegistry is the global analyzer factory registry instance
var GlobalFactoryRegistry = NewFactoryRegistry()

// RegisterAnalyzer registers an analyzer factory in the global registry
func RegisterAnalyzer(language string, factory AnalyzerFactory) {
	GlobalFactoryRegistry.Register(language, factory)
}

// GetAnalyzerFromFactory gets an analyzer from the global factory registry
func GetAnalyzerFromFactory(language string, walker common.FileWalker, logger core.Logger) (common.FullAnalyzer, error) {
	return GlobalFactoryRegistry.GetAnalyzer(language, walker, logger)
}

// GetSupportedLanguagesFromFactory returns supported languages from the global factory registry
func GetSupportedLanguagesFromFactory() []string {
	return GlobalFactoryRegistry.GetSupportedLanguages()
}
