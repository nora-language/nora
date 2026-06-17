package semantic

import (
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/types"
)

// Visibility defines access levels (Public/Private based on capitalization)
type Visibility int

const (
	Private Visibility = iota
	Public
)

type ScopeKind string

const (
	ScopeGlobal   ScopeKind = "global"
	ScopePackage  ScopeKind = "package"
	ScopeFunction ScopeKind = "function"
	ScopeBlock    ScopeKind = "block"
	ScopeLoop     ScopeKind = "loop"
	ScopeSpawn    ScopeKind = "spawn"   // Critical for Nora
	ScopeClosure  ScopeKind = "closure" // For Lambdas
)

// SymbolKind defines what an identifier REPRESENTS.
type SymbolKind int

const (
	SymVar     SymbolKind = iota // let x = 5
	SymFunc                      // fn add()
	SymType                      // type User ...
	SymModule                    // import "math"
	SymParam                     // fn(x: int)
	SymPackage                   // fn(x: int)
	SymVariant                   // Some, None, Active, etc.
)

// String representation for debugging
func (sk SymbolKind) String() string {
	return [...]string{"Variable", "Function", "Type", "Module", "Parameter", "Package", "Variant"}[sk]
}

// Symbol represents a named entity (variable, function, type)
type Symbol struct {
	Name    string
	Type    types.NRType // Interface to your type system
	Kind    SymbolKind   // "var", "const", "func", "type"
	Visible Visibility

	// Definition Context
	DefScope *Scope   // The scope where this was defined
	DefNode  ast.Node // The AST node defining it

	// Lease kind
	LeaseKind types.LeaseKind

	WritePerm bool

	IsInline bool // Indicates if the function is marked for inlining

	// Usage Analysis (Nora)
	IsUsed        bool
	IsCaptured    bool // True if used inside a 'spawn' block but defined outside
	UseCount      int
	Version       int // For tracking mutations (SSA-lite)
	IsInitialized bool

	IsPinned bool
}

type Scope struct {
	Parent      *Scope
	Kind        ScopeKind
	Symbols     map[string]*Symbol
	PackageName string                 // Name of the package if Kind == ScopePackage
	Captures    map[*Symbol]bool       // Symbols captured by this scope (if closure/spawn)
	Bounds      map[*Symbol]*VarBounds // Track bounds of variables for BCE
}

// NewScope creates a child scope
func NewScope(parent *Scope, kind ScopeKind) *Scope {
	return &Scope{
		Parent:   parent,
		Kind:     kind,
		Symbols:  make(map[string]*Symbol),
		Captures: make(map[*Symbol]bool),
		Bounds:   make(map[*Symbol]*VarBounds),
	}
}

// GetBounds recursively searches for the variable bounds upwards through the scope chain.
func (s *Scope) GetBounds(sym *Symbol) *VarBounds {
	if bounds, exists := s.Bounds[sym]; exists {
		return bounds
	}
	if s.Parent != nil {
		return s.Parent.GetBounds(sym)
	}
	return nil
}

type InitState int

const (
	Uninitialized InitState = iota
	Initialized
	PartiallyInitialized
)
