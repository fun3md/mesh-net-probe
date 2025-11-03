

package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-logr/logr"
	stdr "github.com/go-logr/stdr"
)

// Manager provides structured logging capabilities with different backends
type Manager struct {
	logger   logr.Logger
	config   *LoggingConfig
	mu       sync.RWMutex
	handlers map[string]LogHandler
	enabled  bool
}

// LogHandler interface for different logging backends
type LogHandler interface {
	Handle(entry LogEntry) error
	Flush() error
	Close() error
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Level       LogLevel          `json:"level"`
	Message     string            `json:"message"`
	Component   string            `json:"component"`
	Operation   string            `json:"operation"`
	Target      string            `json:"target,omitempty"`
	Error       string            `json:"error,omitempty"`
	Stack       string            `json:"stack,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	ProbeID     string            `json:"probe_id,omitempty"`
	Measurement string            `json:"measurement_id,omitempty"`
}

// LogLevel type alias to the existing LogLevel from logger.go

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level       LogLevel          `json:"level"`
	Format      string            `json:"format"` // "json", "text"
	Output      string            `json:"output"` // "stdout", "stderr", "file"
	FilePath    string            `json:"file_path,omitempty"`
	MaxSize     int64             `json:"max_size,omitempty"` // Max file size in MB
	MaxAge      int               `json:"max_age,omitempty"`  // Max file age in days
	MaxBackups  int               `json:"max_backups,omitempty"`
	EnableStack bool              `json:"enable_stack"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// ComponentLogger provides component-specific logging
type ComponentLogger struct {
	manager   *Manager
	component string
}

// NewManager creates a new logging manager
func NewManager(config *LoggingConfig) (*Manager, error) {
	if config == nil {
		config = &LoggingConfig{
			Level:        InfoLevel,
			Format:       "json",
			Output:       "stdout",
			EnableStack:  false,
			Fields:       make(map[string]interface{}),
		}
	}

	// Configure stdr with Go's standard library
	stdr.SetVerbosity(int(config.Level))

	// Create underlying logger
	var writer io.Writer
	switch config.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	case "file":
		var err error
		writer, err = os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	default:
		writer = os.Stdout
	}

	logger := stdr.New(log.New(writer, "", 0))

	m := &Manager{
		logger:   logger,
		config:   config,
		handlers: make(map[string]LogHandler),
		enabled:  config.Level >= DebugLevel,
	}

	// Add built-in handlers
	if err := m.setupBuiltInHandlers(); err != nil {
		return nil, fmt.Errorf("failed to setup built-in handlers: %w", err)
	}

	return m, nil
}


// setupBuiltInHandlers initializes built-in log handlers
func (m *Manager) setupBuiltInHandlers() error {
	// Add console handler
	consoleHandler := &ConsoleHandler{
		format: m.config.Format,
	}
	m.handlers["console"] = consoleHandler

	// Add file handler if configured
	if m.config.Output == "file" {
		fileHandler, err := NewFileHandler(m.config)
		if err != nil {
			return err
		}
		m.handlers["file"] = fileHandler
	}

	return nil
}

// Debug logs a debug message
func (m *Manager) Debug(ctx context.Context, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, DebugLevel, "", "", "", msg, fields...)
}

// Info logs an info message
func (m *Manager) Info(ctx context.Context, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, InfoLevel, "", "", "", msg, fields...)
}

// Warn logs a warning message
func (m *Manager) Warn(ctx context.Context, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, WarnLevel, "", "", "", msg, fields...)
}

// Error logs an error message
func (m *Manager) Error(ctx context.Context, err error, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, ErrorLevel, "", "", "", msg, fields...)
}

// ErrorWithComponent logs an error message with component information
func (m *Manager) ErrorWithComponent(ctx context.Context, component, operation, target string, err error, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, ErrorLevel, component, operation, target, msg, fields...)
}

// Measurement logs measurement-related data
func (m *Manager) Measurement(ctx context.Context, probeID, measurementID, target string, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, InfoLevel, "measurement", "", target, msg, fields...)
}

// Probe logs probe-related data
func (m *Manager) Probe(ctx context.Context, probeID, operation, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, InfoLevel, "probe", operation, "", msg, fields...)
}

// Config logs configuration-related data
func (m *Manager) Config(ctx context.Context, component, msg string, fields ...interface{}) {
	if !m.enabled {
		return
	}
	m.log(ctx, InfoLevel, component, "", "", msg, fields...)
}

// log performs the actual logging
func (m *Manager) log(ctx context.Context, level LogLevel, component, operation, target, msg string, fields ...interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create log entry
	entry := LogEntry{
		Timestamp:   time.Now(),
		Level:       level,
		Message:     msg,
		Component:   component,
		Operation:   operation,
		Target:      target,
		Fields:      make(map[string]interface{}),
	}

	// Add context fields
	if probeID := getProbeIDFromContext(ctx); probeID != "" {
		entry.ProbeID = probeID
	}

	// Add custom fields
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				entry.Fields[key] = fields[i+1]
			}
		}
	}

	// Add configured fields
	for k, v := range m.config.Fields {
		entry.Fields[k] = v
	}

	// Handle error logging
	if err, ok := extractError(fields); ok {
		entry.Error = err.Error()
		if m.config.EnableStack {
			entry.Stack = string(debug.Stack())
		}
	}

	// Send to all handlers
	for _, handler := range m.handlers {
		if err := handler.Handle(entry); err != nil {
			// Log to stderr if handler fails
			fmt.Fprintf(os.Stderr, "Log handler error: %v\n", err)
		}
	}
}

// extractError extracts an error from fields
func extractError(fields []interface{}) (error, bool) {
	for _, field := range fields {
		if err, ok := field.(error); ok {
			return err, true
		}
	}
	return nil, false
}

// getProbeIDFromContext extracts probe ID from context
func getProbeIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	
	// Try to extract from context value
	if probeID, ok := ctx.Value("probe_id").(string); ok {
		return probeID
	}
	
	return ""
}

// WithComponent creates a component-specific logger
func (m *Manager) WithComponent(component string) *ComponentLogger {
	return &ComponentLogger{
		manager:   m,
		component: component,
	}
}

// Debug logs with component context
func (cl *ComponentLogger) Debug(ctx context.Context, operation, msg string, fields ...interface{}) {
	cl.manager.log(ctx, DebugLevel, cl.component, operation, "", msg, fields...)
}

// Info logs with component context
func (cl *ComponentLogger) Info(ctx context.Context, operation, msg string, fields ...interface{}) {
	cl.manager.log(ctx, InfoLevel, cl.component, operation, "", msg, fields...)
}

// Warn logs with component context
func (cl *ComponentLogger) Warn(ctx context.Context, operation, msg string, fields ...interface{}) {
	cl.manager.log(ctx, WarnLevel, cl.component, operation, "", msg, fields...)
}

// Error logs with component context
func (cl *ComponentLogger) Error(ctx context.Context, err error, operation, msg string, fields ...interface{}) {
	cl.manager.log(ctx, ErrorLevel, cl.component, operation, "", msg, fields...)
}

// Flush flushes all log handlers
func (m *Manager) Flush() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for _, handler := range m.handlers {
		if err := handler.Flush(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("flush errors: %v", errs)
	}
	return nil
}

// Close closes all log handlers
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, handler := range m.handlers {
		if err := handler.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// SetLevel dynamically changes logging level
func (m *Manager) SetLevel(level LogLevel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.Level = level
	m.enabled = level >= DebugLevel

	// Update stdr verbosity
	stdr.SetVerbosity(int(level))
	
	return nil
}

// ConsoleHandler handles console output
type ConsoleHandler struct {
	format string
}

// Handle processes a log entry for console output
func (h *ConsoleHandler) Handle(entry LogEntry) error {
	// For now, just output to stderr/stdout based on level
	switch entry.Level {
	case DebugLevel, InfoLevel:
		fmt.Printf("[%s] %s: %s\n", entry.Level, entry.Component, entry.Message)
	case WarnLevel, ErrorLevel:
		fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", entry.Level, entry.Component, entry.Message)
		if entry.Error != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", entry.Error)
		}
	}
	return nil
}

// Flush flushes console output
func (h *ConsoleHandler) Flush() error {
	return nil
}

// Close closes console handler
func (h *ConsoleHandler) Close() error {
	return nil
}

// FileHandler handles file output
type FileHandler struct {
	file   *os.File
	config *LoggingConfig
}

// NewFileHandler creates a new file handler
func NewFileHandler(config *LoggingConfig) (*FileHandler, error) {
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileHandler{
		file:   file,
		config: config,
	}, nil
}

// Handle processes a log entry for file output
func (h *FileHandler) Handle(entry LogEntry) error {
	// Simple JSON output for file logging
	_, err := h.file.WriteString(fmt.Sprintf("%s\n", marshalLogEntry(entry)))
	return err
}

// Flush flushes file output
func (h *FileHandler) Flush() error {
	return h.file.Sync()
}

// Close closes the file handler
func (h *FileHandler) Close() error {
	return h.file.Close()
}

// marshalLogEntry converts log entry to JSON string
func marshalLogEntry(entry LogEntry) string {
	// This would use proper JSON marshaling in a real implementation
	// For now, return a simple string representation
	return fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s","component":"%s"}`,
		entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message, entry.Component)
}