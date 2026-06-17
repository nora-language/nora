package semantic

import (
	"fmt"
	"unicode"

	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

// Define registers a new symbol in the current scope
func (s *Scope) Define(name string, t types.NRType, kind SymbolKind, node ast.Node) (*Symbol, error) {
	if _, exists := s.Symbols[name]; exists {
		return nil, fmt.Errorf("symbol '%s' already defined in this scope", name)
	}

	if s.Parent != nil {
		if kind == SymVar || kind == SymParam || kind == SymFunc {
			root := s
			for root.Parent != nil {
				root = root.Parent
			}
			if existing, found := root.Symbols[name]; found && existing.Kind == SymType {
				return nil, fmt.Errorf("cannot shadow built-in type '%s' with variable definition", name)
			}
			if existing, found := s.Lookup(name); found && existing.Kind == SymPackage {
				return nil, fmt.Errorf("cannot shadow imported package '%s' with variable definition", name)
			}
		}
	}

	vis := Private
	if len(name) > 0 && unicode.IsUpper(rune(name[0])) {
		vis = Public
	}

	sym := &Symbol{
		Name:          name,
		Type:          t,
		Kind:          kind,
		Visible:       vis,
		DefScope:      s,
		DefNode:       node,
		IsInitialized: true,
	}

	s.Symbols[name] = sym
	return sym, nil
}

// Resolve finds a symbol recursively.
// It also handles "Capture" detection for the Lease system.
func (s *Scope) Resolve(name string) (*Symbol, bool) {
	// 1. Check current scope
	if sym, exists := s.Symbols[name]; exists {
		sym.IsUsed = true
		sym.UseCount++
		return sym, true
	}

	// 2. Check parent scope recursively
	if s.Parent != nil {
		sym, found := s.Parent.Resolve(name)
		if found {
			// --- Nora / LEASE LOGIC ---
			// If we are currently inside a 'spawn' or 'closure' context,
			// and we resolve something from above it, it is a Capture.
			// ONLY local variables (SymVar) and parameters (SymParam) defined in non-global, non-package scopes should be captured.
			isLocalVal := (sym.Kind == SymVar || sym.Kind == SymParam) &&
				(sym.DefScope != nil && sym.DefScope.Kind != ScopePackage && sym.DefScope.Kind != ScopeGlobal)
			if isLocalVal {
				curr := s
				isCaptured := false
				for curr != nil && curr != sym.DefScope {
					if curr.Kind == ScopeClosure {
						isCaptured = true
						curr.Captures[sym] = true
					}
					curr = curr.Parent
				}
				if isCaptured {
					sym.IsCaptured = true
				}
			}
			return sym, true
		}
	}

	return nil, false
}

// ResolveType looks for type definitions (structs, etc.)
// Assumes types are stored in the symbol table with Kind="type"
func (s *Scope) ResolveType(name string) (*Symbol, bool) {
	sym, found := s.Resolve(name)
	if found && sym.Kind == SymType {
		return sym, true
	}
	return nil, false
}

// Lookup finds a symbol recursively WITHOUT marking it as used or captured.
func (s *Scope) Lookup(name string) (*Symbol, bool) {
	if sym, exists := s.Symbols[name]; exists {
		return sym, true
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil, false
}
