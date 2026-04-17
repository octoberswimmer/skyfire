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

// Tests for statement types

func TestParse_SIf(t *testing.T) {
	source := `
		function test(x) {
			if (x > 0) {
				console.log("positive");
			} else if (x < 0) {
				console.log("negative");
			} else {
				console.log("zero");
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Find the if statement inside the function
	var foundIf bool
	var foundElse bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if ifStmt, ok := bodyStmt.Data.(*SIf); ok {
						foundIf = true
						if ifStmt.NoOrNil.Data != nil {
							foundElse = true
						}
					}
				}
			}
		}
	}

	if !foundIf {
		t.Error("expected to find SIf statement")
	}
	if !foundElse {
		t.Error("expected to find else branch")
	}
}

func TestParse_SBlock(t *testing.T) {
	source := `
		function test() {
			{
				let x = 1;
				let y = 2;
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundBlock bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if block, ok := bodyStmt.Data.(*SBlock); ok {
						foundBlock = true
						if len(block.Stmts) != 2 {
							t.Errorf("expected 2 statements in block, got %d", len(block.Stmts))
						}
					}
				}
			}
		}
	}

	if !foundBlock {
		t.Error("expected to find SBlock statement")
	}
}

func TestParse_SFor(t *testing.T) {
	source := `
		function test() {
			for (let i = 0; i < 10; i++) {
				console.log(i);
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundFor bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if forStmt, ok := bodyStmt.Data.(*SFor); ok {
						foundFor = true
						if forStmt.InitOrNil.Data == nil {
							t.Error("expected for loop to have init")
						}
						if forStmt.TestOrNil.Data == nil {
							t.Error("expected for loop to have test")
						}
						if forStmt.UpdateOrNil.Data == nil {
							t.Error("expected for loop to have update")
						}
					}
				}
			}
		}
	}

	if !foundFor {
		t.Error("expected to find SFor statement")
	}
}

func TestParse_SForIn(t *testing.T) {
	source := `
		function test(obj) {
			for (let key in obj) {
				console.log(key);
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundForIn bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if _, ok := bodyStmt.Data.(*SForIn); ok {
						foundForIn = true
					}
				}
			}
		}
	}

	if !foundForIn {
		t.Error("expected to find SForIn statement")
	}
}

func TestParse_SForOf(t *testing.T) {
	source := `
		function test(arr) {
			for (let item of arr) {
				console.log(item);
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundForOf bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if _, ok := bodyStmt.Data.(*SForOf); ok {
						foundForOf = true
					}
				}
			}
		}
	}

	if !foundForOf {
		t.Error("expected to find SForOf statement")
	}
}

func TestParse_SWhile(t *testing.T) {
	source := `
		function test() {
			let i = 0;
			while (i < 10) {
				console.log(i);
				i++;
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundWhile bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if whileStmt, ok := bodyStmt.Data.(*SWhile); ok {
						foundWhile = true
						if whileStmt.Test.Data == nil {
							t.Error("expected while loop to have test condition")
						}
					}
				}
			}
		}
	}

	if !foundWhile {
		t.Error("expected to find SWhile statement")
	}
}

func TestParse_SDoWhile(t *testing.T) {
	source := `
		function test() {
			let i = 0;
			do {
				console.log(i);
				i++;
			} while (i < 10);
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundDoWhile bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if doWhile, ok := bodyStmt.Data.(*SDoWhile); ok {
						foundDoWhile = true
						if doWhile.Test.Data == nil {
							t.Error("expected do-while loop to have test condition")
						}
					}
				}
			}
		}
	}

	if !foundDoWhile {
		t.Error("expected to find SDoWhile statement")
	}
}

func TestParse_SSwitch(t *testing.T) {
	source := `
		function test(x) {
			switch (x) {
				case 1:
					console.log("one");
					break;
				case 2:
					console.log("two");
					break;
				default:
					console.log("other");
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundSwitch bool
	var caseCount int
	var foundDefault bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if switchStmt, ok := bodyStmt.Data.(*SSwitch); ok {
						foundSwitch = true
						caseCount = len(switchStmt.Cases)
						for _, c := range switchStmt.Cases {
							if c.ValueOrNil.Data == nil {
								foundDefault = true
							}
						}
					}
				}
			}
		}
	}

	if !foundSwitch {
		t.Error("expected to find SSwitch statement")
	}
	if caseCount != 3 {
		t.Errorf("expected 3 cases (including default), got %d", caseCount)
	}
	if !foundDefault {
		t.Error("expected to find default case")
	}
}

func TestParse_STry(t *testing.T) {
	source := `
		function test() {
			try {
				riskyOperation();
			} catch (e) {
				console.error(e);
			} finally {
				cleanup();
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundTry bool
	var hasCatch bool
	var hasFinally bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if tryStmt, ok := bodyStmt.Data.(*STry); ok {
						foundTry = true
						if tryStmt.Catch != nil {
							hasCatch = true
						}
						if tryStmt.Finally != nil {
							hasFinally = true
						}
					}
				}
			}
		}
	}

	if !foundTry {
		t.Error("expected to find STry statement")
	}
	if !hasCatch {
		t.Error("expected try to have catch block")
	}
	if !hasFinally {
		t.Error("expected try to have finally block")
	}
}

func TestParse_STry_CatchOnly(t *testing.T) {
	source := `
		function test() {
			try {
				riskyOperation();
			} catch (e) {
				console.error(e);
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundTry bool
	var hasCatch bool
	var hasFinally bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if tryStmt, ok := bodyStmt.Data.(*STry); ok {
						foundTry = true
						hasCatch = tryStmt.Catch != nil
						hasFinally = tryStmt.Finally != nil
					}
				}
			}
		}
	}

	if !foundTry {
		t.Error("expected to find STry statement")
	}
	if !hasCatch {
		t.Error("expected try to have catch block")
	}
	if hasFinally {
		t.Error("expected try to NOT have finally block")
	}
}

func TestParse_STry_FinallyOnly(t *testing.T) {
	source := `
		function test() {
			try {
				riskyOperation();
			} finally {
				cleanup();
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundTry bool
	var hasCatch bool
	var hasFinally bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fn.Fn.Body.Block.Stmts {
					if tryStmt, ok := bodyStmt.Data.(*STry); ok {
						foundTry = true
						hasCatch = tryStmt.Catch != nil
						hasFinally = tryStmt.Finally != nil
					}
				}
			}
		}
	}

	if !foundTry {
		t.Error("expected to find STry statement")
	}
	if hasCatch {
		t.Error("expected try to NOT have catch block")
	}
	if !hasFinally {
		t.Error("expected try to have finally block")
	}
}

func TestParse_SReturn(t *testing.T) {
	source := `
		function test(x) {
			if (x < 0) {
				return -1;
			}
			return x * 2;
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var returnCount int
	var returnWithValue int
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				countReturns(fn.Fn.Body.Block.Stmts, &returnCount, &returnWithValue)
			}
		}
	}

	if returnCount != 2 {
		t.Errorf("expected 2 return statements, got %d", returnCount)
	}
	if returnWithValue != 2 {
		t.Errorf("expected 2 returns with values, got %d", returnWithValue)
	}
}

func countReturns(stmts []Stmt, total *int, withValue *int) {
	for _, stmt := range stmts {
		switch s := stmt.Data.(type) {
		case *SReturn:
			*total++
			if s.ValueOrNil.Data != nil {
				*withValue++
			}
		case *SIf:
			if block, ok := s.Yes.Data.(*SBlock); ok {
				countReturns(block.Stmts, total, withValue)
			}
			if s.NoOrNil.Data != nil {
				if block, ok := s.NoOrNil.Data.(*SBlock); ok {
					countReturns(block.Stmts, total, withValue)
				}
			}
		case *SBlock:
			countReturns(s.Stmts, total, withValue)
		}
	}
}

func TestParse_SReturn_Void(t *testing.T) {
	source := `
		function test(x) {
			if (x < 0) {
				return;
			}
			console.log(x);
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var returnCount int
	var returnWithValue int
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				countReturns(fn.Fn.Body.Block.Stmts, &returnCount, &returnWithValue)
			}
		}
	}

	if returnCount != 1 {
		t.Errorf("expected 1 return statement, got %d", returnCount)
	}
	if returnWithValue != 0 {
		t.Errorf("expected 0 returns with values, got %d", returnWithValue)
	}
}

func TestParse_SThrow(t *testing.T) {
	source := `
		function test(x) {
			if (x < 0) {
				throw new Error("negative value");
			}
			return x;
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundThrow bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				foundThrow = findThrow(fn.Fn.Body.Block.Stmts)
			}
		}
	}

	if !foundThrow {
		t.Error("expected to find SThrow statement")
	}
}

func findThrow(stmts []Stmt) bool {
	for _, stmt := range stmts {
		switch s := stmt.Data.(type) {
		case *SThrow:
			return true
		case *SIf:
			if block, ok := s.Yes.Data.(*SBlock); ok {
				if findThrow(block.Stmts) {
					return true
				}
			}
		case *SBlock:
			if findThrow(s.Stmts) {
				return true
			}
		}
	}
	return false
}

func TestParse_NestedControlFlow(t *testing.T) {
	source := `
		function processItems(items) {
			for (let item of items) {
				if (item.valid) {
					try {
						switch (item.type) {
							case 'a':
								while (item.count > 0) {
									item.count--;
								}
								break;
							default:
								throw new Error("unknown type");
						}
					} catch (e) {
						return null;
					}
				}
			}
			return items;
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Just verify it parses without error and has statements
	var stmtCount int
	for _, part := range result.AST.Parts {
		stmtCount += len(part.Stmts)
	}

	if stmtCount == 0 {
		t.Error("expected statements in parsed AST")
	}
}

func TestParse_ClassMethodsWithControlFlow(t *testing.T) {
	source := `
		class DataProcessor {
			process(data) {
				if (!data) {
					return null;
				}

				for (let i = 0; i < data.length; i++) {
					try {
						this.validate(data[i]);
					} catch (e) {
						throw new Error("validation failed at " + i);
					}
				}

				return data;
			}

			validate(item) {
				switch (typeof item) {
					case 'string':
						return item.length > 0;
					case 'number':
						return !isNaN(item);
					default:
						return false;
				}
			}
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundClass bool
	var methodCount int
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if classStmt, ok := stmt.Data.(*SClass); ok {
				foundClass = true
				for _, prop := range classStmt.Class.Properties {
					if prop.Kind == PropertyMethod {
						methodCount++
					}
				}
			}
		}
	}

	if !foundClass {
		t.Error("expected to find class")
	}
	if methodCount != 2 {
		t.Errorf("expected 2 methods, got %d", methodCount)
	}
}

func TestParse_ArrowFunctionsWithControlFlow(t *testing.T) {
	source := `
		const processor = (items) => {
			for (const item of items) {
				if (item.skip) continue;
				if (item.stop) break;
			}
		};

		const filter = items => items.filter(item => {
			if (!item) return false;
			return item.valid;
		});
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var arrowCount int
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if local, ok := stmt.Data.(*SLocal); ok {
				for _, decl := range local.Decls {
					if _, ok := decl.ValueOrNil.Data.(*EArrow); ok {
						arrowCount++
					}
				}
			}
		}
	}

	if arrowCount != 2 {
		t.Errorf("expected 2 arrow functions, got %d", arrowCount)
	}
}

func TestParse_AsyncAwaitWithControlFlow(t *testing.T) {
	source := `
		async function fetchData(urls) {
			const results = [];
			for (const url of urls) {
				try {
					const response = await fetch(url);
					if (!response.ok) {
						throw new Error("Failed");
					}
					results.push(await response.json());
				} catch (e) {
					console.error(e);
				}
			}
			return results;
		}
	`
	result, err := ParseFile(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var foundAsync bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fn, ok := stmt.Data.(*SFunction); ok {
				if fn.Fn.IsAsync {
					foundAsync = true
				}
			}
		}
	}

	if !foundAsync {
		t.Error("expected to find async function")
	}
}

func TestParseJSX(t *testing.T) {
	source := `<div className="min-h-screen">Hello</div>`

	result, err := Parse(source, Options{JSX: true})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var found bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if exprStmt, ok := stmt.Data.(*SExpr); ok {
				if _, ok := exprStmt.Value.Data.(*EJSXElement); ok {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("expected JSX AST node")
	}
}

func TestParseTSX(t *testing.T) {
	source := `export function Example() {
  return <div className="min-h-screen">Hello</div>
}`

	result, err := Parse(source, Options{TypeScript: true, JSX: true, PreserveUnusedImports: true})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var found bool
	for _, part := range result.AST.Parts {
		for _, stmt := range part.Stmts {
			if fnStmt, ok := stmt.Data.(*SFunction); ok {
				for _, bodyStmt := range fnStmt.Fn.Body.Block.Stmts {
					if ret, ok := bodyStmt.Data.(*SReturn); ok {
						if _, ok := ret.ValueOrNil.Data.(*EJSXElement); ok {
							found = true
						}
					}
				}
			}
		}
	}
	if !found {
		t.Fatal("expected TSX return value to remain as JSX AST")
	}
}
