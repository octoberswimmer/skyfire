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
