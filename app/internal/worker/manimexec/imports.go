package manimexec

import (
	"fmt"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
)

func extractImports(script string) ([]string, error) {
	mod, err := parser.ParseString(script, py.ExecMode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Python script: %w", err)
	}

	var imports []string
	ast.Walk(mod, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.Import:
			for _, alias := range n.Names {
				imports = append(imports, string(alias.Name))
			}
		case *ast.ImportFrom:
			imports = append(imports, string(n.Module))
		}
		return true
	})

	return imports, nil
}
