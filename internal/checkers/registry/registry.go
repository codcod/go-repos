package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/codcod/repos/internal/checkers/ci"
	"github.com/codcod/repos/internal/checkers/compliance"
	"github.com/codcod/repos/internal/checkers/dependencies"
	"github.com/codcod/repos/internal/checkers/git"
	"github.com/codcod/repos/internal/checkers/security"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/commands"
)

// CheckerRegistry manages all available checkers
type CheckerRegistry struct {
	checkers map[string]core.Checker
	mu       sync.RWMutex
}

// NewCheckerRegistry creates a new checker registry with default checkers
func NewCheckerRegistry(executor commands.CommandExecutor) *CheckerRegistry {
	registry := &CheckerRegistry{
		checkers: make(map[string]core.Checker),
	}

	// Register default checkers
	registry.registerDefaultCheckers(executor)

	return registry
}

// registerDefaultCheckers registers all built-in checkers
func (r *CheckerRegistry) registerDefaultCheckers(executor commands.CommandExecutor) {
	// Git checkers
	r.Register(git.NewGitStatusChecker(executor))
	r.Register(git.NewLastCommitChecker(executor))

	// Security checkers
	r.Register(security.NewBranchProtectionChecker(executor))
	r.Register(security.NewVulnerabilityChecker(executor))

	// Dependency checkers
	r.Register(dependencies.NewOutdatedChecker(executor))

	// Compliance checkers
	r.Register(compliance.NewLicenseChecker())

	// CI/CD checkers
	r.Register(ci.NewCIConfigChecker())
}

// Register adds a checker to the registry
func (r *CheckerRegistry) Register(checker core.Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.checkers[checker.ID()] = checker
}

// Unregister removes a checker from the registry
func (r *CheckerRegistry) Unregister(checkerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.checkers, checkerID)
}

// GetChecker returns a specific checker by ID
func (r *CheckerRegistry) GetChecker(checkerID string) (core.Checker, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	checker, exists := r.checkers[checkerID]
	if !exists {
		return nil, fmt.Errorf("checker with ID '%s' not found", checkerID)
	}

	return checker, nil
}

// GetCheckers returns all registered checkers
func (r *CheckerRegistry) GetCheckers() []core.Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	checkers := make([]core.Checker, 0, len(r.checkers))
	for _, checker := range r.checkers {
		checkers = append(checkers, checker)
	}

	return checkers
}

// GetCheckersForRepository returns checkers that support a specific repository
func (r *CheckerRegistry) GetCheckersForRepository(repo core.Repository) []core.Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var supportedCheckers []core.Checker
	for _, checker := range r.checkers {
		if checker.SupportsRepository(repo) {
			supportedCheckers = append(supportedCheckers, checker)
		}
	}

	return supportedCheckers
}

// GetCheckersByCategory returns checkers filtered by category
func (r *CheckerRegistry) GetCheckersByCategory(category string) []core.Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var categoryCheckers []core.Checker
	for _, checker := range r.checkers {
		if checker.Category() == category {
			categoryCheckers = append(categoryCheckers, checker)
		}
	}

	return categoryCheckers
}

// GetEnabledCheckers returns only enabled checkers based on configuration
func (r *CheckerRegistry) GetEnabledCheckers(config map[string]core.CheckerConfig) []core.Checker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var enabledCheckers []core.Checker
	for _, checker := range r.checkers {
		checkerConfig, exists := config[checker.ID()]
		if !exists {
			// If no config exists, use the checker's default config
			if checker.Config().Enabled {
				enabledCheckers = append(enabledCheckers, checker)
			}
		} else if checkerConfig.Enabled {
			enabledCheckers = append(enabledCheckers, checker)
		}
	}

	return enabledCheckers
}

// RunChecker executes a specific checker
func (r *CheckerRegistry) RunChecker(ctx context.Context, checkerID string, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	checker, err := r.GetChecker(checkerID)
	if err != nil {
		return core.CheckResult{}, err
	}

	return checker.Check(ctx, repoCtx)
}

// RunCheckers executes multiple checkers for a repository
func (r *CheckerRegistry) RunCheckers(ctx context.Context, checkerIDs []string, repoCtx core.RepositoryContext) ([]core.CheckResult, error) {
	var results []core.CheckResult
	var errors []error

	for _, checkerID := range checkerIDs {
		result, err := r.RunChecker(ctx, checkerID, repoCtx)
		if err != nil {
			errors = append(errors, fmt.Errorf("checker %s failed: %w", checkerID, err))
			continue
		}
		results = append(results, result)
	}

	// Return combined error if any checkers failed
	if len(errors) > 0 {
		var errorMsg string
		for i, err := range errors {
			if i > 0 {
				errorMsg += "; "
			}
			errorMsg += err.Error()
		}
		return results, fmt.Errorf("some checkers failed: %s", errorMsg)
	}

	return results, nil
}

// RunAllEnabledCheckers runs all enabled checkers for a repository
func (r *CheckerRegistry) RunAllEnabledCheckers(ctx context.Context, repoCtx core.RepositoryContext, config map[string]core.CheckerConfig) ([]core.CheckResult, error) {
	enabledCheckers := r.GetEnabledCheckers(config)

	var checkerIDs []string
	for _, checker := range enabledCheckers {
		if checker.SupportsRepository(repoCtx.Repository) {
			checkerIDs = append(checkerIDs, checker.ID())
		}
	}

	return r.RunCheckers(ctx, checkerIDs, repoCtx)
}

// ListCheckers returns information about all registered checkers
func (r *CheckerRegistry) ListCheckers() []CheckerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var checkerInfos []CheckerInfo
	for _, checker := range r.checkers {
		config := checker.Config()
		checkerInfos = append(checkerInfos, CheckerInfo{
			ID:         checker.ID(),
			Name:       checker.Name(),
			Category:   checker.Category(),
			Enabled:    config.Enabled,
			Severity:   config.Severity,
			Timeout:    config.Timeout,
			Categories: config.Categories,
		})
	}

	return checkerInfos
}

// CheckerInfo contains information about a checker
type CheckerInfo struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Category   string      `json:"category"`
	Enabled    bool        `json:"enabled"`
	Severity   string      `json:"severity"`
	Timeout    interface{} `json:"timeout"`
	Categories []string    `json:"categories"`
}

// GetStats returns statistics about the registry
func (r *CheckerRegistry) GetStats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := RegistryStats{
		TotalCheckers: len(r.checkers),
		Categories:    make(map[string]int),
	}

	for _, checker := range r.checkers {
		category := checker.Category()
		stats.Categories[category]++

		if checker.Config().Enabled {
			stats.EnabledCheckers++
		}
	}

	return stats
}

// RegistryStats contains statistics about the checker registry
type RegistryStats struct {
	TotalCheckers   int            `json:"total_checkers"`
	EnabledCheckers int            `json:"enabled_checkers"`
	Categories      map[string]int `json:"categories"`
}
