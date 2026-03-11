# Skyfire

A fast JavaScript/TypeScript parser library for Go, forked from [esbuild](https://github.com/evanw/esbuild).

## Why Skyfire?

Skyfire exposes esbuild's internal JavaScript parser as a public Go API. This enables:

- **AST access**: Full access to the parsed JavaScript/TypeScript AST
- **Speed**: Leverages esbuild's extremely fast parser
- **ES6+ support**: Handles modern JavaScript including modules, decorators, and class fields
- **Error reporting**: Detailed parse errors with line and column numbers

## Use Cases

- Static analysis of JavaScript/TypeScript code
- Extracting metadata from ES6 modules (imports, exports, decorators)
- Building development tools that need to understand JS/TS structure

## Installation

```bash
go get github.com/octoberswimmer/skyfire
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/octoberswimmer/skyfire/parser"
)

func main() {
    source := `
        import { foo } from 'bar';
        export default class MyClass {
            @api myProperty;
        }
    `

    ast, err := parser.Parse(source, parser.Options{})
    if err != nil {
        fmt.Printf("Parse error: %v\n", err)
        return
    }

    // Access imports, exports, classes, decorators, etc.
    for _, imp := range ast.Imports {
        fmt.Printf("Import: %s\n", imp.Path)
    }
}
```

## License

MIT License - see [LICENSE.md](LICENSE.md)

## Acknowledgments

This project is a fork of [esbuild](https://github.com/evanw/esbuild) by Evan Wallace.
The core parser code is from esbuild with modifications to expose it as a public API.
