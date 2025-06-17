package observability

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/fatih/color"
)

func TestNewStructuredLogger(t *testing.T) {
	logger := NewStructuredLogger(LevelInfo)
	if logger == nil {
		t.Fatal("NewStructuredLogger should not return nil")
	}

	if logger.level != LevelInfo {
		t.Errorf("Expected level %v, got %v", LevelInfo, logger.level)
	}

	if logger.fields == nil {
		t.Error("Fields map should be initialized")
	}
}

func TestLoggerWithPrefix(t *testing.T) {
	logger := NewStructuredLogger(LevelInfo)
	prefixedLogger := logger.WithPrefix("test-prefix")

	if prefixedLogger.prefix != "test-prefix" {
		t.Errorf("Expected prefix 'test-prefix', got '%s'", prefixedLogger.prefix)
	}

	// Should not affect original logger
	if logger.prefix != "" {
		t.Error("Original logger should not have prefix")
	}
}

func TestLoggerWithField(t *testing.T) {
	logger := NewStructuredLogger(LevelInfo)
	fieldLogger := logger.WithField("key", "value")

	if fieldLogger.fields["key"] != "value" {
		t.Errorf("Expected field value 'value', got '%v'", fieldLogger.fields["key"])
	}

	// Should not affect original logger
	if _, exists := logger.fields["key"]; exists {
		t.Error("Original logger should not have the field")
	}
}

func TestLoggerWithFields(t *testing.T) {
	logger := NewStructuredLogger(LevelInfo)
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	fieldLogger := logger.WithFields(fields)

	if fieldLogger.fields["key1"] != "value1" {
		t.Errorf("Expected field value 'value1', got '%v'", fieldLogger.fields["key1"])
	}

	if fieldLogger.fields["key2"] != 42 {
		t.Errorf("Expected field value 42, got '%v'", fieldLogger.fields["key2"])
	}
}

func TestLoggerLevels(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture output
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	logger := NewStructuredLogger(LevelWarn)

	// Debug should not log (below threshold)
	logger.Debug("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not be logged at Warn level")
	}

	// Info should not log (below threshold)
	logger.Info("info message")
	if strings.Contains(buf.String(), "info message") {
		t.Error("Info message should not be logged at Warn level")
	}

	// Warn should log
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn message should be logged at Warn level")
	}

	// Error should log
	logger.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error message should be logged at Warn level")
	}
}

func TestLoggerWithLogFields(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture output
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	logger := NewStructuredLogger(LevelInfo)
	logger.Info("test message",
		core.String("key1", "value1"),
		core.Int("key2", 42))

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Output should contain message")
	}

	if !strings.Contains(output, "key1=value1") {
		t.Error("Output should contain log field key1=value1")
	}

	if !strings.Contains(output, "key2=42") {
		t.Error("Output should contain log field key2=42")
	}
}

func TestStartOperation(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture output
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	logger := NewStructuredLogger(LevelDebug)

	opLogger, done := logger.StartOperation("test-operation")

	// Should have operation in context
	if opLogger.fields["operation"] != "test-operation" {
		t.Errorf("Expected operation field 'test-operation', got '%v'", opLogger.fields["operation"])
	}

	// Complete the operation
	done()

	output := buf.String()
	if !strings.Contains(output, "operation started") {
		t.Error("Output should contain 'operation started'")
	}

	if !strings.Contains(output, "operation completed") {
		t.Error("Output should contain 'operation completed'")
	}

	if !strings.Contains(output, "duration=") {
		t.Error("Output should contain duration field")
	}
}

func TestNewCheckerLogger(t *testing.T) {
	baseLogger := NewStructuredLogger(LevelInfo)
	checkerLogger := NewCheckerLogger("test-checker", baseLogger)

	if checkerLogger.checkerName != "test-checker" {
		t.Errorf("Expected checker name 'test-checker', got '%s'", checkerLogger.checkerName)
	}

	if checkerLogger.fields["checker"] != "test-checker" {
		t.Errorf("Expected checker field 'test-checker', got '%v'", checkerLogger.fields["checker"])
	}
}

func TestCheckerLoggerLogResult(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture output
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	baseLogger := NewStructuredLogger(LevelInfo)
	checkerLogger := NewCheckerLogger("test-checker", baseLogger)

	result := core.RepositoryResult{
		Repository: core.Repository{Name: "test-repo"},
		Status:     core.StatusHealthy,
		Score:      85,
		MaxScore:   100,
		CheckResults: []core.CheckResult{
			{
				Issues:   []core.Issue{{Type: "test-issue"}},
				Warnings: []core.Warning{{Type: "test-warning"}},
			},
		},
		Duration: time.Second,
		EndTime:  time.Now(),
	}

	checkerLogger.LogResult(result)

	output := buf.String()
	if !strings.Contains(output, "check completed successfully") {
		t.Error("Output should contain success message")
	}

	if !strings.Contains(output, "status=healthy") {
		t.Error("Output should contain status")
	}

	if !strings.Contains(output, "score=85") {
		t.Error("Output should contain score")
	}

	if !strings.Contains(output, "issues=1") {
		t.Error("Output should contain issues count")
	}

	if !strings.Contains(output, "warnings=1") {
		t.Error("Output should contain warnings count")
	}
}

func TestLoggerLogResultWithWarning(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture output
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	baseLogger := NewStructuredLogger(LevelInfo)
	checkerLogger := NewCheckerLogger("test-checker", baseLogger)

	result := core.RepositoryResult{
		Status: core.StatusWarning,
		Score:  70,
	}

	checkerLogger.LogResult(result)

	output := buf.String()
	if !strings.Contains(output, "check completed with warnings") {
		t.Error("Output should contain warning message")
	}
}

func TestLoggerLogResultWithCritical(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture output
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	baseLogger := NewStructuredLogger(LevelInfo)
	checkerLogger := NewCheckerLogger("test-checker", baseLogger)

	result := core.RepositoryResult{
		Status: core.StatusCritical,
		Score:  30,
	}

	checkerLogger.LogResult(result)

	output := buf.String()
	if !strings.Contains(output, "check completed with critical issues") {
		t.Error("Output should contain critical message")
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

func TestFormatContextFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]interface{}
		expected string
	}{
		{
			name:     "empty fields",
			fields:   map[string]interface{}{},
			expected: "",
		},
		{
			name:     "single field",
			fields:   map[string]interface{}{"key": "value"},
			expected: "context={key=value}",
		},
		{
			name: "multiple fields",
			fields: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expected: "context={", // We just check it contains the prefix
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := formatContextFields(test.fields)
			if test.name == "multiple fields" {
				if !strings.HasPrefix(result, test.expected) {
					t.Errorf("Expected result to start with '%s', got '%s'", test.expected, result)
				}
				if !strings.Contains(result, "key1=value1") {
					t.Error("Result should contain key1=value1")
				}
				if !strings.Contains(result, "key2=42") {
					t.Error("Result should contain key2=42")
				}
			} else {
				if result != test.expected {
					t.Errorf("Expected '%s', got '%s'", test.expected, result)
				}
			}
		})
	}
}

func TestFormatLogFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   []core.Field
		expected string
	}{
		{
			name:     "empty fields",
			fields:   []core.Field{},
			expected: "",
		},
		{
			name:     "single field",
			fields:   []core.Field{core.String("key", "value")},
			expected: "fields={key=value}",
		},
		{
			name: "multiple fields",
			fields: []core.Field{
				core.String("key1", "value1"),
				core.Int("key2", 42),
			},
			expected: "fields={key1=value1, key2=42}",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := formatLogFields(test.fields)
			if result != test.expected {
				t.Errorf("Expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}

func BenchmarkStructuredLogger(b *testing.B) {
	// Redirect output to discard for benchmarking
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	logger := NewStructuredLogger(LevelInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", core.String("iteration", string(rune(i))))
	}
}

func BenchmarkCheckerLogger(b *testing.B) {
	// Redirect output to discard for benchmarking
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	baseLogger := NewStructuredLogger(LevelInfo)
	checkerLogger := NewCheckerLogger("benchmark-checker", baseLogger)

	result := core.RepositoryResult{
		Status: core.StatusHealthy,
		Score:  85,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkerLogger.LogResult(result)
	}
}
