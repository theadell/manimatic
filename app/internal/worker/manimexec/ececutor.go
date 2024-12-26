package manimexec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type ExecutionResult struct {
	OutputPath string
	WorkingDir string
	Stdout     string
	Stderr     string
}

type Executor struct {
	baseDir  string
	security SecurityConfig
	quality  Quality
	timeout  time.Duration
}

func NewExecutor(baseDir string, opts ...Option) *Executor {
	e := &Executor{
		baseDir:  baseDir,
		security: defaultSecurityConfig,
		quality:  QualityLow,
		timeout:  DefaultTimeout,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

type Option func(*Executor)

func WithQuality(q Quality) Option {
	return func(e *Executor) {
		e.quality = q
	}
}

func WithTimeout(t time.Duration) Option {
	return func(e *Executor) {
		e.timeout = t
	}
}

func WithSecurityConfig(sc SecurityConfig) Option {
	return func(e *Executor) {
		e.security = sc
	}
}
func (e *Executor) ExecuteScript(ctx context.Context, script string, sessionID string) (*ExecutionResult, error) {
	// Validate script size
	if len(script) > MaxScriptSize {
		return nil, newSizeError(
			fmt.Sprintf("Script size %d exceeds limit %d", len(script), MaxScriptSize),
			ErrScriptTooLarge,
		)
	}

	// Validate script security
	if err := e.validateScript(script); err != nil {
		return nil, newSecurityError("Script contains forbidden patterns or imports", err)
	}

	// Create working directory
	compilationID := uuid.New().String()
	workDir, err := os.MkdirTemp(e.baseDir, fmt.Sprintf("%s_%s", sessionID, compilationID))
	if err != nil {
		return nil, newSystemError("Failed to create working directory", err)
	}

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(workDir)
	}

	// Setup cancellation
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Ensure cleanup on error
	var success bool
	defer func() {
		if !success {
			cleanup()
		}
	}()

	// Write script to file
	scriptPath, err := e.writeScript(workDir, script)
	if err != nil {
		return nil, newSystemError("Failed to write script to file", err)
	}

	// Prepare output path
	outputPath := filepath.Join(workDir, "output.mp4")

	// Execute the script
	result, err := e.runManimProcess(ctx, scriptPath, outputPath)
	if err != nil {
		return nil, err
	}

	result.WorkingDir = workDir
	success = true
	return result, nil
}

func (e *Executor) validateScript(script string) error {

	imports, err := extractImports(script)
	if err != nil {
		return fmt.Errorf("import validation failed: %w", err)
	}

	for _, imp := range imports {
		if !slices.Contains(e.security.AllowedImports, imp) {
			return fmt.Errorf("import not allowed: %s", imp)
		}
	}

	for _, pattern := range e.security.ForbiddenPatterns {
		if pattern.MatchString(script) {
			return fmt.Errorf("forbidden pattern detected: %s", pattern.String())
		}
	}

	return nil
}

func (e *Executor) writeScript(workDir, script string) (string, error) {
	scriptFile, err := os.CreateTemp(workDir, ScriptFilePrefix+"*.py")
	if err != nil {
		return "", fmt.Errorf("failed to create script file: %w", err)
	}
	defer scriptFile.Close()

	if _, err := scriptFile.Write([]byte(script)); err != nil {
		return "", fmt.Errorf("failed to write script: %w", err)
	}

	return scriptFile.Name(), nil
}

func (e *Executor) runManimProcess(ctx context.Context, scriptPath, outputPath string) (*ExecutionResult, error) {
	// Prepare command
	cmd := exec.CommandContext(ctx, "manim",
		string(e.quality),
		"-o", outputPath,
		scriptPath,
	)

	// Set up process attributes
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Set up output buffers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, newSystemError("Failed to start manim process", err)
	}

	// Create a channel for the command completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for completion or cancellation
	var result *ExecutionResult
	select {
	case <-ctx.Done():
		// Kill the process group
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done // Wait for the process to be killed
		return nil, newTimeoutError()

	case err := <-done:
		if err != nil {
			return nil, newCompilationError(
				"Manim compilation failed",
				stdout.String(),
				stderr.String(),
				err,
			)
		}

		// Check output size
		if stdout.Len() > MaxOutputSize || stderr.Len() > MaxOutputSize {
			return nil, newSizeError(
				fmt.Sprintf("Output size exceeds limit of %d bytes", MaxOutputSize),
				ErrOutputTooLarge,
			)
		}

		// Verify output file exists and is accessible
		// Corner case: when no animation is played, a png is generated
		path := outputPath
		if err := checkOutputFile(outputPath); err != nil {
			pngPath := outputPath + ".png"
			if err := checkOutputFile(pngPath); err != nil {
				return nil, newCompilationError(
					"Manim compilation completed but no output file was created",
					stdout.String(),
					stderr.String(),
					fmt.Errorf("neither video nor image output found"),
				)
			}
			path = pngPath
		}

		result = &ExecutionResult{
			OutputPath: path,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
		}
	}

	return result, nil
}
func checkOutputFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %w", err)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}
	return nil
}
