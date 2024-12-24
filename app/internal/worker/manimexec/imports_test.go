package manimexec

import (
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected []string
		wantErr  bool
	}{
		{
			name:     "simple import",
			script:   `import manim`,
			expected: []string{"manim"},
		},
		{
			name:     "multiple imports on one line",
			script:   `import numpy, math`,
			expected: []string{"numpy", "math"},
		},
		{
			name:     "from import",
			script:   `from manim import Scene`,
			expected: []string{"manim"},
		},
		{
			name:     "import with alias",
			script:   `import numpy as np`,
			expected: []string{"numpy"},
		},
		{
			name:     "mixed import styles",
			script:   "import manim\nfrom numpy import array\nimport math as m\n",
			expected: []string{"manim", "numpy", "math"},
		},
		{
			name:    "malformed Python",
			script:  `imp[ort manim`,
			wantErr: true,
		},
		{
			name:     "empty script",
			script:   ``,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports, err := extractImports(tt.script)

			if tt.wantErr {
				if err == nil {
					t.Errorf("extractImports() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Errorf("extractImports() error = %v, wantErr = false", err)
				return
			}

			// Sort both slices for deterministic comparison
			sort.Strings(imports)
			expected := make([]string, len(tt.expected))
			copy(expected, tt.expected)
			sort.Strings(expected)

			if !slices.Equal(imports, expected) {
				t.Errorf("extractImports() = %v, want %v", imports, expected)
			}
		})
	}
}

// TestExtractImportsLargeScript tests handling of a large script
func TestExtractImportsLargeScript(t *testing.T) {
	var scriptBuilder strings.Builder
	scriptBuilder.WriteString("import manim\n")
	scriptBuilder.WriteString("from numpy import array\n")
	for i := 0; i < 1000; i++ {
		scriptBuilder.WriteString("# Comment line\n")
	}
	scriptBuilder.WriteString("import math\n")

	imports, err := extractImports(scriptBuilder.String())
	if err != nil {
		t.Errorf("extractImports() error = %v", err)
		return
	}

	expected := []string{"manim", "numpy", "math"}
	sort.Strings(imports)
	sort.Strings(expected)

	if !reflect.DeepEqual(imports, expected) {
		t.Errorf("extractImports() = %v, want %v", imports, expected)
	}
}

func TestExtractImportsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected []string
		wantErr  bool
	}{
		{
			name: "import in string literal",
			script: `x = "import math"
import numpy
`,
			expected: []string{"numpy"},
		},
		{
			name: "comment that looks like import",
			script: `# import math
import numpy
`,
			expected: []string{"numpy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports, err := extractImports(tt.script)

			if tt.wantErr {
				if err == nil {
					t.Errorf("extractImports() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Errorf("extractImports() error = %v, wantErr = false", err)
				return
			}

			// Sort both slices for deterministic comparison
			sort.Strings(imports)
			expected := make([]string, len(tt.expected))
			copy(expected, tt.expected)
			sort.Strings(expected)

			if !reflect.DeepEqual(imports, expected) {
				t.Errorf("extractImports() = %v, want %v", imports, expected)
			}
		})
	}
}
