// Package analyzers provides initialization for all language analyzers
package analyzers

import (
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers/common"
	go_analyzer "github.com/codcod/repos/internal/health/analyzers/go"
	java_analyzer "github.com/codcod/repos/internal/health/analyzers/java"
	javascript_analyzer "github.com/codcod/repos/internal/health/analyzers/javascript"
	python_analyzer "github.com/codcod/repos/internal/health/analyzers/python"
	"github.com/codcod/repos/internal/health/analyzers/registry"
)

// GoAnalyzerAdapter adapts the existing GoAnalyzer to the new interface
type GoAnalyzerAdapter struct {
	*go_analyzer.GoAnalyzer
}

// NewGoAnalyzerAdapter creates a new Go analyzer adapter
func NewGoAnalyzerAdapter(walker common.FileWalker, logger core.Logger) common.FullAnalyzer {
	analyzer := go_analyzer.NewGoAnalyzer(walker, logger)
	return &GoAnalyzerAdapter{GoAnalyzer: analyzer}
}

// PythonAnalyzerAdapter adapts the existing PythonAnalyzer to the new interface
type PythonAnalyzerAdapter struct {
	*python_analyzer.PythonAnalyzer
}

// NewPythonAnalyzerAdapter creates a new Python analyzer adapter
func NewPythonAnalyzerAdapter(walker common.FileWalker, logger core.Logger) common.FullAnalyzer {
	analyzer := python_analyzer.NewPythonAnalyzer(walker, logger)
	return &PythonAnalyzerAdapter{PythonAnalyzer: analyzer}
}

// JavaAnalyzerAdapter adapts the existing JavaAnalyzer to the new interface
type JavaAnalyzerAdapter struct {
	*java_analyzer.JavaAnalyzer
}

// NewJavaAnalyzerAdapter creates a new Java analyzer adapter
func NewJavaAnalyzerAdapter(walker common.FileWalker, logger core.Logger) common.FullAnalyzer {
	analyzer := java_analyzer.NewJavaAnalyzer(walker, logger)
	return &JavaAnalyzerAdapter{JavaAnalyzer: analyzer}
}

// JavaScriptAnalyzerAdapter adapts the existing JavaScriptAnalyzer to the new interface
type JavaScriptAnalyzerAdapter struct {
	*javascript_analyzer.JavaScriptAnalyzer
}

// NewJavaScriptAnalyzerAdapter creates a new JavaScript analyzer adapter
func NewJavaScriptAnalyzerAdapter(walker common.FileWalker, logger core.Logger) common.FullAnalyzer {
	analyzer := javascript_analyzer.NewJavaScriptAnalyzer(walker, logger)
	return &JavaScriptAnalyzerAdapter{JavaScriptAnalyzer: analyzer}
}

// init automatically registers all available analyzers when the package is imported
func init() {
	RegisterAllAnalyzers()
}

// RegisterAllAnalyzers registers all available analyzers in the global registry
func RegisterAllAnalyzers() {
	registry.RegisterAnalyzer("go", NewGoAnalyzerAdapter)
	registry.RegisterAnalyzer("python", NewPythonAnalyzerAdapter)
	registry.RegisterAnalyzer("java", NewJavaAnalyzerAdapter)
	registry.RegisterAnalyzer("javascript", NewJavaScriptAnalyzerAdapter)
}

// GetAnalyzer gets an analyzer for the specified language
func GetAnalyzer(language string, logger core.Logger) (common.FullAnalyzer, error) {
	walker := common.NewDefaultFileWalker()
	return registry.GetAnalyzerFromFactory(language, walker, logger)
}

// GetSupportedLanguages returns all supported languages
func GetSupportedLanguages() []string {
	return registry.GetSupportedLanguagesFromFactory()
}
