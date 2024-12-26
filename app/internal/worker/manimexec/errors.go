package manimexec

import (
	"fmt"
	"manimatic/internal/api/events"
)

// ExecutionError represents a structured error from manim execution
type ExecutionError struct {
	Kind    ErrorKind // Type of error that occurred
	Message string    // User-friendly error message
	Stdout  string    // Standard output from execution if available
	Stderr  string    // Standard error from execution if available
	Line    int       // Line number where error occurred (if available)
	Cause   error     // Original error that caused this
}

// ErrorKind represents different categories of execution errors
type ErrorKind int

const (
	ErrorKindSecurity    ErrorKind = iota // Security validation failed
	ErrorKindSize                         // Script or output size exceeded limits
	ErrorKindTimeout                      // Execution timed out
	ErrorKindCompilation                  // Manim compilation failed
	ErrorKindSystem                       // System-level error (IO, etc)
)

func (k ErrorKind) String() string {
	switch k {
	case ErrorKindSecurity:
		return "Security Validation Error"
	case ErrorKindSize:
		return "Size Limit Error"
	case ErrorKindTimeout:
		return "Timeout Error"
	case ErrorKindCompilation:
		return "Compilation Error"
	case ErrorKindSystem:
		return "System Error"
	default:
		return "Unknown Error"
	}
}

// Error implements the error interface
func (e *ExecutionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *ExecutionError) Unwrap() error {
	return e.Cause
}

// Helper functions to create specific error types
func newSecurityError(message string, cause error) *ExecutionError {
	return &ExecutionError{
		Kind:    ErrorKindSecurity,
		Message: message,
		Cause:   cause,
	}
}

func newSizeError(message string, cause error) *ExecutionError {
	return &ExecutionError{
		Kind:    ErrorKindSize,
		Message: message,
		Cause:   cause,
	}
}

func newTimeoutError() *ExecutionError {
	return &ExecutionError{
		Kind:    ErrorKindTimeout,
		Message: "Script execution timed out",
		Cause:   ErrExecutionTimeout,
	}
}

func newCompilationError(message string, stdout, stderr string, cause error) *ExecutionError {
	return &ExecutionError{
		Kind:    ErrorKindCompilation,
		Message: message,
		Stdout:  stdout,
		Stderr:  stderr,
		Cause:   cause,
	}
}

func newSystemError(message string, cause error) *ExecutionError {
	return &ExecutionError{
		Kind:    ErrorKindSystem,
		Message: message,
		Cause:   cause,
	}
}

func (e *ExecutionError) ToCompileError() events.CompileError {
	return events.CompileError{
		Message: fmt.Sprintf("%s: %s", e.Kind, e.Message),
		Stdout:  e.Stdout,
		Stderr:  e.Stderr,
		Line:    e.Line,
	}
}
