package manimexec

import (
	"errors"
	"time"
)

const (
	MaxScriptSize    = 1_024 * 1_024
	DefaultTimeout   = 30 * time.Second
	MaxOutputSize    = 10 * 1_024 * 1_024 // 10MB
	TempDirPrefix    = "manim_exec_"
	ScriptFilePrefix = "scene_"
	defaultBaseDir   = "/manim/worker"
)

type Quality string

const (
	QualityLow    Quality = "-ql"
	QualityMedium Quality = "-qm"
	QualityHigh   Quality = "-qh"
)

var (
	ErrScriptTooLarge   = errors.New("script exceeds maximum size limit")
	ErrExecutionTimeout = errors.New("script execution timed out")
	ErrOutputTooLarge   = errors.New("output exceeds maximum size")
)
