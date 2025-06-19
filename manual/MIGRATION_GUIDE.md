# Health Analyzers Migration Guide

## Overview
This guide provides instructions for completing the migration from the legacy analyzer system to the new refactored architecture.

## Current State

### ✅ Completed
- All language analyzers refactored to use new interfaces
- Common utilities and base implementations created
- Factory/registry pattern implemented with automatic initialization
- File system abstraction layer established
- Test helpers and mocking infrastructure created
- Package compilation verified
- All analyzer tests migrated to use new mock infrastructure
- Legacy function references removed from command line interface
- Health module updated to use new factory system
- Automatic analyzer registration implemented via init() function
- All tests passing for analyzers, registry, and health modules

### 🔄 Remaining (Optional Cleanup)
- Full removal of LegacyAnalyzer interface (currently still used by adapter for backward compatibility)
- Complete removal of compatibility layer once all consumers migrate

## Migration Steps

### Step 1: Update Health Module (Priority: High)

**Current Issue:** The health module still uses the old registry system:
```go
// In internal/health/health.go
func NewAnalyzerRegistry(fs core.FileSystem, logger core.Logger) *AnalyzerRegistry {
    return analyzer_registry.NewRegistryWithStandardAnalyzers(fs, logger)
}
```

**Recommended Solution:** Update to use new factory system:
```go
// In internal/health/health.go
func NewAnalyzerRegistry(logger core.Logger) map[string]common.FullAnalyzer {
    registry := make(map[string]common.FullAnalyzer)
    
    for _, lang := range analyzers.GetSupportedLanguages() {
        if analyzer, err := analyzers.GetAnalyzer(lang, logger); err == nil {
            registry[lang] = analyzer
        }
    }
    
    return registry
}
```

### Step 2: Update Analyzer Tests (Priority: Medium)

**Current Issue:** Tests in `go/analyzer_test.go` and others use old interfaces.

**Required Changes:**
1. Update imports to include `"github.com/codcod/repos/internal/health/analyzers/common"`
2. Replace `filesystem.NewOSFileSystem()` with `common.NewMockFileWalker()`
3. Replace local `MockLogger` with `&common.MockLogger{}`
4. Update method calls:
   - `analyzer.Name()` → `analyzer.Language()` 
   - `analyzer.SupportedExtensions()` → `analyzer.FileExtensions()`

**Example Fix:**
```go
// Before
func TestAnalyzer(t *testing.T) {
    logger := &MockLogger{}
    fs := filesystem.NewOSFileSystem()
    analyzer := NewGoAnalyzer(fs, logger)
    
    if analyzer.Name() != "go-analyzer" {
        t.Error("wrong name")
    }
}

// After  
func TestAnalyzer(t *testing.T) {
    logger := &common.MockLogger{}
    walker := common.NewMockFileWalker()
    analyzer := NewGoAnalyzer(walker, logger)
    
    if analyzer.Language() != "go" {
        t.Error("wrong language")
    }
}
```

### Step 3: Remove Legacy Code (Priority: Low)

**After migration complete:**
1. Remove `LegacyAnalyzer` interface from core package
2. Remove compatibility layer in `registry/registry.go`
3. Clean up unused imports and old test utilities

## Breaking Changes

### Constructor Signatures
```go
// Old
NewGoAnalyzer(fs core.FileSystem, logger core.Logger) *GoAnalyzer

// New
NewGoAnalyzer(walker common.FileWalker, logger core.Logger) *GoAnalyzer
```

### Interface Methods
```go
// Old
analyzer.Name() string
analyzer.SupportedExtensions() []string

// New  
analyzer.Language() string
analyzer.FileExtensions() []string
```

### Factory Usage
```go
// Old
registry := analyzer_registry.NewRegistryWithStandardAnalyzers(fs, logger)
analyzer := registry.GetAnalyzer("python")

// New
analyzer, err := analyzers.GetAnalyzer("python", logger)
```

## Testing Migration

### Old Test Pattern
```go
func TestOld(t *testing.T) {
    fs := filesystem.NewOSFileSystem()
    logger := &LocalMockLogger{}
    analyzer := NewAnalyzer(fs, logger)
    
    // Test with real file system...
}
```

### New Test Pattern
```go
func TestNew(t *testing.T) {
    walker := common.NewMockFileWalker()
    logger := &common.MockLogger{}
    
    // Setup test data
    walker.AddFile("/test/main.go", []byte("package main"))
    
    analyzer := NewAnalyzer(walker, logger)
    
    // Test with controlled mock data...
}
```

## Rollback Plan

If issues arise during migration:

1. **Quick Fix:** Temporarily restore old registry functionality by updating `NewRegistryWithStandardAnalyzers()` to create adapters:
   ```go
   func NewRegistryWithStandardAnalyzers(fs core.FileSystem, logger core.Logger) *Registry {
       registry := NewRegistry()
       
       // Create adapters that bridge new analyzers to old interface
       walker := &FileSystemWalkerAdapter{fs: fs}
       goAnalyzer := go_analyzer.NewGoAnalyzer(walker, logger)
       registry.Register(&LegacyAnalyzerAdapter{goAnalyzer})
       
       return registry
   }
   ```

2. **Full Rollback:** Revert to git commit before refactoring began.

## Validation Checklist

Before completing migration, verify:

- [x] All language analyzers compile and run
- [x] Health module uses new factory system  
- [x] All tests pass with new mock infrastructure
- [x] Error handling uses new standardized types
- [x] Documentation reflects new architecture
- [x] Factory initialization happens automatically via init() function
- [x] CLI commands work properly with new analyzer system
- [ ] No references to legacy `NewRegistryWithStandardAnalyzers` remain (optional cleanup)
- [ ] LegacyAnalyzer interface removed (waiting for complete consumer migration)

## Support

For questions or issues during migration:
1. Review the `REFACTORING_SUMMARY.md` for architecture details
2. Check `common/test_helpers_test.go` for testing examples
3. Reference `python/analyzer_test.go` for updated test patterns

The refactored architecture provides a solid foundation for analyzer development while maintaining backward compatibility during the transition period.
