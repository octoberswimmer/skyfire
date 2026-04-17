package cssparser

import (
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	result, err := ParseFile(`@media (min-width: 640px) { .foo { color: red; } }`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(result.AST.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(result.AST.Rules))
	}
	if _, ok := result.AST.Rules[0].Data.(*RAtMedia); !ok {
		t.Fatalf("got %T, want *RAtMedia", result.AST.Rules[0].Data)
	}
}

func TestParseErrors(t *testing.T) {
	result, err := ParseFile(`/* unterminated`)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected error details")
	}
}

func TestPrint(t *testing.T) {
	result, err := ParseFile(`
@custom-variant dark (&:is(.dark *));
.foo {
  color: red;
}
`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	css := Print(result.AST, PrintOptions{})
	if !strings.Contains(css, "@custom-variant") {
		t.Fatalf("printed CSS missing at-rule: %q", css)
	}
	if !strings.Contains(css, ".foo") {
		t.Fatalf("printed CSS missing selector: %q", css)
	}
}

func TestPrintMinifyWhitespace(t *testing.T) {
	result, err := ParseFile(`
.foo {
  color: red;
  margin: 0 auto;
}
`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	css := Print(result.AST, PrintOptions{MinifyWhitespace: true})
	if strings.Contains(css, "\n") || strings.Contains(css, "  ") {
		t.Fatalf("expected minified CSS, got %q", css)
	}
	if !strings.Contains(css, ".foo{color: red;margin: 0 auto}") {
		t.Fatalf("unexpected minified CSS: %q", css)
	}
}

func TestImportPath(t *testing.T) {
	result, err := ParseFile(`@import "./theme.css" layer(base) supports(display: grid) screen and (min-width: 640px);`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	rule, ok := result.AST.Rules[0].Data.(*RAtImport)
	if !ok {
		t.Fatalf("got %T, want *RAtImport", result.AST.Rules[0].Data)
	}
	path, ok := ImportPath(result.AST, rule)
	if !ok {
		t.Fatal("expected import path")
	}
	if path != "./theme.css" {
		t.Fatalf("unexpected import path %q", path)
	}
	if _, ok := ImportPath(result.AST, nil); ok {
		t.Fatal("expected nil import rule to be rejected")
	}
}

func TestPrintTokens(t *testing.T) {
	result, err := ParseFile(`@theme inline { --color-background: var(--background); @keyframes enter { from { opacity: 0; } } }`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	unknown, ok := result.AST.Rules[0].Data.(*RUnknownAt)
	if !ok {
		t.Fatalf("got %T, want *RUnknownAt", result.AST.Rules[0].Data)
	}
	if len(unknown.Block) == 0 || unknown.Block[0].Children == nil {
		t.Fatalf("expected block children")
	}
	printed := PrintTokens(*unknown.Block[0].Children)
	if !strings.Contains(printed, `--color-background:`) || !strings.Contains(printed, `var(--background)`) {
		t.Fatalf("unexpected token output: %q", printed)
	}
	if !strings.Contains(printed, `@keyframes`) || !strings.Contains(printed, `enter`) {
		t.Fatalf("expected nested at-rule in token output: %q", printed)
	}
	if !strings.Contains(printed, `opacity:`) || !strings.Contains(printed, `0;`) {
		t.Fatalf("expected nested declarations in token output: %q", printed)
	}
}

func TestPrintTokensPreservesHashColors(t *testing.T) {
	result, err := ParseFile(`.x { color: #212121; }`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	selector, ok := result.AST.Rules[0].Data.(*RSelector)
	if !ok || len(selector.Rules) == 0 {
		t.Fatalf("unexpected first rule shape: %T", result.AST.Rules[0].Data)
	}
	decl, ok := selector.Rules[0].Data.(*RDeclaration)
	if !ok {
		t.Fatalf("expected declaration")
	}
	if printed := strings.TrimSpace(PrintTokens(decl.Value)); printed != `#212121` {
		t.Fatalf("unexpected token output: %q", printed)
	}
}

func TestExportsTokenKinds(t *testing.T) {
	result, err := ParseFile(`.x { color: rgb(1 2 3 / 50%); width: 1px; opacity: 50%; } @media (min-width: 1px) {}`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	selector, ok := result.AST.Rules[0].Data.(*RSelector)
	if !ok || len(selector.Rules) < 3 {
		t.Fatalf("unexpected first rule shape: %T", result.AST.Rules[0].Data)
	}
	declColor, ok := selector.Rules[0].Data.(*RDeclaration)
	if !ok || len(declColor.Value) != 1 || declColor.Value[0].Kind != TokenFunction {
		t.Fatalf("expected function token, got %#v", declColor.Value)
	}
	declWidth, ok := selector.Rules[1].Data.(*RDeclaration)
	if !ok || len(declWidth.Value) != 1 || declWidth.Value[0].Kind != TokenDimension {
		t.Fatalf("expected dimension token, got %#v", declWidth.Value)
	}
	declOpacity, ok := selector.Rules[2].Data.(*RDeclaration)
	if !ok || len(declOpacity.Value) != 1 || declOpacity.Value[0].Kind != TokenPercentage {
		t.Fatalf("expected percentage token, got %#v", declOpacity.Value)
	}
	media, ok := result.AST.Rules[1].Data.(*RAtMedia)
	if !ok || len(media.Queries) == 0 {
		t.Fatalf("expected media query rule")
	}
	plain, ok := media.Queries[0].Data.(*MQPlainOrBoolean)
	if !ok || len(plain.ValueOrNil) != 1 || plain.ValueOrNil[0].Kind != TokenDimension {
		t.Fatalf("expected dimension token in media query, got %#v", plain.ValueOrNil)
	}
}

func TestExportsWhitespaceFlags(t *testing.T) {
	result, err := ParseFile(`:root{--a:calc(100% - var(--gap));}`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	selector, ok := result.AST.Rules[0].Data.(*RSelector)
	if !ok || len(selector.Rules) == 0 {
		t.Fatalf("unexpected first rule shape: %T", result.AST.Rules[0].Data)
	}
	decl, ok := selector.Rules[0].Data.(*RDeclaration)
	if !ok || len(decl.Value) != 1 {
		t.Fatalf("expected declaration value, got %#v", selector.Rules)
	}
	fn := decl.Value[0]
	if fn.Kind != TokenFunction || fn.Children == nil {
		t.Fatalf("expected function token, got %#v", fn)
	}
	children := *fn.Children
	if len(children) < 2 {
		t.Fatalf("expected function children, got %#v", children)
	}
	if children[1].Whitespace&WhitespaceBefore == 0 {
		t.Fatalf("expected whitespace-before flag on operator token, got %#v", children[1].Whitespace)
	}
	if children[1].Whitespace&WhitespaceAfter == 0 {
		t.Fatalf("expected whitespace-after flag on operator token, got %#v", children[1].Whitespace)
	}
}

func TestExportsMediaQueryAliasSurface(t *testing.T) {
	result, err := ParseFile(`@media not screen and (min-width: 640px), (400px <= width <= 1200px), (color) {}`)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	media, ok := result.AST.Rules[0].Data.(*RAtMedia)
	if !ok {
		t.Fatalf("got %T, want *RAtMedia", result.AST.Rules[0].Data)
	}
	if len(media.Queries) != 3 {
		t.Fatalf("got %d queries, want 3", len(media.Queries))
	}
	notQuery, ok := media.Queries[0].Data.(*MQType)
	if !ok {
		t.Fatalf("expected MQType for first query, got %T", media.Queries[0].Data)
	}
	if notQuery.Op != MQTypeOpNot || notQuery.Type != "screen" {
		t.Fatalf("unexpected MQType values: %#v", notQuery)
	}
	if _, ok := notQuery.AndOrNull.Data.(*MQPlainOrBoolean); !ok {
		t.Fatalf("expected MQPlainOrBoolean tail, got %T", notQuery.AndOrNull.Data)
	}
	rangeQuery, ok := media.Queries[1].Data.(*MQRange)
	if !ok {
		t.Fatalf("expected MQRange for second query, got %T", media.Queries[1].Data)
	}
	if rangeQuery.Name != "width" || len(rangeQuery.Before) != 1 || len(rangeQuery.After) != 1 {
		t.Fatalf("unexpected MQRange contents: %#v", rangeQuery)
	}
	plainQuery, ok := media.Queries[2].Data.(*MQPlainOrBoolean)
	if !ok {
		t.Fatalf("expected MQPlainOrBoolean for third query, got %T", media.Queries[2].Data)
	}
	if plainQuery.Name != "color" || plainQuery.ValueOrNil != nil {
		t.Fatalf("unexpected plain query contents: %#v", plainQuery)
	}
}
