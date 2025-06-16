package health

import (
	"time"
)

// CheckResult represents a structured result from a health check
type CheckResult struct {
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Status      HealthStatus           `json:"status"`
	Score       int                    `json:"score"`
	Issues      []Issue                `json:"issues"`
	Warnings    []Warning              `json:"warnings"`
	Metrics     map[string]interface{} `json:"metrics"`
	Metadata    map[string]string      `json:"metadata"`
	Duration    time.Duration          `json:"duration"`
	LastChecked time.Time              `json:"last_checked"`
}

// Issue represents a specific problem found during health checking
type Issue struct {
	Type       string                 `json:"type"`
	Severity   Severity               `json:"severity"`
	Message    string                 `json:"message"`
	Location   *Location              `json:"location,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Warning represents a warning found during health checking
type Warning struct {
	Message    string    `json:"message"`
	Location   *Location `json:"location,omitempty"`
	Suggestion string    `json:"suggestion,omitempty"`
}

// Location represents a location in source code
type Location struct {
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`
}

// ResultBuilder helps build complex CheckResult objects
type ResultBuilder struct {
	result *CheckResult
}

// NewResultBuilder creates a new result builder
func NewResultBuilder(name, category string) *ResultBuilder {
	return &ResultBuilder{
		result: &CheckResult{
			Name:        name,
			Category:    category,
			Status:      HealthStatusHealthy,
			Issues:      make([]Issue, 0),
			Warnings:    make([]Warning, 0),
			Metrics:     make(map[string]interface{}),
			Metadata:    make(map[string]string),
			LastChecked: time.Now(),
		},
	}
}

// AddIssue adds an issue to the result
func (b *ResultBuilder) AddIssue(issue Issue) *ResultBuilder {
	b.result.Issues = append(b.result.Issues, issue)
	b.updateStatus()
	return b
}

// AddWarning adds a warning to the result
func (b *ResultBuilder) AddWarning(warning Warning) *ResultBuilder {
	b.result.Warnings = append(b.result.Warnings, warning)
	return b
}

// AddMetric adds a metric to the result
func (b *ResultBuilder) AddMetric(key string, value interface{}) *ResultBuilder {
	b.result.Metrics[key] = value
	return b
}

// AddMetadata adds metadata to the result
func (b *ResultBuilder) AddMetadata(key, value string) *ResultBuilder {
	b.result.Metadata[key] = value
	return b
}

// WithDuration sets the duration of the check
func (b *ResultBuilder) WithDuration(duration time.Duration) *ResultBuilder {
	b.result.Duration = duration
	return b
}

// WithStatus sets the status of the check
func (b *ResultBuilder) WithStatus(status HealthStatus) *ResultBuilder {
	b.result.Status = status
	return b
}

// Build returns the final CheckResult
func (b *ResultBuilder) Build() CheckResult {
	b.calculateScore()
	return *b.result
}

// updateStatus updates the overall status based on issues
func (b *ResultBuilder) updateStatus() {
	hasCritical := false
	hasWarning := false

	for _, issue := range b.result.Issues {
		switch issue.Severity {
		case SeverityCritical:
			hasCritical = true
		case SeverityWarning:
			hasWarning = true
		}
	}

	if hasCritical {
		b.result.Status = HealthStatusCritical
	} else if hasWarning || len(b.result.Warnings) > 0 {
		b.result.Status = HealthStatusWarning
	} else {
		b.result.Status = HealthStatusHealthy
	}
}

// calculateScore calculates a numeric score based on issues and warnings
func (b *ResultBuilder) calculateScore() {
	baseScore := 100

	for _, issue := range b.result.Issues {
		switch issue.Severity {
		case SeverityCritical:
			baseScore -= 20
		case SeverityWarning:
			baseScore -= 10
		case SeverityInfo:
			baseScore -= 5
		}
	}

	// Warnings reduce score by 3 points each
	baseScore -= len(b.result.Warnings) * 3

	if baseScore < 0 {
		baseScore = 0
	}

	b.result.Score = baseScore
}

// NewIssue creates a new Issue
func NewIssue(issueType string, severity Severity, message string) Issue {
	return Issue{
		Type:     issueType,
		Severity: severity,
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// WithLocation adds location information to an issue
func (i Issue) WithLocation(location Location) Issue {
	i.Location = &location
	return i
}

// WithSuggestion adds a suggestion to an issue
func (i Issue) WithSuggestion(suggestion string) Issue {
	i.Suggestion = suggestion
	return i
}

// WithContext adds context information to an issue
func (i Issue) WithContext(key string, value interface{}) Issue {
	if i.Context == nil {
		i.Context = make(map[string]interface{})
	}
	i.Context[key] = value
	return i
}

// NewWarning creates a new Warning
func NewWarning(message string) Warning {
	return Warning{
		Message: message,
	}
}

// WithLocation adds location information to a warning
func (w Warning) WithLocation(location Location) Warning {
	w.Location = &location
	return w
}

// WithSuggestion adds a suggestion to a warning
func (w Warning) WithSuggestion(suggestion string) Warning {
	w.Suggestion = suggestion
	return w
}

// NewLocation creates a new Location
func NewLocation(filePath string, line, column int) Location {
	return Location{
		FilePath: filePath,
		Line:     line,
		Column:   column,
	}
}

// WithEnd adds end position to a location
func (l Location) WithEnd(endLine, endColumn int) Location {
	l.EndLine = endLine
	l.EndColumn = endColumn
	return l
}
