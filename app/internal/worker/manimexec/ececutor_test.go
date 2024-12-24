package manimexec

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

type testScript struct {
	name    string
	content string
	wantErr error
}

var testScripts = []testScript{
	{
		name: "valid_simple",
		content: `
from manim import Scene, Circle
class TestScene(Scene):
    def construct(self):
        circle = Circle()
        self.play(Create(circle))
`,
		wantErr: nil,
	},
	{
		name: "invalid_import",
		content: `
import os
from manim import Scene
class TestScene(Scene):
    def construct(self):
        pass
`,
		wantErr: fmt.Errorf("forbidden pattern detected: (?i)import\\s+os"),
	},
	{
		name:    "script_too_large",
		content: strings.Repeat("a", MaxScriptSize+1),
		wantErr: ErrScriptTooLarge,
	},
	{
		name: "valid_with_numpy",
		content: `
from manim import Scene
import numpy as np
class TestScene(Scene):
    def construct(self):
        arr = np.array([1, 2, 3])
`,
		wantErr: nil,
	},
}

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		opts    []Option
		want    *Executor
	}{
		{
			name:    "default_settings",
			baseDir: "/tmp",
			opts:    nil,
			want: &Executor{
				baseDir:  "/tmp",
				security: defaultSecurityConfig,
				quality:  QualityLow,
				timeout:  DefaultTimeout,
			},
		},
		{
			name:    "custom_settings",
			baseDir: "/tmp",
			opts: []Option{
				WithQuality(QualityHigh),
				WithTimeout(60 * time.Second),
			},
			want: &Executor{
				baseDir:  "/tmp",
				security: defaultSecurityConfig,
				quality:  QualityHigh,
				timeout:  60 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExecutor(tt.baseDir, tt.opts...)
			if got.baseDir != tt.want.baseDir {
				t.Errorf("baseDir = %v, want %v", got.baseDir, tt.want.baseDir)
			}
			if got.quality != tt.want.quality {
				t.Errorf("quality = %v, want %v", got.quality, tt.want.quality)
			}
			if got.timeout != tt.want.timeout {
				t.Errorf("timeout = %v, want %v", got.timeout, tt.want.timeout)
			}
		})
	}
}

func TestExecutor_validateScript(t *testing.T) {
	executor := NewExecutor("/tmp")

	for _, tt := range testScripts {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validateScript(tt.content)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("validateScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("validateScript() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateScript(t *testing.T) {
	executor := NewExecutor("/tmp")

	tests := []struct {
		name    string
		script  string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid imports",
			script: `
import manim
import numpy
from math import sin
`,
			wantErr: false,
		},
		{
			name: "forbidden import",
			script: `
import manim
import os
`,
			wantErr: true,
			errMsg:  "forbidden pattern detected",
		},
		{
			name: "non-allowed import",
			script: `
import manim
import pandas
`,
			wantErr: true,
			errMsg:  "import not allowed: pandas",
		},
		{
			name: "multiple valid imports",
			script: `
import manim
from numpy import array
import math as m
`,
			wantErr: false,
		},
		{
			name: "hidden malicious import",
			script: `
import manim
x = "__import__('os')"
`,
			wantErr: true,
			errMsg:  "forbidden pattern detected",
		},
		{
			name: "eval attempt",
			script: `
import manim
eval("import os")
`,
			wantErr: true,
			errMsg:  "forbidden pattern detected",
		},
		{
			name: "exec attempt",
			script: `
import manim
exec("import os")
`,
			wantErr: true,
			errMsg:  "forbidden pattern detected",
		},
		{
			name: "file operations",
			script: `
import manim
open('test.txt', 'w')
`,
			wantErr: true,
			errMsg:  "forbidden pattern detected",
		},
		{
			name: "complex import pattern",
			script: `
from manim import Scene
from numpy import array as arr
import math
`,
			wantErr: false,
		},
		{
			name: "case insensitive check",
			script: `
import manim
IMPORT os
`,
			wantErr: true,
			errMsg:  "forbidden pattern detected",
		},
		{
			name: "custom security config",
			script: `
import custom_module
`,
			wantErr: true,
			errMsg:  "import not allowed: custom_module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validateScript(tt.script)
			if tt.wantErr {
				if err == nil {
					t.Errorf("%s: validateScript() error = nil, wantErr %v", tt.name, tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateScript() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateScript() unexpected error = %v", err)
			}
		})
	}
}

func TestExecutor_writeScript(t *testing.T) {

	tempDir, err := os.MkdirTemp("", "manim_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	executor := NewExecutor(tempDir)
	script := "test script content"

	scriptPath, err := executor.writeScript(tempDir, script)
	if err != nil {
		t.Fatalf("writeScript() error = %v", err)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Script file was not created")
	}

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read script file: %v", err)
	}
	if string(content) != script {
		t.Errorf("Script content = %v, want %v", string(content), script)
	}
}
