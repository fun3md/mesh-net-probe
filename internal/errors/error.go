package errors

import (
	"context"
	"fmt"
	"time"

	"github.com/mesh-net-probe/probe/internal/logger"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	NetworkError       ErrorType = "network"      // Network-related errors
	ConfigurationError ErrorType = "config"       // Configuration-related errors
	PermissionError    ErrorType = "permission"   // Permission-related errors
	ResourceError      ErrorType = "resource"     // Resource-related errors
	TimeoutError       ErrorType = "timeout"      // Timeout-related errors
	ValidationError    ErrorType = "validation"   // Validation-related errors
	InternalError      ErrorType = "internal"     // Internal system errors
)

// Severity represents error severity levels
type Severity string

const (
	Low      Severity = "low"       // Non-critical errors
	Medium   Severity = "medium"    // Moderate impact errors
	High     Severity = "high"      // High impact errors
	Critical Severity = "critical"  // Critical system errors
)

// MeshError represents a structured error with metadata
type MeshError struct {
	Type        ErrorType       `json:"type"`         // Error type/category
	Code        string          `json:"code"`         // Unique error code
	Message     string          `json:"message"`      // Human-readable error message
	Cause       error           `json:"cause"`        // Underlying cause
	Context     []interface{}   `json:"context"`      // Additional context
	Severity    Severity        `json:"severity"`     // Error severity
	Recoverable bool            `json:"recoverable"`  // Whether error is recoverable
	Timestamp   time.Time       `json:"timestamp"`    // When error occurred
	TraceID     string          `json:"trace_id"`     // OpenTelemetry trace ID
	SpanID      string          `json:"span_id"`      // OpenTelemetry span ID
	Component   string          `json:"component"`    // System component that generated error
}

// Error returns the error message implementing the error interface
func (e *MeshError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As compatibility
func (e *MeshError) Unwrap() error {
	return e.Cause
}

// CreateError creates a new MeshError
func CreateError(
	errorType ErrorType,
	code string,
	message string,
	opts ...ErrorOption,
) *MeshError {
	err := &MeshError{
		Type:        errorType,
		Code:        code,
		Message:     message,
		Severity:    Medium,
		Recoverable: false,
		Timestamp:   time.Now(),
		Component:   "unknown",
	}

	for _, opt := range opts {
		opt(err)
	}

	return err
}

// ErrorOption is a functional option for configuring MeshError
type ErrorOption func(*MeshError)

// WithCause sets the underlying cause of the error
func WithCause(cause error) ErrorOption {
	return func(e *MeshError) {
		e.Cause = cause
	}
}

// WithContext adds context information to the error
func WithContext(key string, value interface{}) ErrorOption {
	return func(e *MeshError) {
		e.Context = append(e.Context, key, value)
	}
}

// WithSeverity sets the error severity
func WithSeverity(severity Severity) ErrorOption {
	return func(e *MeshError) {
		e.Severity = severity
	}
}

// WithRecoverable marks the error as recoverable
func WithRecoverable(recoverable bool) ErrorOption {
	return func(e *MeshError) {
		e.Recoverable = recoverable
	}
}

// WithTraceInfo adds OpenTelemetry trace information
func WithTraceInfo(traceID, spanID string) ErrorOption {
	return func(e *MeshError) {
		e.TraceID = traceID
		e.SpanID = spanID
	}
}

// WithComponent sets the component that generated the error
func WithComponent(component string) ErrorOption {
	return func(e *MeshError) {
		e.Component = component
	}
}

// Helper functions for common error types

// NewNetworkError creates a network error
func NewNetworkError(code string, message string, opts ...ErrorOption) *MeshError {
	return CreateError(NetworkError, code, message, opts...)
}

// NewMeshConfigError creates a mesh configuration error
func NewMeshConfigError(code string, message string, opts ...ErrorOption) *MeshError {
	return CreateError(ConfigurationError, code, message, opts...)
}

// NewPermissionError creates a permission error
func NewPermissionError(code string, message string, opts ...ErrorOption) *MeshError {
	return CreateError(PermissionError, code, message, opts...)
}

// NewResourceError creates a resource error
func NewResourceError(code string, message string, opts ...ErrorOption) *MeshError {
	return CreateError(ResourceError, code, message, opts...)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(code string, message string, opts ...ErrorOption) *MeshError {
	return CreateError(TimeoutError, code, message, opts...)
}

// NewValidationError creates a validation error
func NewValidationError(code string, message string, opts ...ErrorOption) *MeshError {
	return CreateError(ValidationError, code, message, opts...)
}

// NewInternalError creates an internal error
func NewInternalError(code string, message string, opts ...ErrorOption) *MeshError {
	return CreateError(InternalError, code, message, opts...)
}

// ErrorHandler provides structured error handling capabilities
type ErrorHandler interface {
	// HandleError processes and logs an error
	HandleError(ctx context.Context, err error) error

	// HandleAndLogError processes an error and logs it
	HandleAndLogError(ctx context.Context, err error, logger logger.Logger) error

	// IsErrorType checks if an error is of a specific type
	IsErrorType(err error, errorType ErrorType) bool

	// GetSeverity determines the severity of an error
	GetSeverity(err error) Severity
}

// defaultErrorHandler implements the ErrorHandler interface
type defaultErrorHandler struct {
	minSeverity Severity
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(minSeverity Severity) ErrorHandler {
	return &defaultErrorHandler{
		minSeverity: minSeverity,
	}
}

// HandleError processes and logs an error
func (h *defaultErrorHandler) HandleError(ctx context.Context, err error) error {
	// Check if error is a MeshError
	if meshErr, ok := err.(*MeshError); ok {
		// Log the error if severity meets threshold
		if h.shouldLog(meshErr.Severity) {
			logger := h.getLogger(ctx)
			logger.Error(ctx, meshErr, "Mesh error occurred",
				"error_code", meshErr.Code,
				"error_type", meshErr.Type,
				"severity", meshErr.Severity,
				"recoverable", meshErr.Recoverable,
			)
		}
		return meshErr
	}

	// Handle standard errors
	return err
}

// HandleAndLogError processes an error and logs it
func (h *defaultErrorHandler) HandleAndLogError(ctx context.Context, err error, logger logger.Logger) error {
	// Check if error is a MeshError
	if meshErr, ok := err.(*MeshError); ok {
		// Log with structured data
		logger.Error(ctx, meshErr, "Mesh error occurred",
			"error_code", meshErr.Code,
			"error_type", meshErr.Type,
			"severity", meshErr.Severity,
			"recoverable", meshErr.Recoverable,
			"component", meshErr.Component,
		)
		return meshErr
	}

	// Log generic error
	logger.Error(ctx, err, "Generic error occurred")
	return err
}

// IsErrorType checks if an error is of a specific type
func (h *defaultErrorHandler) IsErrorType(err error, errorType ErrorType) bool {
	if meshErr, ok := err.(*MeshError); ok {
		return meshErr.Type == errorType
	}
	return false
}

// GetSeverity determines the severity of an error
func (h *defaultErrorHandler) GetSeverity(err error) Severity {
	if meshErr, ok := err.(*MeshError); ok {
		return meshErr.Severity
	}
	return Medium // Default severity for non-MeshErrors
}

// shouldLog determines if an error should be logged based on severity
func (h *defaultErrorHandler) shouldLog(severity Severity) bool {
	severityLevels := map[Severity]int{
		Low:      0,
		Medium:   1,
		High:     2,
		Critical: 3,
	}

	return severityLevels[severity] >= severityLevels[h.minSeverity]
}

// getLogger retrieves logger from context or creates a default one
func (h *defaultErrorHandler) getLogger(ctx context.Context) logger.Logger {
	if loggerInstance, ok := logger.FromContext(ctx); ok {
		return loggerInstance
	}
	return logger.GetGlobalLogger()
}

// Error codes for common mesh probe errors
const (
	// Network error codes
	ICMPNoResponse       = "ICMP_001" // No ICMP response received
	ICMPTimeout          = "ICMP_002" // ICMP request timed out
	ICMPPermissionDenied = "ICMP_003" // ICMP permission denied
	ICMPInvalidPacket    = "ICMP_004" // Invalid ICMP packet format

	// Configuration error codes
	ConfigInvalidFormat    = "CONFIG_001" // Invalid configuration format
	ConfigMissingRequired  = "CONFIG_002" // Missing required configuration
	ConfigValidationFailed = "CONFIG_003" // Configuration validation failed

	// Resource error codes
	ResourceExhausted = "RESOURCE_001" // System resources exhausted
	MemoryAllocationFailed = "RESOURCE_002" // Memory allocation failed

	// Internal error codes
	InternalUnexpected = "INTERNAL_001" // Unexpected internal error
	InternalStateCorrupt = "INTERNAL_002" // Internal state corruption detected
)