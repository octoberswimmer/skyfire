package parser

import (
	"os"
	"testing"
)

func TestParseSimple(t *testing.T) {
	source := `
		import { foo } from 'bar';
		export default class MyClass {
			myProperty = 1;
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.AST.Parts) == 0 {
		t.Fatal("Expected at least one part in AST")
	}
}

func TestParseWithDecorators(t *testing.T) {
	source := `
		import { LightningElement, api } from 'lwc';
		export default class MyComponent extends LightningElement {
			@api myProperty;
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.AST.Parts) == 0 {
		t.Fatal("Expected at least one part in AST")
	}
}

func TestParsePropertyListMap(t *testing.T) {
	// This file failed to parse with gotreesitter
	content, err := os.ReadFile(os.ExpandEnv("$HOME/git/dreamhouse-lwc/force-app/main/default/lwc/propertyListMap/propertyListMap.js"))
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	result, parseErr := ParseFile(string(content))
	if parseErr != nil {
		t.Fatalf("Parse failed: %v", parseErr)
	}

	if len(result.AST.Parts) == 0 {
		t.Fatal("Expected at least one part in AST")
	}

	// Verify we got import records
	if len(result.AST.ImportRecords) == 0 {
		t.Fatal("Expected import records")
	}

	t.Logf("Parsed successfully: %d parts, %d import records",
		len(result.AST.Parts), len(result.AST.ImportRecords))
}

func TestParseErrors(t *testing.T) {
	source := `
		class {  // Missing class name
			foo() {}
		}
	`
	result, err := ParseFile(source)
	if err == nil {
		t.Fatal("Expected parse error")
	}

	if len(result.Errors) == 0 {
		t.Fatal("Expected error details")
	}

	t.Logf("Got expected error: %s (line %d, col %d)",
		result.Errors[0].Message, result.Errors[0].Line, result.Errors[0].Column)
}

func TestParseTypeScriptTypeAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		propName string
		wantType string
	}{
		{
			name: "simple string type",
			source: `class Foo {
				name: string
			}`,
			propName: "name",
			wantType: "string",
		},
		{
			name: "number type",
			source: `class Foo {
				count: number
			}`,
			propName: "count",
			wantType: "number",
		},
		{
			name: "array type",
			source: `class Foo {
				items: string[]
			}`,
			propName: "items",
			wantType: "string[]",
		},
		{
			name: "generic type",
			source: `class Foo {
				data: Map<string, number>
			}`,
			propName: "data",
			wantType: "Map<string, number>",
		},
		{
			name: "union type",
			source: `class Foo {
				value: string | number
			}`,
			propName: "value",
			wantType: "string | number",
		},
		{
			name: "type with initializer",
			source: `class Foo {
				name: string = 'hello'
			}`,
			propName: "name",
			wantType: "string",
		},
		{
			name: "no type annotation",
			source: `class Foo {
				name = 'hello'
			}`,
			propName: "name",
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFile(tt.source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Find the class and property
			var gotType string
			for _, part := range result.AST.Parts {
				for _, stmt := range part.Stmts {
					if classStmt, ok := stmt.Data.(*SClass); ok {
						for _, prop := range classStmt.Class.Properties {
							if str, ok := prop.Key.Data.(*EString); ok {
								if stringFromUTF16(str.Value) == tt.propName {
									gotType = prop.TSTypeAnnotation
									break
								}
							}
						}
					}
				}
			}

			if gotType != tt.wantType {
				t.Errorf("got type %q, want %q", gotType, tt.wantType)
			}
		})
	}
}

func stringFromUTF16(data []uint16) string {
	runes := make([]rune, len(data))
	for i, v := range data {
		runes[i] = rune(v)
	}
	return string(runes)
}

func TestParseFile_PreservesUnusedImports(t *testing.T) {
	// This import is unused in the code but should be preserved by ParseFile
	source := `import { unusedFunction } from 'some-module';
export default class Foo {}`

	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the import statement
	var foundImport bool
	var foundBinding string
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if imp, ok := stmt.Data.(*SImport); ok {
				if imp.Items != nil {
					for _, item := range *imp.Items {
						foundImport = true
						foundBinding = item.Alias
					}
				}
			}
		}
	}

	if !foundImport {
		t.Fatal("Expected unused import to be preserved in AST")
	}
	if foundBinding != "unusedFunction" {
		t.Errorf("Expected binding 'unusedFunction', got %q", foundBinding)
	}
}

func TestParse_WithoutPreserveUnusedImports_RemovesUnused(t *testing.T) {
	// This import is unused and should be removed when PreserveUnusedImports is false
	source := `import { unusedFunction } from 'some-module';
export default class Foo {}`

	result, err := Parse(source, Options{TypeScript: true, PreserveUnusedImports: false})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// The import statement should be removed
	var foundImport bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if _, ok := stmt.Data.(*SImport); ok {
				foundImport = true
			}
		}
	}

	if foundImport {
		t.Fatal("Expected unused import to be removed from AST when PreserveUnusedImports is false")
	}
}

func TestParse_WithPreserveUnusedImports_KeepsUnused(t *testing.T) {
	// This import is unused but should be preserved when PreserveUnusedImports is true
	source := `import { unusedFunction } from 'some-module';
export default class Foo {}`

	result, err := Parse(source, Options{TypeScript: true, PreserveUnusedImports: true})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the import statement
	var foundImport bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if imp, ok := stmt.Data.(*SImport); ok {
				if imp.Items != nil && len(*imp.Items) > 0 {
					foundImport = true
				}
			}
		}
	}

	if !foundImport {
		t.Fatal("Expected unused import to be preserved in AST when PreserveUnusedImports is true")
	}
}
