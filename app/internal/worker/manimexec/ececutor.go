package manimexec

import (
	"bytes"
	"context"
	"fmt"
	"manimatic/internal/worker/manimexec/security"
	"os"
	"os/exec"
	"path/filepath"
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
	baseDir   string
	quality   Quality
	timeout   time.Duration
	workerUID uint32 // TODO
	workerGID uint32 // TODO
	validator *security.Validator
}

func NewExecutor(opts ...Option) *Executor {
	e := &Executor{
		baseDir:   defaultBaseDir,
		quality:   QualityLow,
		timeout:   DefaultTimeout,
		workerUID: 10001,
		workerGID: 10001,
		validator: security.NewValidator(nil),
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

func WithBaseDir(dir string) Option {
	return func(e *Executor) {
		e.baseDir = dir
	}
}

func WithWorkerUID(uid uint32) Option {
	return func(e *Executor) {
		e.workerUID = uid
	}
}
func WithWorkerGID(gid uint32) Option {
	return func(e *Executor) {
		e.workerUID = gid
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
	if err := e.validator.ValidateScript(script); err != nil {
		if valErr, ok := err.(*security.ValidationError); ok {
			message := fmt.Sprintf("This code cannot be executed. %s", valErr.Error())
			return nil, newSecurityError(message, err)
		}
		return nil, newSecurityError("This code cannot be executed.", err)
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
		"render",
		"--media_dir", "/manim/worker",
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
