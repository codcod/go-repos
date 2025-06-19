package logging

import (
	"errors"
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	if collector == nil {
		t.Fatal("NewMetricsCollector should not return nil")
	}

	if collector.counters == nil {
		t.Error("Counters map should be initialized")
	}

	if collector.gauges == nil {
		t.Error("Gauges map should be initialized")
	}

	if collector.histograms == nil {
		t.Error("Histograms map should be initialized")
	}

	if collector.timers == nil {
		t.Error("Timers map should be initialized")
	}

	if collector.startTime.IsZero() {
		t.Error("Start time should be set")
	}
}

func TestCounterOperations(t *testing.T) {
	collector := NewMetricsCollector()

	// Test initial value
	if collector.GetCounter("test") != 0 {
		t.Error("Initial counter value should be 0")
	}

	// Test increment
	collector.IncrementCounter("test")
	if collector.GetCounter("test") != 1 {
		t.Errorf("Expected counter value 1, got %d", collector.GetCounter("test"))
	}

	// Test add to counter
	collector.AddToCounter("test", 5)
	if collector.GetCounter("test") != 6 {
		t.Errorf("Expected counter value 6, got %d", collector.GetCounter("test"))
	}
}

func TestGaugeOperations(t *testing.T) {
	collector := NewMetricsCollector()

	// Test initial value
	if collector.GetGauge("test") != 0 {
		t.Error("Initial gauge value should be 0")
	}

	// Test set gauge
	collector.SetGauge("test", 42.5)
	if collector.GetGauge("test") != 42.5 {
		t.Errorf("Expected gauge value 42.5, got %f", collector.GetGauge("test"))
	}
}

func TestHistogramOperations(t *testing.T) {
	collector := NewMetricsCollector()

	// Test recording values
	collector.RecordHistogram("test", 10.0)
	collector.RecordHistogram("test", 20.0)
	collector.RecordHistogram("test", 30.0)

	histogram := collector.GetHistogram("test")
	if histogram == nil {
		t.Fatal("Histogram should not be nil")
	}

	summary := histogram.Summary()
	if summary.Count != 3 {
		t.Errorf("Expected count 3, got %d", summary.Count)
	}

	if summary.Mean != 20.0 {
		t.Errorf("Expected mean 20.0, got %f", summary.Mean)
	}
}

func TestTimerOperations(t *testing.T) {
	collector := NewMetricsCollector()

	timer := collector.StartTimer("test")
	if timer == nil {
		t.Fatal("Timer should not be nil")
	}

	// Sleep for a short time
	time.Sleep(10 * time.Millisecond)

	duration := timer.Stop()
	if duration <= 0 {
		t.Error("Duration should be positive")
	}

	retrievedTimer := collector.GetTimer("test")
	if retrievedTimer == nil {
		t.Error("Retrieved timer should not be nil")
	}
}

func TestRecordCheckResult(t *testing.T) {
	collector := NewMetricsCollector()

	result := core.CheckResult{
		Status:   core.StatusHealthy,
		Score:    85,
		Duration: 100 * time.Millisecond,
	}

	collector.RecordCheckResult(result)

	// Check counters
	if collector.GetCounter("checks_executed") != 1 {
		t.Error("checks_executed counter should be 1")
	}

	if collector.GetCounter("checks_healthy") != 1 {
		t.Error("checks_healthy counter should be 1")
	}

	// Check histograms
	scoreHist := collector.GetHistogram("check_scores")
	if scoreHist == nil {
		t.Error("check_scores histogram should exist")
	}

	durationHist := collector.GetHistogram("check_duration_ms")
	if durationHist == nil {
		t.Error("check_duration_ms histogram should exist")
	}
}

func TestGetSummary(t *testing.T) {
	collector := NewMetricsCollector()

	// Record some data
	collector.IncrementCounter("test_counter")
	collector.SetGauge("test_gauge", 42.0)
	collector.RecordHistogram("test_histogram", 10.0)

	summary := collector.GetSummary()

	if summary.StartTime.IsZero() {
		t.Error("Summary should have start time")
	}

	if summary.TotalDuration <= 0 {
		t.Error("Summary should have positive duration")
	}

	if summary.Counters["test_counter"] != 1 {
		t.Error("Summary should include test_counter")
	}

	if summary.Gauges["test_gauge"] != 42.0 {
		t.Error("Summary should include test_gauge")
	}

	if _, exists := summary.Histograms["test_histogram"]; !exists {
		t.Error("Summary should include test_histogram")
	}
}

func TestMeasureOperation(t *testing.T) {
	collector := NewMetricsCollector()

	// Test successful operation
	err := collector.MeasureOperation("test_op", func() error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if collector.GetCounter("test_op_success") != 1 {
		t.Error("Success counter should be 1")
	}

	// Test failed operation
	testErr := errors.New("test error")
	err = collector.MeasureOperation("test_op_fail", func() error {
		return testErr
	})

	if err != testErr {
		t.Error("Should return the original error")
	}

	if collector.GetCounter("test_op_fail_errors") != 1 {
		t.Error("Error counter should be 1")
	}
}
