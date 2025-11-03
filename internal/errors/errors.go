package errors

import (
	"fmt"
	"os"
	"time"
)

// Standardized error codes for mesh probe system
const (
	// Configuration errors
	ErrCodeConfigInvalid = "CONFIG_INVALID"
	ErrCodeConfigMissing = "CONFIG_MISSING"
	ErrCodeConfigRead    = "CONFIG_READ"

	// Network/ICMP errors
	ErrCodeICMPTimeout    = "ICMP_TIMEOUT"
	ErrCodeICMPNoResponse = "ICMP_NO_RESPONSE"
	ErrCodeICMPInvalid    = "ICMP_INVALID"
	ErrCodeICMPPermission = "ICMP_PERMISSION"

	// Platform errors
	ErrCodePlatformUnsupported = "PLATFORM_UNSUPPORTED"
	ErrCodePlatformCapability  = "PLATFORM_CAPABILITY"

	// Measurement errors
	ErrCodeMeasurementPrecision = "MEASUREMENT_PRECISION"
	ErrCodeMeasurementData      = "MEASUREMENT_DATA"

	// System errors
	ErrCodeSystemResource   = "SYSTEM_RESOURCE"
	ErrCodeSystemPermission = "SYSTEM_PERMISSION"
)

// Error represents a structured error in the mesh probe system
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Cause     error  `json:"cause,omitempty"`
	Suggested string `json:"suggested,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As compatibility
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new structured error
func NewError(code, message, suggested string, cause error) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Suggested: suggested,
	}
}

// NewICMPTimeoutError creates an ICMP timeout error
func NewICMPTimeoutError(target string, timeout time.Duration) *Error {
	return NewError(
		ErrCodeICMPTimeout,
		fmt.Sprintf("ICMP timeout for target %s after %v", target, timeout),
		fmt.Sprintf("Check network connectivity to %s or increase timeout", target),
		nil,
	)
}

// NewICMPPermissionError creates an ICMP permission error
func NewICMPPermissionError() *Error {
	return NewError(
		ErrCodeICMPPermission,
		"Insufficient permissions for ICMP operations",
		"Run probe with elevated privileges or configure network capabilities",
		nil,
	)
}

// NewConfigError creates a configuration error
func NewConfigError(message, suggested string, cause error) *Error {
	return NewError(
		ErrCodeConfigInvalid,
		fmt.Sprintf("Configuration error: %s", message),
		suggested,
		cause,
	)
}

// NewPlatformError creates a platform compatibility error
func NewPlatformError(platform string) *Error {
	return NewError(
		ErrCodePlatformUnsupported,
		fmt.Sprintf("Platform %s is not supported", platform),
		"Ensure probe is deployed on supported platforms: Linux, macOS, Windows",
		nil,
	)
}

// Structured logging interface
type Logger interface {
	Error(err *Error)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

// StandardLogger implements basic structured logging
type StandardLogger struct{}

func (l *StandardLogger) Error(err *Error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
	if err.Suggested != "" {
		fmt.Fprintf(os.Stderr, "SUGGESTION: %s\n", err.Suggested)
	}
	if err.Cause != nil {
		fmt.Fprintf(os.Stderr, "CAUSE: %v\n", err.Cause)
	}
}

func (l *StandardLogger) Warn(msg string) {
	fmt.Printf("WARN: %s\n", msg)
}

func (l *StandardLogger) Info(msg string) {
	fmt.Printf("INFO: %s\n", msg)
}

func (l *StandardLogger) Debug(msg string) {
	fmt.Printf("DEBUG: %s\n", msg)
}