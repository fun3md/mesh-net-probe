package logger

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

// Logger provides structured logging with OpenTelemetry integration
type Logger interface {
	// Debug logs debug level messages
	Debug(ctx context.Context, msg string, keysAndValues ...interface{})

	// Info logs info level messages
	Info(ctx context.Context, msg string, keysAndValues ...interface{})

	// Warn logs warning level messages
	Warn(ctx context.Context, msg string, keysAndValues ...interface{})

	// Error logs error level messages
	Error(ctx context.Context, err error, msg string, keysAndValues ...interface{})

	// WithValues creates a child logger with additional values
	WithValues(keysAndValues ...interface{}) Logger

	// WithName creates a named logger
	WithName(name string) Logger

	// WithTraceInfo adds trace context to the logger
	WithTraceInfo(span trace.Span, traceID, spanID string) Logger

	// Enabled returns if logging is enabled for the specified level
	Enabled(ctx context.Context, level LogLevel) bool
}

// LogLevel represents logging levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "unknown"
	}
}

// contextKey is used to store logger in context
type contextKey struct{}

// defaultLogger implements the Logger interface using logr
type defaultLogger struct {
	logr.Logger
	enabledLevels map[LogLevel]bool
}

// NewLogger creates a new structured logger
func NewLogger(level LogLevel) Logger {
	// Use stdr as the underlying implementation
	logger := stdr.New(nil)
	
	// Set log level (convert LogLevel to int)
	logger = logger.V(int(level))
	
	// Create enabled levels map
	enabledLevels := make(map[LogLevel]bool)
	for i := DebugLevel; i <= ErrorLevel; i++ {
		enabledLevels[i] = i <= level
	}

	return &defaultLogger{
		Logger:        logger,
		enabledLevels: enabledLevels,
	}
}

// NewFromLogr creates a logger from an existing logr instance
func NewFromLogr(logrInstance logr.Logger, enabledLevels map[LogLevel]bool) Logger {
	return &defaultLogger{
		Logger:        logrInstance,
		enabledLevels: enabledLevels,
	}
}

// Debug logs debug level messages
func (l *defaultLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	// Add context information
	enhancedValues := l.enhanceValues(ctx, keysAndValues...)
	l.Logger.Info(msg, enhancedValues...)
}

// Info logs info level messages
func (l *defaultLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	// Add context information
	enhancedValues := l.enhanceValues(ctx, keysAndValues...)
	l.Logger.Info(msg, enhancedValues...)
}

// Warn logs warning level messages
func (l *defaultLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	// Add context information
	enhancedValues := l.enhanceValues(ctx, keysAndValues...)
	l.Logger.Info(msg, enhancedValues...)
}

// Error logs error level messages
func (l *defaultLogger) Error(ctx context.Context, err error, msg string, keysAndValues ...interface{}) {
	// Add context information
	enhancedValues := l.enhanceValues(ctx, keysAndValues...)
	// Add error information
	enhancedValues = append(enhancedValues, "error", err)
	l.Logger.Info(msg, enhancedValues...)
}

// WithValues creates a child logger with additional values
func (l *defaultLogger) WithValues(keysAndValues ...interface{}) Logger {
	childLogger := l.Logger.WithValues(keysAndValues...)
	return &defaultLogger{
		Logger:        childLogger,
		enabledLevels: l.enabledLevels,
	}
}

// WithName creates a named logger
func (l *defaultLogger) WithName(name string) Logger {
	childLogger := l.Logger.WithName(name)
	return &defaultLogger{
		Logger:        childLogger,
		enabledLevels: l.enabledLevels,
	}
}

// WithTraceInfo adds trace context to the logger
func (l *defaultLogger) WithTraceInfo(span trace.Span, traceID, spanID string) Logger {
	childLogger := l.Logger.WithValues(
		"trace_id", traceID,
		"span_id", spanID,
		"trace_flags", span.SpanContext().TraceFlags(),
	)
	return &defaultLogger{
		Logger:        childLogger,
		enabledLevels: l.enabledLevels,
	}
}

// Enabled returns if logging is enabled for the specified level
func (l *defaultLogger) Enabled(ctx context.Context, level LogLevel) bool {
	return l.enabledLevels[level]
}

// enhanceValues adds context information to the log values
func (l *defaultLogger) enhanceValues(ctx context.Context, keysAndValues ...interface{}) []interface{} {
	enhanced := make([]interface{}, 0, len(keysAndValues)+4)
	
	// Add timestamp
	enhanced = append(enhanced, "timestamp", time.Now().Format(time.RFC3339Nano))
	
	// Add trace information if available
	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		enhanced = append(enhanced, "trace_id", traceID)
	}
	
	// Add all provided values
	enhanced = append(enhanced, keysAndValues...)
	
	return enhanced
}

// getTraceIDFromContext extracts trace ID from context
func getTraceIDFromContext(ctx context.Context) string {
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		return span.TraceID().String()
	}
	return ""
}

// FromContext retrieves a logger from context
func FromContext(ctx context.Context) (Logger, bool) {
	logger, ok := ctx.Value(contextKey{}).(Logger)
	return logger, ok
}

// WithContext adds a logger to context
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// Global logger instance
var globalLogger Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(level LogLevel) {
	globalLogger = NewLogger(level)
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		InitGlobalLogger(InfoLevel)
	}
	return globalLogger
}