// Package parser provides a JavaScript/TypeScript parser with full AST access.
// It wraps esbuild's internal parser with a simpler, public API.
package parser

import (
	"github.com/octoberswimmer/skyfire/internal/config"
	"github.com/octoberswimmer/skyfire/internal/js_ast"
	"github.com/octoberswimmer/skyfire/internal/js_parser"
	"github.com/octoberswimmer/skyfire/internal/logger"
)

// AST represents the parsed JavaScript/TypeScript abstract syntax tree.
type AST = js_ast.AST

// Stmt represents a statement in the AST.
type Stmt = js_ast.Stmt

// Expr represents an expression in the AST.
type Expr = js_ast.Expr

// Class represents a JavaScript class declaration.
type Class = js_ast.Class

// Property represents a class property or method.
type Property = js_ast.Property

// Decorator represents a decorator on a class, method, or property.
type Decorator = js_ast.Decorator

// Part represents a part of the AST (used for code splitting).
type Part = js_ast.Part

// Scope represents a lexical scope in the AST.
type Scope = js_ast.Scope

// NamedImport represents a named import binding.
type NamedImport = js_ast.NamedImport

// NamedExport represents a named export.
type NamedExport = js_ast.NamedExport

// Options configures the parser behavior.
type Options struct {
	// TypeScript enables TypeScript parsing.
	TypeScript bool

	// JSX enables JSX parsing.
	JSX bool
}

// Error represents a parse error with location information.
type Error struct {
	Message string
	Line    int
	Column  int
}

func (e *Error) Error() string {
	return e.Message
}

// ParseResult contains the parsing result.
type ParseResult struct {
	AST    AST
	Source string
	Errors []Error
}

// Parse parses JavaScript or TypeScript source code and returns the AST.
func Parse(source string, opts Options) (*ParseResult, error) {
	log := logger.NewDeferLog(logger.DeferLogAll, nil)
	src := logger.Source{
		Contents: source,
	}

	configOpts := &config.Options{}
	if opts.TypeScript {
		configOpts.TS.Parse = true
	}
	parserOpts := js_parser.OptionsFromConfig(configOpts)

	ast, ok := js_parser.Parse(log, src, parserOpts)

	result := &ParseResult{
		AST:    ast,
		Source: source,
	}

	// Collect any errors
	msgs := log.Done()
	for _, msg := range msgs {
		if msg.Kind == logger.Error {
			line := 0
			col := 0
			if msg.Data.Location != nil {
				line = msg.Data.Location.Line
				col = msg.Data.Location.Column
			}
			result.Errors = append(result.Errors, Error{
				Message: msg.Data.Text,
				Line:    line,
				Column:  col,
			})
		}
	}

	if !ok {
		if len(result.Errors) > 0 {
			return result, &result.Errors[0]
		}
		return result, &Error{Message: "parse failed"}
	}

	return result, nil
}

// ParseFile is a convenience function that parses source code with sensible defaults.
// It enables TypeScript support by default.
func ParseFile(source string) (*ParseResult, error) {
	return Parse(source, Options{TypeScript: true})
}
