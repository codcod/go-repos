package health

import (
	"fmt"
	"log"
	"time"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// NewField creates a new log field
func NewField(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// SimpleLogger is a basic implementation of Logger
type SimpleLogger struct {
	fields []Field
}

// NewSimpleLogger creates a new simple logger
func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{
		fields: make([]Field, 0),
	}
}

func (l *SimpleLogger) Debug(msg string, fields ...Field) {
	l.log("DEBUG", msg, fields...)
}

func (l *SimpleLogger) Info(msg string, fields ...Field) {
	l.log("INFO", msg, fields...)
}

func (l *SimpleLogger) Warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields...)
}

func (l *SimpleLogger) Error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields...)
}

func (l *SimpleLogger) With(fields ...Field) Logger {
	newFields := append(l.fields, fields...)
	return &SimpleLogger{fields: newFields}
}

func (l *SimpleLogger) log(level, msg string, fields ...Field) {
	allFields := append(l.fields, fields...)

	logMsg := fmt.Sprintf("[%s] %s", level, msg)

	if len(allFields) > 0 {
		logMsg += " |"
		for _, field := range allFields {
			logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
		}
	}

	log.Println(logMsg)
}

// CheckerLogger wraps a logger with checker-specific context
type CheckerLogger struct {
	logger  Logger
	checker string
}

// NewCheckerLogger creates a new checker-specific logger
func NewCheckerLogger(checker string, logger Logger) *CheckerLogger {
	if logger == nil {
		logger = NewSimpleLogger()
	}

	return &CheckerLogger{
		logger:  logger.With(String("checker", checker)),
		checker: checker,
	}
}

func (l *CheckerLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

func (l *CheckerLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

func (l *CheckerLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

func (l *CheckerLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

func (l *CheckerLogger) With(fields ...Field) Logger {
	return &CheckerLogger{
		logger:  l.logger.With(fields...),
		checker: l.checker,
	}
}

// StartOperation creates a logger with operation context and returns a completion function
func (l *CheckerLogger) StartOperation(operation string) (Logger, func()) {
	start := time.Now()
	opLogger := l.With(
		String("operation", operation),
		String("started_at", start.Format(time.RFC3339)),
	)

	opLogger.Debug("operation started")

	return opLogger, func() {
		duration := time.Since(start)
		opLogger.Info("operation completed", Duration("duration", duration))
	}
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (NoOpLogger) Debug(msg string, fields ...Field) {}
func (NoOpLogger) Info(msg string, fields ...Field)  {}
func (NoOpLogger) Warn(msg string, fields ...Field)  {}
func (NoOpLogger) Error(msg string, fields ...Field) {}
func (l NoOpLogger) With(fields ...Field) Logger     { return l }
