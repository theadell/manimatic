package security

import (
	"fmt"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
)

type SecurityConfig struct {
	AllowedImports  map[string]bool
	BlockedBuiltins map[string]bool
	BlockedAttrs    map[string]bool
	ProtectedNames  map[string]bool
}

func DefaultConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedImports: map[string]bool{
			"manim":  true,
			"numpy":  true,
			"math":   true,
			"typing": true,
			"abc":    true,
			"enum":   true,
			"colour": true,
		},
		BlockedBuiltins: map[string]bool{
			"eval":       true,
			"exec":       true,
			"open":       true,
			"__import__": true,
			"getattr":    true,
			"setattr":    true,
		},
		BlockedAttrs: map[string]bool{
			"__dict__":         true,
			"__class__":        true,
			"__bases__":        true,
			"__globals__":      true,
			"__builtins__":     true,
			"__subclasses__":   true,
			"__getattribute__": true,
			"__setattr__":      true,
		},
		ProtectedNames: map[string]bool{
			"open":       true,
			"__import__": true,
			"eval":       true,
			"exec":       true,
		},
	}
}

type ValidationError struct {
	Line    int
	Col     int
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("at line %d, column %d: %s", e.Line, e.Col, e.Message)
}

type Validator struct {
	config *SecurityConfig
}

func NewValidator(config *SecurityConfig) *Validator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Validator{config: config}
}

func (v *Validator) ValidateScript(script string) error {
	mod, err := parser.ParseString(script, py.ExecMode)
	if err != nil {
		return fmt.Errorf("failed to parse Python script: %w", err)
	}

	var validationErr error
	ast.Walk(mod, func(node ast.Ast) bool {
		if err := v.validateNode(node); err != nil {
			validationErr = err
			return false
		}
		return true
	})

	return validationErr
}

func (v *Validator) validateNode(node ast.Ast) error {
	switch n := node.(type) {
	case *ast.Import:
		return v.validateImport(n)
	case *ast.ImportFrom:
		return v.validateImportFrom(n)
	case *ast.Call:
		return v.validateCall(n)
	case *ast.Attribute:
		return v.validateAttribute(n)
	case *ast.Assign:
		return v.validateAssign(n)
	case *ast.With:
		return v.validateWith(n)
	}
	return nil
}

func (v *Validator) validateImport(node *ast.Import) error {
	for _, alias := range node.Names {
		name := string(alias.Name)
		if !v.config.AllowedImports[name] {
			return &ValidationError{
				Line:    node.Lineno,
				Col:     node.ColOffset,
				Message: fmt.Sprintf("import of '%s' is not allowed", name),
			}
		}
	}
	return nil
}

func (v *Validator) validateImportFrom(node *ast.ImportFrom) error {
	moduleName := string(node.Module)
	if !v.config.AllowedImports[moduleName] {
		return &ValidationError{
			Line:    node.Lineno,
			Col:     node.ColOffset,
			Message: fmt.Sprintf("import from '%s' is not allowed", moduleName),
		}
	}
	return nil
}

func (v *Validator) validateCall(node *ast.Call) error {
	// Check for direct calls to blocked builtins
	if name, ok := node.Func.(*ast.Name); ok {
		funcName := string(name.Id)
		if v.config.BlockedBuiltins[funcName] {
			return &ValidationError{
				Line:    node.Lineno,
				Col:     node.ColOffset,
				Message: fmt.Sprintf("call to '%s' is not allowed", funcName),
			}
		}
	}
	return nil
}

func (v *Validator) validateAttribute(node *ast.Attribute) error {
	attrName := string(node.Attr)
	if v.config.BlockedAttrs[attrName] {
		return &ValidationError{
			Line:    node.Lineno,
			Col:     node.ColOffset,
			Message: fmt.Sprintf("access to attribute '%s' is not allowed", attrName),
		}
	}

	if attr, ok := node.Value.(*ast.Attribute); ok {
		if err := v.validateAttribute(attr); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateAssign(node *ast.Assign) error {
	for _, target := range node.Targets {
		if name, ok := target.(*ast.Name); ok {
			if v.config.ProtectedNames[string(name.Id)] {
				return &ValidationError{
					Line:    node.Lineno,
					Col:     node.ColOffset,
					Message: fmt.Sprintf("assignment to protected name '%s' is not allowed", name.Id),
				}
			}
		}
	}
	return nil
}

func (v *Validator) validateWith(node *ast.With) error {
	// Check context manager expressions
	for _, item := range node.Items {
		// Look for potential file operations
		if call, ok := item.ContextExpr.(*ast.Call); ok {
			if err := v.validateCall(call); err != nil {
				return err
			}
		}
		// Check attribute access in context manager
		if attr, ok := item.ContextExpr.(*ast.Attribute); ok {
			if err := v.validateAttribute(attr); err != nil {
				return err
			}
		}
	}
	return nil
}
