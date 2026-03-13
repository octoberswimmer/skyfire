package parser

import (
	"github.com/octoberswimmer/skyfire/internal/ast"
	"github.com/octoberswimmer/skyfire/internal/js_ast"
	"github.com/octoberswimmer/skyfire/internal/logger"
)

// Re-export commonly needed types for AST traversal

// Statement types
type (
	SImport        = js_ast.SImport
	SExportDefault = js_ast.SExportDefault
	SExportClause  = js_ast.SExportClause
	SExportFrom    = js_ast.SExportFrom
	SExportStar    = js_ast.SExportStar
	SLocal         = js_ast.SLocal
	SClass         = js_ast.SClass
	SFunction      = js_ast.SFunction
	SExpr          = js_ast.SExpr
)

// Expression types
type (
	EIdentifier       = js_ast.EIdentifier
	EImportIdentifier = js_ast.EImportIdentifier
	ECall             = js_ast.ECall
	EString           = js_ast.EString
	ENumber           = js_ast.ENumber
	EBoolean          = js_ast.EBoolean
	ENull             = js_ast.ENull
	EObject           = js_ast.EObject
	EArray            = js_ast.EArray
	EArrow            = js_ast.EArrow
	EFunction         = js_ast.EFunction
	EDot              = js_ast.EDot
	EIndex            = js_ast.EIndex
	ENew              = js_ast.ENew
	EBinary           = js_ast.EBinary
	EUnary            = js_ast.EUnary
	ETemplate         = js_ast.ETemplate
	EIf               = js_ast.EIf
)

// Binding types
type (
	Binding         = js_ast.Binding
	BIdentifier     = js_ast.BIdentifier
	BArray          = js_ast.BArray
	BObject         = js_ast.BObject
	PropertyBinding = js_ast.PropertyBinding
	ArrayBinding    = js_ast.ArrayBinding
)

// Other types
type (
	ClauseItem   = js_ast.ClauseItem
	PropertyKind = js_ast.PropertyKind
	Ref          = ast.Ref
	LocRef       = ast.LocRef
	ImportRecord = ast.ImportRecord
	Symbol       = ast.Symbol
	Loc          = logger.Loc
	Range        = logger.Range
	Path         = logger.Path
)

// PropertyKind constants
const (
	PropertyField             = js_ast.PropertyField
	PropertyMethod            = js_ast.PropertyMethod
	PropertyGetter            = js_ast.PropertyGetter
	PropertySetter            = js_ast.PropertySetter
	PropertyAutoAccessor      = js_ast.PropertyAutoAccessor
	PropertySpread            = js_ast.PropertySpread
	PropertyDeclareOrAbstract = js_ast.PropertyDeclareOrAbstract
	PropertyClassStaticBlock  = js_ast.PropertyClassStaticBlock
)

// PropertyFlags constants
type PropertyFlags = js_ast.PropertyFlags

const (
	PropertyIsComputed      = js_ast.PropertyIsComputed
	PropertyIsStatic        = js_ast.PropertyIsStatic
	PropertyWasShorthand    = js_ast.PropertyWasShorthand
	PropertyPreferQuotedKey = js_ast.PropertyPreferQuotedKey
)
