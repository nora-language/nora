package semantic

import "github.com/DwiYI/Project-Nora/pkg/types"

// ModuleType represents an imported package.
// It resides in 'semantic' because it holds a *Scope.
type ModuleType struct {
	Ident   string
	Exports *Scope // Direct access to Scope is now fine!
}

// --- Implement types.NRType Interface ---

func (m *ModuleType) Name() string {
	return m.Ident
}

func (m *ModuleType) GetKind() types.Kind {
	return types.KindModule
}

// Modules are compile-time artifacts, not runtime values.
func (m *ModuleType) IsLeased() bool {
	return false
}

func (m *ModuleType) Size() int {
	return 0
}
