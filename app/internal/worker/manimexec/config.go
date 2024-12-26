package manimexec

import (
	"errors"
	"regexp"
	"time"
)

const (
	MaxScriptSize    = 1_024 * 1_024
	DefaultTimeout   = 30 * time.Second
	MaxOutputSize    = 10 * 1_024 * 1_024 // 10MB
	TempDirPrefix    = "manim_exec_"
	ScriptFilePrefix = "scene_"
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

type SecurityConfig struct {
	AllowedImports    []string
	ForbiddenPatterns []*regexp.Regexp
}

var defaultSecurityConfig = SecurityConfig{
	AllowedImports: []string{
		"manim",
		"numpy",
		"math",
		"random",
		"decimal",
		"sympy",
		"scipy",
		"fractions",
	},
	ForbiddenPatterns: []*regexp.Regexp{
		// general
		regexp.MustCompile(`(?i)import\s+os`),
		regexp.MustCompile(`(?i)import\s+sys`),
		regexp.MustCompile(`(?i)import\s+subprocess`),
		regexp.MustCompile(`(?i)import\s+pathlib`),
		regexp.MustCompile(`(?i)__import__`),
		regexp.MustCompile(`(?i)eval\(`),
		regexp.MustCompile(`(?i)exec\(`),
		regexp.MustCompile(`(?i)open\(`),
		regexp.MustCompile(`(?i)file\(`),
		regexp.MustCompile(`(?i)glob\(`),
		regexp.MustCompile(`(?i)importlib`),

		// Module introspection (applicable to any allowed module)
		regexp.MustCompile(`(?i)__class__`),
		regexp.MustCompile(`(?i)__bases__`),
		regexp.MustCompile(`(?i)__subclasses__`),
		regexp.MustCompile(`(?i)__globals__`),
		regexp.MustCompile(`(?i)__builtins__`),
		regexp.MustCompile(`(?i)__dict__`),

		// Dangerous operations from allowed modules
		regexp.MustCompile(`(?i)numpy\.load\(`),
		regexp.MustCompile(`(?i)numpy\.save\(`),
		regexp.MustCompile(`(?i)sympy\.preview\(`),
	},
}