package manimexec

import (
	"bytes"
	"context"
	"fmt"
	"manimatic/internal/config"
	"manimatic/internal/worker/manimexec/security"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type ManimTask struct {
	ScriptPath string
	OutputPath string
	MediaDir   string
	Quality    string
}

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
	workerUID uint32
	workerGID uint32
	validator *security.Validator
}

func MustNewExecutor(cfg *config.Config) *Executor {

	tasksDir := MustTaskDir(cfg.Worker.BaseDir)

	return &Executor{
		baseDir:   tasksDir,
		quality:   QualityLow,
		timeout:   DefaultTimeout,
		workerUID: 10001,
		workerGID: 10001,
		validator: security.NewValidator(nil),
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
		// os.RemoveAll(workDir)
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

	absoluteOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for output: %w", err)
	}

	// Execute the script
	result, err := e.runManimProcess(ctx, scriptPath, absoluteOutputPath)
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

	cmd := exec.CommandContext(ctx, "manim",
		"render",
		"--media_dir", e.baseDir,
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

func MustTaskDir(baseDir string) string {
	tasksDir := filepath.Join(baseDir, "tasks")
	if err := ensureDirExists(tasksDir); err != nil {
		panic(fmt.Errorf("failed to initialize tasks directory: %w", err))
	}
	return tasksDir
}

func ensureDirExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
		} else {
			return fmt.Errorf("failed to access directory %s: %w", path, err)
		}
	}
	return nil
}
