// Package cssparser provides a CSS parser with AST access and simple printing.
// It wraps Skyfire's internal CSS parser with a smaller public API.
package cssparser

import (
	"strings"

	"github.com/octoberswimmer/skyfire/internal/ast"
	"github.com/octoberswimmer/skyfire/internal/config"
	"github.com/octoberswimmer/skyfire/internal/css_ast"
	"github.com/octoberswimmer/skyfire/internal/css_lexer"
	"github.com/octoberswimmer/skyfire/internal/css_parser"
	"github.com/octoberswimmer/skyfire/internal/css_printer"
	"github.com/octoberswimmer/skyfire/internal/logger"
)

// AST represents the parsed CSS abstract syntax tree.
type AST = css_ast.AST

// Rule represents a CSS rule in the AST.
type Rule = css_ast.Rule

// R represents the interface implemented by all CSS rule variants.
type R = css_ast.R

// Token represents a CSS token in the AST.
type Token = css_ast.Token

// TokenKind represents the kind of a CSS token.
type TokenKind = css_lexer.T

const (
	TokenAtKeyword  TokenKind = css_lexer.TAtKeyword
	TokenDimension  TokenKind = css_lexer.TDimension
	TokenFunction   TokenKind = css_lexer.TFunction
	TokenHash       TokenKind = css_lexer.THash
	TokenIdent      TokenKind = css_lexer.TIdent
	TokenNumber     TokenKind = css_lexer.TNumber
	TokenPercentage TokenKind = css_lexer.TPercentage
)

// ImportRecord represents a CSS import record in the AST.
type ImportRecord = ast.ImportRecord

// WhitespaceFlags represents the presence of whitespace before or after a token.
type WhitespaceFlags = css_ast.WhitespaceFlags

const (
	// WhitespaceBefore indicates whitespace before a token.
	WhitespaceBefore = css_ast.WhitespaceBefore

	// WhitespaceAfter indicates whitespace after a token.
	WhitespaceAfter = css_ast.WhitespaceAfter
)

// RAtCharset represents a parsed "@charset" rule.
type RAtCharset = css_ast.RAtCharset

// RAtImport represents a parsed "@import" rule.
type RAtImport = css_ast.RAtImport

// RAtKeyframes represents a parsed "@keyframes" rule.
type RAtKeyframes = css_ast.RAtKeyframes

// RKnownAt represents a known at-rule with nested rules.
type RKnownAt = css_ast.RKnownAt

// RUnknownAt represents an unknown at-rule.
type RUnknownAt = css_ast.RUnknownAt

// RSelector represents a selector rule.
type RSelector = css_ast.RSelector

// RQualified represents a qualified rule.
type RQualified = css_ast.RQualified

// RDeclaration represents a declaration rule.
type RDeclaration = css_ast.RDeclaration

// RBadDeclaration represents a declaration that could not be parsed.
type RBadDeclaration = css_ast.RBadDeclaration

// RComment represents a comment rule.
type RComment = css_ast.RComment

// RAtLayer represents an "@layer" rule.
type RAtLayer = css_ast.RAtLayer

// RAtMedia represents an "@media" rule.
type RAtMedia = css_ast.RAtMedia

// MediaQuery represents a CSS media query in the AST.
type MediaQuery = css_ast.MediaQuery

// MQ represents the interface implemented by media query variants.
type MQ = css_ast.MQ

// MQType represents a media type query.
type MQType = css_ast.MQType

// MQTypeOp represents the qualifier on a media type query.
type MQTypeOp = css_ast.MQTypeOp

const (
	MQTypeOpNone MQTypeOp = css_ast.MQTypeOpNone
	MQTypeOpNot  MQTypeOp = css_ast.MQTypeOpNot
	MQTypeOpOnly MQTypeOp = css_ast.MQTypeOpOnly
)

// MQNot represents a negated media query.
type MQNot = css_ast.MQNot

// MQBinary represents a binary media query expression.
type MQBinary = css_ast.MQBinary

// MQBinaryOp represents the boolean operator between media query terms.
type MQBinaryOp = css_ast.MQBinaryOp

const (
	MQBinaryOpAnd MQBinaryOp = css_ast.MQBinaryOpAnd
	MQBinaryOpOr  MQBinaryOp = css_ast.MQBinaryOpOr
)

// MQArbitraryTokens represents an arbitrary-token media query.
type MQArbitraryTokens = css_ast.MQArbitraryTokens

// MQPlainOrBoolean represents a plain or boolean media feature query.
type MQPlainOrBoolean = css_ast.MQPlainOrBoolean

// MQRange represents a range media feature query.
type MQRange = css_ast.MQRange

// MQCmp represents a media query comparison operator.
type MQCmp = css_ast.MQCmp

// RAtScope represents an "@scope" rule.
type RAtScope = css_ast.RAtScope

// Options configures parser behavior.
type Options struct {
	MinifyWhitespace bool
}

// PrintOptions configures CSS printing behavior.
type PrintOptions struct {
	MinifyWhitespace bool
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

// Parse parses CSS source code and returns the AST.
func Parse(source string, opts Options) (*ParseResult, error) {
	log := logger.NewDeferLog(logger.DeferLogAll, nil)
	src := logger.Source{
		Contents: source,
	}
	configOpts := &config.Options{
		MinifyWhitespace: opts.MinifyWhitespace,
	}
	tree := css_parser.Parse(log, src, css_parser.OptionsFromConfig(config.LoaderCSS, configOpts))
	result := &ParseResult{
		AST:    tree,
		Source: source,
	}

	msgs := log.Done()
	for _, msg := range msgs {
		if msg.Kind != logger.Error {
			continue
		}
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

	if len(result.Errors) > 0 {
		return result, &result.Errors[0]
	}
	return result, nil
}

// ParseFile is a convenience function that parses CSS source code with
// default options.
func ParseFile(source string) (*ParseResult, error) {
	return Parse(source, Options{})
}

// Print prints the AST back to CSS source.
func Print(tree AST, opts PrintOptions) string {
	symbols := ast.NewSymbolMap(1)
	symbols.SymbolsForSource[0] = tree.Symbols
	result := css_printer.Print(tree, symbols, css_printer.Options{
		MinifyWhitespace: opts.MinifyWhitespace,
	})
	return string(result.CSS)
}

// ImportPath returns the path text referenced by an @import rule.
func ImportPath(tree AST, rule *RAtImport) (string, bool) {
	if rule == nil {
		return "", false
	}
	index := int(rule.ImportRecordIndex)
	if index < 0 || index >= len(tree.ImportRecords) {
		return "", false
	}
	return tree.ImportRecords[index].Path.Text, true
}

// PrintTokens prints a CSS token list back to source text.
func PrintTokens(tokens []Token) string {
	var builder strings.Builder
	printTokens(&builder, tokens)
	return builder.String()
}

func printTokens(builder *strings.Builder, tokens []Token) {
	for _, token := range tokens {
		if token.Whitespace&WhitespaceBefore != 0 {
			builder.WriteByte(' ')
		}
		printToken(builder, token)
		if token.Whitespace&WhitespaceAfter != 0 {
			builder.WriteByte(' ')
		}
	}
}

func printToken(builder *strings.Builder, token Token) {
	if token.Children == nil {
		if token.Kind == TokenAtKeyword {
			builder.WriteByte('@')
		}
		if token.Kind == TokenHash {
			builder.WriteByte('#')
		}
		builder.WriteString(token.Text)
		return
	}
	switch token.Text {
	case "{", "[", "(":
		builder.WriteString(token.Text)
		printTokens(builder, *token.Children)
		switch token.Text {
		case "{":
			builder.WriteByte('}')
		case "[":
			builder.WriteByte(']')
		default:
			builder.WriteByte(')')
		}
	default:
		if token.Kind == TokenAtKeyword {
			builder.WriteByte('@')
		}
		builder.WriteString(token.Text)
		builder.WriteByte('(')
		printTokens(builder, *token.Children)
		builder.WriteByte(')')
	}
}
