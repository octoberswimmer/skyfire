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
	SExportEquals  = js_ast.SExportEquals
	SLazyExport    = js_ast.SLazyExport
	SLocal         = js_ast.SLocal
	SClass         = js_ast.SClass
	SFunction      = js_ast.SFunction
	SExpr          = js_ast.SExpr
	SIf            = js_ast.SIf
	SBlock         = js_ast.SBlock
	SFor           = js_ast.SFor
	SForIn         = js_ast.SForIn
	SForOf         = js_ast.SForOf
	SWhile         = js_ast.SWhile
	SDoWhile       = js_ast.SDoWhile
	SSwitch        = js_ast.SSwitch
	STry           = js_ast.STry
	SReturn        = js_ast.SReturn
	SThrow         = js_ast.SThrow
	SEmpty         = js_ast.SEmpty
	SComment       = js_ast.SComment
	SDebugger      = js_ast.SDebugger
	SDirective     = js_ast.SDirective
	STypeScript    = js_ast.STypeScript
	SEnum          = js_ast.SEnum
	SNamespace     = js_ast.SNamespace
	SLabel         = js_ast.SLabel
	SWith          = js_ast.SWith
	SBreak         = js_ast.SBreak
	SContinue      = js_ast.SContinue
)

// Expression types
type (
	EIdentifier           = js_ast.EIdentifier
	EImportIdentifier     = js_ast.EImportIdentifier
	EPrivateIdentifier    = js_ast.EPrivateIdentifier
	ENameOfSymbol         = js_ast.ENameOfSymbol
	ECall                 = js_ast.ECall
	EString               = js_ast.EString
	ENumber               = js_ast.ENumber
	EBoolean              = js_ast.EBoolean
	ENull                 = js_ast.ENull
	EObject               = js_ast.EObject
	EArray                = js_ast.EArray
	EArrow                = js_ast.EArrow
	EFunction             = js_ast.EFunction
	EClass                = js_ast.EClass
	EDot                  = js_ast.EDot
	EIndex                = js_ast.EIndex
	ENew                  = js_ast.ENew
	EBinary               = js_ast.EBinary
	EUnary                = js_ast.EUnary
	ETemplate             = js_ast.ETemplate
	EIf                   = js_ast.EIf
	ESpread               = js_ast.ESpread
	EJSXElement           = js_ast.EJSXElement
	EJSXText              = js_ast.EJSXText
	EAnnotation           = js_ast.EAnnotation
	EAwait                = js_ast.EAwait
	EYield                = js_ast.EYield
	ERequireString        = js_ast.ERequireString
	ERequireResolveString = js_ast.ERequireResolveString
	EImportString         = js_ast.EImportString
	EImportCall           = js_ast.EImportCall
)

// Binding types
type (
	Binding         = js_ast.Binding
	BIdentifier     = js_ast.BIdentifier
	BArray          = js_ast.BArray
	BObject         = js_ast.BObject
	BMissing        = js_ast.BMissing
	PropertyBinding = js_ast.PropertyBinding
	ArrayBinding    = js_ast.ArrayBinding
)

// Other types
type (
	ClauseItem    = js_ast.ClauseItem
	Decl          = js_ast.Decl
	Arg           = js_ast.Arg
	Fn            = js_ast.Fn
	FnBody        = js_ast.FnBody
	TemplatePart  = js_ast.TemplatePart
	PropertyKind  = js_ast.PropertyKind
	Ref           = ast.Ref
	LocRef        = ast.LocRef
	ImportRecord  = ast.ImportRecord
	Symbol        = ast.Symbol
	Loc           = logger.Loc
	Range         = logger.Range
	Path          = logger.Path
	Case          = js_ast.Case
	Catch         = js_ast.Catch
	Finally       = js_ast.Finally
	EnumValue     = js_ast.EnumValue
	NamedImportEx = js_ast.NamedImport
	NamedExportEx = js_ast.NamedExport
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
