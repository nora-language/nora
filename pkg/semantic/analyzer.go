package semantic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/nora-language/nora/pkg/diag"
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

// SemanticInfo holds the "Side Tables" - the results of analysis.
// This decouples the AST from the Semantic logic (No import cycles!)
type SemanticInfo struct {
	Defs               map[*ast.Identifier]*Symbol // Usage -> Definition
	Types              map[ast.Node]types.NRType   // Node -> Type
	Scopes             map[ast.Node]*Scope         // Node -> Scope
	Uses               map[*ast.Identifier]*Symbol
	DeadSyms           map[*Symbol]ast.Node
	Instances          map[*ast.FunctionStatement]map[string]*ast.FunctionStatement // genericFn -> typeArgString -> specializedFn
	MonomorphizedNames map[*ast.CallExpression]string
	SpecTypes          map[string]types.NRType
	FieldSymbols       map[*types.StructType]map[string]*Symbol
	MethodSymbols      map[types.NRType]map[string]*Symbol
}

// Add to SemanticInfo struct
// DeadSyms map[*Symbol]ast.Node // Maps symbol to the node that killed it (for error reporting)

func (si *SemanticInfo) Kill(sym *Symbol, killer ast.Node) {
	if sym.Kind != SymVar && sym.Kind != SymParam {
		return
	}
	if sym.Type != nil && sym.Type.GetKind() == types.KindFunction {
		return
	}
	// Safety: Lazy initialization if NewAnalyzer wasn't used correctly
	if si.DeadSyms == nil {
		si.DeadSyms = make(map[*Symbol]ast.Node)
	}
	si.DeadSyms[sym] = killer
}

func (si *SemanticInfo) IsDead(sym *Symbol) bool {
	if si.DeadSyms == nil {
		return false
	}
	_, dead := si.DeadSyms[sym]
	return dead
}

func (si *SemanticInfo) Alive(sym *Symbol) {
	if si.DeadSyms != nil {
		delete(si.DeadSyms, sym)
	}
}

// SemanticError is deprecated, use diag.Diagnostic

type SemanticAnalyzer struct {
	// The Root Scope (Universe) containing built-ins like i32, str, print
	GlobalScope    *Scope
	CurrentScope   *Scope
	Diagnostics    *diag.Collection
	InitStates     map[*Symbol]InitState
	Context        context.Context // Added for cancellation
	SemanticInfo   SemanticInfo
	Loader         PackageLoader // <--- New Dependency
	analyzingTypes map[*Symbol]bool

	// Context for analysis
	CurrentFunction   *ast.FunctionStatement
	CurrentLambda     *ast.LambdaExpression
	AllowUnsafe       bool              // <--- Set by compiler flag
	AllowedUnsafeDirs []string          // <--- Set from Project Manifest
	PackageScopes     map[string]*Scope // <--- Track scopes by package name
	LoadedDirs        map[string]bool   // <--- Track loaded directories to avoid recursion
	ProcessedFiles    map[string]bool   // <--- Prevent duplicate symbol collection
	AnalyzedFiles     map[string]bool   // <--- Prevent duplicate deep analysis
	CollectingSymbols bool              // <--- True during CollectSymbols passes

	FuncScopes           map[*ast.FunctionStatement]*Scope // <--- Track original scopes where functions are defined
	depth                int                               // Recursion depth for infinite loop detection
	DebugMode            bool
	inSpawn              int
	inParallel           int
	inScope              int
	callGraph            map[*ast.FunctionStatement]map[*ast.FunctionStatement]bool
	spawnedFunctions     map[*ast.FunctionStatement]bool
	mutatedGlobalsInFunc map[*ast.FunctionStatement]map[*Symbol]bool
	TypeCollectionOnly   bool
}

func (sa *SemanticAnalyzer) invalidateBounds(sym *Symbol) {
	if sym == nil {
		return
	}
	scope := sa.CurrentScope
	for scope != nil {
		if _, exists := scope.Bounds[sym]; exists {
			delete(scope.Bounds, sym)
		}
		scope = scope.Parent
	}
}

func (sa *SemanticAnalyzer) getRootSymbol(node ast.Expression) *Symbol {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.Identifier:
		return sa.SemanticInfo.Uses[n]
	case *ast.SelectorExpression:
		return sa.getRootSymbol(n.Left)
	case *ast.IndexExpression:
		return sa.getRootSymbol(n.Left)
	case *ast.PrefixExpression:
		return sa.getRootSymbol(n.Right)
	}
	return nil
}

func (sa *SemanticAnalyzer) validateConcurrentGlobalMutations() {
	concurrentFuncs := make(map[*ast.FunctionStatement]bool)
	var queue []*ast.FunctionStatement
	for f := range sa.spawnedFunctions {
		concurrentFuncs[f] = true
		queue = append(queue, f)
	}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for neighbor := range sa.callGraph[curr] {
			if !concurrentFuncs[neighbor] {
				concurrentFuncs[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	for f := range concurrentFuncs {
		for gSym := range sa.mutatedGlobalsInFunc[f] {
			sa.AddError(f.Pos(), "cannot mutate global variable '%s' inside concurrent function '%s'", gSym.Name, f.Name.Value)
		}
	}
}

func NewAnalyzer() *SemanticAnalyzer {
	global := NewScope(nil, ScopeGlobal)

	global.Define("i32", types.I32, SymType, nil)
	global.Define("int", types.Int, SymType, nil)
	global.Define("str", types.String, SymType, nil)
	global.Define("bool", types.Bool, SymType, nil)
	global.Define("void", types.Void, SymType, nil)
	global.Define("ptr", types.Ptr, SymType, nil)
	global.Define("fiber", types.Fiber, SymType, nil)
	global.Define("f64", types.F64, SymType, nil)
	global.Define("f32", types.F32, SymType, nil)
	global.Define("i64", types.I64, SymType, nil)
	global.Define("u64", types.U64, SymType, nil)
	global.Define("u32", types.U32, SymType, nil)
	global.Define("u16", types.U16, SymType, nil)
	global.Define("u8", types.U8, SymType, nil)
	global.Define("i16", types.I16, SymType, nil)
	global.Define("i8", types.I8, SymType, nil)
	global.Define("byte", types.Byte, SymType, nil)

	parkType := &types.FunctionType{
		Params:      []types.NRType{},
		ParamLeases: []types.LeaseKind{},
		Return:      types.Void,
	}
	global.Define("park", parkType, SymFunc, nil)

	resumeType := &types.FunctionType{
		Params:      []types.NRType{types.Fiber},
		ParamLeases: []types.LeaseKind{types.LeaseRead},
		Return:      types.Void,
	}
	global.Define("resume", resumeType, SymFunc, nil)
	global.Define("make", types.Void, SymFunc, nil)   // Poly-builtin
	global.Define("append", types.Void, SymFunc, nil) // Poly-builtin

	// --- Built-in Error Handling: Result[T, E] ---
	resultType := &types.SumType{
		TypeName:      "Result",
		CoreIntrinsic: "Result",
		TypeParams:    []*types.TypeParam{{Name: "T"}, {Name: "E"}},
		Variants:      make(map[string]*types.Variant),
	}
	resultType.Variants["Ok"] = &types.Variant{
		Name: "Ok",
		Tag:  0,
		Fields: map[string]types.NRType{
			"Value": &types.GenericType{TypeParam: "T"},
		},
		FieldNames: []string{"Value"},
	}
	resultType.Variants["Err"] = &types.Variant{
		Name: "Err",
		Tag:  1,
		Fields: map[string]types.NRType{
			"Error": &types.GenericType{TypeParam: "E"},
		},
		FieldNames: []string{"Error"},
	}
	global.Define("Result", resultType, SymType, nil)

	// Define Ok and Err as variant constructors in global scope
	okCtor := &types.FunctionType{
		Params:      []types.NRType{&types.GenericType{TypeParam: "T"}},
		ParamLeases: []types.LeaseKind{types.LeaseMove},
		Return:      resultType,
	}
	global.Define("Ok", okCtor, SymVariant, nil)

	errCtor := &types.FunctionType{
		Params:      []types.NRType{&types.GenericType{TypeParam: "E"}},
		ParamLeases: []types.LeaseKind{types.LeaseMove},
		Return:      resultType,
	}
	global.Define("Err", errCtor, SymVariant, nil)

	// --- Built-in Option: Option[T] ---
	optionType := &types.SumType{
		TypeName:      "Option",
		CoreIntrinsic: "Option",
		TypeParams:    []*types.TypeParam{{Name: "T"}},
		Variants:      make(map[string]*types.Variant),
	}
	optionType.Variants["Some"] = &types.Variant{
		Name: "Some",
		Tag:  0,
		Fields: map[string]types.NRType{
			"val": &types.GenericType{TypeParam: "T"},
		},
		FieldNames: []string{"val"},
	}
	optionType.Variants["None"] = &types.Variant{
		Name:   "None",
		Tag:    1,
		Fields: make(map[string]types.NRType),
	}
	global.Define("Option", optionType, SymType, nil)

	someCtor := &types.FunctionType{
		Params:      []types.NRType{&types.GenericType{TypeParam: "T"}},
		ParamLeases: []types.LeaseKind{types.LeaseMove},
		Return:      optionType,
	}
	global.Define("Some", someCtor, SymVariant, nil)
	global.Define("None", optionType, SymVariant, nil)

	return &SemanticAnalyzer{
		GlobalScope:          global,
		CurrentScope:         global,
		Diagnostics:          &diag.Collection{},
		InitStates:           make(map[*Symbol]InitState),
		PackageScopes:        make(map[string]*Scope),
		FuncScopes:           make(map[*ast.FunctionStatement]*Scope),
		LoadedDirs:           make(map[string]bool),
		ProcessedFiles:       make(map[string]bool),
		AnalyzedFiles:        make(map[string]bool),
		analyzingTypes:       make(map[*Symbol]bool),
		callGraph:            make(map[*ast.FunctionStatement]map[*ast.FunctionStatement]bool),
		spawnedFunctions:     make(map[*ast.FunctionStatement]bool),
		mutatedGlobalsInFunc: make(map[*ast.FunctionStatement]map[*Symbol]bool),

		SemanticInfo: SemanticInfo{
			Defs:               make(map[*ast.Identifier]*Symbol),
			Types:              make(map[ast.Node]types.NRType),
			Scopes:             make(map[ast.Node]*Scope),
			Uses:               make(map[*ast.Identifier]*Symbol),
			DeadSyms:           make(map[*Symbol]ast.Node),
			Instances:          make(map[*ast.FunctionStatement]map[string]*ast.FunctionStatement),
			MonomorphizedNames: make(map[*ast.CallExpression]string),
			SpecTypes:          make(map[string]types.NRType),
			FieldSymbols:       make(map[*types.StructType]map[string]*Symbol),
			MethodSymbols:      make(map[types.NRType]map[string]*Symbol),
		},
		DebugMode: false,
	}
}

func (sa *SemanticAnalyzer) debug(format string, args ...interface{}) {
	if sa.DebugMode {
		fmt.Fprintf(os.Stderr, "[Semantic] "+format+"\n", args...)
	}
}

func (sa *SemanticAnalyzer) AddError(pos token.Position, format string, args ...interface{}) {
	sa.ReportErrorWithHint(pos, fmt.Sprintf(format, args...), "")
}

func (sa *SemanticAnalyzer) ReportErrorWithHint(pos token.Position, message string, hint string) {
	sa.ReportError(pos, message, nil, hint)
}

func (sa *SemanticAnalyzer) ReportErrorWithNotes(pos token.Position, message string, notes []string) {
	sa.ReportError(pos, message, notes, "")
}

func (sa *SemanticAnalyzer) ReportError(pos token.Position, message string, notes []string, hint string) {
	if len(sa.Diagnostics.Diagnostics) > 100 {
		return // Prevent diagnostic storm
	}
	sa.Diagnostics.Add(diag.Diagnostic{
		Range: diag.Range{
			Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
			End:   diag.Position{Line: pos.Line, Column: pos.Column + 1, Offset: pos.Offset + 1},
		},
		Severity: diag.Error,
		Message:  message,
		Source:   "Semantic",
		File:     pos.Filename,
		Notes:    notes,
		Hint:     hint,
	})
}

func (sa *SemanticAnalyzer) GetPackageName(file *ast.File) string {
	if file == nil {
		return "main"
	}
	for _, stmt := range file.Statements {
		if pkg, ok := stmt.(*ast.PackageStatement); ok && pkg != nil && pkg.Name != nil {
			return pkg.Name.Value
		}
	}
	// fmt.Printf("[DEBUG] GetPackageName: no package stmt in %s, returning main\n", file.Name)
	return "main"
}

func (sa *SemanticAnalyzer) GetPackageScope(packageName string) *Scope {
	if scope, ok := sa.PackageScopes[packageName]; ok {
		return scope
	}
	scope := NewScope(sa.GlobalScope, ScopePackage)
	scope.PackageName = packageName
	sa.PackageScopes[packageName] = scope
	return scope
}

// -------------------------------------------------------------------------
// PASS 1: COLLECT SYMBOLS (The "Harvest")
// -------------------------------------------------------------------------
// This pass only looks at top-level definitions (Functions, Structs).
// It registers their names so that Pass 2 can see them regardless of file order.
func (sa *SemanticAnalyzer) CollectSymbols(node ast.Node) {
	if sa.Context != nil && sa.Context.Err() != nil {
		return
	}
	if ast.IsNil(node) {
		return
	}

	sa.depth++
	defer func() { sa.depth-- }()
	if sa.depth > 500 {
		panic("infinite recursion detected in CollectSymbols")
	}

	switch n := node.(type) {
	case *ast.Program:
		// Step 1: Collect ONLY type definitions from all files.
		// Use a manual loop because n.Files can grow as we discover sibling files.
		sa.CollectingSymbols = true
		sa.TypeCollectionOnly = true
		for i := 0; i < len(n.Files); i++ {
			if sa.Context != nil && sa.Context.Err() != nil {
				return
			}
			sa.CollectSymbols(n.Files[i])
		}

		// Step 2: Collect all imports, functions, etc. from all files.
		sa.TypeCollectionOnly = false
		// Reset ProcessedFiles so files can be scanned again for imports/functions
		sa.ProcessedFiles = make(map[string]bool)
		for i := 0; i < len(n.Files); i++ {
			sa.CollectSymbols(n.Files[i])
		}
		sa.CollectingSymbols = false

		// --- HANDLING IMPORTS ---
	case *ast.ImportStatement:
		pkgPath := n.PathValue()
		if sa.Loader == nil {
			sa.AddError(n.Path.Pos(), "could not load package %q: no package loader configured", pkgPath)
			return
		}
		pkgScope, err := sa.Loader.Load(pkgPath)
		if err != nil {
			sa.AddError(n.Path.Pos(), "could not load package %q: %v", pkgPath, err)
			return
		}

		var name string
		var nameNode *ast.Identifier

		if n.Alias != nil {
			name = n.Alias.Value
			nameNode = n.Alias
		} else {
			parts := strings.Split(pkgPath, "/")
			name = parts[len(parts)-1]

			// Create a virtual identifier so we have a key for Defs
			nameNode = &ast.Identifier{
				Token: n.Path.Token,
				Value: name,
			}
		}

		modType := &ModuleType{Ident: name, Exports: pkgScope}

		// --- FIX: Change SymModule to SymPackage to match your test ---
		sym, err := sa.CurrentScope.Define(name, modType, SymPackage, n)
		if err != nil {
			if existing, found := sa.CurrentScope.Resolve(name); found && existing.Kind == SymPackage {
				var existingPath string
				if existingImport, ok := existing.DefNode.(*ast.ImportStatement); ok {
					existingPath = existingImport.PathValue()
				}
				if existingPath == pkgPath {
					sa.SemanticInfo.Defs[nameNode] = existing
					// Create a pathSym pointing to the actual package file,
					// not the existing symbol (whose DefNode is the sibling's ImportStatement)
					pathSym := &Symbol{
						Name:    name,
						Type:    modType,
						Kind:    SymPackage,
						DefNode: n, // Default: points to this import statement
					}
					for _, s := range pkgScope.Symbols {
						if s.DefNode != nil && s.DefNode.Pos().Filename != "" {
							pathSym.DefNode = &ast.Identifier{
								Token: token.Token{
									Position: token.Position{
										Filename: s.DefNode.Pos().Filename,
										Line:     1,
										Column:   1,
									},
									Literal: "",
								},
								Value: "",
							}
							break
						}
					}
					sa.SemanticInfo.Uses[n.Path] = pathSym
					return
				}
				sa.ReportErrorWithHint(n.Token.Position,
					fmt.Sprintf("import name conflict: package %q conflicts with already imported package %q", pkgPath, existingPath),
					fmt.Sprintf("Provide a custom alias, e.g.: import custom_%s %q", name, pkgPath))
				return
			}
			sa.AddError(n.Token.Position, "%s", err.Error())
			return
		}

		sa.SemanticInfo.Defs[nameNode] = sym

		// Create a symbol for the path literal that points to the package file if possible
		pathSym := &Symbol{
			Name:    name,
			Type:    modType,
			Kind:    SymPackage,
			DefNode: n, // Default: points to import statement
		}
		for _, s := range pkgScope.Symbols {
			if s.DefNode != nil && s.DefNode.Pos().Filename != "" {
				// Point to the first file found in the package
				pathSym.DefNode = &ast.Identifier{
					Token: token.Token{
						Position: token.Position{
							Filename: s.DefNode.Pos().Filename,
							Line:     1,
							Column:   1,
						},
						Literal: "",
					},
					Value: "",
				}
				break
			}
		}
		sa.SemanticInfo.Uses[n.Path] = pathSym

	case *ast.File:
		if n.Name == "" {
			return
		}
		normName := normalizePath(n.Name)
		if sa.ProcessedFiles[normName] {
			return
		}
		sa.ProcessedFiles[normName] = true

		packageName := sa.GetPackageName(n)

		prevScope := sa.CurrentScope
		sa.CurrentScope = sa.GetPackageScope(packageName)

		// Record scope for the package statement if it exists
		for _, stmt := range n.Statements {
			if pkg, ok := stmt.(*ast.PackageStatement); ok {
				sa.SemanticInfo.Scopes[pkg] = sa.CurrentScope
			}
		}

		// Pass 1a: Collect all Types AND Imports first (so methods can resolve receivers)
		for _, stmt := range n.Statements {
			switch stmt.(type) {
			case *ast.TypeStatement, *ast.ImportStatement:
				sa.CollectSymbols(stmt)
			}
		}

		if sa.TypeCollectionOnly {
			// Auto-load other files in the same directory for multi-file packages to discover them
			if sa.Loader != nil && n.Name != "" {
				dir := filepath.Dir(n.Name)
				if !filepath.IsAbs(dir) && !strings.HasPrefix(dir, ".") {
					dir = "./" + dir
				}
				normDir := normalizePath(dir)
				if !sa.LoadedDirs[normDir] {
					sa.LoadedDirs[normDir] = true
					sa.Loader.Load(dir)
				}
			}
			sa.CurrentScope = prevScope
			return
		}

		// Pass 1b: Collect Functions and other top-level symbols
		for _, stmt := range n.Statements {
			switch stmt.(type) {
			case *ast.TypeStatement, *ast.ImportStatement:
				// Skip, already collected in Pass 1a
			default:
				sa.CollectSymbols(stmt)
			}
		}

		// Auto-load other files in the same directory for multi-file packages
		// (Must be done AFTER collecting this file's symbols so sibling files can resolve this file's types!)
		if sa.Loader != nil && n.Name != "" {
			dir := filepath.Dir(n.Name)
			if !filepath.IsAbs(dir) && !strings.HasPrefix(dir, ".") {
				dir = "./" + dir
			}
			normDir := normalizePath(dir)
			if !sa.LoadedDirs[normDir] {
				sa.LoadedDirs[normDir] = true
				sa.Loader.Load(dir)
			}
		}
		sa.CurrentScope = prevScope

	// CASE: WRAPPER (Unwrap ExpressionStatement to see if it holds a function)
	case *ast.ExpressionStatement:
		sa.CollectSymbols(n.Expression)

		// --- PASS 1: COLLECT SYMBOLS ---
		// In CollectSymbols method...
	case *ast.FunctionStatement:
		if sa.FuncScopes == nil {
			sa.FuncScopes = make(map[*ast.FunctionStatement]*Scope)
		}
		sa.FuncScopes[n] = sa.CurrentScope

		// 1. Temporarily put TypeParameters in scope so receivers and return types can resolve them
		prevScope := sa.CurrentScope
		// hasImplicitParams := false
		if len(n.TypeParameters) > 0 {
			sa.CurrentScope = NewScope(prevScope, ScopeFunction)
			// Pass 1: Define all TypeParams in the new scope
			for _, tp := range n.TypeParameters {
				sa.CurrentScope.Define(tp.Name.Value, &types.GenericType{TypeParam: tp.Name.Value, Constraint: nil}, SymType, tp)
			}
			// Pass 2: Resolve constraints using the fully populated scope
			for _, tp := range n.TypeParameters {
				if tp.Constraint != nil {
					constraint := sa.resolveTypeNode(tp.Constraint)
					if constraint == types.Void {
						if sa.DebugMode {
							println("[DEBUG CollectSymbols] Function:", n.Name.Value, "tp:", tp.Name.Value, "resolved to void!")
						}
					}
					if sym, found := sa.CurrentScope.Lookup(tp.Name.Value); found {
						if gt, ok := sym.Type.(*types.GenericType); ok {
							gt.Constraint = constraint
						}
					}
				}
			}
		}

		// Extract implicit generic type parameters from the receiver
		if n.Receiver != nil {
			var baseExp ast.Expression = n.Receiver.Type
			for {
				if pref, ok := baseExp.(*ast.PrefixExpression); ok {
					baseExp = pref.Right
				} else {
					break
				}
			}
			if idxExpr, ok := baseExp.(*ast.IndexExpression); ok {
				for _, idx := range idxExpr.Indices {
					if ident, ok := idx.(*ast.Identifier); ok {
						if _, exists := sa.CurrentScope.Resolve(ident.Value); !exists {
							if sa.CurrentScope == prevScope {
								sa.CurrentScope = NewScope(prevScope, ScopeFunction)
							}
							gt := &types.GenericType{TypeParam: ident.Value, Constraint: types.Any}
							sa.CurrentScope.Define(ident.Value, gt, SymType, ident)
							n.IsGenericTemplate = true
						}
					}
				}
			}
		}

		// 2. Resolve Receiver Type (if any)
		var receiverType types.NRType
		if n.Receiver != nil {
			receiverType = sa.resolveTypeNode(n.Receiver.Type)
			delete(sa.SemanticInfo.Types, n.Receiver.Type) // Keep the cache clean for Pass 2 recursive resolution!
		}

		if n.Name != nil && n.Name.Value != "" {
			// 3. Resolve Function Type
			fnType := sa.resolveFunctionType(n)

			// Phase 7: Skip body collection for standalone generic functions (monomorphized later)
			// But for methods, we still want to attach them to the struct!
			if len(n.TypeParameters) > 0 && n.Receiver == nil {
				sa.CurrentScope = prevScope
				sym, err := sa.CurrentScope.Define(n.Name.Value, types.Void, SymFunc, n)
				if err != nil {
					if existing, ok := sa.CurrentScope.Symbols[n.Name.Value]; ok && existing.DefNode == n {
						sym = existing
					}
				}
				if sym != nil {
					if n.IsExport || n.IsPublic {
						sym.Visible = Public
					}
					sa.SemanticInfo.Defs[n.Name] = sym
				}
				return
			}

			if receiverType != nil && receiverType != types.ErrorType {
				// Method Definition
				baseType := receiverType
				if pt, ok := receiverType.(*types.PointerType); ok {
					baseType = pt.Base
				}

				// Methods are defined globally with mangled name for C
				rName := receiverType.Name()
				rName = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(rName, "@"), "#"), "&")
				mangledName := rName + "_" + n.Name.Value
				sym, _ := sa.CurrentScope.Define(mangledName, fnType, SymFunc, n)
				sa.SemanticInfo.Defs[n.Name] = sym

				if st, ok := baseType.(*types.StructType); ok {
					if n.Name.Value == "Next" || n.Name.Value == "Take" {
						if sa.DebugMode {
							fmt.Printf("DEBUG Adding method %s to struct %s (%p)\n", n.Name.Value, st.TypeName, st)
						}
					}
					st.Methods[n.Name.Value] = fnType

					if sa.SemanticInfo.MethodSymbols[st] == nil {
						sa.SemanticInfo.MethodSymbols[st] = make(map[string]*Symbol)
					}
					sa.SemanticInfo.MethodSymbols[st][n.Name.Value] = sym
				} else if sumT, ok := baseType.(*types.SumType); ok {
					if sumT.Methods == nil {
						sumT.Methods = make(map[string]types.NRType)
					}
					sumT.Methods[n.Name.Value] = fnType

					if sa.SemanticInfo.MethodSymbols[sumT] == nil {
						sa.SemanticInfo.MethodSymbols[sumT] = make(map[string]*Symbol)
					}
					sa.SemanticInfo.MethodSymbols[sumT][n.Name.Value] = sym
				} else if primT, ok := baseType.(*types.PrimitiveType); ok {
					if primT.Methods == nil {
						primT.Methods = make(map[string]types.NRType)
					}
					primT.Methods[n.Name.Value] = fnType

					if sa.SemanticInfo.MethodSymbols[primT] == nil {
						sa.SemanticInfo.MethodSymbols[primT] = make(map[string]*Symbol)
					}
					sa.SemanticInfo.MethodSymbols[primT][n.Name.Value] = sym
				} else {
					sa.AddError(n.Receiver.Pos(), "cannot define methods on non-struct/non-sum type %s", receiverType.Name())
				}
			} else {
				// Standard Global Function
				sym, err := sa.CurrentScope.Define(n.Name.Value, fnType, SymFunc, n)
				if err != nil {
					if existing, ok := sa.CurrentScope.Symbols[n.Name.Value]; ok {
						if existing.DefNode == n {
							sym = existing
						} else {
							sa.AddError(n.Token.Position, "%s", err.Error())
							sa.CurrentScope = prevScope
							return
						}
					} else {
						sa.AddError(n.Token.Position, "%s", err.Error())
						sa.CurrentScope = prevScope
						return
					}
				}
				if n.IsExport || n.IsPublic {
					sym.Visible = Public
				}
				sa.SemanticInfo.Defs[n.Name] = sym
			}
		}
		if n.Name != nil && n.Name.Value != "" {
			if sym := sa.SemanticInfo.Defs[n.Name]; sym != nil {
				if ast.GetAttribute(n.Attributes, "inline") != nil {
					sym.IsInline = true
				}
			}
		}
		sa.CurrentScope = prevScope

	case *ast.TypeStatement:
		// Don't re-collect if we've already done so (Pass 1b protection)
		if !sa.TypeCollectionOnly && n.Name != nil && sa.SemanticInfo.Defs[n.Name] != nil {
			return
		}

		// 1. Determine if it's a struct or interface
		if _, ok := n.Value.(*ast.StructLiteral); ok {
			if n.Name.Value == "" {
				return
			}
			structType := types.NewStructType(n.Name.Value)
			if ast.GetAttribute(n.Attributes, "shared") != nil {
				structType.IsShared = true
			}
			if attr := ast.GetAttribute(n.Attributes, "core_intrinsic"); attr != nil && len(attr.Args) > 0 {
				structType.CoreIntrinsic = attr.Args[0]
			}
			for _, tp := range n.TypeParameters {
				structType.TypeParams = append(structType.TypeParams, &types.TypeParam{
					Name: tp.Name.Value,
				})
			}
			sym, err := sa.CurrentScope.Define(n.Name.Value, structType, SymType, n)

			if err != nil {
				if existing, ok := sa.CurrentScope.Symbols[n.Name.Value]; ok && existing.DefNode == n {
					sym = existing
				} else {
					sa.AddError(n.Name.Pos(), "%s", err.Error())
				}
			}
			if sym != nil {
				sa.SemanticInfo.Defs[n.Name] = sym
				if n.IsPublic {
					sym.Visible = Public
				}
			}
		} else if _, ok := n.Value.(*ast.InterfaceLiteral); ok {

			if n.Name.Value == "" {
				return
			}
			protocolType := &types.ProtocolType{
				ProtocolName: n.Name.Value,
				Methods:      make(map[string]*types.FunctionType),
			}
			if ast.GetAttribute(n.Attributes, "shared") != nil {
				protocolType.IsShared = true
			}
			for _, tp := range n.TypeParameters {
				protocolType.TypeParams = append(protocolType.TypeParams, &types.TypeParam{
					Name: tp.Name.Value,
				})
			}
			sym, err := sa.CurrentScope.Define(n.Name.Value, protocolType, SymType, n)

			if err != nil {
				if existing, ok := sa.CurrentScope.Symbols[n.Name.Value]; ok && existing.DefNode == n {
					sym = existing
				} else {
					sa.AddError(n.Name.Pos(), "%s", err.Error())
				}
			}
			if sym != nil {
				sa.SemanticInfo.Defs[n.Name] = sym
				if n.IsPublic {
					sym.Visible = Public
				}
			}
		} else if _, ok := n.Value.(*ast.SumTypeLiteral); ok {
			if n.Name.Value == "" {
				return
			}
			sumType := &types.SumType{
				TypeName: n.Name.Value,
			}
			if attr := ast.GetAttribute(n.Attributes, "core_intrinsic"); attr != nil && len(attr.Args) > 0 {
				sumType.CoreIntrinsic = attr.Args[0]
			}
			for _, tp := range n.TypeParameters {
				sumType.TypeParams = append(sumType.TypeParams, &types.TypeParam{
					Name: tp.Name.Value,
				})
			}
			sym, err := sa.CurrentScope.Define(n.Name.Value, sumType, SymType, n)

			if err != nil {
				if existing, ok := sa.CurrentScope.Symbols[n.Name.Value]; ok && existing.DefNode == n {
					sym = existing
				} else {
					sa.AddError(n.Name.Pos(), "%s", err.Error())
				}
			}
			if sym != nil {
				sa.SemanticInfo.Defs[n.Name] = sym
				if n.IsPublic {
					sym.Visible = Public
				}
			}
		}
	case *ast.VarStatement:
		if sa.CurrentScope.Kind == ScopePackage {
			sym, err := sa.CurrentScope.Define(n.Name.Value, types.Void, SymVar, n)
			if err == nil {
				sa.SemanticInfo.Defs[n.Name] = sym
				if n.IsPublic {
					sym.Visible = Public
				}
			}
		}
	}
}

// -------------------------------------------------------------------------
// PASS 2: ANALYZE (The "Deep Dive")

// -------------------------------------------------------------------------
// This pass enters bodies, checks types, and tracks variable usage.

func (sa *SemanticAnalyzer) analyzePattern(pattern ast.Expression, targetType types.NRType, isMove bool) {
	if ident, ok := pattern.(*ast.Identifier); ok {
		if ident.Value == "_" {
			return
		}
		sumType, isSum := targetType.(*types.SumType)
		isVariant := false
		if isSum {
			if _, exists := sumType.Variants[ident.Value]; exists {
				isVariant = true
			}
		}
		if !isVariant {
			resolvedType := targetType
			if !isMove && types.IsOwnedType(targetType) {
				resolvedType = &types.PointerType{Base: targetType, Leased: true, Kind: types.LeaseRead}
			}
			sym, _ := sa.CurrentScope.Define(ident.Value, resolvedType, SymVar, nil)
			if sym != nil {
				if pt, ok := resolvedType.(*types.PointerType); ok && pt.Leased {
					if !isMove {
						sym.LeaseKind = types.LeaseRead
					} else {
						sym.LeaseKind = pt.Kind
					}
				} else if isMove {
					sym.LeaseKind = types.LeaseMove
				} else {
					sym.LeaseKind = types.LeaseRead
				}
				sa.InitStates[sym] = Initialized
			}
			sa.SemanticInfo.Defs[ident] = sa.CurrentScope.Symbols[ident.Value]
		}
	} else if call, ok := pattern.(*ast.CallExpression); ok {
		sumType, isSum := targetType.(*types.SumType)
		if variantIdent, ok := call.Function.(*ast.Identifier); ok && isSum {
			variant, exists := sumType.Variants[variantIdent.Value]
			if !exists {
				sa.AddError(variantIdent.Pos(), "variant %s not found in %s", variantIdent.Value, sumType.Name())
			} else {
				if len(call.Arguments) != len(variant.Fields) {
					sa.AddError(call.Pos(), "variant %s expects %d fields, got %d", variant.Name, len(variant.Fields), len(call.Arguments))
				} else {
					fieldNames := sa.sortedFieldNames(variant.Fields)
					for j, arg := range call.Arguments {
						fieldType := variant.Fields[fieldNames[j]]
						sa.analyzePattern(arg.Value, fieldType, isMove)
					}
				}
			}
		} else if typeIdent, ok := call.Function.(*ast.Identifier); ok {
			_, isProto := types.UnwrapLease(targetType).(*types.ProtocolType)
			if isProto {
				// Protocol/Any downcast pattern matching
				concreteType := sa.tryResolveAsType(typeIdent)
				if concreteType == nil || concreteType == types.ErrorType {
					sa.AddError(typeIdent.Pos(), "unknown type %s for downcast pattern", typeIdent.Value)
				} else {
					if len(call.Arguments) != 1 {
						sa.AddError(call.Pos(), "downcast pattern expects exactly 1 argument to bind, got %d", len(call.Arguments))
					} else {
						// We pass the concreteType down. We preserve the lease kind of the target type.
						if pt, ok := targetType.(*types.PointerType); ok && pt.Leased {
							concreteType = &types.PointerType{Base: concreteType, Leased: true, Kind: pt.Kind}
						} else if !isMove {
							concreteType = &types.PointerType{Base: concreteType, Leased: true, Kind: types.LeaseRead}
						}
						sa.analyzePattern(call.Arguments[0].Value, concreteType, isMove)
					}
				}
			} else {
				typeName := "unknown"
				if targetType != nil {
					typeName = targetType.Name()
				}
				sa.AddError(call.Pos(), "invalid pattern for type %s", typeName)
			}
		}
	} else if structLit, ok := pattern.(*ast.StructLiteral); ok {
		ut := types.UnwrapLease(targetType)
		// If it's a pointer to a struct, unwrap the pointer to match against the struct fields
		if pt, ok := ut.(*types.PointerType); ok && !pt.IsArray {
			ut = pt.Base
		}
		structType, isStruct := ut.(*types.StructType)
		if !isStruct {
			typeName := "unknown"
			if targetType != nil {
				typeName = targetType.Name()
			}
			sa.AddError(structLit.Pos(), "expected struct pattern for type %s", typeName)
			return
		}
		if ident, ok := structLit.Name.(*ast.Identifier); ok {
			targetName := structType.TypeName
			if structType.BaseType != nil {
				targetName = structType.BaseType.TypeName
			}
			if ident.Value != targetName {
				sa.AddError(structLit.Name.Pos(), "expected struct %s, got %s", targetName, ident.Value)
			}
		}
		for _, field := range structLit.Fields {
			fieldType, ok := structType.Fields[field.Name.Value]
			if !ok {
				sa.AddError(field.Pos(), "field %s not found in struct %s", field.Name.Value, structType.Name())
				continue
			}
			if field.Value != nil {
				sa.analyzePattern(field.Value, fieldType, isMove)
			}
		}
	} else {
		sa.Analyze(pattern)
		patType := sa.SemanticInfo.Types[pattern]
		if patType != nil && targetType != nil && !types.IsAssignable(targetType, patType) {
			if _, ok := targetType.(*types.ProtocolType); ok {
				sa.checkInterfaceCompatibility(pattern, targetType)
			} else {
				sa.AddError(pattern.Pos(), "pattern type mismatch: expected %s, got %s", targetType.Name(), patType.Name())
			}
		}
	}
}

func (sa *SemanticAnalyzer) extractMatchedVariant(pattern ast.Expression, sumType *types.SumType) (string, bool) {
	if ident, ok := pattern.(*ast.Identifier); ok {
		if _, exists := sumType.Variants[ident.Value]; exists {
			return ident.Value, true
		}
	} else if call, ok := pattern.(*ast.CallExpression); ok {
		if variantIdent, ok := call.Function.(*ast.Identifier); ok {
			if _, exists := sumType.Variants[variantIdent.Value]; exists {
				return variantIdent.Value, true
			}
		}
	}
	return "", false
}

func (sa *SemanticAnalyzer) Analyze(node ast.Node) {
	if node == nil {
		return
	}

	if sa.Context != nil && sa.Context.Err() != nil {
		return
	}
	if ast.IsNil(node) {
		return
	}

	if sa.SemanticInfo.Types[node] != nil && sa.SemanticInfo.Types[node] != types.ErrorType {
		reanalyze := false
		switch node.(type) {
		case *ast.BlockStatement, *ast.IfExpression, *ast.MatchExpression,
			*ast.ReturnStatement, *ast.VarStatement, *ast.AssignmentStatement,
			*ast.ExpressionStatement, *ast.ForStatement, *ast.WhileStatement:
			reanalyze = true
		}
		if !reanalyze {
			return
		}
	}

	sa.depth++
	defer func() { sa.depth-- }()
	if sa.depth > 1000 {
		panic("infinite recursion detected in Analyze")
	}

	switch n := node.(type) {

	// --- ENTRY POINT ---
	case *ast.Program:

		// 1. Setup Package Scope (Shared by all files in the entry package)
		// Note: sa.CurrentScope is currently GlobalScope

		// 2. Run Pass 1 (Harvest declarations from ALL files)
		sa.CollectSymbols(n)

		// Pass 1.5: Analyze all TypeStatements across ALL files first,
		// ensuring all struct/enum definitions are fully populated before body analysis.
		for _, f := range n.Files {
			sa.AnalyzeFileTypes(f)
		}

		// 3. Run Pass 2 (Analyze bodies)
		for _, f := range n.Files {
			sa.Analyze(f)
		}

		sa.validateConcurrentGlobalMutations()

		// 4. Run Unused Symbols Check (Imports, Types, Functions)
		sa.checkUnusedProgramSymbols(n)
	// --- 1. HANDLE FILES (Iterate statements) ---
	case *ast.File:
		if n.Name == "" {
			return
		}
		normName := normalizePath(n.Name)
		if sa.AnalyzedFiles[normName] {
			return
		}
		sa.AnalyzedFiles[normName] = true

		packageName := sa.GetPackageName(n)
		prevScope := sa.CurrentScope
		sa.CurrentScope = sa.GetPackageScope(packageName)

		if packageName == "main" {
			var symNames []string
			for k := range sa.CurrentScope.Symbols {
				symNames = append(symNames, k)
			}
			// fmt.Fprintf(os.Stderr, "[DEBUG Analyze File] Package main symbols: %v\n", symNames)
		}

		for _, stmt := range n.Statements {
			sa.Analyze(stmt)
		}
		sa.CurrentScope = prevScope
	// --- 2. HANDLE WRAPPERS (Unwrap ExpressionStatement) ---
	// Top-level functions (like main) and expressions (like "unknown()")
	// are wrapped in this. We must drill down.
	case *ast.ExpressionStatement:
		sa.Analyze(n.Expression)
		sa.SemanticInfo.Types[n] = sa.SemanticInfo.Types[n.Expression]

	// --- 3. HANDLE LITERALS (Essential for Type Inference) ---
	// If we don't handle these, 'let a = 20' sees a nil type and fails inference.
	case *ast.FloatLiteral:
		if n.Suffix != "" {
			t, ok := types.LookupPrimitive(n.Suffix)
			if ok {
				sa.SemanticInfo.Types[n] = t
				return
			}
		}
		sa.SemanticInfo.Types[n] = types.F64

		// --- OPERATORS ---
	case *ast.GroupedExpression:
		sa.Analyze(n.Expression)
		sa.SemanticInfo.Types[n] = sa.SemanticInfo.Types[n.Expression]

	case *ast.RangeExpression:
		sa.Analyze(n.Start)
		sa.Analyze(n.End)

		startType := sa.SemanticInfo.Types[n.Start]
		endType := sa.SemanticInfo.Types[n.End]

		if startType == nil || endType == nil {
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		if !types.Equals(startType, endType) {
			sa.AddError(n.Token.Position, "range bounds must have exactly the same type, got %s and %s", startType.Name(), endType.Name())
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		if pt, ok := startType.(*types.PrimitiveType); !ok || (!strings.HasPrefix(pt.Name(), "i") && !strings.HasPrefix(pt.Name(), "u")) {
			sa.AddError(n.Token.Position, "range bounds must be integers, got %s", startType.Name())
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		// Instantiate Range
		rangeSym, _ := sa.CurrentScope.Lookup("Range")
		if rangeSym == nil || rangeSym.Type == nil {
			if mainScope := sa.GetPackageScope("main"); mainScope != nil {
				rangeSym, _ = mainScope.Lookup("Range")
			}
		}
		if rangeSym == nil || rangeSym.Type == nil {
			sa.AddError(n.Token.Position, "compiler error: Range type not found in scope (required for range syntax)")
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		if baseStruct, ok := rangeSym.Type.(*types.StructType); ok {
			sa.SemanticInfo.Types[n] = baseStruct
		} else {
			sa.SemanticInfo.Types[n] = types.ErrorType
		}

	case *ast.InfixExpression:
		// 1. Analyze Operands
		sa.Analyze(n.Left)
		sa.Analyze(n.Right)

		leftType := sa.SemanticInfo.Types[n.Left]
		rightType := sa.SemanticInfo.Types[n.Right]

		// Safety check
		if leftType == nil || rightType == nil {
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		// Unwrap leases for comparison/arithmetic
		// EXCEPT when comparing against 'none' (types.Ptr), as leases are pointers.
		leftBase := leftType
		if pt, ok := leftType.(*types.PointerType); ok && pt.Leased && !pt.IsArray && rightType != types.Ptr {
			leftBase = pt.Base
		}
		rightBase := rightType
		if pt, ok := rightType.(*types.PointerType); ok && pt.Leased && !pt.IsArray && leftType != types.Ptr {
			rightBase = pt.Base
		}

		// Contextual promotion for unsuffixed IntegerLiteral in infix expressions
		if lit, ok := n.Right.(*ast.IntegerLiteral); ok && lit.Suffix == "" {
			if leftBase != nil && leftBase.GetKind() == types.KindPrimitive && leftBase.Name() != "i32" && types.IsAssignable(leftBase, types.I32) {
				sa.SemanticInfo.Types[lit] = leftBase
				rightBase = leftBase
				rightType = leftBase
			}
		} else if lit, ok := n.Left.(*ast.IntegerLiteral); ok && lit.Suffix == "" {
			if rightBase != nil && rightBase.GetKind() == types.KindPrimitive && rightBase.Name() != "i32" && types.IsAssignable(rightBase, types.I32) {
				sa.SemanticInfo.Types[lit] = rightBase
				leftBase = rightBase
				leftType = rightBase
			}
		}

		// 2. Check Compatibility based on Operator
		switch n.Operator {

		// Arithmetic & Bitwise: Result type == Operand type (e.g. i32 + i32 = i32)
		case "+", "-", "*", "/", "%", "&", "|", "^", "<<", ">>":
			// 1. Check for struct operator overloads
			if st, ok := leftBase.(*types.StructType); ok {
				methodName := ""
				switch n.Operator {
				case "+":
					methodName = "add"
				case "-":
					methodName = "sub"
				case "*":
					methodName = "mul"
				case "/":
					methodName = "div"
				case "%":
					methodName = "mod"
				case "&":
					methodName = "bitand"
				case "|":
					methodName = "bitor"
				case "^":
					methodName = "bitxor"
				case "<<":
					methodName = "shl"
				case ">>":
					methodName = "shr"
				}
				if methodType, exists := st.Methods[methodName]; exists {
					if ft, ok := methodType.(*types.FunctionType); ok && len(ft.Params) == 1 {
						expectedBase := ft.Params[0]
						if pt, ok := expectedBase.(*types.PointerType); ok && pt.Leased {
							expectedBase = pt.Base
						}
						if types.IsAssignable(rightType, expectedBase) || types.IsAssignable(rightBase, expectedBase) {
							sa.SemanticInfo.Types[n] = ft.Return
						} else {
							sa.AddError(n.Token.Position, "operator method '%s' expects %s, got %s", methodName, ft.Params[0].Name(), rightType.Name())
							sa.SemanticInfo.Types[n] = types.ErrorType
						}
						return
					} else {
						sa.AddError(n.Token.Position, "invalid signature for operator method '%s'", methodName)
						sa.SemanticInfo.Types[n] = types.ErrorType
						return
					}
				}
			}

			// 2. Fallback to primitive arithmetic
			if !types.Equals(leftBase, rightBase) {
				// Special Case: String Concatenation with Booleans
				if n.Operator == "+" && (leftBase == types.String || rightBase == types.String) {
					if (leftBase == types.String && rightBase == types.Bool) ||
						(leftBase == types.Bool && rightBase == types.String) {
						sa.SemanticInfo.Types[n] = types.String
						return
					}
				}
				// Implicit promotion for i8 and i32 arithmetic
				if (leftBase == types.I32 && rightBase == types.I8) || (leftBase == types.I8 && rightBase == types.I32) {
					sa.SemanticInfo.Types[n] = types.I32
					return
				}

				sa.AddError(n.Token.Position, "type mismatch: %s %s %s", leftBase.Name(), n.Operator, rightBase.Name())
				sa.SemanticInfo.Types[n] = types.ErrorType
			} else {
				if leftBase.GetKind() != types.KindPrimitive && leftBase.GetKind() != types.KindPointer && leftBase.GetKind() != types.KindGeneric {
					sa.AddError(n.Token.Position, "operator '%s' not supported for type %s without operator method", n.Operator, leftBase.Name())
					sa.SemanticInfo.Types[n] = types.ErrorType
				} else {
					sa.SemanticInfo.Types[n] = leftBase
				}
			}

		// Logical: Result type == Boolean
		case "&&", "||":
			if leftBase != types.Bool || rightBase != types.Bool {
				sa.AddError(n.Token.Position, "logical operators expect booleans, got %s and %s", leftBase.Name(), rightBase.Name())
			}
			sa.SemanticInfo.Types[n] = types.Bool

		// Comparison: Result type == Boolean
		case "<", ">", "<=", ">=":
			if st, ok := leftBase.(*types.StructType); ok {
				if methodType, exists := st.Methods["cmp"]; exists {
					if ft, ok := methodType.(*types.FunctionType); ok && len(ft.Params) == 1 && ft.Return == types.I32 {
						expectedBase := ft.Params[0]
						if pt, ok := expectedBase.(*types.PointerType); ok && pt.Leased {
							expectedBase = pt.Base
						}
						if types.IsAssignable(rightType, expectedBase) || types.IsAssignable(rightBase, expectedBase) {
							sa.SemanticInfo.Types[n] = types.Bool
							return
						} else {
							sa.AddError(n.Token.Position, "method 'cmp' expects %s, got %s", ft.Params[0].Name(), rightType.Name())
							sa.SemanticInfo.Types[n] = types.ErrorType
							return
						}
					} else {
						sa.AddError(n.Token.Position, "invalid signature for 'cmp' method")
						sa.SemanticInfo.Types[n] = types.ErrorType
						return
					}
				}
			}

			if !types.IsAssignable(leftBase, rightBase) && !types.IsAssignable(leftType, rightType) {
				sa.AddError(n.Token.Position, "cannot compare mismatching types: %s %s %s", leftBase.Name(), n.Operator, rightBase.Name())
			} else {
				if leftBase.GetKind() != types.KindPrimitive && leftBase.GetKind() != types.KindPointer && leftBase.GetKind() != types.KindGeneric {
					sa.AddError(n.Token.Position, "operator '%s' not supported for type %s without 'cmp' method", n.Operator, leftBase.Name())
				}
			}
			sa.SemanticInfo.Types[n] = types.Bool

		case "==", "!=":
			if st, ok := leftBase.(*types.StructType); ok {
				if methodType, exists := st.Methods["eq"]; exists {
					if ft, ok := methodType.(*types.FunctionType); ok && len(ft.Params) == 1 && ft.Return == types.Bool {
						expectedBase := ft.Params[0]
						if pt, ok := expectedBase.(*types.PointerType); ok && pt.Leased {
							expectedBase = pt.Base
						}
						if !types.IsAssignable(rightType, expectedBase) && !types.IsAssignable(rightBase, expectedBase) {
							sa.AddError(n.Token.Position, "method 'eq' expects %s, got %s", ft.Params[0].Name(), rightType.Name())
						}
					} else {
						sa.AddError(n.Token.Position, "invalid signature for 'eq' method (must return bool and take 1 argument)")
					}
				}
			}
			if !types.IsAssignable(leftBase, rightBase) && !types.IsAssignable(leftType, rightType) {
				sa.AddError(n.Token.Position, "cannot compare mismatching types: %s %s %s", leftBase.Name(), n.Operator, rightBase.Name())
			}
			sa.SemanticInfo.Types[n] = types.Bool

		default:
			sa.AddError(n.Token.Position, "unknown operator: %s", n.Operator)
			sa.SemanticInfo.Types[n] = types.ErrorType
		}

	case *ast.StructLiteral:
		// 1. Resolve the Struct Type (e.g., "User" or "Box[i32]")
		if n.Name != nil {
			if typeNode, ok := n.Name.(ast.TypeNode); ok {
				resolvedType := sa.resolveTypeNode(typeNode)
				// Nora: Auto-specialize generic struct names in monomorphized context
				if st, ok := resolvedType.(*types.StructType); ok && len(st.TypeParams) > 0 {
					if sa.CurrentFunction != nil && !sa.CurrentFunction.IsGenericTemplate {
						args := []types.NRType{}
						for _, tp := range st.TypeParams {
							if tSym, exists := sa.CurrentScope.Resolve(tp.Name); exists && tSym.Kind == SymType {
								args = append(args, tSym.Type)
							}
						}
						if len(args) == len(st.TypeParams) {
							resolvedType = sa.specializeStructType(st, args, n.Pos())
						}
					}
				}

				if resolvedType != types.ErrorType {
					sa.SemanticInfo.Types[n] = resolvedType
				} else {
					sa.AddError(n.Name.Pos(), "invalid or undefined struct type")
				}
			} else {
				sa.AddError(n.Name.Pos(), "expected type name for struct literal")
			}
		}

		// 2. Analyze Fields
		for _, field := range n.Fields {
			if field.Value != nil {
				sa.Analyze(field.Value)
			} else if expr, ok := field.Type.(ast.Expression); ok {
				// Fallback for some legacy parser cases where value might be in Type
				sa.Analyze(expr)
			}
		}

		// 3. Enforce all fields are initialized (Pathway B)
		if stType, ok := sa.SemanticInfo.Types[n].(*types.StructType); ok {
			providedFields := make(map[string]bool)
			for _, field := range n.Fields {
				providedFields[field.Name.Value] = true
			}
			for fieldName := range stType.Fields {
				if !providedFields[fieldName] {
					sa.AddError(n.Token.Position, "missing field '%s' in struct literal of type '%s'", fieldName, stType.Name())
				}
			}
		}

	case *ast.PinStatement:
		for _, target := range n.Targets {
			sym, exists := sa.CurrentScope.Resolve(target.Value)
			if !exists {
				sa.AddError(target.Pos(), "undefined variable: '%s'", target.Value)
				continue
			}
			sa.SemanticInfo.Uses[target] = sym
			sym.IsPinned = true
		}
		sa.SemanticInfo.Types[n] = types.Void

		// We look for TypeStatements (e.g., "type User = ...")
	case *ast.NoneLiteral:
		sa.SemanticInfo.Types[n] = types.Ptr

	case *ast.PrefixExpression:
		sa.Analyze(n.Right)
		rightType := sa.SemanticInfo.Types[n.Right]
		if rightType == nil {
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		switch n.Operator {
		case "!", "-", "~":
			rightBase := rightType
			if pt, ok := rightType.(*types.PointerType); ok && !pt.IsArray {
				rightBase = pt.Base
			}

			if st, ok := rightBase.(*types.StructType); ok {
				methodName := ""
				switch n.Operator {
				case "!":
					methodName = "not"
				case "-":
					methodName = "neg"
				case "~":
					methodName = "bitnot"
				}
				if methodType, exists := st.Methods[methodName]; exists {
					if ft, ok := methodType.(*types.FunctionType); ok && len(ft.Params) == 0 {
						sa.SemanticInfo.Types[n] = ft.Return
						return
					} else {
						sa.AddError(n.Token.Position, "invalid signature for operator method '%s' (should take 0 parameters)", methodName)
						sa.SemanticInfo.Types[n] = types.ErrorType
						return
					}
				}
			}

			// Fallbacks
			if n.Operator == "!" {
				if rightType != types.Bool && rightType.GetKind() != types.KindGeneric {
					sa.AddError(n.Token.Position, "cannot apply '!' to %s without 'not' method", rightType.Name())
					sa.SemanticInfo.Types[n] = types.ErrorType
				} else {
					sa.SemanticInfo.Types[n] = types.Bool
				}
			} else if n.Operator == "-" {
				name := rightType.Name()
				isNumeric := name == "int" || name == "i64" || name == "i32" || name == "i16" || name == "i8" ||
					name == "u64" || name == "u32" || name == "u16" || name == "u8" || name == "byte" ||
					name == "f64" || name == "f32" || rightType.GetKind() == types.KindGeneric
				if !isNumeric {
					sa.AddError(n.Token.Position, "cannot apply '-' to %s without 'neg' method", rightType.Name())
					sa.SemanticInfo.Types[n] = types.ErrorType
				} else {
					sa.SemanticInfo.Types[n] = rightType
				}
			} else if n.Operator == "~" {
				if rightType != types.I32 && rightType != types.Int && rightType.GetKind() != types.KindGeneric {
					sa.AddError(n.Token.Position, "cannot apply '~' to %s without 'bitnot' method", rightType.Name())
					sa.SemanticInfo.Types[n] = types.ErrorType
				} else {
					sa.SemanticInfo.Types[n] = rightType
				}
			}
		case "#":
			sa.SemanticInfo.Types[n] = &types.PointerType{Base: rightType, Leased: true, Kind: types.LeaseRead}
		case "&":
			sa.SemanticInfo.Types[n] = &types.PointerType{Base: rightType, Leased: true, Kind: types.LeaseWrite}
			if sa.CurrentFunction != nil {
				if rootSym := sa.getRootSymbol(n.Right); rootSym != nil {
					if rootSym.Kind == SymVar && rootSym.DefScope != nil && (rootSym.DefScope.Kind == ScopePackage || rootSym.DefScope.Kind == ScopeGlobal) {
						if sa.mutatedGlobalsInFunc[sa.CurrentFunction] == nil {
							sa.mutatedGlobalsInFunc[sa.CurrentFunction] = make(map[*Symbol]bool)
						}
						sa.mutatedGlobalsInFunc[sa.CurrentFunction][rootSym] = true
					}
				}
			}
			if sa.inSpawn > 0 || sa.inParallel > 0 {
				if rootSym := sa.getRootSymbol(n.Right); rootSym != nil {
					if rootSym.Kind == SymVar && rootSym.DefScope != nil && (rootSym.DefScope.Kind == ScopePackage || rootSym.DefScope.Kind == ScopeGlobal) {
						sa.AddError(n.Token.Position, "cannot take mutable lease of global variable '%s' inside a concurrent context", rootSym.Name)
					}
				}
			}
		case "@":
			// In Nora, @x is move lease. Only for owned types.
			actualType := rightType
			if pt, ok := rightType.(*types.PointerType); ok && pt.Leased && pt.Kind != types.LeaseMove {
				actualType = pt.Base
			}
			if types.IsOwnedType(actualType) {
				sa.SemanticInfo.Types[n] = &types.PointerType{Base: actualType, Leased: true, Kind: types.LeaseMove}
			} else {
				sa.SemanticInfo.Types[n] = rightType
			}
			if id, ok := n.Right.(*ast.Identifier); ok {
				if sym := sa.SemanticInfo.Uses[id]; sym != nil {
					sa.SemanticInfo.Kill(sym, n)
				}
			}
		default:
			sa.SemanticInfo.Types[n] = types.ErrorType
		}

	case *ast.BreakStatement:
		sa.SemanticInfo.Types[n] = types.Void

	case *ast.ContinueStatement:
		sa.SemanticInfo.Types[n] = types.Void

	case *ast.TypeStatement:
		// 1. Retrieve the symbol we defined in Pass 1
		sym, exists := sa.CurrentScope.Lookup(n.Name.Value)
		if !exists {
			return // Should have been caught in Pass 1
		}

		// 3. Check if the value is actually a struct literal
		if structLit, ok := n.Value.(*ast.StructLiteral); ok {
			structType := sym.Type.(*types.StructType)
			// Add TypeParams to temporary scope for field resolution
			if len(n.TypeParameters) > 0 {
				sa.CurrentScope = NewScope(sa.CurrentScope, ScopeBlock)

				// Pass 1: Define all TypeParams in the new scope
				for _, tp := range n.TypeParameters {
					sym, _ := sa.CurrentScope.Define(tp.Name.Value, &types.GenericType{TypeParam: tp.Name.Value, Constraint: nil}, SymType, tp)
					if sym != nil {
						sa.SemanticInfo.Defs[tp.Name] = sym
					}
				}

				// Pass 2: Resolve constraints using the fully populated scope
				for i, tp := range n.TypeParameters {
					if i < len(structType.TypeParams) {
						var constraint types.NRType = types.Any
						if tp.Constraint != nil {
							constraint = sa.resolveTypeNode(tp.Constraint)
						}
						structType.TypeParams[i].Constraint = constraint

						if sym, found := sa.CurrentScope.Lookup(tp.Name.Value); found {
							if gt, ok := sym.Type.(*types.GenericType); ok {
								gt.Constraint = constraint
							}
						}
					}
				}
			}

			for _, fieldDef := range structLit.Fields {
				fieldName := fieldDef.Name.Value
				if _, hasMethod := structType.Methods[fieldName]; hasMethod {
					sa.AddError(fieldDef.Name.Pos(), "name conflict in struct '%s': field '%s' conflicts with method '%s'", structType.TypeName, fieldName, fieldName)
				}
				resolvedType := sa.resolveTypeNode(fieldDef.Type)
				if _, exists := structType.Fields[fieldName]; !exists {
					structType.FieldNames = append(structType.FieldNames, fieldName)
				}
				structType.Fields[fieldName] = resolvedType

				fieldSym := &Symbol{
					Name:    fieldName,
					Type:    resolvedType,
					Kind:    SymVar,
					DefNode: fieldDef,
				}
				if sa.SemanticInfo.FieldSymbols[structType] == nil {
					sa.SemanticInfo.FieldSymbols[structType] = make(map[string]*Symbol)
				}
				sa.SemanticInfo.FieldSymbols[structType][fieldName] = fieldSym
			}

			// Restore scope after field resolution
			// (e.g. Node_T created while evaluating #Node[T] field).
			for _, spec := range sa.SemanticInfo.SpecTypes {
				if specStruct, ok := spec.(*types.StructType); ok && specStruct.BaseType == structType {
					// Re-build substitution mapping from TypeArgs
					subs := make(map[string]types.NRType)
					for i, tp := range structType.TypeParams {
						if i < len(specStruct.TypeArgs) {
							subs[tp.Name] = specStruct.TypeArgs[i]
						}
					}

					// Update fields with substituted types
					for fName, fType := range structType.Fields {
						specStruct.Fields[fName] = sa.substituteType(fType, subs)
					}

					// Update methods (if any were added)
					for mName, mType := range structType.Methods {
						specStruct.Methods[mName] = sa.substituteType(mType, subs)
					}
				}
			}

			if len(n.TypeParameters) > 0 {
				sa.CurrentScope = sa.CurrentScope.Parent
			}
		} else if interfaceLit, ok := n.Value.(*ast.InterfaceLiteral); ok {
			protocolType := sym.Type.(*types.ProtocolType)

			// Add TypeParams to temporary scope for method resolution
			if len(n.TypeParameters) > 0 {
				sa.CurrentScope = NewScope(sa.CurrentScope, ScopeBlock)

				// Pass 1: Define all TypeParams in the new scope
				for _, tp := range n.TypeParameters {
					sym, _ := sa.CurrentScope.Define(tp.Name.Value, &types.GenericType{TypeParam: tp.Name.Value, Constraint: nil}, SymType, tp)
					if sym != nil {
						sa.SemanticInfo.Defs[tp.Name] = sym
					}
				}

				// Pass 2: Resolve constraints using the fully populated scope
				for i, tp := range n.TypeParameters {
					if i < len(protocolType.TypeParams) {
						var constraint types.NRType = types.Any
						if tp.Constraint != nil {
							constraint = sa.resolveTypeNode(tp.Constraint)
						}
						protocolType.TypeParams[i].Constraint = constraint

						if sym, found := sa.CurrentScope.Lookup(tp.Name.Value); found {
							if gt, ok := sym.Type.(*types.GenericType); ok {
								gt.Constraint = constraint
							}
						}
					}
				}
			}

			// 1. Resolve Embedded Interfaces
			for _, embed := range interfaceLit.Embedded {
				embeddedType := sa.resolveTypeNode(embed)
				if proto, ok := embeddedType.(*types.ProtocolType); ok {
					// Copy methods from embedded interface with duplicate check
					for name, method := range proto.Methods {
						if existingMethod, exists := protocolType.Methods[name]; exists {
							if !types.Equals(existingMethod, method) {
								sa.AddError(embed.Pos(), "interface embedding conflict: method '%s' defined in embedded interfaces has incompatible signatures", name)
							}
						} else {
							protocolType.Methods[name] = method
						}
					}
				} else {
					sa.AddError(embed.Pos(), "cannot embed non-interface type '%s'", embeddedType.Name())
				}
			}

			// 2. Resolve Methods
			for _, m := range interfaceLit.Methods {
				mType := &types.FunctionType{
					Params:      []types.NRType{},
					ParamLeases: []types.LeaseKind{},
				}
				if m.Receiver != nil {
					mType.Receiver = sa.resolveTypeNode(m.Receiver.Type)
				}
				for _, p := range m.Parameters {
					pType := sa.resolveTypeNode(p.Type)
					mType.Params = append(mType.Params, pType)

					// Resolve lease kind for interface method parameters
					lease := types.LeaseRead
					if p.LeaseKind != ast.LeaseRead {
						lease = types.LeaseKind(p.LeaseKind)
					} else if pref, ok := p.Type.(*ast.PrefixExpression); ok {
						if pref.Operator == "&" {
							lease = types.LeaseWrite
						} else if pref.Operator == "@" {
							lease = types.LeaseMove
						} else if pref.Operator == "#" {
							lease = types.LeaseRead
						}
					} else if pType != nil && pType.GetKind() != types.KindFunction && (pType.GetKind() == types.KindGeneric || types.IsOwnedType(pType)) {
						lease = types.LeaseMove
					}
					p.LeaseKind = ast.LeaseKind(lease)
					mType.ParamLeases = append(mType.ParamLeases, lease)
				}
				mType.Return = sa.resolveTypeNode(m.ReturnType)
				methodSym := &Symbol{
					Name:    m.Name.Value,
					Type:    mType,
					Kind:    SymFunc,
					DefNode: m.Name,
				}
				if sa.SemanticInfo.MethodSymbols[protocolType] == nil {
					sa.SemanticInfo.MethodSymbols[protocolType] = make(map[string]*Symbol)
				}
				sa.SemanticInfo.MethodSymbols[protocolType][m.Name.Value] = methodSym
				if existingMethod, exists := protocolType.Methods[m.Name.Value]; exists {
					if !types.Equals(existingMethod, mType) {
						sa.AddError(m.Name.Pos(), "interface embedding conflict: method '%s' has incompatible signatures with embedded method", m.Name.Value)
					}
				}
				protocolType.Methods[m.Name.Value] = mType
			}

			if len(n.TypeParameters) > 0 {
				sa.CurrentScope = sa.CurrentScope.Parent
			}
		} else if sumLit, ok := n.Value.(*ast.SumTypeLiteral); ok {
			sumType := sym.Type.(*types.SumType)
			// Add TypeParams to temporary scope for variant resolution
			if len(n.TypeParameters) > 0 {
				sa.CurrentScope = NewScope(sa.CurrentScope, ScopeBlock)

				// Pass 1: Define all TypeParams in the new scope
				for _, tp := range n.TypeParameters {
					sym, _ := sa.CurrentScope.Define(tp.Name.Value, &types.GenericType{TypeParam: tp.Name.Value, Constraint: nil}, SymType, tp)
					if sym != nil {
						sa.SemanticInfo.Defs[tp.Name] = sym
					}
				}

				// Pass 2: Resolve constraints using the fully populated scope
				for i, tp := range n.TypeParameters {
					if i < len(sumType.TypeParams) {
						var constraint types.NRType = types.Any
						if tp.Constraint != nil {
							constraint = sa.resolveTypeNode(tp.Constraint)
						}
						sumType.TypeParams[i].Constraint = constraint

						if sym, found := sa.CurrentScope.Lookup(tp.Name.Value); found {
							if gt, ok := sym.Type.(*types.GenericType); ok {
								gt.Constraint = constraint
							}
						}
					}
				}
			}

			sumType.Variants = make(map[string]*types.Variant)
			for i, vDef := range sumLit.Variants {
				variant := &types.Variant{
					Name:   vDef.Name.Value,
					Tag:    i,
					Fields: make(map[string]types.NRType),
				}
				for _, fDef := range vDef.Fields {
					fName := fDef.Name.Value
					fType := sa.resolveTypeNode(fDef.Type)
					if _, exists := variant.Fields[fName]; !exists {
						variant.FieldNames = append(variant.FieldNames, fName)
					}
					variant.Fields[fName] = fType
				}
				sumType.Variants[variant.Name] = variant

				// Define variant constructor in scope
				targetScope := sa.CurrentScope
				if len(sumType.TypeParams) > 0 {
					targetScope = sa.CurrentScope.Parent
				}

				if len(variant.Fields) == 0 {
					targetScope.Define(variant.Name, sumType, SymVariant, vDef)
				} else {
					// Variant constructor function type
					params := []types.NRType{}
					leases := []types.LeaseKind{}
					for _, fName := range variant.FieldNames {
						fType := variant.Fields[fName]
						params = append(params, fType)

						lease := types.LeaseMove
						if pt, ok := fType.(*types.PointerType); ok && pt.Leased {
							lease = pt.Kind
						}
						leases = append(leases, lease)
					}
					ctorType := &types.FunctionType{
						Params:      params,
						ParamLeases: leases,
						Return:      sumType,
					}
					targetScope.Define(variant.Name, ctorType, SymVariant, vDef)
				}
			}

			// FIX: Re-sync any SpecTypes that were created prematurely during variant resolution
			for _, spec := range sa.SemanticInfo.SpecTypes {
				if specSum, ok := spec.(*types.SumType); ok && specSum.BaseType == sumType {
					subs := make(map[string]types.NRType)
					for i, tp := range sumType.TypeParams {
						if i < len(specSum.TypeArgs) {
							subs[tp.Name] = specSum.TypeArgs[i]
						}
					}
					// Update variants
					for vName, vBase := range sumType.Variants {
						vSpec := &types.Variant{
							Name:       vBase.Name,
							Tag:        vBase.Tag,
							Fields:     make(map[string]types.NRType),
							FieldNames: vBase.FieldNames,
						}
						for fName, fType := range vBase.Fields {
							vSpec.Fields[fName] = sa.substituteType(fType, subs)
						}
						specSum.Variants[vName] = vSpec
					}
				}
			}

			if len(n.TypeParameters) > 0 {
				sa.CurrentScope = sa.CurrentScope.Parent
			}
			sym.Type = sumType
		} else {
			sa.AddError(n.Value.Pos(), "expected struct, interface, or enum definition")
			return
		}

	// --- TOP LEVEL ---
	case *ast.PackageStatement:
		// Just map the node to the current package scope we created in *ast.Program
		sa.SemanticInfo.Scopes[n] = sa.CurrentScope

	// --- FUNCTIONS ---
	case *ast.FunctionStatement:
		sa.AnalyzeFunctionStatement(n)

	case *ast.LambdaExpression:
		sa.AnalyzeLambdaExpression(n)

	case *ast.IntegerLiteral:
		if n.Suffix != "" {
			t, ok := types.LookupPrimitive(n.Suffix)
			if ok {
				sa.SemanticInfo.Types[n] = t
				return
			}
		}
		if n.Value < -2147483648 || n.Value > 2147483647 {
			sa.SemanticInfo.Types[n] = types.I64
		} else {
			sa.SemanticInfo.Types[n] = types.I32
		}

	case *ast.ImaginaryLiteral:
		sa.SemanticInfo.Types[n] = types.F64

	case *ast.StringLiteral:
		sa.SemanticInfo.Types[n] = types.String

	case *ast.InterpolatedString:
		for _, part := range n.Parts {
			sa.Analyze(part)
		}
		sa.SemanticInfo.Types[n] = types.String

	case *ast.Boolean:
		sa.SemanticInfo.Types[n] = types.Bool

	case *ast.RuneLiteral:
		sa.SemanticInfo.Types[n] = types.I32

	case *ast.ArrayLiteral:
		var elemType types.NRType = types.ErrorType
		if len(n.Elements) > 0 {
			sa.Analyze(n.Elements[0])
			elemType = sa.SemanticInfo.Types[n.Elements[0]]
			for i := 1; i < len(n.Elements); i++ {
				sa.Analyze(n.Elements[i])
				t := sa.SemanticInfo.Types[n.Elements[i]]
				if !types.Equals(elemType, t) {
					sa.AddError(n.Elements[i].Pos(), "array element type mismatch: expected %s, got %s", elemType.Name(), t.Name())
				}
			}
		}
		sa.SemanticInfo.Types[n] = &types.ListType{ElementType: elemType}

	case *ast.MapLiteral:
		var keyType types.NRType = types.ErrorType
		var valType types.NRType = types.ErrorType
		// Type inference from first pair
		first := true
		for k, v := range n.Pairs {
			sa.Analyze(k)
			sa.Analyze(v)
			kt := sa.SemanticInfo.Types[k]
			vt := sa.SemanticInfo.Types[v]
			if first {
				keyType = kt
				valType = vt
				first = false
			} else {
				if !types.Equals(keyType, kt) {
					sa.AddError(k.Pos(), "map key type mismatch")
				}
				if !types.Equals(valType, vt) {
					sa.AddError(v.Pos(), "map value type mismatch")
				}
			}
		}
		sa.SemanticInfo.Types[n] = &types.MapType{Key: keyType, Value: valType}

	case *ast.MatchExpression:
		sa.Analyze(n.Target)
		targetType := sa.SemanticInfo.Types[n.Target]
		if targetType != nil {
			targetType = types.UnwrapLease(targetType)
		}
		if targetType == nil {
			targetType = types.ErrorType
		}

		var resultType types.NRType = types.Void

		// Save ownership state and initialization states before branching
		origDead := make(map[*Symbol]ast.Node)
		for k, v := range sa.SemanticInfo.DeadSyms {
			origDead[k] = v
		}
		origInit := make(map[*Symbol]InitState)
		for k, v := range sa.InitStates {
			origInit[k] = v
		}
		allBranchesDead := []map[*Symbol]ast.Node{}
		allBranchesInit := []map[*Symbol]InitState{}
		allBranchesTerminated := []bool{}

		for i, c := range n.Cases {
			// Restore for this case
			sa.SemanticInfo.DeadSyms = make(map[*Symbol]ast.Node)
			for k, v := range origDead {
				sa.SemanticInfo.DeadSyms[k] = v
			}
			for k := range sa.InitStates {
				sa.InitStates[k] = origInit[k]
			}

			sa.CurrentScope = NewScope(sa.CurrentScope, ScopeBlock)

			// Pattern Handling
			origTargetType := sa.SemanticInfo.Types[n.Target]
			isTargetLeased := false
			if origTargetType != nil && origTargetType.IsLeased() {
				isTargetLeased = true
			} else if pref, ok := n.Target.(*ast.PrefixExpression); ok && (pref.Operator == "#" || pref.Operator == "&") {
				isTargetLeased = true
			} else if id, ok := n.Target.(*ast.Identifier); ok {
				if sym, exists := sa.CurrentScope.Lookup(id.Value); exists && sym != nil {
					if sym.Type != nil && sym.Type.IsLeased() {
						isTargetLeased = true
					}
				}
			}

			isMove := !isTargetLeased
			if pref, ok := n.Target.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				isMove = true
			}
			sa.analyzePattern(c.Pattern, targetType, isMove)

			sa.Analyze(c.Body)
			branchType := sa.SemanticInfo.Types[c.Body]

			if i == 0 {
				resultType = branchType
			} else {
				// Verify all branches have compatible types
				isResVoid := resultType == nil || resultType == types.Void
				isBrVoid := branchType == nil || branchType == types.Void
				var matchMatch bool
				if isResVoid && isBrVoid {
					matchMatch = true
				} else {
					matchMatch = types.Equals(resultType, branchType)
				}
				if !matchMatch {
					resName := "void"
					if resultType != nil {
						resName = resultType.Name()
					}
					brName := "void"
					if branchType != nil {
						brName = branchType.Name()
					}
					sa.AddError(c.Body.Pos(), "match branch type mismatch: expected %s, got %s", resName, brName)
				}
			}

			// Record dead syms after this branch
			branchDead := make(map[*Symbol]ast.Node)
			for k, v := range sa.SemanticInfo.DeadSyms {
				branchDead[k] = v
			}
			allBranchesDead = append(allBranchesDead, branchDead)

			// Record initialization states after this branch
			branchInit := make(map[*Symbol]InitState)
			for k, v := range sa.InitStates {
				branchInit[k] = v
			}
			allBranchesInit = append(allBranchesInit, branchInit)
			allBranchesTerminated = append(allBranchesTerminated, sa.terminates(c.Body))

			sa.CurrentScope = sa.CurrentScope.Parent
		}

		// Merge all branch deaths back to current state
		sa.SemanticInfo.DeadSyms = make(map[*Symbol]ast.Node)
		for k, v := range origDead {
			sa.SemanticInfo.DeadSyms[k] = v
		}
		for _, bd := range allBranchesDead {
			for k, v := range bd {
				sa.SemanticInfo.DeadSyms[k] = v
			}
		}

		// Merge all branch initialization states
		nonTerminatingCount := 0
		for _, term := range allBranchesTerminated {
			if !term {
				nonTerminatingCount++
			}
		}

		for sym := range origInit {
			if nonTerminatingCount == 0 {
				sa.InitStates[sym] = origInit[sym]
				continue
			}

			allInit := true
			anyInitOrPartial := false

			for idx, bInit := range allBranchesInit {
				if allBranchesTerminated[idx] {
					continue
				}
				state := bInit[sym]
				if state == Initialized {
					anyInitOrPartial = true
				} else {
					allInit = false
				}
				if state == PartiallyInitialized {
					anyInitOrPartial = true
				}
			}

			if allInit {
				sa.InitStates[sym] = Initialized
			} else if anyInitOrPartial {
				sa.InitStates[sym] = PartiallyInitialized
			} else {
				sa.InitStates[sym] = Uninitialized
			}
		}

		sa.SemanticInfo.Types[n] = resultType

		// Exhaustiveness Check for Sum Types and Any/Protocols
		if sumType, isSum := targetType.(*types.SumType); isSum {
			coveredVariants := make(map[string]bool)
			hasWildcard := false

			for _, c := range n.Cases {
				if ident, ok := c.Pattern.(*ast.Identifier); ok && ident.Value == "_" {
					hasWildcard = true
					continue
				}

				if varName, ok := sa.extractMatchedVariant(c.Pattern, sumType); ok {
					coveredVariants[varName] = true
				}
			}

			if !hasWildcard {
				var missing []string
				var allVariants []string
				for k := range sumType.Variants {
					allVariants = append(allVariants, k)
				}
				sort.Strings(allVariants)

				for _, name := range allVariants {
					if !coveredVariants[name] {
						missing = append(missing, name)
					}
				}

				if len(missing) > 0 {
					sa.AddError(n.Pos(), "match statement on type '%s' is not exhaustive; missing variants: %s (or use a wildcard '_')", targetType.Name(), strings.Join(missing, ", "))
				}
			}
		} else {
			_, isProto := types.UnwrapLease(targetType).(*types.ProtocolType)
			if isProto {
				hasWildcard := false
				for _, c := range n.Cases {
					if ident, ok := c.Pattern.(*ast.Identifier); ok && ident.Value == "_" {
						hasWildcard = true
						break
					}
				}
				if !hasWildcard {
					sa.AddError(n.Pos(), "match statement on type '%s' is not exhaustive; missing fallback wildcard '_'", targetType.Name())
				}
			}
		}

	case *ast.IfExpression:
		sa.Analyze(n.Condition)
		condType := sa.SemanticInfo.Types[n.Condition]
		if condType != types.Bool && condType != types.ErrorType {
			sa.AddError(n.Condition.Pos(), "if condition must be boolean, got %s", condType.Name())
		}

		// Save ownership state (DeadSyms) and initialization states before branching
		origDead := make(map[*Symbol]ast.Node)
		for k, v := range sa.SemanticInfo.DeadSyms {
			origDead[k] = v
		}
		origInit := make(map[*Symbol]InitState)
		for k, v := range sa.InitStates {
			origInit[k] = v
		}

		sa.Analyze(n.Consequence)
		consequenceDead := make(map[*Symbol]ast.Node)
		for k, v := range sa.SemanticInfo.DeadSyms {
			consequenceDead[k] = v
		}
		consequenceInit := make(map[*Symbol]InitState)
		for k, v := range sa.InitStates {
			consequenceInit[k] = v
		}

		// Restore for Alternative
		sa.SemanticInfo.DeadSyms = make(map[*Symbol]ast.Node)
		for k, v := range origDead {
			sa.SemanticInfo.DeadSyms[k] = v
		}
		for k := range sa.InitStates {
			sa.InitStates[k] = origInit[k]
		}

		if n.Alternative != nil {
			sa.Analyze(n.Alternative)
			alternativeDead := make(map[*Symbol]ast.Node)
			for k, v := range sa.SemanticInfo.DeadSyms {
				alternativeDead[k] = v
			}
			alternativeInit := make(map[*Symbol]InitState)
			for k, v := range sa.InitStates {
				alternativeInit[k] = v
			}

			// Restore for Merge
			sa.SemanticInfo.DeadSyms = make(map[*Symbol]ast.Node)
			for k, v := range origDead {
				sa.SemanticInfo.DeadSyms[k] = v
			}
			for k := range sa.InitStates {
				sa.InitStates[k] = origInit[k]
			}

			// Merge: Conservative ownership.
			// If a branch terminates, it doesn't contribute its deaths to the main path!
			if !sa.terminates(n.Consequence) {
				for k, v := range consequenceDead {
					sa.SemanticInfo.DeadSyms[k] = v
				}
			}
			if !sa.terminates(n.Alternative) {
				for k, v := range alternativeDead {
					sa.SemanticInfo.DeadSyms[k] = v
				}
			}

			// Merge: flow-sensitive initialization
			tTerm := sa.terminates(n.Consequence)
			aTerm := sa.terminates(n.Alternative)

			for sym := range origInit {
				tState := consequenceInit[sym]
				aState := alternativeInit[sym]

				if tTerm && aTerm {
					sa.InitStates[sym] = origInit[sym]
				} else if tTerm {
					sa.InitStates[sym] = aState
				} else if aTerm {
					sa.InitStates[sym] = tState
				} else {
					if tState == Initialized && aState == Initialized {
						sa.InitStates[sym] = Initialized
					} else if tState == Initialized || aState == Initialized || tState == PartiallyInitialized || aState == PartiallyInitialized {
						sa.InitStates[sym] = PartiallyInitialized
					} else {
						sa.InitStates[sym] = Uninitialized
					}
				}
			}

			consequenceType := sa.SemanticInfo.Types[n.Consequence]
			alternativeType := sa.SemanticInfo.Types[n.Alternative]
			conName := "void"
			if consequenceType != nil {
				conName = consequenceType.Name()
			}
			altName := "void"
			if alternativeType != nil {
				altName = alternativeType.Name()
			}
			isConVoid := consequenceType == nil || consequenceType == types.Void
			isAltVoid := alternativeType == nil || alternativeType == types.Void
			var typesMatch bool
			if isConVoid && isAltVoid {
				typesMatch = true
			} else {
				typesMatch = types.Equals(consequenceType, alternativeType)
			}
			if !typesMatch {
				sa.AddError(n.Token.Position, "if-else branches have mismatching types: %s and %s", conName, altName)
				sa.SemanticInfo.Types[n] = types.ErrorType
			} else {
				sa.SemanticInfo.Types[n] = consequenceType
			}
		} else {
			// Merge consequence only if it doesn't return
			if !sa.terminates(n.Consequence) {
				for k, v := range consequenceDead {
					sa.SemanticInfo.DeadSyms[k] = v
				}
			}

			// Merge: flow-sensitive initialization (no else branch)
			tTerm := sa.terminates(n.Consequence)
			for sym := range origInit {
				tState := consequenceInit[sym]
				if tTerm {
					sa.InitStates[sym] = origInit[sym]
				} else {
					if origInit[sym] == Uninitialized && tState == Initialized {
						sa.InitStates[sym] = PartiallyInitialized
					} else {
						sa.InitStates[sym] = origInit[sym]
					}
				}
			}

			sa.SemanticInfo.Types[n] = sa.SemanticInfo.Types[n.Consequence]
		}

	case *ast.ReturnStatement:
		if sa.CurrentFunction == nil && sa.CurrentLambda == nil {
			sa.AddError(n.Token.Position, "return statement outside of function")
			return
		}

		var returnType types.NRType = types.Void
		if n.ReturnValue != nil {
			sa.Analyze(n.ReturnValue)
			returnType = sa.SemanticInfo.Types[n.ReturnValue]

			if sa.hasRestrictedClosure(returnType, nil) {
				sa.AddError(n.Token.Position, "closure captures local lease and cannot escape the function")
			}
		}

		var expectedType types.NRType
		var funcName string
		if sa.CurrentLambda != nil {
			expectedType = sa.resolveTypeNode(sa.CurrentLambda.ReturnType)
			funcName = "lambda"
		} else {
			expectedType = sa.resolveTypeNode(sa.CurrentFunction.ReturnType)
			funcName = sa.CurrentFunction.Name.Value
		}

		if expectedType == nil {
			panic("EXPECTED TYPE IS NIL INTERFACE: func=" + funcName)
		}
		if pt, ok := expectedType.(*types.PointerType); ok && pt.Base == nil {
			panic("EXPECTED TYPE HAS NIL BASE: func=" + funcName)
		}
		if returnType == nil {
			panic("RETURN TYPE IS NIL INTERFACE: func=" + funcName)
		}
		if pt, ok := returnType.(*types.PointerType); ok && pt.Base == nil {
			panic("RETURN TYPE HAS NIL BASE: func=" + funcName)
		}
		if expectedType == nil {
			expectedType = sa.resolveTypeNode(sa.CurrentFunction.ReturnType)
		}
		sa.checkImplicitMoveLoad(n.ReturnValue, expectedType)

		if !types.IsAssignable(expectedType, returnType) {
			if _, ok := expectedType.(*types.ProtocolType); ok {
				sa.checkInterfaceCompatibility(n.ReturnValue, expectedType)
			} else {
				sa.AddError(n.Token.Position, "cannot return value of type %s from function returning %s", returnType.Name(), expectedType.Name())
			}
		}

	// --- BLOCKS ---
	case *ast.BlockStatement:
		prevScope := sa.CurrentScope
		sa.CurrentScope = NewScope(prevScope, ScopeBlock)
		sa.SemanticInfo.Scopes[n] = sa.CurrentScope

		var lastType types.NRType = types.Void
		for i, stmt := range n.Statements {
			sa.Analyze(stmt)
			if t, ok := sa.SemanticInfo.Types[stmt]; ok {
				lastType = t
			} else {
				lastType = types.Void
			}

			// If it's an ExpressionStatement, and it's NOT the last statement, its value is discarded.
			if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok && i < len(n.Statements)-1 {
				if sa.containsOwnedLease(lastType, nil) {
					sa.AddError(exprStmt.Pos(), "cannot discard owned value of type '%s'. it must be assigned, moved, or passed to a function to avoid memory leaks", lastType.Name())
				}
			}
		}
		sa.SemanticInfo.Types[n] = lastType
		sa.checkUnusedSymbolsInScope(sa.CurrentScope)
		sa.CurrentScope = prevScope

	case *ast.VarStatement:
		var rhsType types.NRType = types.Void
		if n.Value != nil {
			// 1. Analyze the Right-Hand Side (Value)
			sa.Analyze(n.Value)

			// 2. Determine the Type of the RHS
			rhsType = sa.SemanticInfo.Types[n.Value]
			if rhsType == nil {
				rhsType = types.ErrorType
			}
		}

		// 3. Resolve the Final Variable Type
		var finalType types.NRType

		if n.Type != nil {
			explicitType := sa.resolveTypeNode(n.Type)
			if n.Value != nil {
				sa.checkImplicitMoveLoad(n.Value, explicitType)
				if !types.IsAssignable(explicitType, rhsType) {
					// Interface compatibility check
					if _, ok := explicitType.(*types.ProtocolType); ok {
						sa.checkInterfaceCompatibility(n.Value, explicitType)
					} else {
						sa.AddError(n.Value.Pos(), "type mismatch: cannot assign %s to %s", rhsType.Name(), explicitType.Name())
					}
				}
			}
			finalType = explicitType
		} else {
			if n.Value == nil {
				sa.AddError(n.Name.Pos(), "variable '%s' requires a type or an initializer", n.Name.Value)
				finalType = types.ErrorType
			} else {
				finalType = rhsType
			}
		}

		if finalType == types.ErrorType || finalType == nil {
			sa.AddError(n.Name.Pos(), "cannot infer type for variable '%s'", n.Name.Value)
			finalType = types.ErrorType
		}
		// 5. Define Symbol in Scope
		var sym *Symbol
		var err error
		if existing, ok := sa.CurrentScope.Symbols[n.Name.Value]; ok && existing.DefNode == n {
			sym = existing
			sym.Type = finalType // Update from Void to actual type
		} else {
			sym, err = sa.CurrentScope.Define(n.Name.Value, finalType, SymVar, n)
		}

		if err != nil {
			sa.AddError(n.Name.Pos(), "%s", err.Error())
		} else {
			sa.SemanticInfo.Defs[n.Name] = sym
			sym.IsInitialized = (n.Value != nil)
			if n.Value != nil {
				sa.InitStates[sym] = Initialized
			} else {
				sa.InitStates[sym] = Uninitialized
			}
			sym.WritePerm = true
			if n.IsPublic {
				sym.Visible = Public
			}

			// Determine LeaseKind for the new variable
			if pt, ok := finalType.(*types.PointerType); ok && pt.Leased {
				sym.LeaseKind = pt.Kind
			} else if pref, ok := n.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				sym.LeaseKind = types.LeaseMove
			} else if _, ok := n.Value.(*ast.AllocExpression); ok {
				sym.LeaseKind = types.LeaseMove
			} else if id, ok := n.Value.(*ast.Identifier); ok {
				if rhsSym := sa.SemanticInfo.Uses[id]; rhsSym != nil {
					sym.LeaseKind = rhsSym.LeaseKind
				} else {
					sym.LeaseKind = types.LeaseMove // Fallback for things like built-ins
				}
			} else {
				// Literals, Allocations, and Function Calls provide ownership
				// --- FIX: Primitives are copy-types (LeaseRead), others are owned (LeaseMove) ---
				// EXCEPTION: String literals are static and shouldn't be moved/dropped.
				isLiteral := false
				if _, ok := n.Value.(*ast.StringLiteral); ok {
					isLiteral = true
				}

				if types.IsOwnedType(finalType) && !isLiteral {
					sym.LeaseKind = types.LeaseMove
				} else {
					sym.LeaseKind = types.LeaseRead
				}
				sa.debug("Var %s (type %s) assigned LeaseKind %v (isLiteral=%v, isOwned=%v)",
					sym.Name, finalType.Name(), sym.LeaseKind, isLiteral, types.IsOwnedType(finalType))
			}

			// --- Nora: Implicit Move Tracking ---
			if id, ok := n.Value.(*ast.Identifier); ok {
				if rhsSym := sa.SemanticInfo.Uses[id]; rhsSym != nil {
					if types.IsOwnedType(rhsSym.Type) && !rhsSym.Type.IsLeased() && rhsSym.LeaseKind != types.LeaseRead && rhsSym.LeaseKind != types.LeaseWrite {
						sa.SemanticInfo.Kill(rhsSym, n)
					}
				}
			}
		}

		if sa.CurrentScope.Kind == ScopePackage || sa.CurrentScope.Kind == ScopeGlobal {
			if sa.hasRestrictedClosure(finalType, nil) {
				sa.AddError(n.Token.Position, "closure captures local lease and cannot be assigned to global/package scope")
			}
		}

		sa.SemanticInfo.Types[n] = types.Void
	case *ast.AssignmentStatement:
		if ast.IsNil(n.Left) {
			sa.AddError(n.Pos(), "missing assignment target")
			return
		}
		if ast.IsNil(n.Value) {
			sa.AddError(n.Pos(), "missing assignment value")
			return
		}

		// 1. Analyze RHS first
		sa.Analyze(n.Value)
		// Only analyze LHS if it's NOT a simple identifier (to avoid false use-after-move errors during revival)
		if _, ok := n.Left.(*ast.Identifier); !ok {
			sa.Analyze(n.Left)
		}

		if sa.CurrentFunction != nil {
			if _, ok := n.Left.(*ast.Identifier); !ok {
				if rootSym := sa.getRootSymbol(n.Left); rootSym != nil {
					if rootSym.Kind == SymVar && rootSym.DefScope != nil && (rootSym.DefScope.Kind == ScopePackage || rootSym.DefScope.Kind == ScopeGlobal) {
						if sa.mutatedGlobalsInFunc[sa.CurrentFunction] == nil {
							sa.mutatedGlobalsInFunc[sa.CurrentFunction] = make(map[*Symbol]bool)
						}
						sa.mutatedGlobalsInFunc[sa.CurrentFunction][rootSym] = true
					}
				}
			}
		}

		if sa.inSpawn > 0 || sa.inParallel > 0 {
			if _, ok := n.Left.(*ast.Identifier); !ok {
				if rootSym := sa.getRootSymbol(n.Left); rootSym != nil {
					if rootSym.Kind == SymVar && rootSym.DefScope != nil && (rootSym.DefScope.Kind == ScopePackage || rootSym.DefScope.Kind == ScopeGlobal) {
						sa.AddError(n.Left.Pos(), "cannot mutate global variable '%s' inside a concurrent context", rootSym.Name)
						return
					}
				}
			}
		}

		// 2. Resolve the target type and check permissions
		var targetType types.NRType
		if ident, ok := n.Left.(*ast.Identifier); ok {
			sym, exists := sa.CurrentScope.Resolve(ident.Value)
			if !exists {
				sa.AddError(ident.Pos(), "undefined variable: '%s'", ident.Value)
				return
			}
			sa.invalidateBounds(sym)
			if sa.CurrentFunction != nil {
				if sym.Kind == SymVar && sym.DefScope != nil && (sym.DefScope.Kind == ScopePackage || sym.DefScope.Kind == ScopeGlobal) {
					if sa.mutatedGlobalsInFunc[sa.CurrentFunction] == nil {
						sa.mutatedGlobalsInFunc[sa.CurrentFunction] = make(map[*Symbol]bool)
					}
					sa.mutatedGlobalsInFunc[sa.CurrentFunction][sym] = true
				}
			}
			if sa.inSpawn > 0 || sa.inParallel > 0 {
				if sym.Kind == SymVar && sym.DefScope != nil && (sym.DefScope.Kind == ScopePackage || sym.DefScope.Kind == ScopeGlobal) {
					sa.AddError(ident.Pos(), "cannot mutate global variable '%s' inside a concurrent context", ident.Value)
					return
				}
			}
			targetType = sym.Type
			// Link AST usage
			sa.SemanticInfo.Uses[ident] = sym
			sa.SemanticInfo.Defs[ident] = sym

			// Update LeaseKind if it's a move
			if pref, ok := n.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				sym.LeaseKind = types.LeaseMove
			}

			// Nora: Revival - variable is no longer dead after re-assignment
			sa.SemanticInfo.Alive(sym)
			sym.IsInitialized = true
			sa.InitStates[sym] = Initialized

			// --- Nora: FROZEN SHARING CHECK ---
			// We allow modification (re-assignment) of captured variables
			// ONLY if they have been moved (so the original copy is gone),
			// OR if they are channels (which are designed for shared communication).
			if sym.IsCaptured && !sa.SemanticInfo.IsDead(sym) && sym.Type.GetKind() != types.KindChan {
				sa.AddError(ident.Pos(), "cannot modify frozen variable '%s' shared with another fiber", ident.Value)
				return
			}

			// Check: Is it a variable or parameter?
			if sym.Kind != SymVar && sym.Kind != SymParam {
				sa.AddError(ident.Pos(), "cannot assign to '%s' (it is a %s)", ident.Value, sym.Kind)
				return
			}

			if sa.hasRestrictedClosure(sa.SemanticInfo.Types[n.Value], nil) {
				if sym.DefScope != nil && (sym.DefScope.Kind == ScopePackage || sym.DefScope.Kind == ScopeGlobal) {
					sa.AddError(ident.Pos(), "closure captures local lease and cannot be assigned to global/package scope")
				}
			}

			// Check: Write permission
			if !sym.WritePerm {
				if sym.Kind == SymVar {
					sym.WritePerm = true // PROMOTE
				} else {
					sa.AddError(ident.Pos(), "cannot assign to read lease parameter '%s' (did you mean '#%s'?)", ident.Value, ident.Value)
					return
				}
			}
		} else if idx, ok := n.Left.(*ast.IndexExpression); ok {
			if !sa.checkWritePermission(idx) {
				sa.AddError(idx.Pos(), "cannot modify element through a read-only lease")
				return
			}
			leftType := sa.SemanticInfo.Types[idx.Left]
			actualType := sa.unwrapToCollection(leftType)

			if lt, ok := actualType.(*types.ListType); ok {
				targetType = lt.ElementType
			} else if mt, ok := actualType.(*types.MapType); ok {
				targetType = mt.Value
			} else if actualType != nil && actualType.Name() == "str" {
				if sa.isUnsafeAllowed(idx.Pos()) && sa.hasUnsafeAttr() {
					targetType = types.I32
				} else {
					sa.AddError(idx.Pos(), "cannot assign to immutable string index")
					return
				}
			} else if pt, ok := actualType.(*types.PointerType); ok && pt.IsArray {
				targetType = pt.Base
			} else {
				baseActualType := actualType
				if pt, ok := actualType.(*types.PointerType); ok && !pt.IsArray {
					baseActualType = pt.Base
				}
				if st, ok := baseActualType.(*types.StructType); ok {
					if methodType, exists := st.Methods["index_mut"]; exists {
						if ft, ok := methodType.(*types.FunctionType); ok && len(ft.Params) == 1 {
							targetType = ft.Return
							if pt, ok := targetType.(*types.PointerType); ok && pt.Leased && pt.Kind == types.LeaseWrite {
								targetType = pt.Base
							}
						}
					}
					if targetType == nil {
						sa.AddError(idx.Pos(), "cannot assign to struct index without valid 'index_mut' method returning a mutable lease")
						return
					}
				} else {
					typeName := "unknown"
					if actualType != nil {
						typeName = actualType.Name()
					}
					sa.AddError(idx.Pos(), "cannot index into non-collection type: %s", typeName)
					return
				}
			}
		} else if sel, ok := n.Left.(*ast.SelectorExpression); ok {
			sa.Analyze(sel.Left)
			if !sa.checkWritePermission(sel) {
				sa.AddError(sel.Pos(), "cannot assign to field through a read-only lease pointer")
				return
			}
			leftType := sa.SemanticInfo.Types[sel.Left]

			// Auto-dereference for pointers (recursive)
			for {
				if pt, ok := leftType.(*types.PointerType); ok {
					leftType = pt.Base
				} else {
					break
				}
			}

			if st, ok := leftType.(*types.StructType); ok {
				if sel.Field == nil {
					return
				}
				fieldType, exists := st.Fields[sel.Field.Value]
				if !exists {
					sa.AddError(sel.Field.Pos(), "struct '%s' has no field '%s'", st.Name(), sel.Field.Value)
					return
				}
				targetType = fieldType

				// Check write permission of base object
				if ident, ok := sel.Left.(*ast.Identifier); ok {
					sym := sa.SemanticInfo.Uses[ident]
					if sym != nil && !sym.WritePerm {
						sa.AddError(sel.Pos(), "cannot assign to field of read-only object '%s'", ident.Value)
						return
					}
				}
			} else {
				sa.AddError(sel.Pos(), "cannot assign to field of non-struct type %s", leftType.Name())
				return
			}
		} else {
			sa.AddError(n.Left.Pos(), "illegal assignment target")
			return
		}

		// 3. Type Compatibility
		rhsType := sa.SemanticInfo.Types[n.Value]
		sa.checkImplicitMoveLoad(n.Value, targetType)
		if !types.IsAssignable(targetType, rhsType) {
			sa.checkInterfaceCompatibility(n.Value, targetType)
		}

		if ident, ok := n.Left.(*ast.Identifier); ok {
			if sym, exists := sa.CurrentScope.Resolve(ident.Value); exists {
				sym.Version++
			}
		}

		// --- Nora: Implicit Move Tracking ---
		if id, ok := n.Value.(*ast.Identifier); ok {
			if sym := sa.SemanticInfo.Uses[id]; sym != nil {
				// Kill if it's an owned type and we are moving it (LeaseMove or not a lease)
				if types.IsOwnedType(sym.Type) && (sym.LeaseKind == types.LeaseMove || !sym.Type.IsLeased()) && sym.LeaseKind != types.LeaseRead && sym.LeaseKind != types.LeaseWrite {
					sa.SemanticInfo.Kill(sym, n)
				}
			}
		}
		sa.SemanticInfo.Types[n] = types.Void

	case *ast.IndexExpression:
		sa.Analyze(n.Left)
		if len(n.Indices) > 0 {
			sa.Analyze(n.Indices[0])

			// [NEW] BCE Range Analysis check
			if ident, ok := n.Indices[0].(*ast.Identifier); ok {
				if sym, exists := sa.CurrentScope.Resolve(ident.Value); exists {
					if bounds := sa.CurrentScope.GetBounds(sym); bounds != nil && bounds.MinBound >= 0 {
						syntheticLen := &ast.CallExpression{
							Function:  &ast.Identifier{Value: "len"},
							Arguments: []*ast.ArgumentsExpression{{Value: n.Left}},
						}
						if IsSemanticallyEquivalent(bounds.MaxSymbol, syntheticLen) {
							n.NoBoundsCheck = true
						}
					}
				}
			}
		}

		leftType := sa.SemanticInfo.Types[n.Left]
		actualType := sa.unwrapToCollection(leftType)

		// Handle Slicing: arr[start:end]
		if len(n.Indices) == 1 {
			if _, ok := n.Indices[0].(*ast.SliceExpression); ok {
				isArr := false
				if _, ok := actualType.(*types.ListType); ok {
					isArr = true
				} else if pt, ok := actualType.(*types.PointerType); ok && pt.IsArray {
					isArr = true
				} else if actualType != nil && actualType.Name() == "str" {
					isArr = true
				}

				if isArr {
					sa.SemanticInfo.Types[n] = actualType
					return
				} else {
					typeName := "unknown"
					if actualType != nil {
						typeName = actualType.Name()
					}
					sa.AddError(n.Pos(), "cannot slice non-array type %s", typeName)
					sa.SemanticInfo.Types[n] = types.ErrorType
					return
				}
			}
		}
		if lt, ok := actualType.(*types.ListType); ok {
			if len(n.Indices) > 0 {
				idxExpr := n.Indices[0]
				idxType := sa.SemanticInfo.Types[idxExpr]
				if idxType != nil && idxType != types.ErrorType && !isIntegerType(idxType) {
					sa.AddError(idxExpr.Pos(), "index must be an integer (got %s)", idxType.Name())
				}
			}
			sa.SemanticInfo.Types[n] = lt.ElementType
			return
		} else if mt, ok := actualType.(*types.MapType); ok {
			if len(n.Indices) > 0 {
				idxExpr := n.Indices[0]
				idxType := sa.SemanticInfo.Types[idxExpr]
				if idxType != nil && idxType != types.ErrorType && !types.IsAssignable(mt.Key, idxType) {
					sa.AddError(idxExpr.Pos(), "cannot use type %s as map key of type %s", idxType.Name(), mt.Key.Name())
				}
			}
			sa.SemanticInfo.Types[n] = mt.Value
			return
		} else if actualType != nil && actualType.Name() == "str" {
			if len(n.Indices) > 0 {
				idxExpr := n.Indices[0]
				idxType := sa.SemanticInfo.Types[idxExpr]
				if idxType != nil && idxType != types.ErrorType && !isIntegerType(idxType) {
					sa.AddError(idxExpr.Pos(), "index must be an integer (got %s)", idxType.Name())
				}
			}
			// Strings are indexable in Nora (returns byte/i8)
			sa.SemanticInfo.Types[n] = types.I8
			return
		} else if ft, ok := actualType.(*types.FunctionType); ok && ft.Return != nil {
			// Generic Variant Instantiation: Some[i32]
			// Generic Variant Instantiation: Some[i32]
			if st, ok := ft.Return.(*types.SumType); ok && len(st.TypeParams) > 0 {
				// We reuse resolveTypeNode logic for simplicity or just call it
				specType := sa.resolveTypeNode(n)
				sa.SemanticInfo.Types[n] = specType
				return
			}
			sa.SemanticInfo.Types[n] = actualType // Fallback
			return
		}

		baseActualType := actualType
		if pt, ok := actualType.(*types.PointerType); ok && !pt.IsArray {
			baseActualType = pt.Base
		}

		if st, ok := baseActualType.(*types.SumType); ok && len(st.TypeParams) > 0 {
			// Generic Type instantiation: None[i32]
			specType := sa.resolveTypeNode(n)
			sa.SemanticInfo.Types[n] = specType
			return
		} else if st, ok := baseActualType.(*types.StructType); ok {
			if len(st.TypeParams) > 0 {
				// Generic Type instantiation: Node[i32]
				specType := sa.resolveTypeNode(n)
				sa.SemanticInfo.Types[n] = specType
				return
			} else if methodType, exists := st.Methods["index"]; exists {
				if ft, ok := methodType.(*types.FunctionType); ok && len(ft.Params) == 1 {
					if len(n.Indices) > 0 {
						idxExpr := n.Indices[0]
						idxType := sa.SemanticInfo.Types[idxExpr]
						expectedKeyType := ft.Params[0]
						if pt, ok := expectedKeyType.(*types.PointerType); ok && pt.Leased {
							expectedKeyType = pt.Base
						}
						if !types.IsAssignable(expectedKeyType, idxType) {
							sa.AddError(idxExpr.Pos(), "cannot use type %s as index key of type %s", idxType.Name(), expectedKeyType.Name())
						}
					}
					sa.SemanticInfo.Types[n] = ft.Return
					return
				}
			}
		} else if pt, ok := actualType.(*types.PointerType); ok {
			if len(n.Indices) > 0 {
				idxExpr := n.Indices[0]
				idxType := sa.SemanticInfo.Types[idxExpr]
				if idxType != nil && idxType != types.ErrorType && !isIntegerType(idxType) {
					sa.AddError(idxExpr.Pos(), "index must be an integer (got %s)", idxType.Name())
				}
			}
			sa.SemanticInfo.Types[n] = pt.Base
			return
		}

		if actualType != nil {
			sa.AddError(n.Left.Pos(), "cannot index into non-collection type: %s", actualType.Name())
		}
		sa.SemanticInfo.Types[n] = types.ErrorType

		// --- EXPRESSIONS ---
	case *ast.Identifier:
		// This is called for 'io' in 'io.Print'
		sym, exists := sa.CurrentScope.Resolve(n.Value)
		if !exists {
			sa.AddError(n.Token.Position, "undefined identifier: %s", n.Value)
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}
		// This fills the map your test is checking!
		sa.SemanticInfo.Uses[n] = sym
		sa.SemanticInfo.Types[n] = sym.Type

		// Definite Initialization Check
		if sym.Kind == SymVar {
			isGlobal := false
			if sym.DefScope != nil && (sym.DefScope.Kind == ScopePackage || sym.DefScope.Kind == "global") {
				isGlobal = true
			}
			if !isGlobal {
				state, tracked := sa.InitStates[sym]
				if !tracked || state == Uninitialized || state == PartiallyInitialized {
					sa.AddError(n.Token.Position, "use of possibly uninitialized variable '%s'", n.Value)
					sa.SemanticInfo.Types[n] = types.ErrorType
					return
				}
			}
		}

		// --- FIX: THE MISSING LIVENESS CHECK ---
		// If the variable was previously "Moved" (killed), we cannot use it again.
		if sa.SemanticInfo.IsDead(sym) {
			killer := sa.SemanticInfo.DeadSyms[sym]
			msg := fmt.Sprintf("use of moved value '%s'", n.Value)
			notes := []string{}
			if killer != nil {
				notes = append(notes, fmt.Sprintf("value moved here at %s:%d:%d", killer.Pos().Filename, killer.Pos().Line, killer.Pos().Column))
			}
			sa.ReportErrorWithNotes(n.Token.Position, msg, notes)
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

	case *ast.CallExpression:
		// SPECIAL CASE: Type cast (e.g. i64(x), MyType(y), my_pkg.MyType(z), MyGenericType[T](w))
		if castType := sa.tryResolveAsType(n.Function); castType != nil && castType != types.ErrorType {
			if len(n.Arguments) != 1 {
				sa.AddError(n.Token.Position, "type cast expects exactly 1 argument, got %d", len(n.Arguments))
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}
			sa.Analyze(n.Arguments[0].Value)
			argType := sa.SemanticInfo.Types[n.Arguments[0].Value]
			if argType == nil || argType == types.ErrorType {
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}
			if st, ok := castType.(*types.StructType); ok && len(st.TypeParams) > 0 && len(n.TypeArguments) > 0 {
				argTypes := []types.NRType{}
				for _, ta := range n.TypeArguments {
					argTypes = append(argTypes, sa.resolveTypeNode(ta))
				}
				castType = sa.specializeStructType(st, argTypes, n.Pos())
			} else if st, ok := castType.(*types.SumType); ok && len(st.TypeParams) > 0 && len(n.TypeArguments) > 0 {
				argTypes := []types.NRType{}
				for _, ta := range n.TypeArguments {
					argTypes = append(argTypes, sa.resolveTypeNode(ta))
				}
				castType = sa.specializeSumType(st, argTypes, n.Pos())
			}
			if prim, isPrim := castType.(*types.PrimitiveType); isPrim {
				unwrapped := types.UnwrapLease(argType)
				if _, isProto := unwrapped.(*types.ProtocolType); isProto {
					kind := types.LeaseRead
					if pt, ok := argType.(*types.PointerType); ok {
						if pt.Kind == types.LeaseWrite || pt.Kind == types.LeaseMove {
							kind = types.LeaseWrite
						}
					} else if !argType.IsLeased() {
						kind = types.LeaseWrite
					}
					sa.SemanticInfo.Types[n] = &types.PointerType{
						Base:   prim,
						Leased: true,
						Kind:   kind,
					}
					return
				}

				// Ensure argument is also a primitive or pointer-like type that can be casted
				_, isArgPrim := argType.(*types.PrimitiveType)
				_, isArgPtr := argType.(*types.PointerType)
				_, isArgFn := argType.(*types.FunctionType)

				isValidFnCast := isArgFn && prim.Name() == "ptr"

				if !isArgPrim && !isArgPtr && !isValidFnCast {
					sa.AddError(n.Arguments[0].Pos(), "cannot cast %s to %s", argType.Name(), prim.Name())
					sa.SemanticInfo.Types[n] = types.ErrorType
					return
				}
				sa.SemanticInfo.Types[n] = prim
				return
			} else if structType, isStruct := castType.(*types.StructType); isStruct {
				unwrapped := types.UnwrapLease(argType)
				if _, isProto := unwrapped.(*types.ProtocolType); isProto {
					kind := types.LeaseRead
					if pt, ok := argType.(*types.PointerType); ok {
						if pt.Kind == types.LeaseWrite || pt.Kind == types.LeaseMove {
							kind = types.LeaseWrite
						}
					} else if !argType.IsLeased() {
						kind = types.LeaseWrite
					}
					sa.SemanticInfo.Types[n] = &types.PointerType{
						Base:   structType,
						Leased: true,
						Kind:   kind,
					}
					return
				}
			} else if gt, isGeneric := castType.(*types.GenericType); isGeneric {
				unwrapped := types.UnwrapLease(argType)
				if _, isProto := unwrapped.(*types.ProtocolType); isProto {
					kind := types.LeaseRead
					if pt, ok := argType.(*types.PointerType); ok {
						if pt.Kind == types.LeaseWrite || pt.Kind == types.LeaseMove {
							kind = types.LeaseWrite
						}
					} else if !argType.IsLeased() {
						kind = types.LeaseWrite
					}
					sa.SemanticInfo.Types[n] = &types.PointerType{
						Base:   gt,
						Leased: true,
						Kind:   kind,
					}
					return
				}
			}
		}

		// SPECIAL CASE: panic()
		if ident, ok := n.Function.(*ast.Identifier); ok && ident.Value == "panic" {
			if len(n.Arguments) != 1 {
				sa.AddError(n.Token.Position, "panic() expects 1 argument")
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}
			sa.Analyze(n.Arguments[0].Value)
			argType := sa.SemanticInfo.Types[n.Arguments[0].Value]
			if !types.Equals(argType, types.String) {
				sa.AddError(n.Arguments[0].Pos(), "panic() expects a string argument, got %s", argType.Name())
			}
			sa.SemanticInfo.Types[n] = types.Void
			return
		}

		// SPECIAL CASE: len()
		if ident, ok := n.Function.(*ast.Identifier); ok && ident.Value == "len" {
			if len(n.Arguments) != 1 {
				sa.AddError(n.Token.Position, "len() expects 1 argument")
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}
			sa.Analyze(n.Arguments[0].Value)
			argType := sa.SemanticInfo.Types[n.Arguments[0].Value]
			actualType := sa.unwrapToCollection(argType)

			isIterable := false
			if _, ok := actualType.(*types.ListType); ok {
				isIterable = true
			} else if pt, ok := actualType.(*types.PointerType); ok && pt.IsArray {
				isIterable = true
			} else if actualType != nil && actualType.Name() == "str" {
				isIterable = true
			}

			if isIterable {
				sa.SemanticInfo.Types[n] = types.I32
				return
			}
			sa.AddError(n.Arguments[0].Pos(), "len() expects an array or collection")
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		// 1. Analyze the Function Expression
		sa.Analyze(n.Function)

		// 2. Check if calling a Generic function or Variant
		var fnStmt *ast.FunctionStatement
		if ident, ok := n.Function.(*ast.Identifier); ok {
			sym, exists := sa.CurrentScope.Resolve(ident.Value)
			if exists && sym.Kind == SymFunc {
				var fs *ast.FunctionStatement
				if f, ok := sym.DefNode.(*ast.FunctionStatement); ok {
					fs = f
				} else if ext, ok := sym.DefNode.(*ast.ExternStatement); ok {
					fs = ext.Function
				}
				if fs != nil && len(fs.TypeParameters) > 0 {
					fnStmt = fs
				}
			}
		} else if sel, ok := n.Function.(*ast.SelectorExpression); ok {
			if sym, ok := sa.SemanticInfo.Uses[sel.Field]; ok && sym.Kind == SymFunc {
				var fs *ast.FunctionStatement
				if f, ok := sym.DefNode.(*ast.FunctionStatement); ok {
					fs = f
				} else if ext, ok := sym.DefNode.(*ast.ExternStatement); ok {
					fs = ext.Function
				}
				if fs != nil {
					if len(fs.TypeParameters) > 0 {
						fnStmt = fs
					} else {
						// Even if not generic itself, it might be a method of a specialized struct
						leftType := sa.SemanticInfo.Types[sel.Left]
						if st, ok := leftType.(*types.StructType); ok {
							// Find the specialized method type
							if methodSyms, ok := sa.SemanticInfo.MethodSymbols[st]; ok {
								if methodSym, ok := methodSyms[sel.Field.Value]; ok {
									sa.SemanticInfo.Types[n.Function] = methodSym.Type
								}
							}
						} else if sumT, ok := leftType.(*types.SumType); ok {
							// Find the specialized method type for SumType
							if methodSyms, ok := sa.SemanticInfo.MethodSymbols[sumT]; ok {
								if methodSym, ok := methodSyms[sel.Field.Value]; ok {
									sa.SemanticInfo.Types[n.Function] = methodSym.Type
								}
							}
						}
					}
				}
			}
		}

		if fnStmt != nil {
			sa.handleGenericCall(n, fnStmt)
			return
		}

		fnType := sa.SemanticInfo.Types[n.Function]

		// Handle Generic Variant Constructor (e.g. Result.Ok(10) or Result.Ok[i32, ErrorCode](10))
		if ft, ok := fnType.(*types.FunctionType); ok {
			if st, ok := ft.Return.(*types.SumType); ok && len(st.TypeParams) > 0 {
				var specializedTypes []types.NRType

				if len(n.TypeArguments) > 0 {
					// Explicit: Result.Ok[i32, ErrorCode](10)
					for _, ta := range n.TypeArguments {
						specializedTypes = append(specializedTypes, sa.resolveTypeNode(ta))
					}
				} else if len(n.Arguments) > 0 {
					// Inference: Result.Ok(10) -> T = i32
					sa.Analyze(n.Arguments[0].Value)
					argType := sa.SemanticInfo.Types[n.Arguments[0].Value]
					specializedTypes = []types.NRType{argType}
				}

				if len(specializedTypes) > 0 {
					specFn := sa.specializeFunctionType(ft, st, specializedTypes)
					sa.SemanticInfo.Types[n.Function] = specFn
					sa.SemanticInfo.Types[n] = specFn.Return
					sa.verifyCallArguments(n, specFn)
					return
				}
			}
		}

		// Anti-Cascading Checks
		if fnType == types.ErrorType {
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}
		if fnType == nil {
			sa.AddError(n.Function.Pos(), "cannot call expression of unknown type")
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		// 2. Ensure it is actually a Function
		if ident, ok := n.Function.(*ast.Identifier); ok && ident.Value == "make" {
			sa.analyzeMakeCall(n)
			return
		}
		if ident, ok := n.Function.(*ast.Identifier); ok && ident.Value == "append" {
			sa.analyzeAppendCall(n)
			return
		}

		// Check for Compiler Intrinsics (Builtins) like print, println, scan
		var isBuiltin bool
		var fnSym *Symbol
		if ident, ok := n.Function.(*ast.Identifier); ok {
			fnSym = sa.SemanticInfo.Uses[ident]
		} else if sel, ok := n.Function.(*ast.SelectorExpression); ok {
			fnSym = sa.SemanticInfo.Uses[sel.Field]
		}

		if fnSym != nil && fnSym.Kind == SymFunc {
			if fs, ok := fnSym.DefNode.(*ast.FunctionStatement); ok {
				// Functions with [macro] or [builtin] are treated as variadic by the analyzer
				isBuiltin = ast.GetAttribute(fs.Attributes, "macro") != nil || ast.GetAttribute(fs.Attributes, "builtin") != nil
			}
		}

		if isBuiltin {
			for _, arg := range n.Arguments {
				sa.Analyze(arg.Value)
			}
			sa.SemanticInfo.Types[n] = types.Void
			return
		}

		unwrappedFnType := types.UnwrapLease(fnType)
		functionType, isFunc := unwrappedFnType.(*types.FunctionType)
		if !isFunc {
			sa.AddError(n.Function.Pos(), "cannot call non-function type '%s'", fnType.Name())
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		sa.verifyCallArguments(n, functionType)

		// 5. Set Result Type
		if functionType.Return != nil {
			sa.SemanticInfo.Types[n] = functionType.Return
		} else {
			sa.SemanticInfo.Types[n] = types.Void
		}

		if sa.CurrentFunction != nil && fnSym != nil && fnSym.Kind == SymFunc {
			if fs, ok := fnSym.DefNode.(*ast.FunctionStatement); ok {
				if sa.callGraph[sa.CurrentFunction] == nil {
					sa.callGraph[sa.CurrentFunction] = make(map[*ast.FunctionStatement]bool)
				}
				sa.callGraph[sa.CurrentFunction][fs] = true
			}
		}

	case *ast.SpawnExpression:
		if n.Call == nil {
			sa.AddError(n.Token.Position, "spawn requires a function call")
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		if n.MonitorChannel != nil {
			sa.Analyze(n.MonitorChannel)
			monitorType := sa.SemanticInfo.Types[n.MonitorChannel]
			// We expect a chan[str]
			if !types.Equals(monitorType, &types.ChanType{Elem: types.String}) {
				sa.AddError(n.MonitorChannel.Pos(), "spawn monitor channel must be of type chan[str], got %s", monitorType.Name())
			}
		}

		// Create a Spawn scope to indicate we are crossing a fiber boundary
		prevScope := sa.CurrentScope
		sa.CurrentScope = NewScope(prevScope, ScopeSpawn)
		sa.SemanticInfo.Scopes[n] = sa.CurrentScope

		sa.inSpawn++
		sa.Analyze(n.Call)
		sa.inSpawn--

		if sa.hasRestrictedClosure(sa.SemanticInfo.Types[n.Call.Function], nil) {
			sa.AddError(n.Call.Function.Pos(), "cannot spawn a closure that captures a local lease")
		}

		sa.CurrentScope = prevScope

		var spawnedFnSym *Symbol
		if ident, ok := n.Call.Function.(*ast.Identifier); ok {
			spawnedFnSym = sa.SemanticInfo.Uses[ident]
		} else if sel, ok := n.Call.Function.(*ast.SelectorExpression); ok {
			spawnedFnSym = sa.SemanticInfo.Uses[sel.Field]
		}
		if spawnedFnSym != nil && spawnedFnSym.Kind == SymFunc {
			if fs, ok := spawnedFnSym.DefNode.(*ast.FunctionStatement); ok {
				sa.spawnedFunctions[fs] = true
			}
		}

		// Enforce Lease Rules for Fibers
		// No sharing across fibers. Disallow read leases (#T) and mutable leases (&T)
		fnType := sa.SemanticInfo.Types[n.Call.Function]
		if ft, ok := fnType.(*types.FunctionType); ok {
			for i, lease := range ft.ParamLeases {
				if i < len(n.Call.Arguments) {
					argPos := n.Call.Arguments[i].Pos()
					argType := sa.SemanticInfo.Types[n.Call.Arguments[i].Value]

					// EXEMPTION: Channels, nursery contexts, sync primitives, and SharedData are completely safe to pass by read-only lease (#)
					isExempt := false
					if argType != nil {
						baseType := argType
						for {
							if pt, ok := baseType.(*types.PointerType); ok {
								baseType = pt.Base
							} else {
								break
							}
						}
						if argType.GetKind() == types.KindChan || argType.GetKind() == types.KindFunction {
							isExempt = true
						} else if st, ok := baseType.(*types.StructType); ok && st.IsShared {
							isExempt = true
						} else if proto, ok := baseType.(*types.ProtocolType); ok && proto.IsShared {
							isExempt = true
						} else if pt, ok := argType.(*types.PointerType); ok {
							if pt.Base.GetKind() == types.KindChan || pt.Base.GetKind() == types.KindFunction {
								isExempt = true
							}
						}
					}

					if !sa.hasUnsafeAttr() {
						if lease == types.LeaseWrite && ft.Params[i] != nil && ft.Params[i].IsLeased() {
							sa.AddError(argPos, "cannot pass mutable lease (write) across fiber boundary")
						} else if lease == types.LeaseRead && ft.Params[i] != nil && ft.Params[i].IsLeased() {
							if !isExempt {
								sa.AddError(argPos, "cannot pass read lease across fiber boundary")
							}
						}
					}
				}
			}
			if ft.IsMethod {
				isReceiverLeased := false
				if ft.Receiver != nil {
					isReceiverLeased = ft.Receiver.IsLeased()
				}

				isExempt := false
				if ft.Receiver != nil {
					baseType := ft.Receiver
					for {
						if pt, ok := baseType.(*types.PointerType); ok {
							baseType = pt.Base
						} else {
							break
						}
					}
					if ft.Receiver.GetKind() == types.KindChan || ft.Receiver.GetKind() == types.KindFunction {
						isExempt = true
					} else if st, ok := baseType.(*types.StructType); ok && st.IsShared {
						isExempt = true
					} else if proto, ok := baseType.(*types.ProtocolType); ok && proto.IsShared {
						isExempt = true
					} else if pt, ok := ft.Receiver.(*types.PointerType); ok {
						if pt.Base.GetKind() == types.KindChan {
							isExempt = true
						}
					}
				}

				if !sa.hasUnsafeAttr() {
					if ft.ReceiverLease == types.LeaseWrite && isReceiverLeased {
						sa.AddError(n.Call.Function.Pos(), "cannot pass mutable lease (write) across fiber boundary")
					} else if ft.ReceiverLease == types.LeaseRead && isReceiverLeased {
						if !isExempt {
							sa.AddError(n.Call.Function.Pos(), "cannot call method on read lease across fiber boundary")
						}
					}
				}
			}
		}

		sa.SemanticInfo.Types[n] = types.Fiber

	// --- DOT NOTATION (e.g., math.Pi or user.id) ---
	case *ast.DeferStatement:
		sa.Analyze(n.Call)
		sa.SemanticInfo.Types[n] = types.Void

	case *ast.AllocExpression:
		sa.analyzeAllocExpression(n)

	case *ast.SliceExpression:
		if n.Start != nil {
			sa.Analyze(n.Start)
		}
		if n.End != nil {
			sa.Analyze(n.End)
		}
		sa.SemanticInfo.Types[n] = types.Void

	case *ast.ScopeExpression:
		sa.inScope++
		sa.Analyze(n.Body)
		sa.inScope--
	case *ast.ParallelExpression:
		sa.inParallel++
		sa.Analyze(n.Body)
		sa.inParallel--

		isResult := false
		var errType types.NRType

		for _, stmt := range n.Body.Statements {
			stmtType := sa.SemanticInfo.Types[stmt]
			if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
				stmtType = sa.SemanticInfo.Types[exprStmt.Expression]
			}

			if stmtType != nil {
				if st, ok := stmtType.(*types.SumType); ok && st.CoreIntrinsic == "Result" {
					isResult = true
					if errVar, exists := st.Variants["Err"]; exists {
						for _, fType := range errVar.Fields {
							errType = fType
							break
						}
					}
				}
			}
		}

		var resultSumType *types.SumType
		if sym, exists := sa.CurrentScope.Resolve("Result"); exists {
			if st, ok := sym.Type.(*types.SumType); ok {
				resultSumType = st
			}
		}

		if isResult && resultSumType != nil {
			if errType == nil {
				errType = types.String
			}
			specialized := sa.specializeSumType(resultSumType, []types.NRType{types.Bool, errType}, n.Pos())
			sa.SemanticInfo.Types[n] = specialized
		} else {
			sa.SemanticInfo.Types[n] = types.Void
		}

	case *ast.TryExpression:
		sa.Analyze(n.Value)
		valType := sa.SemanticInfo.Types[n.Value]

		st, ok := valType.(*types.SumType)
		if !ok {
			sa.AddError(n.Pos(), "cannot use '?' operator on non-SumType %s", valType.Name())
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		isResult := st.CoreIntrinsic == "Result"
		isOption := st.CoreIntrinsic == "Option"

		if !isResult && !isOption {
			sa.AddError(n.Pos(), "cannot use '?' operator on type %s (must be Result or Option)", st.TypeName)
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		// 2. Resolve Unwrapped Value Type (Ok or Some)
		var okValueType types.NRType = types.Void
		variantName := "Ok"
		if isOption {
			variantName = "Some"
		}

		variant, exists := st.Variants[variantName]
		if !exists {
			sa.AddError(n.Pos(), "%s type missing '%s' variant", st.TypeName, variantName)
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}
		// Take the first field as the "value"
		for _, ft := range variant.Fields {
			okValueType = ft
			break
		}

		// 3. Verify function return type context
		if sa.CurrentFunction == nil {
			sa.AddError(n.Pos(), "'?' operator can only be used inside a function")
			sa.SemanticInfo.Types[n] = okValueType
			return
		}

		fnSym := sa.SemanticInfo.Defs[sa.CurrentFunction.Name]
		if fnSym != nil {
			if ft, ok := fnSym.Type.(*types.FunctionType); ok {
				retSum, ok := ft.Return.(*types.SumType)
				if !ok {
					sa.AddError(n.Pos(), "function must return Result or Option to use '?' operator")
				} else {
					if isResult && retSum.CoreIntrinsic != "Result" {
						sa.AddError(n.Pos(), "cannot use '?' on Result in a function returning %s", retSum.TypeName)
					} else if isOption && retSum.CoreIntrinsic != "Option" {
						sa.AddError(n.Pos(), "cannot use '?' on Option in a function returning %s", retSum.TypeName)
					}

					// Verify error type matching for Result
					if isResult && retSum.CoreIntrinsic == "Result" {
						errVar := st.Variants["Err"]
						retErrVar := retSum.Variants["Err"]
						if errVar != nil && retErrVar != nil {
							errT := errVar.Fields["Error"]
							retErrT := retErrVar.Fields["Error"]
							if errT != nil && retErrT != nil && !types.Equals(errT, retErrT) {
								sa.AddError(n.Pos(), "error type mismatch: cannot bubble up %s to function returning %s", errT.Name(), retErrT.Name())
							}
						}
					}
				}
			}
		}

		sa.SemanticInfo.Types[n] = okValueType

	case *ast.SelectorExpression:
		sa.Analyze(n.Left) // Resolves 'io' and fills sa.SemanticInfo.Uses[n.Left]

		leftType := sa.SemanticInfo.Types[n.Left]
		if leftType == nil || leftType == types.ErrorType {
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}

		// SPECIAL CASE: unchecked_get and unchecked_set for arrays/slices
		if n.Field != nil && (n.Field.Value == "unchecked_get" || n.Field.Value == "unchecked_set") {
			actual := sa.unwrapToCollection(leftType)
			if lt, ok := actual.(*types.ListType); ok {
				if n.Field.Value == "unchecked_get" {
					sa.SemanticInfo.Types[n] = &types.FunctionType{Params: []types.NRType{types.I32}, Return: lt.ElementType}
				} else {
					sa.SemanticInfo.Types[n] = &types.FunctionType{Params: []types.NRType{types.I32, lt.ElementType}, Return: types.Void}
				}
				return
			} else if pt, ok := actual.(*types.PointerType); ok && pt.IsArray {
				if n.Field.Value == "unchecked_get" {
					sa.SemanticInfo.Types[n] = &types.FunctionType{Params: []types.NRType{types.I32}, Return: pt.Base}
				} else {
					sa.SemanticInfo.Types[n] = &types.FunctionType{Params: []types.NRType{types.I32, pt.Base}, Return: types.Void}
				}
				return
			}
		}

		// Auto-dereference for pointers (recursive)
		for {
			if pt, ok := leftType.(*types.PointerType); ok {
				leftType = pt.Base
			} else {
				break
			}
		}
		sa.ensureMethodsSpecialized(leftType)
		sa.SemanticInfo.Types[n] = types.ErrorType // Default
		if n.Field == nil {
			return
		}

		switch t := leftType.(type) {
		case *ModuleType:
			// Look up the member (e.g., "Print") in the package's exported scope
			sym, exists := t.Exports.Resolve(n.Field.Value)
			if !exists {
				sa.AddError(n.Field.Pos(), "package '%s' has no member '%s'", t.Name(), n.Field.Value)
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}

			// --- ENFORCE VISIBILITY ---
			if sym.Visible != Public {
				sa.AddError(n.Field.Pos(), "member '%s' of package '%s' is private", n.Field.Value, t.Name())
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}

			// Fill Uses for the field (e.g., 'Print')
			sa.SemanticInfo.Uses[n.Field] = sym
			sa.SemanticInfo.Types[n] = sym.Type

		// CASE B: Struct Field Access (e.g., user.id)
		case *types.StructType:
			// 1. Look up field
			fieldType, exists := t.Fields[n.Field.Value]
			if exists {
				sa.SemanticInfo.Types[n] = fieldType
				base := t
				if t.BaseType != nil {
					base = t.BaseType
				}
				if syms, ok := sa.SemanticInfo.FieldSymbols[base]; ok {
					if sym, ok := syms[n.Field.Value]; ok {
						sa.SemanticInfo.Uses[n.Field] = sym
					}
				}
				return
			}

			// 2. Look up method
			methodType, exists := t.Methods[n.Field.Value]
			if !exists && t.BaseType != nil {
				methodType, exists = t.BaseType.Methods[n.Field.Value]
			}

			if exists {
				sa.SemanticInfo.Types[n] = methodType

				// Find symbol (may be on base type)
				if syms, ok := sa.SemanticInfo.MethodSymbols[t]; ok {
					if sym, ok := syms[n.Field.Value]; ok && sym != nil {
						sa.SemanticInfo.Uses[n.Field] = sym
						// Use the specialized type if available (e.g. Node_i32.insert_front)
						sa.SemanticInfo.Types[n] = sym.Type
					}
				}
				if sa.SemanticInfo.Uses[n.Field] == nil && t.BaseType != nil {
					if syms, ok := sa.SemanticInfo.MethodSymbols[t.BaseType]; ok {
						if sym, ok := syms[n.Field.Value]; ok && sym != nil {
							sa.SemanticInfo.Uses[n.Field] = sym
							sa.SemanticInfo.Types[n] = sym.Type
						}
					}
				}
				return
			}

			keys := make([]string, 0, len(t.Methods))
			for k := range t.Methods {
				keys = append(keys, k)
			}
			sa.AddError(n.Field.Pos(), "struct '%s' (ptr: %p) has no field or method '%s'. Methods: %v", t.Name(), t, n.Field.Value, keys)
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		// CASE C: Error handling
		case *types.SumType:
			// 1. Check if it's a method call
			methodType, exists := t.Methods[n.Field.Value]
			if !exists && t.BaseType != nil {
				methodType, exists = t.BaseType.Methods[n.Field.Value]
			}

			if exists {
				sa.SemanticInfo.Types[n] = methodType

				if syms, ok := sa.SemanticInfo.MethodSymbols[t]; ok {
					if sym, ok := syms[n.Field.Value]; ok && sym != nil {
						sa.SemanticInfo.Uses[n.Field] = sym
						sa.SemanticInfo.Types[n] = sym.Type
					}
				}
				if sa.SemanticInfo.Uses[n.Field] == nil && t.BaseType != nil {
					if syms, ok := sa.SemanticInfo.MethodSymbols[t.BaseType]; ok {
						if sym, ok := syms[n.Field.Value]; ok && sym != nil {
							sa.SemanticInfo.Uses[n.Field] = sym
							sa.SemanticInfo.Types[n] = sym.Type
						}
					}
				}
				return
			}

			if t.Name() == "Result_JsonValue_str" || strings.HasPrefix(t.Name(), "Result") {
				// fmt.Printf("[DEBUG] sum type method lookup for %s: method = %s, BaseType = %p (%s), Methods = %v\n", t.Name(), n.Field.Value, t.BaseType, t.BaseType.Name(), t.Methods)
				if t.BaseType != nil {
					if sa.DebugMode {
						fmt.Printf("  BaseType Methods: %v\n", t.BaseType.Methods)
					}
					if syms, ok := sa.SemanticInfo.MethodSymbols[t.BaseType]; ok {
						if sa.DebugMode {
							fmt.Printf("  BaseType MethodSymbols exist! keys: ")
							for k := range syms {
								fmt.Printf("%s ", k)
							}
							fmt.Println()
						}
					} else {
						if sa.DebugMode {
							fmt.Printf("  BaseType MethodSymbols do NOT exist in sa.SemanticInfo.MethodSymbols for pointer %p\n", t.BaseType)
						}
					}
				}
			}
			variant, exists := t.Variants[n.Field.Value]
			if !exists {
				sa.AddError(n.Field.Pos(), "sum type %s has no variant or method %s", t.Name(), n.Field.Value)
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}
			// Return a constructor function type
			params := []types.NRType{}
			fieldNames := sa.sortedFieldNames(variant.Fields)
			for _, fName := range fieldNames {
				params = append(params, variant.Fields[fName])
			}
			if len(params) == 0 {
				sa.SemanticInfo.Types[n] = t
			} else {
				sa.SemanticInfo.Types[n] = &types.FunctionType{
					Params: params,
					Return: t,
				}
			}
			if sym, ok := sa.CurrentScope.Resolve(n.Field.Value); ok {
				sa.SemanticInfo.Uses[n.Field] = sym
			}
		case *types.ProtocolType:
			method, exists := t.Methods[n.Field.Value]
			if exists {
				sa.SemanticInfo.Types[n] = method
				if syms, ok := sa.SemanticInfo.MethodSymbols[t]; ok {
					if sym, ok := syms[n.Field.Value]; ok {
						sa.SemanticInfo.Uses[n.Field] = sym
					}
				}
				return
			}
			sa.AddError(n.Field.Pos(), "interface '%s' has no method '%s'", t.Name(), n.Field.Value)
			sa.SemanticInfo.Types[n] = types.ErrorType

		case *types.GenericType:
			// Resolve constraint (must be a protocol)
			if t.Constraint != nil {
				// Unwrap potential leases/pointers from the constraint (e.g. #Printable)
				constraint := types.UnwrapLease(t.Constraint)
				if proto, ok := constraint.(*types.ProtocolType); ok {
					method, exists := proto.Methods[n.Field.Value]
					if exists {
						sa.SemanticInfo.Types[n] = method
						if syms, ok := sa.SemanticInfo.MethodSymbols[proto]; ok {
							if sym, ok := syms[n.Field.Value]; ok {
								sa.SemanticInfo.Uses[n.Field] = sym
							}
						}
					}
					return
				}
			}
			sa.AddError(n.Field.Pos(), "type parameter '%s' has no field or method '%s'", t.TypeParam, n.Field.Value)
			sa.SemanticInfo.Types[n] = types.ErrorType

		case *types.PrimitiveType:
			methodType, exists := t.Methods[n.Field.Value]
			if exists {
				sa.SemanticInfo.Types[n] = methodType
				if syms, ok := sa.SemanticInfo.MethodSymbols[t]; ok {
					if sym, ok := syms[n.Field.Value]; ok && sym != nil {
						sa.SemanticInfo.Uses[n.Field] = sym
						sa.SemanticInfo.Types[n] = sym.Type
					}
				}
				return
			}
			sa.AddError(n.Field.Pos(), "primitive type '%s' has no method '%s'", t.Name(), n.Field.Value)
			sa.SemanticInfo.Types[n] = types.ErrorType

		default:
			if leftType.GetKind() == types.KindChan && n.Field.Value == "clone" {
				sa.SemanticInfo.Types[n] = &types.FunctionType{
					Params: []types.NRType{},
					Return: leftType,
				}
				return
			}

			sa.AddError(n.Token.Position, "type '%s' does not support field access", t.Name())
			sa.SemanticInfo.Types[n] = types.ErrorType
		}

	case *ast.SelectStatement:
		for _, sc := range n.Cases {
			if sc.Condition != nil {
				sa.Analyze(sc.Condition)
				// Verification: Must be a channel operation (Send or Recv)
				// For now, we trust the user or add detailed checks later.
			}
			sa.Analyze(sc.Body)
		}

	case *ast.ForStatement:
		prevScope := sa.CurrentScope
		sa.CurrentScope = NewScope(prevScope, ScopeLoop)
		sa.SemanticInfo.Scopes[n] = sa.CurrentScope

		// Save initialization states before loop
		origInit := make(map[*Symbol]InitState)
		for k, v := range sa.InitStates {
			origInit[k] = v
		}

		// Handle For-In loops
		if n.Value != nil {
			sa.Analyze(n.Iterable)
			iterType := sa.SemanticInfo.Types[n.Iterable]
			actualType := sa.unwrapToCollection(iterType)

			var keyType types.NRType = types.I32 // Default key type (for array indices)

			var valType types.NRType = types.ErrorType

			switch t := actualType.(type) {
			case *types.ListType:
				valType = t.ElementType
			case *types.MapType:
				keyType = t.Key
				valType = t.Value
			case *types.PointerType:
				if t.IsArray {
					valType = t.Base
				} else {
					sa.AddError(n.Iterable.Pos(), "cannot iterate over non-array pointer type %s", actualType.Name())
				}
			case *types.StructType:
				if n.Key != nil {
					sa.AddError(n.Iterable.Pos(), "iterator loops cannot have both key and value variables, just one")
				}
				sa.ensureMethodsSpecialized(t)
				if method, exists := t.Methods["Next"]; exists {
					// Create fake AST nodes for the implicit Next() call so monomorphization works
					selExpr := &ast.SelectorExpression{
						Token: n.Token,
						Left:  n.Iterable,
						Field: &ast.Identifier{Token: n.Token, Value: "Next"},
					}
					callExpr := &ast.CallExpression{
						Token:     n.Token,
						Function:  selExpr,
						Arguments: []*ast.ArgumentsExpression{},
					}
					n.NextCall = callExpr
					sa.Analyze(callExpr)

					if fnType, ok := method.(*types.FunctionType); ok && len(fnType.Params) == 0 {
						if sumType, ok := fnType.Return.(*types.SumType); ok {
							baseName := sumType.Name()
							if sumType.BaseType != nil {
								baseName = sumType.BaseType.Name()
							}
							if baseName == "Option" && len(sumType.TypeArgs) == 1 {
								valType = sumType.TypeArgs[0]
							} else {
								sa.AddError(n.Iterable.Pos(), "iterator Next() method must return Option[T]")
							}
						} else {
							sa.AddError(n.Iterable.Pos(), "iterator Next() method must return Option[T]")
						}
					} else {
						sa.AddError(n.Iterable.Pos(), "iterator Next() method must have no arguments (besides receiver), found %d params", len(method.(*types.FunctionType).Params))
					}
				} else {
					sa.AddError(n.Iterable.Pos(), "cannot iterate over struct type %s without a Next() method", actualType.Name())
				}
			default:
				if actualType != nil && actualType.Name() == "str" {
					valType = types.I8
				} else {
					sa.AddError(n.Iterable.Pos(), "cannot iterate over type %s", actualType.Name())
				}
			}

			if n.Key != nil {
				sym, _ := sa.CurrentScope.Define(n.Key.Value, keyType, SymVar, n.Key)
				sa.SemanticInfo.Defs[n.Key] = sym
				sa.InitStates[sym] = Initialized // Loop variables are initialized
				// BCE for for-in loop key: key goes from 0 to len(iterable)
				sa.CurrentScope.Bounds[sym] = &VarBounds{
					MinBound:  0,
					MaxSymbol: n.Iterable,
				}
			}
			if n.Value != nil {
				sym, _ := sa.CurrentScope.Define(n.Value.Value, valType, SymVar, n.Value)
				sa.SemanticInfo.Defs[n.Value] = sym
				sa.InitStates[sym] = Initialized

				if rangeExpr, isRange := n.Iterable.(*ast.RangeExpression); isRange {
					sa.CurrentScope.Bounds[sym] = &VarBounds{
						MinBound:  0, // Just an approximation, we can't easily resolve ast.Expression to int at compile time yet
						MaxSymbol: rangeExpr.End,
					}
				}
			}

			sa.Analyze(n.Body)
		} else {
			// Infinite loop
			sa.Analyze(n.Body)
		}

		sa.CurrentScope = prevScope

		// Merge: loop body might execute 0 times
		for sym := range origInit {
			if origInit[sym] == Uninitialized && sa.InitStates[sym] == Initialized {
				sa.InitStates[sym] = PartiallyInitialized
			} else if origInit[sym] == Uninitialized && sa.InitStates[sym] == PartiallyInitialized {
				sa.InitStates[sym] = PartiallyInitialized
			} else {
				sa.InitStates[sym] = origInit[sym]
			}
		}

	case *ast.WhileStatement:
		origInit := make(map[*Symbol]InitState)
		for k, v := range sa.InitStates {
			origInit[k] = v
		}

		sa.Analyze(n.Condition)

		// BCE tracking for while loop
		if infix, ok := n.Condition.(*ast.InfixExpression); ok && infix.Operator == "<" {
			if id, ok := infix.Left.(*ast.Identifier); ok {
				sym := sa.SemanticInfo.Uses[id]
				if sym == nil {
					sym = sa.SemanticInfo.Defs[id]
				}
				if sym != nil {
					sa.CurrentScope.Bounds[sym] = &VarBounds{
						MinBound:  0, // Assuming starting from 0 or safely within bounds
						MaxSymbol: infix.Right,
					}
				}
			}
		}

		sa.Analyze(n.Body)

		// Merge: loop body might execute 0 times
		for sym := range origInit {
			if origInit[sym] == Uninitialized && sa.InitStates[sym] == Initialized {
				sa.InitStates[sym] = PartiallyInitialized
			} else if origInit[sym] == Uninitialized && sa.InitStates[sym] == PartiallyInitialized {
				sa.InitStates[sym] = PartiallyInitialized
			} else {
				sa.InitStates[sym] = origInit[sym]
			}
		}

	case *ast.SendExpression:
		sa.Analyze(n.Left)
		sa.Analyze(n.Right)
		leftType := sa.SemanticInfo.Types[n.Left]

		// Unwrap leases
		actualType := leftType
		if pt, ok := leftType.(*types.PointerType); ok && pt.Leased {
			actualType = pt.Base
		}

		chType, ok := actualType.(*types.ChanType)
		if !ok {
			sa.AddError(n.Left.Pos(), "cannot send to non-channel type '%s'", leftType.Name())
			return
		}
		rightType := sa.SemanticInfo.Types[n.Right]
		if !types.IsAssignable(chType.Elem, rightType) {
			sa.AddError(n.Right.Pos(), "type mismatch: channel expects '%s', got '%s'", chType.Elem.Name(), rightType.Name())
		}
		sa.SemanticInfo.Types[n] = types.Void

	case *ast.ReceiveExpression:
		sa.Analyze(n.Value)
		valType := sa.SemanticInfo.Types[n.Value]

		// Unwrap leases
		actualType := valType
		if pt, ok := valType.(*types.PointerType); ok && pt.Leased {
			actualType = pt.Base
		}

		chType, ok := actualType.(*types.ChanType)
		if !ok {
			sa.AddError(n.Value.Pos(), "cannot receive from non-channel type '%s'", valType.Name())
			sa.SemanticInfo.Types[n] = types.ErrorType
			return
		}
		sa.SemanticInfo.Types[n] = chType.Elem
	}
}

func (sa *SemanticAnalyzer) AnalyzeLambdaExpression(n *ast.LambdaExpression) {
	var capturesLease bool
	if n.Body != nil {
		origInitStates := sa.InitStates
		origDeadSyms := sa.SemanticInfo.DeadSyms

		sa.SemanticInfo.DeadSyms = make(map[*Symbol]ast.Node)
		sa.InitStates = make(map[*Symbol]InitState)
		for k, v := range origInitStates {
			sa.InitStates[k] = v
		}

		prevScope := sa.CurrentScope
		fnScope := NewScope(prevScope, ScopeClosure) // Mark as closure to capture variables
		sa.CurrentScope = fnScope
		sa.SemanticInfo.Scopes[n] = fnScope

		prevLambda := sa.CurrentLambda
		sa.CurrentLambda = n

		for _, p := range n.Parameters {
			pType := sa.resolveTypeNode(p.Type)
			sym, err := sa.CurrentScope.Define(p.Name.Value, pType, SymParam, p)
			if err != nil {
				sa.AddError(p.Name.Pos(), "%s", err.Error())
			}
			sa.SemanticInfo.Defs[p.Name] = sym
			if sym != nil {
				sa.InitStates[sym] = Initialized

				lease := types.LeaseRead
				if pref, ok := p.Type.(*ast.PrefixExpression); ok {
					switch pref.Operator {
					case "&":
						lease = types.LeaseWrite
					case "@":
						lease = types.LeaseMove
					case "#":
						lease = types.LeaseRead
					}
				}
				sym.LeaseKind = lease
				if lease == types.LeaseWrite || lease == types.LeaseMove {
					sym.WritePerm = true
				}
			}
		}

		sa.Analyze(n.Body)

		for sym := range fnScope.Captures {
			if sa.hasLease(sym.Type, nil) {
				capturesLease = true
				break
			}
		}

		sa.InitStates = origInitStates
		sa.SemanticInfo.DeadSyms = origDeadSyms
		sa.CurrentScope = prevScope
		sa.CurrentLambda = prevLambda
	}

	// Resolve the type
	fnType := &types.FunctionType{
		Params:        []types.NRType{},
		ParamLeases:   []types.LeaseKind{},
		CapturesLease: capturesLease,
	}
	for _, p := range n.Parameters {
		fnType.Params = append(fnType.Params, sa.resolveTypeNode(p.Type))

		lease := types.LeaseRead
		if pref, ok := p.Type.(*ast.PrefixExpression); ok {
			switch pref.Operator {
			case "#":
				lease = types.LeaseWrite
			case "@":
				lease = types.LeaseMove
			}
		}
		fnType.ParamLeases = append(fnType.ParamLeases, lease)
	}

	if n.ReturnType != nil {
		fnType.Return = sa.resolveTypeNode(n.ReturnType)
	} else {
		fnType.Return = types.Void // TODO: type inference if omitted
	}
	sa.SemanticInfo.Types[n] = fnType
}

func (sa *SemanticAnalyzer) AnalyzeFunctionStatement(n *ast.FunctionStatement) {
	if n.IsExport && len(n.TypeParameters) > 0 {
		sa.AddError(n.Token.Position, "cannot export generic functions")
	}

	if n.IsExport && n.Receiver != nil {
		sa.AddError(n.Token.Position, "cannot export methods (functions with receivers)")
	}

	if n.Body != nil {
		// Backup caller's initialization state and liveness state
		origInitStates := sa.InitStates
		origDeadSyms := sa.SemanticInfo.DeadSyms

		// --- Nora: Fresh liveness and initialization tracking for this function ---
		sa.SemanticInfo.DeadSyms = make(map[*Symbol]ast.Node)
		sa.InitStates = make(map[*Symbol]InitState)

		// 1. Create a fresh scope for the function body
		prevScope := sa.CurrentScope
		fnScope := NewScope(prevScope, ScopeFunction)
		sa.CurrentScope = fnScope
		sa.SemanticInfo.Scopes[n] = fnScope

		// 1.2. Define Type Parameters in scope (as GenericTypes)
		if len(n.TypeParameters) > 0 {
			for _, tp := range n.TypeParameters {
				var constraint types.NRType = types.Any
				if tp.Constraint != nil {
					constraint = sa.resolveTypeNode(tp.Constraint)
				}
				gt := &types.GenericType{TypeParam: tp.Name.Value, Constraint: constraint}
				sym, _ := sa.CurrentScope.Define(tp.Name.Value, gt, SymType, tp)
				if sym != nil {
					sa.SemanticInfo.Defs[tp.Name] = sym
				}
			}
		}

		// 1.3. Extract implicit generic type parameters from the receiver
		if n.Receiver != nil {
			var baseExp ast.Expression = n.Receiver.Type
			for {
				if pref, ok := baseExp.(*ast.PrefixExpression); ok {
					baseExp = pref.Right
				} else {
					break
				}
			}
			if idxExpr, ok := baseExp.(*ast.IndexExpression); ok {
				for _, idx := range idxExpr.Indices {
					if ident, ok := idx.(*ast.Identifier); ok {
						// Check if it's already defined (e.g. i32)
						if _, exists := sa.CurrentScope.Resolve(ident.Value); !exists {
							if sa.DebugMode {
								fmt.Println("Found implicit generic parameter:", ident.Value)
							}
							// It's not defined, so it must be an implicit generic parameter!
							gt := &types.GenericType{TypeParam: ident.Value, Constraint: types.Any}
							sym, _ := sa.CurrentScope.Define(ident.Value, gt, SymType, ident)
							if sym != nil {
								sa.SemanticInfo.Defs[ident] = sym
							}
						}
					}
				}
			}
		}

		// 1.5. Define Receiver in scope (if any)
		if n.Receiver != nil {
			rType := sa.resolveTypeNode(n.Receiver.Type)
			sym, err := sa.CurrentScope.Define(n.Receiver.Name.Value, rType, SymParam, n.Receiver)
			if err != nil {
				sa.AddError(n.Receiver.Name.Pos(), "%s", err.Error())
			}
			sa.SemanticInfo.Defs[n.Receiver.Name] = sym
			if sym != nil {
				sa.InitStates[sym] = Initialized
				sym.LeaseKind = types.LeaseKind(n.Receiver.LeaseKind)
				if sym.LeaseKind == types.LeaseWrite || sym.LeaseKind == types.LeaseMove {
					sym.WritePerm = true
				}
			}
		}

		for _, p := range n.Parameters {
			pType := sa.resolveTypeNode(p.Type)
			sym, err := sa.CurrentScope.Define(p.Name.Value, pType, SymParam, p)
			if err != nil {
				sa.AddError(p.Name.Pos(), "%s", err.Error())
			}
			sa.SemanticInfo.Defs[p.Name] = sym
			if sym != nil {
				sa.InitStates[sym] = Initialized

				// Extract lease from type prefix if present
				lease := types.LeaseRead
				if pref, ok := p.Type.(*ast.PrefixExpression); ok {
					switch pref.Operator {
					case "#":
						lease = types.LeaseRead
					case "&":
						lease = types.LeaseWrite
					case "@":
						lease = types.LeaseMove
					}
				} else if pType != nil && pType.GetKind() != types.KindFunction && (pType.GetKind() == types.KindGeneric || types.IsOwnedType(pType)) {
					lease = types.LeaseMove
				}
				sym.LeaseKind = lease
				if sym.LeaseKind == types.LeaseWrite || sym.LeaseKind == types.LeaseMove {
					sym.WritePerm = true
				}
			}
		}

		// 3. Analyze Body
		oldFn := sa.CurrentFunction
		sa.CurrentFunction = n
		sa.Analyze(n.Body)

		// Check if the last statement of a void function is a discarded owned lease
		retType := types.NRType(types.Void)
		if n.ReturnType != nil {
			retType = sa.resolveTypeNode(n.ReturnType)
		}
		if retType == types.Void || retType == nil {
			if n.Body != nil && len(n.Body.Statements) > 0 {
				lastStmt := n.Body.Statements[len(n.Body.Statements)-1]
				if exprStmt, ok := lastStmt.(*ast.ExpressionStatement); ok {
					if lastType, ok := sa.SemanticInfo.Types[lastStmt]; ok {
						if sa.containsOwnedLease(lastType, nil) {
							sa.AddError(exprStmt.Pos(), "cannot discard owned value of type '%s'. it must be assigned, moved, or passed to a function to avoid memory leaks", lastType.Name())
						}
					}
				}
			}
		}

		sa.CurrentFunction = oldFn

		// 4. Restore Scope
		sa.checkUnusedSymbolsInScope(sa.CurrentScope)
		sa.CurrentScope = prevScope

		// Restore caller's initialization state and liveness state
		sa.InitStates = origInitStates
		sa.SemanticInfo.DeadSyms = origDeadSyms
	} else {
		// Bodiless function/method: resolve signature in Pass 2 to cache type nodes
		sa.resolveFunctionType(n)
	}
}

func (sa *SemanticAnalyzer) terminates(node ast.Node) bool {
	if node == nil {
		return false
	}
	switch n := node.(type) {
	case *ast.ReturnStatement:
		return true
	case *ast.BlockStatement:
		if len(n.Statements) == 0 {
			return false
		}
		// A block terminates if its last statement terminates
		return sa.terminates(n.Statements[len(n.Statements)-1])
	case *ast.IfExpression:
		// If both branches terminate, the whole IF terminates
		if n.Alternative == nil {
			return false
		}
		return sa.terminates(n.Consequence) && sa.terminates(n.Alternative)
	case *ast.ExpressionStatement:
		return sa.terminates(n.Expression)
	}
	return false
}

func (sa *SemanticAnalyzer) tryResolveAsType(n ast.Expression) types.NRType {
	switch t := n.(type) {
	case *ast.Identifier:
		if prim, ok := types.LookupPrimitive(t.Value); ok {
			return prim
		}
		sym, exists := sa.CurrentScope.Resolve(t.Value)
		if exists {
			// fmt.Printf("[DEBUG tryResolveAsType] resolved identifier %s: kind=%d (SymType=%d), type=%T\n", t.Value, sym.Kind, SymType, sym.Type)
			if sym.Kind == SymType || sym.Kind == SymVariant {
				return sym.Type
			}
		} else {
			// fmt.Printf("[DEBUG tryResolveAsType] failed to resolve identifier %s\n", t.Value)
		}
	case *ast.SelectorExpression:
		if ident, ok := t.Left.(*ast.Identifier); ok {
			sym, exists := sa.CurrentScope.Resolve(ident.Value)
			if exists && sym.Kind == SymPackage {
				if modType, ok := sym.Type.(*ModuleType); ok {
					if pkgSym, exists := modType.Exports.Resolve(t.Field.Value); exists && (pkgSym.Kind == SymType || pkgSym.Kind == SymVariant) {
						return pkgSym.Type
					}
				}
			}
		}
	case *ast.IndexExpression:
		baseType := sa.tryResolveAsType(t.Left)
		if baseType != nil && baseType != types.ErrorType {
			if tn, ok := n.(ast.TypeNode); ok {
				return sa.resolveTypeNode(tn)
			}
		}
	}
	return nil
}

func (sa *SemanticAnalyzer) resolveTypeNode(n ast.TypeNode) types.NRType {
	if n == nil {
		return types.Void
	}

	if sa.SemanticInfo.Types[n] != nil {
		return sa.SemanticInfo.Types[n]
	}

	if sa.Context != nil && sa.Context.Err() != nil {
		return types.ErrorType
	}
	if ast.IsNil(n) {
		return types.Void
	}

	switch t := n.(type) {

	case *ast.GroupedExpression:
		tn, ok := t.Expression.(ast.TypeNode)
		if !ok {
			sa.AddError(t.Pos(), "invalid type syntax in grouped expression: %T", t.Expression)
			return types.ErrorType
		}
		res := sa.resolveTypeNode(tn)
		sa.SemanticInfo.Types[n] = res
		return res

	// Case 1: Simple Name (e.g., "i32", "User")
	case *ast.Identifier:
		// A. Check if it's a built-in primitive or collection marker
		if prim, ok := types.LookupPrimitive(t.Value); ok {
			sa.SemanticInfo.Types[n] = prim
			return prim
		}
		if t.Value == "Map" || t.Value == "List" {
			if sym, exists := sa.CurrentScope.Resolve(t.Value); exists && (sym.Kind == SymType || sym.Kind == SymVariant) {
				// Let standard scope resolution handle it
			} else {
				return &types.PrimitiveType{KindName: t.Value}
			}
		}

		// A.2. Check if it's a specialized type in SpecTypes
		if specType, ok := sa.SemanticInfo.SpecTypes[t.Value]; ok {
			sa.SemanticInfo.Types[n] = specType
			return specType
		}

		sym, exists := sa.CurrentScope.Resolve(t.Value)
		if !exists {
			sa.AddError(t.Pos(), "undefined type: '%s'", t.Value)
			return types.ErrorType
		}

		// C. Ensure the symbol we found is actually a TYPE or VARIANT
		if sym.Kind != SymType && sym.Kind != SymVariant {
			sa.AddError(t.Pos(), "'%s' is not a type or variant (it is a %s)", t.Value, sym.Kind)
			return types.ErrorType
		}

		res := sym.Type
		if res == nil || res == types.ErrorType {
			// fmt.Printf("[DEBUG] resolveTypeNode: identifier %s resolved to <error>\n", t.Value)
		}

		if res == types.Void && sym.DefNode != nil {
			res = sa.forceAnalyzeType(sym)
			if res == types.ErrorType {
				sa.AddError(t.Pos(), "circular type dependency detected for '%s'", t.Value)
				return types.ErrorType
			}
		}

		// [FIX] Don't cache GenericType if it might be re-resolved with constraints later
		if _, ok := res.(*types.GenericType); ok {
			sa.SemanticInfo.Uses[t] = sym
			return res
		}

		sa.SemanticInfo.Uses[t] = sym
		sa.SemanticInfo.Types[n] = res
		return res

	case *ast.ChanType:
		inner := sa.resolveTypeNode(t.Value)
		res := &types.ChanType{Elem: inner}
		sa.SemanticInfo.Types[n] = res
		return res

	case *ast.PrefixExpression:
		// Leases (# and @) are now treated as PointerType
		// they are ALWAYS pointers internally in the C generator.
		if t.Operator == "#" || t.Operator == "&" || t.Operator == "@" {
			tn, ok := t.Right.(ast.TypeNode)
			if !ok {
				panic(fmt.Sprintf("resolveTypeNode PrefixExpression Right is NOT a TypeNode! operator: %s, type: %T, value: %v", t.Operator, t.Right, t.Right))
			}
			inner := sa.resolveTypeNode(tn)
			kind := types.LeaseRead // default for '#'
			if t.Operator == "&" {
				kind = types.LeaseWrite
			}
			if t.Operator == "@" {
				kind = types.LeaseMove
				// Validate: You cannot own a borrow or an already owned type
				if pt, ok := inner.(*types.PointerType); ok && pt.Leased {
					sa.AddError(t.Pos(), "invalid lease combination: cannot own a borrowed or already owned type")
					sa.SemanticInfo.Types[n] = types.ErrorType
					return types.ErrorType
				}
			}
			res := &types.PointerType{Base: inner, Leased: true, Kind: kind}
			if pt, ok := inner.(*types.PointerType); ok && pt.IsArray {
				res.IsArray = true
				res.Base = pt.Base // Pass through element type for array leases
			}
			sa.SemanticInfo.Types[n] = res
			return res
		}

		sa.AddError(t.Pos(), "invalid type operator: %s", t.Operator)
		return types.ErrorType

	case *ast.IndexExpression:
		// Generic type instantiation: Option[i32] or Array type: []i32
		baseType := sa.resolveTypeNode(t.Left.(ast.TypeNode))
		if len(t.Indices) == 0 {
			// [FIX] T[] is a ListType in Nora
			res := &types.ListType{ElementType: baseType}
			sa.SemanticInfo.Types[n] = res
			return res
		}

		// Build arg types from all indices
		argTypes := []types.NRType{}
		for _, idx := range t.Indices {
			if tn, ok := idx.(ast.TypeNode); ok {
				argTypes = append(argTypes, sa.resolveTypeNode(tn))
			}
		}
		// If no type arguments were found, but there are indices, it might be an array size like i32[10]
		if len(argTypes) == 0 && len(t.Indices) == 0 {
			return types.ErrorType
		}

		// Case 0: Built-in Collection types (Map[K, V], List[T])
		if ident, ok := t.Left.(*ast.Identifier); ok {
			if ident.Value == "Map" && len(argTypes) == 2 {
				res := &types.MapType{Key: argTypes[0], Value: argTypes[1]}
				sa.SemanticInfo.Types[n] = res
				return res
			}
			if ident.Value == "List" && len(argTypes) == 1 {
				res := &types.ListType{ElementType: argTypes[0]}
				sa.SemanticInfo.Types[n] = res
				return res
			}
		}

		if len(t.Indices) > 0 {
			// Check if the first index is a type node or an expression
			if _, ok := t.Indices[0].(ast.TypeNode); !ok {
				// It's an expression (like [10]), so this is an array type T[n]
				res := &types.PointerType{Base: baseType, IsArray: true}
				sa.SemanticInfo.Types[n] = res
				return res
			}
		}

		// Case A: Base is a SumType (e.g. Option[i32])
		if st, ok := baseType.(*types.SumType); ok && len(st.TypeParams) > 0 {
			specialized := sa.specializeSumType(st, argTypes, n.Pos())
			sa.SemanticInfo.Types[n] = specialized
			return specialized
		}

		// Case B: Base is a StructType (e.g. Box[i32])
		if st, ok := baseType.(*types.StructType); ok && len(st.TypeParams) > 0 {
			specialized := sa.specializeStructType(st, argTypes, n.Pos())
			if specialized == nil {
				return types.ErrorType
			}
			sa.SemanticInfo.Types[n] = specialized
			return specialized
		}

		// Case C: Base is a ProtocolType (e.g. Iterator[i32])
		if pt, ok := baseType.(*types.ProtocolType); ok && len(pt.TypeParams) > 0 {
			specialized := sa.specializeProtocolType(pt, argTypes, n.Pos())
			if specialized == nil {
				return types.ErrorType
			}
			sa.SemanticInfo.Types[n] = specialized
			return specialized
		}

		// Case B: Base is a FunctionType (e.g. Some[i32])
		if ft, ok := baseType.(*types.FunctionType); ok {
			if st, ok := ft.Return.(*types.SumType); ok && len(st.TypeParams) > 0 {
				// Specialize the Return type first
				specRet := sa.specializeSumType(st, argTypes, n.Pos())

				// Specialize the Parameters
				newParams := []types.NRType{}
				// Substitute params using map based on type params
				subs := make(map[string]types.NRType)
				for i, tp := range st.TypeParams {
					if i < len(argTypes) {
						subs[tp.Name] = argTypes[i]
					}
				}
				for _, p := range ft.Params {
					newParams = append(newParams, sa.substituteType(p, subs))
				}

				specFn := &types.FunctionType{
					Params: newParams,
					ParamLeases: ft.ParamLeases,
					Return: specRet,
				}
				sa.SemanticInfo.Types[n] = specFn
				return specFn
			}
		}

		return types.ErrorType

	case *ast.SelectorExpression:
		sa.Analyze(t)
		res := sa.SemanticInfo.Types[t]

		if res == types.Void {
			if sym, ok := sa.SemanticInfo.Uses[t.Field]; ok && sym.DefNode != nil {
				res = sa.forceAnalyzeType(sym)
				if res == types.ErrorType {
					sa.AddError(t.Field.Pos(), "circular type dependency detected for '%s'", t.Field.Value)
					return types.ErrorType
				}
				sa.SemanticInfo.Types[t] = res
			}
		}

		if id, ok := t.Left.(*ast.Identifier); ok && id.Value == "hash" {
			if sa.DebugMode {
				println("[DEBUG resolveTypeNode] SelectorExpression:", id.Value, ".", t.Field.Value)
				if res != nil {
					println("  - res kind:", res.GetKind(), "name:", res.Name(), "type:", fmt.Sprintf("%T", res))
				} else {
					println("  - res is nil!")
				}
			}
		}

		if res == nil {
			return types.ErrorType
		}
		return res

	case *ast.FunctionType:
		paramTypes := []types.NRType{}
		paramLeases := []types.LeaseKind{}
		for _, p := range t.Parameters {
			pt := sa.resolveTypeNode(p)
			paramTypes = append(paramTypes, pt)
			// Extract lease from PrefixExpression if any
			lease := types.LeaseRead
			if pref, ok := p.(*ast.PrefixExpression); ok {
				switch pref.Operator {
				case "#":
					lease = types.LeaseRead
				case "&":
					lease = types.LeaseWrite
				case "@":
					lease = types.LeaseMove
				}
			} else if pt != nil && pt.GetKind() != types.KindFunction && (pt.GetKind() == types.KindGeneric || types.IsOwnedType(pt)) {
				lease = types.LeaseMove
			}
			paramLeases = append(paramLeases, lease)
		}
		var retType types.NRType = types.Void
		if t.ReturnType != nil {
			retType = sa.resolveTypeNode(t.ReturnType)
		}
		return &types.FunctionType{
			Params:      paramTypes,
			ParamLeases: paramLeases,
			Return:      retType,
		}

	default:
		sa.AddError(n.Pos(), "invalid type syntax: %T", n)
		return types.ErrorType
	}
}

func (sa *SemanticAnalyzer) findTypeSymbol(t types.NRType) *Symbol {
	for _, sym := range sa.SemanticInfo.Defs {
		if sym == nil {
			continue
		}
		if sym.Kind == SymType && sym.Type == t {
			return sym
		}
	}
	for _, scope := range sa.PackageScopes {
		for _, sym := range scope.Symbols {
			if sym.Kind == SymType && sym.Type == t {
				return sym
			}
		}
	}
	return nil
}

func (sa *SemanticAnalyzer) specializeSumType(st *types.SumType, argTypes []types.NRType, pos token.Position) *types.SumType {
	// 1. Identity Check: if args are exactly the type params, return the original
	isIdentity := len(argTypes) == len(st.TypeParams)
	if isIdentity {
		for i, ta := range argTypes {
			if gt, ok := ta.(*types.GenericType); !ok || gt.TypeParam != st.TypeParams[i].Name {
				isIdentity = false
				break
			} else {
				baseConstraint := st.TypeParams[i].Constraint
				gtConstraint := gt.Constraint
				if baseConstraint == nil {
					baseConstraint = types.Any
				}
				if gtConstraint == nil {
					gtConstraint = types.Any
				}
				if !types.Equals(baseConstraint, gtConstraint) {
					isIdentity = false
					break
				}
			}
		}
	}

	if isIdentity {
		return st
	}

	if len(st.Variants) == 0 {
		if sym := sa.findTypeSymbol(st); sym != nil && sym.DefNode != nil {
			if ts, ok := sym.DefNode.(*ast.TypeStatement); ok {
				if !sa.analyzingTypes[sym] {
					sa.analyzingTypes[sym] = true
					prevScope := sa.CurrentScope
					if sym.DefScope != nil {
						sa.CurrentScope = sym.DefScope
					}
					sa.Analyze(ts)
					sa.CurrentScope = prevScope
					delete(sa.analyzingTypes, sym)
				}
			}
		}
	}
	pkgName := ""
	for _, sym := range sa.SemanticInfo.Defs {
		if sym == nil {
			continue
		}
		if sym.Kind == SymType && sym.Type == st {
			if sym.DefScope != nil {
				s := sym.DefScope
				for s != nil {
					if s.Kind == ScopePackage {
						pkgName = s.PackageName
						break
					}
					s = s.Parent
				}
			}
			break
		}
	}

	nameParts := []string{st.TypeName}
	subs := make(map[string]types.NRType)
	for i, arg := range argTypes {
		nameParts = append(nameParts, arg.Name())
		if i < len(st.TypeParams) {
			tp := st.TypeParams[i]
			subs[tp.Name] = arg
		}
	}

	// Pass 2: Check constraints with fully populated substitutions
	for i, arg := range argTypes {
		if i < len(st.TypeParams) {
			tp := st.TypeParams[i]
			if tp.Constraint != nil {
				constraintType := sa.substituteType(tp.Constraint, subs)
				if !sa.checkConstraint(arg, constraintType, pos) {
					return nil
				}
			}
		}
	}
	hashSuffix := types.GetHashSuffix(st.TypeName, argTypes)
	specialName := st.TypeName + "_" + hashSuffix
	if pkgName != "" && pkgName != "main" {
		safePkg := strings.ReplaceAll(pkgName, "/", "_")
		safePkg = strings.ReplaceAll(safePkg, ".", "_")
		specialName = safePkg + "_" + specialName
	}
	specialName = sanitizeCIdentifier(specialName)

	// Check if already specialized
	if existing, ok := sa.SemanticInfo.SpecTypes[specialName]; ok {
		if existingST, ok := existing.(*types.SumType); ok {
			return existingST
		}
	}

	specialized := &types.SumType{
		TypeName:      specialName,
		Variants:      make(map[string]*types.Variant),
		TypeParams:    []*types.TypeParam{},
		TypeArgs:      argTypes,
		BaseType:      st,
		Methods:       make(map[string]types.NRType),
		CoreIntrinsic: st.CoreIntrinsic,
	}

	// Register BEFORE mapping fields to break self-referential recursion.
	sa.SemanticInfo.SpecTypes[specialName] = specialized

	for name, variant := range st.Variants {
		newVariant := &types.Variant{
			Name: variant.Name, Tag: variant.Tag,
			Fields:     make(map[string]types.NRType),
			FieldNames: variant.FieldNames,
		}
		newVariant.FieldNames = append([]string(nil), variant.FieldNames...)
		for fName, fType := range variant.Fields {
			newVariant.Fields[fName] = sa.substituteType(fType, subs)
		}
		specialized.Variants[name] = newVariant
	}

	// Monomorphize methods for SumType
	hasGenericArgs := false
	for _, arg := range argTypes {
		if sa.hasGeneric(arg) {
			hasGenericArgs = true
			break
		}
	}

	for mName, mType := range st.Methods {
		specialized.Methods[mName] = mType
		if !hasGenericArgs {
			if methodSyms, ok := sa.SemanticInfo.MethodSymbols[st]; ok {
				if methodSym, ok := methodSyms[mName]; ok {
					if methodFn, ok := methodSym.DefNode.(*ast.FunctionStatement); ok {
						if len(methodFn.TypeParameters) == len(argTypes) {
							if mName == "drop" || mName == "eq" {
								sa.Monomorphize(methodFn, argTypes, nil, specialized)
							}
						}
					}
				}
			}
		}
	}
	return specialized
}

func (sa *SemanticAnalyzer) specializeStructType(base *types.StructType, argTypes []types.NRType, pos token.Position) types.NRType {
	if len(base.Fields) == 0 {
		if sym := sa.findTypeSymbol(base); sym != nil && sym.DefNode != nil {
			if ts, ok := sym.DefNode.(*ast.TypeStatement); ok {
				if !sa.analyzingTypes[sym] {
					sa.analyzingTypes[sym] = true
					prevScope := sa.CurrentScope
					if sym.DefScope != nil {
						sa.CurrentScope = sym.DefScope
					}
					sa.Analyze(ts)
					sa.CurrentScope = prevScope
					delete(sa.analyzingTypes, sym)
				}
			}
		}
	}
	// 1. Identity Check: if args are exactly the type params, return the original
	isIdentity := len(argTypes) == len(base.TypeParams)
	if isIdentity {
		for i, ta := range argTypes {
			if gt, ok := ta.(*types.GenericType); !ok || gt.TypeParam != base.TypeParams[i].Name {
				if sa.DebugMode {
					fmt.Printf("[DEBUG identity] failed because ta is not gt or name mismatch: ta=%#v, name=%s\n", ta, base.TypeParams[i].Name)
				}
				isIdentity = false
				break
			} else {
				// Must also check if constraints are compatible.
				baseConstraint := base.TypeParams[i].Constraint
				gtConstraint := gt.Constraint
				if baseConstraint == nil {
					baseConstraint = types.Any
				}
				if gtConstraint == nil {
					gtConstraint = types.Any
				}
				if !types.Equals(baseConstraint, gtConstraint) {
					if gtConstraint == types.Any || gtConstraint == nil {
						gt.Constraint = baseConstraint
					} else {
						if sa.DebugMode {
							fmt.Printf("[DEBUG identity] failed because constraint mismatch: baseName=%s, gtName=%s\n", baseConstraint.Name(), gtConstraint.Name())
						}
						isIdentity = false
						break
					}
				}
			}
		}
	}
	if isIdentity {
		return base
	}

	pkgName := ""
	for _, sym := range sa.SemanticInfo.Defs {
		if sym == nil {
			continue
		}
		if sym.Kind == SymType && sym.Type == base {
			if sym.DefScope != nil {
				s := sym.DefScope
				for s != nil {
					if s.Kind == ScopePackage {
						pkgName = s.PackageName
						break
					}
					s = s.Parent
				}
			}
			break
		}
	}

	nameParts := []string{base.TypeName}
	subs := make(map[string]types.NRType)
	for i, arg := range argTypes {
		nameParts = append(nameParts, arg.Name())
		if i < len(base.TypeParams) {
			tp := base.TypeParams[i]
			subs[tp.Name] = arg
		}
	}

	// Pass 2: Check constraints with fully populated substitutions
	for i, arg := range argTypes {
		if i < len(base.TypeParams) {
			tp := base.TypeParams[i]
			if tp.Constraint != nil {
				constraintType := sa.substituteType(tp.Constraint, subs)
				if !sa.checkConstraint(arg, constraintType, pos) {
					return types.ErrorType
				}
			}
		}
	}
	hashSuffix := types.GetHashSuffix(base.TypeName, argTypes)
	specialName := base.TypeName + "_" + hashSuffix
	if pkgName != "" && pkgName != "main" {
		safePkg := strings.ReplaceAll(pkgName, "/", "_")
		safePkg = strings.ReplaceAll(safePkg, ".", "_")
		specialName = safePkg + "_" + specialName
	}
	specialName = sanitizeCIdentifier(specialName)

	// 3. Check Cache
	if existing, ok := sa.SemanticInfo.SpecTypes[specialName]; ok {
		return existing
	}

	// 4. Create specialized struct
	specialized := types.NewStructType(specialName)
	specialized.TypeParams = []*types.TypeParam{} // No longer generic
	specialized.TypeArgs = argTypes
	specialized.BaseType = base
	specialized.IsShared = base.IsShared
	specialized.CoreIntrinsic = base.CoreIntrinsic

	// Register BEFORE mapping fields to break self-referential recursion.
	sa.SemanticInfo.SpecTypes[specialName] = specialized

	// 5. Map fields
	specialized.FieldNames = append([]string(nil), base.FieldNames...)
	for fName, fType := range base.Fields {
		specialized.Fields[fName] = sa.substituteType(fType, subs)
	}

	// 6. Specializing methods
	hasGenericArgs := false
	for _, arg := range argTypes {
		if sa.hasGeneric(arg) {
			hasGenericArgs = true
			break
		}
	}

	for mName, mType := range base.Methods {
		specialized.Methods[mName] = sa.substituteType(mType, subs)
		if !hasGenericArgs {
			if methodSyms, ok := sa.SemanticInfo.MethodSymbols[base]; ok {
				if methodSym, ok := methodSyms[mName]; ok {
					if methodFn, ok := methodSym.DefNode.(*ast.FunctionStatement); ok {
						if sa.DebugMode {
							fmt.Printf("  Monomorphizing method: %s\n", mName)
						}
						if len(methodFn.TypeParameters) <= len(argTypes) {
							if mName == "drop" || mName == "eq" {
								sa.Monomorphize(methodFn, argTypes, nil, specialized)
							}
							// other methods are DEFERRED to ensureMethodsSpecialized or explicitly called
						} else {
							if sa.DebugMode {
								fmt.Printf("  Skipping Monomorphize: len(TypeParameters) %d > len(argTypes) %d\n", len(methodFn.TypeParameters), len(argTypes))
							}
						}
					} else {
						if sa.DebugMode {
							fmt.Printf("  Method node is not FunctionStatement: %T\n", methodSym.DefNode)
						}
					}
				} else {
					if sa.DebugMode {
						fmt.Printf("  Method symbol not found in MethodSymbols\n")
					}
				}
			} else {
				if sa.DebugMode {
					fmt.Printf("  Base struct not in MethodSymbols\n")
				}
			}
		}
	}

	return specialized
}

func (sa *SemanticAnalyzer) specializeProtocolType(base *types.ProtocolType, argTypes []types.NRType, pos token.Position) *types.ProtocolType {
	if len(base.Methods) == 0 {
		if sym := sa.findTypeSymbol(base); sym != nil && sym.DefNode != nil {
			if ts, ok := sym.DefNode.(*ast.TypeStatement); ok {
				if !sa.analyzingTypes[sym] {
					sa.analyzingTypes[sym] = true
					prevScope := sa.CurrentScope
					if sym.DefScope != nil {
						sa.CurrentScope = sym.DefScope
					}
					sa.Analyze(ts)
					sa.CurrentScope = prevScope
					delete(sa.analyzingTypes, sym)
				}
			}
		}
	}
	isIdentity := len(argTypes) == len(base.TypeParams)
	if isIdentity {
		for i, ta := range argTypes {
			if gt, ok := ta.(*types.GenericType); !ok || gt.TypeParam != base.TypeParams[i].Name {
				isIdentity = false
				break
			} else {
				// Must also check if constraints are compatible.
				baseConstraint := base.TypeParams[i].Constraint
				gtConstraint := gt.Constraint
				if baseConstraint == nil {
					baseConstraint = types.Any
				}
				if gtConstraint == nil {
					gtConstraint = types.Any
				}
				if !types.Equals(baseConstraint, gtConstraint) {
					isIdentity = false
					break
				}
			}
		}
	}
	if isIdentity {
		return base
	}

	pkgName := ""
	for _, sym := range sa.SemanticInfo.Defs {
		if sym == nil {
			continue
		}
		if sym.Kind == SymType && sym.Type == base {
			if sym.DefScope != nil {
				s := sym.DefScope
				for s != nil {
					if s.Kind == ScopePackage {
						pkgName = s.PackageName
						break
					}
					s = s.Parent
				}
			}
			break
		}
	}

	nameParts := []string{base.ProtocolName}
	subs := make(map[string]types.NRType)
	for i, arg := range argTypes {
		nameParts = append(nameParts, arg.Name())
		if i < len(base.TypeParams) {
			tp := base.TypeParams[i]
			subs[tp.Name] = arg
		}
	}
	hashSuffix := types.GetHashSuffix(base.ProtocolName, argTypes)
	specialName := base.ProtocolName + "_" + hashSuffix
	if pkgName != "" && pkgName != "main" {
		safePkg := strings.ReplaceAll(pkgName, "/", "_")
		safePkg = strings.ReplaceAll(safePkg, ".", "_")
		specialName = safePkg + "_" + specialName
	}
	specialName = sanitizeCIdentifier(specialName)

	if existing, ok := sa.SemanticInfo.SpecTypes[specialName]; ok {
		if proto, ok := existing.(*types.ProtocolType); ok {
			return proto
		}
	}

	specialized := &types.ProtocolType{
		ProtocolName: specialName,
		Methods:      make(map[string]*types.FunctionType),
		TypeParams:   []*types.TypeParam{},
		TypeArgs:     argTypes,
		BaseType:     base,
	}

	sa.SemanticInfo.SpecTypes[specialName] = specialized

	for mName, mType := range base.Methods {
		specialized.Methods[mName] = sa.substituteType(mType, subs).(*types.FunctionType)
	}

	return specialized
}

func (sa *SemanticAnalyzer) substituteType(t types.NRType, subs map[string]types.NRType) types.NRType {
	if t == nil {
		return nil
	}

	switch kt := t.(type) {
	case *types.GenericType:
		if arg, ok := subs[kt.TypeParam]; ok {
			return arg
		}
		return kt
	case *types.PointerType:
		return &types.PointerType{
			Base:    sa.substituteType(kt.Base, subs),
			IsArray: kt.IsArray,
			Leased:  kt.Leased,
			Kind:    kt.Kind,
		}
	case *types.ListType:
		return &types.ListType{
			ElementType: sa.substituteType(kt.ElementType, subs),
		}
	case *types.MapType:
		return &types.MapType{
			Key:   sa.substituteType(kt.Key, subs),
			Value: sa.substituteType(kt.Value, subs),
		}
	case *types.FunctionType:
		newParams := []types.NRType{}
		for _, p := range kt.Params {
			newParams = append(newParams, sa.substituteType(p, subs))
		}
		var newReceiver types.NRType
		if kt.Receiver != nil {
			newReceiver = sa.substituteType(kt.Receiver, subs)
		}
		return &types.FunctionType{
			Params:      newParams,
			ParamLeases: kt.ParamLeases,
			Return:      sa.substituteType(kt.Return, subs),
			Receiver:    newReceiver,
			IsVariadic:  kt.IsVariadic,
		}
	case *types.StructType:
		// If this is a specialized generic struct (e.g. Node_T with BaseType=Node,
		// TypeArgs=[GenericType{T}]), try to substitute the TypeArgs and re-specialize
		// from the base. This handles self-referential generics like Node[T].next: *Node[T].
		if kt.BaseType != nil && len(kt.TypeArgs) > 0 {
			newArgs := make([]types.NRType, len(kt.TypeArgs))
			changed := false
			for i, arg := range kt.TypeArgs {
				substituted := sa.substituteType(arg, subs)
				newArgs[i] = substituted
				if !types.Equals(substituted, arg) {
					changed = true
				}
			}
			if !changed {
				return kt
			}
			return sa.specializeStructType(kt.BaseType, newArgs, token.Position{})
		} else if len(kt.TypeParams) > 0 {
			// This is the BASE generic struct (e.g. Node[T])
			// Try to substitute its TypeParams and specialize it.
			newArgs := make([]types.NRType, len(kt.TypeParams))
			changed := false
			for i, tp := range kt.TypeParams {
				if arg, ok := subs[tp.Name]; ok {
					newArgs[i] = arg
					changed = true
				} else {
					newArgs[i] = &types.GenericType{TypeParam: tp.Name}
				}
			}
			if changed {
				return sa.specializeStructType(kt, newArgs, token.Position{})
			}
		}
		return kt
	case *types.SumType:
		if kt.BaseType != nil && len(kt.TypeArgs) > 0 {
			newArgs := make([]types.NRType, len(kt.TypeArgs))
			changed := false
			for i, arg := range kt.TypeArgs {
				substituted := sa.substituteType(arg, subs)
				newArgs[i] = substituted
				if !types.Equals(substituted, arg) {
					changed = true
				}
			}
			if !changed {
				return kt
			}
			return sa.specializeSumType(kt.BaseType, newArgs, token.Position{})
		} else if len(kt.TypeParams) > 0 {
			// This is the BASE generic sum type (e.g. Option[T])
			newArgs := make([]types.NRType, len(kt.TypeParams))
			changed := false
			for i, tp := range kt.TypeParams {
				if arg, ok := subs[tp.Name]; ok {
					newArgs[i] = arg
					changed = true
				} else {
					newArgs[i] = &types.GenericType{TypeParam: tp.Name}
				}
			}
			if changed {
				return sa.specializeSumType(kt, newArgs, token.Position{})
			}
		}
		return kt
	case *types.ProtocolType:
		if kt.BaseType != nil && len(kt.TypeArgs) > 0 {
			newArgs := make([]types.NRType, len(kt.TypeArgs))
			changed := false
			for i, arg := range kt.TypeArgs {
				substituted := sa.substituteType(arg, subs)
				newArgs[i] = substituted
				if !types.Equals(substituted, arg) {
					changed = true
				}
			}
			if !changed {
				return kt
			}
			return sa.specializeProtocolType(kt.BaseType, newArgs, token.Position{})
		} else if len(kt.TypeParams) > 0 {
			newArgs := make([]types.NRType, len(kt.TypeParams))
			changed := false
			for i, tp := range kt.TypeParams {
				if arg, ok := subs[tp.Name]; ok {
					newArgs[i] = arg
					changed = true
				} else {
					newArgs[i] = &types.GenericType{TypeParam: tp.Name}
				}
			}
			if changed {
				return sa.specializeProtocolType(kt, newArgs, token.Position{})
			}
		}
		return kt
	default:
		return t
	}
}

func (sa *SemanticAnalyzer) specializeFunctionType(ft *types.FunctionType, st *types.SumType, argTypes []types.NRType) *types.FunctionType {
	specRet := sa.specializeSumType(st, argTypes, token.Position{})
	newParams := []types.NRType{}

	subs := make(map[string]types.NRType)
	for i, tp := range st.TypeParams {
		if i < len(argTypes) {
			subs[tp.Name] = argTypes[i]
		}
	}

	for _, p := range ft.Params {
		newParams = append(newParams, sa.substituteType(p, subs))
	}
	return &types.FunctionType{
		Params:      newParams,
		ParamLeases: ft.ParamLeases,
		Return:      specRet,
		IsVariadic:  ft.IsVariadic,
	}
}

func (sa *SemanticAnalyzer) resolveFunctionType(n *ast.FunctionStatement) *types.FunctionType {
	prevScope := sa.CurrentScope
	if len(n.TypeParameters) > 0 {
		sa.CurrentScope = NewScope(prevScope, ScopeBlock)
		// Pass 1: Define all TypeParams in the new scope
		for _, tp := range n.TypeParameters {
			gt := &types.GenericType{TypeParam: tp.Name.Value, Constraint: nil}
			sym, _ := sa.CurrentScope.Define(tp.Name.Value, gt, SymType, tp)
			if sym != nil {
				sa.SemanticInfo.Defs[tp.Name] = sym
			}
		}
		// Pass 2: Resolve constraints using the fully populated scope
		for _, tp := range n.TypeParameters {
			if tp.Constraint != nil {
				constraint := sa.resolveTypeNode(tp.Constraint)
				if sym, found := sa.CurrentScope.Lookup(tp.Name.Value); found {
					if gt, ok := sym.Type.(*types.GenericType); ok {
						gt.Constraint = constraint
					}
				}
			}
		}
	}
	defer func() {
		sa.CurrentScope = prevScope
	}()

	// 1. Resolve Parameters
	paramTypes := []types.NRType{}
	paramLeases := []types.LeaseKind{}
	isVariadic := false

	for _, p := range n.Parameters {
		if p.IsVariadic {
			isVariadic = true
			paramTypes = append(paramTypes, types.Void)
			paramLeases = append(paramLeases, types.LeaseRead)
			continue
		}

		resolved := sa.resolveTypeNode(p.Type)

		// Infer lease from type node
		lease := types.LeaseRead
		if p.LeaseKind != ast.LeaseRead {
			lease = types.LeaseKind(p.LeaseKind)
		} else if pref, ok := p.Type.(*ast.PrefixExpression); ok {
			if pref.Operator == "&" {
				lease = types.LeaseWrite
			} else if pref.Operator == "@" {
				lease = types.LeaseMove
			} else if pref.Operator == "#" {
				lease = types.LeaseRead
			}
		} else if resolved != nil && resolved.GetKind() != types.KindFunction && (resolved.GetKind() == types.KindGeneric || types.IsOwnedType(resolved)) {
			// Generic type parameters and owned types default to LeaseMove (ownership transfer)
			lease = types.LeaseMove
		}
		p.LeaseKind = ast.LeaseKind(lease)

		paramTypes = append(paramTypes, resolved)
		paramLeases = append(paramLeases, lease)
	}

	// 2. Resolve Return Type
	returnType := sa.resolveTypeNode(n.ReturnType)
	// if strings.Contains(n.Name.Value, "IsSome") || strings.Contains(n.Name.Value, "Len") {
	// 	fmt.Fprintf(os.Stderr, "[DEBUG resolveFunctionType] Method: %s, ReturnType AST: %T (%s), Resolved ReturnType: %s\n", n.Name.Value, n.ReturnType, n.ReturnType.String(), returnType.Name())
	// }

	// 3. Resolve Receiver
	var receiverType types.NRType
	var receiverLease types.LeaseKind
	if n.Receiver != nil {
		receiverLease = types.LeaseRead
		if pref, ok := n.Receiver.Type.(*ast.PrefixExpression); ok {
			if pref.Operator == "&" {
				receiverLease = types.LeaseWrite
			} else if pref.Operator == "@" {
				receiverLease = types.LeaseMove
			}
		}
		n.Receiver.LeaseKind = ast.LeaseKind(receiverLease)
		receiverType = sa.resolveTypeNode(n.Receiver.Type)
	}

	return &types.FunctionType{
		Params:        paramTypes,
		ParamLeases:   paramLeases,
		Return:        returnType,
		IsVariadic:    isVariadic,
		IsMethod:      receiverType != nil,
		Receiver:      receiverType,
		ReceiverLease: receiverLease,
	}
}

func (sa *SemanticAnalyzer) verifyCallArguments(n *ast.CallExpression, functionType *types.FunctionType) {
	// 3. Check Argument Count
	expectedCount := len(functionType.Params)
	if functionType.IsVariadic {
		if len(n.Arguments) < expectedCount-1 {
			sa.AddError(n.Token.Position, "wrong number of arguments: expected at least %d, got %d",
				expectedCount-1, len(n.Arguments))
		}
	} else {
		if len(n.Arguments) != expectedCount {
			sa.AddError(n.Token.Position, "wrong number of arguments: expected %d, got %d",
				expectedCount, len(n.Arguments))
		}
	}

	// 4. Analyze and Verify Arguments
	for i, arg := range n.Arguments {
		if arg == nil || arg.Value == nil {
			continue
		}
		sa.Analyze(arg.Value)
		if i >= len(functionType.Params) {
			continue
		}

		expectedType := functionType.Params[i]
		var expectedLease types.LeaseKind
		if i < len(functionType.ParamLeases) {
			expectedLease = functionType.ParamLeases[i]
		} else {
			expectedLease = types.LeaseRead
		}

		argType := sa.SemanticInfo.Types[arg.Value]
		if argType == nil {
			argType = types.ErrorType
		}

		// A. Type Check
		if functionType.IsVariadic && i >= expectedCount-1 {
			// Variadic part: allow any type
		} else {
			sa.checkInterfaceCompatibility(arg.Value, expectedType)
		}

		// B. Lease Contract Enforcement

		// DYNAMIC RESOLUTION: If the expected type itself is a leased pointer
		// (e.g. explicitly defined or instantiated from a generic), override the lease.
		if pt, ok := expectedType.(*types.PointerType); ok && pt.Leased {
			expectedLease = pt.Kind
		}

		var argSym *Symbol
		if expectedLease == types.LeaseWrite {
			argSym = sa.getRootSymbol(arg.Value)
		} else {
			if ident, ok := arg.Value.(*ast.Identifier); ok {
				argSym = sa.SemanticInfo.Uses[ident]
			} else if pref, ok := arg.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if ident, ok := pref.Right.(*ast.Identifier); ok {
					argSym = sa.SemanticInfo.Uses[ident]
				}
			}
		}

		switch expectedLease {
		case types.LeaseWrite:
			if argSym == nil {
				sa.AddError(arg.Pos(), "cannot pass a literal/temporary to a mutating parameter")
			} else {
				if !argSym.WritePerm && argSym.Kind == SymVar {
					argSym.WritePerm = true
				}
				if !argSym.WritePerm {
					sa.AddError(arg.Pos(), "current lease does not allow writing to '%s'", argSym.Name)
				}
			}
		case types.LeaseMove:
			// Special Case: Allow literals like 'none' to be passed to move-lease parameters
			if _, ok := arg.Value.(*ast.NoneLiteral); ok {
				continue
			}
			if pref, ok := arg.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if _, ok := pref.Right.(*ast.NoneLiteral); ok {
					continue
				}
				if _, ok := pref.Right.(*ast.SelectorExpression); ok {
					// Topological Solver will handle advanced field-level move tracking
					continue
				}
				if _, ok := pref.Right.(*ast.IndexExpression); ok {
					continue
				}
			}

			if argSym == nil {
				// Temporary values (R-values) can be consumed safely.
				// There is no symbol to track or kill.
			} else {
				if pt, ok := argSym.Type.(*types.PointerType); ok && pt.Leased && pt.Kind != types.LeaseMove {
					if types.IsOwnedType(types.UnwrapLease(argSym.Type)) {
						sa.AddError(arg.Pos(), "cannot move borrowed value '%s'", argSym.Name)
						return
					}
				}
				if argSym.IsPinned {
					sa.AddError(arg.Pos(), "illegal move: variable '%s' is currently pinned and cannot be consumed", argSym.Name)
					return
				}
				if sa.SemanticInfo.IsDead(argSym) {
					killer := sa.SemanticInfo.DeadSyms[argSym]
					// Robust check: the killer might be the argument itself or its parent prefix expr
					isMoveExpr := false
					if pref, ok := arg.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
						isMoveExpr = true
					}
					if killer != arg.Value && !isMoveExpr {
						sa.AddError(arg.Pos(), "use of moved value '%s'", argSym.Name)
					}
				}
				sa.SemanticInfo.Kill(argSym, n)
			}
		case types.LeaseRead:
			if argSym != nil {
				if sa.SemanticInfo.IsDead(argSym) {
					// EXCEPTION: If the argument itself is a Move prefix (@var), it's not a use-after-move
					isMoveExpr := false
					if pref, ok := arg.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
						isMoveExpr = true
					}
					if !isMoveExpr {
						sa.AddError(arg.Pos(), "use of moved value '%s'", argSym.Name)
					}
				}
			}
		}
	}
}

func (sa *SemanticAnalyzer) analyzeMakeCall(n *ast.CallExpression) {
	if len(n.Arguments) < 1 {
		sa.AddError(n.Token.Position, "make requires at least 1 argument")
		sa.SemanticInfo.Types[n] = types.ErrorType
		return
	}

	// 1. Resolve the type being 'made'
	// In Nora, we allow types as first-class expressions in 'make'
	typeExpr := n.Arguments[0].Value
	var targetType types.NRType

	if tn, ok := typeExpr.(ast.TypeNode); ok {
		targetType = sa.resolveTypeNode(tn)
	} else {
		sa.AddError(typeExpr.Pos(), "first argument to make must be a type")
		sa.SemanticInfo.Types[n] = types.ErrorType
		return
	}

	// 2. Set the return type of the call
	sa.SemanticInfo.Types[n] = targetType

	// 3. Handle specific 'make' logic
	switch t := targetType.(type) {
	case *types.ChanType:
		if len(n.Arguments) > 2 {
			sa.AddError(n.Token.Position, "make(chan) takes 1 or 2 arguments")
		}
		if len(n.Arguments) == 2 {
			sa.Analyze(n.Arguments[1].Value)
			capType := sa.SemanticInfo.Types[n.Arguments[1].Value]
			if !types.Equals(capType, types.Int) && !types.Equals(capType, types.I32) {
				sa.AddError(n.Arguments[1].Pos(), "channel capacity must be an integer, got %s", capType.Name())
			}
		}
	case *types.MapType:
		if len(n.Arguments) > 1 {
			sa.AddError(n.Token.Position, "make(Map) takes exactly 1 argument (the type)")
		}
	case *types.ListType:
		if len(n.Arguments) > 2 {
			sa.AddError(n.Token.Position, "make(List) takes 1 or 2 arguments")
		}
		if len(n.Arguments) == 2 {
			sa.Analyze(n.Arguments[1].Value)
			sizeType := sa.SemanticInfo.Types[n.Arguments[1].Value]
			if !types.Equals(sizeType, types.Int) && !types.Equals(sizeType, types.I32) {
				sa.AddError(n.Arguments[1].Pos(), "list size must be an integer, got %s", sizeType.Name())
			}
		}
	default:
		sa.AddError(n.Arguments[0].Pos(), "cannot use make with type %s", t.Name())
	}
}

func (sa *SemanticAnalyzer) analyzeAppendCall(n *ast.CallExpression) {
	if len(n.Arguments) != 2 {
		sa.AddError(n.Token.Position, "append requires exactly 2 arguments: list and item")
		sa.SemanticInfo.Types[n] = types.ErrorType
		return
	}

	sa.Analyze(n.Arguments[0].Value)
	listType := sa.SemanticInfo.Types[n.Arguments[0].Value]

	if lt, ok := listType.(*types.ListType); ok {
		sa.Analyze(n.Arguments[1].Value)
		itemType := sa.SemanticInfo.Types[n.Arguments[1].Value]

		if !types.IsAssignable(lt.ElementType, itemType) {
			// Check interface compatibility (Fat Pointer casting)
			sa.checkInterfaceCompatibility(n.Arguments[1].Value, lt.ElementType)
		}
		sa.SemanticInfo.Types[n] = lt

		// Create a virtual FunctionType to help the RAII solver know that append moves its arguments
		sa.SemanticInfo.Types[n.Function] = &types.FunctionType{
			Params:      []types.NRType{lt, lt.ElementType},
			ParamLeases: []types.LeaseKind{types.LeaseMove, types.LeaseMove},
			Return:      lt,
		}
	} else {
		sa.AddError(n.Arguments[0].Pos(), "first argument to append must be a list, got %s", listType.Name())
		sa.SemanticInfo.Types[n] = types.ErrorType
	}
}

func (sa *SemanticAnalyzer) handleGenericCall(n *ast.CallExpression, fnStmt *ast.FunctionStatement) {
	// pos := n.Pos()
	// fmt.Printf("[DEBUG] handleGenericCall: function = %s, typeArgsCount = %d, argsCount = %d, pos = %s:%d:%d\n", n.Function.String(), len(n.TypeArguments), len(n.Arguments), pos.Filename, pos.Line, pos.Column)
	var receiverType types.NRType
	if sel, ok := n.Function.(*ast.SelectorExpression); ok {
		receiverType = sa.SemanticInfo.Types[sel.Left]
	}

	// 1. Resolve Type Arguments
	typeArgs := []types.NRType{}
	if len(n.TypeArguments) == 0 {
		typeArgs = sa.inferTypeArguments(fnStmt, n)
		if typeArgs == nil {
			return // Inference failed
		}
	} else {
		for _, ta := range n.TypeArguments {
			typeArgs = append(typeArgs, sa.resolveTypeNode(ta))
		}
	}

	if len(typeArgs) != len(fnStmt.TypeParameters) {
		sa.AddError(n.Pos(), "generic function '%s' expects %d type arguments, got %d",
			fnStmt.Name.Value, len(fnStmt.TypeParameters), len(typeArgs))
		sa.SemanticInfo.Types[n] = types.ErrorType
		return
	}

	// Verify Constraints
	subs := make(map[string]types.NRType)
	for i, tp := range fnStmt.TypeParameters {
		if i < len(typeArgs) {
			subs[tp.Name.Value] = typeArgs[i]
		}
	}

	for i, tp := range fnStmt.TypeParameters {
		if tp.Constraint != nil {
			constraintType := sa.resolveTypeNode(tp.Constraint)
			constraintType = sa.substituteType(constraintType, subs)
			if !sa.checkConstraint(typeArgs[i], constraintType, n.Pos()) {
				sa.SemanticInfo.Types[n] = types.ErrorType
				return
			}
		}
	}

	inst := sa.Monomorphize(fnStmt, typeArgs, n, receiverType)
	if inst == nil {
		sa.SemanticInfo.Types[n] = types.ErrorType
		return
	}

	effectiveKey := "_" + types.GetHashSuffix(fnStmt.Name.Value, typeArgs)
	// if fnStmt.Name.Value == "Insert" {
	// 	var names []string
	// 	for _, ta := range typeArgs {
	// 		if ta != nil {
	// 			names = append(names, ta.Name())
	// 		} else {
	// 			names = append(names, "nil")
	// 		}
	// 	}
	// 	println("[DEBUG-MONO] Insert call: typeArgs = " + strings.Join(names, ", ") + ", key = " + effectiveKey + ", file = " + n.Pos().Filename)
	// }

	fnName := fnStmt.Name.Value + effectiveKey
	if receiverType == nil {
		pkgName := ""
		var fnSym *Symbol
		if ident, ok := n.Function.(*ast.Identifier); ok {
			if sym, ok := sa.SemanticInfo.Uses[ident]; ok {
				fnSym = sym
			} else if sym, exists := sa.CurrentScope.Resolve(ident.Value); exists {
				fnSym = sym
			}
		} else if sel, ok := n.Function.(*ast.SelectorExpression); ok {
			if sym, ok := sa.SemanticInfo.Uses[sel.Field]; ok {
				fnSym = sym
			}
		}
		if fnSym != nil && fnSym.DefScope != nil {
			s := fnSym.DefScope
			for s != nil {
				if s.Kind == ScopePackage {
					pkgName = s.PackageName
					break
				}
				s = s.Parent
			}
		}
		if pkgName != "" && pkgName != "main" {
			safePkg := strings.ReplaceAll(pkgName, "/", "_")
			safePkg = strings.ReplaceAll(safePkg, ".", "_")
			fnName = safePkg + "_" + fnName
		}
	} else {
		if sel, ok := n.Function.(*ast.SelectorExpression); ok {
			receiverType := sa.SemanticInfo.Types[sel.Left]
			if receiverType != nil {
				fnName = receiverType.Name() + "_" + fnName
			}
		}
	}
	if fnStmt.Name.Value == "nr_serialize_json" || fnStmt.Name.Value == "nr_deserialize_json" {
		fnName = sa.getMangledSerializationFuncName(fnStmt.Name.Value, typeArgs[0])
	} else {
		fnName = sanitizeCIdentifier(fnName)
	}
	sa.SemanticInfo.MonomorphizedNames[n] = fnName

	// Type of the call is the return type of the specialized function
	if sym, ok := sa.SemanticInfo.Defs[inst.Name]; ok && sym != nil {
		if ft, ok := sym.Type.(*types.FunctionType); ok {
			sa.SemanticInfo.Types[n.Function] = ft
			sa.SemanticInfo.Types[n] = ft.Return

			// --- FIX: Analyze and verify arguments against the specialized signature ---
			for _, arg := range n.Arguments {
				sa.Analyze(arg.Value)
			}
			sa.verifyCallArguments(n, ft)
			return
		}
	}
	sa.SemanticInfo.Types[n.Function] = types.ErrorType
	sa.SemanticInfo.Types[n] = types.ErrorType
}

func (sa *SemanticAnalyzer) Monomorphize(fnStmt *ast.FunctionStatement, typeArgs []types.NRType, n *ast.CallExpression, receiverType types.NRType) *ast.FunctionStatement {
	var recTypeArgs []types.NRType
	if receiverType != nil {
		underlying := receiverType
		for {
			if pt, ok := underlying.(*types.PointerType); ok {
				underlying = pt.Base
				continue
			}
			break
		}
		if st, ok := underlying.(*types.StructType); ok {
			recTypeArgs = st.TypeArgs
		} else if sum, ok := underlying.(*types.SumType); ok {
			recTypeArgs = sum.TypeArgs
		}
	}
	// 1. Generate Unique Key
	typeKey := ""
	for _, ta := range typeArgs {
		typeKey += "_" + ta.Name()
	}

	// 2. Check if instance already exists
	if sa.SemanticInfo.Instances[fnStmt] == nil {
		sa.SemanticInfo.Instances[fnStmt] = make(map[string]*ast.FunctionStatement)
	}

	if inst := sa.SemanticInfo.Instances[fnStmt][typeKey]; inst != nil {
		// Ensure it's registered in Defs even if returning from cache
		if _, ok := sa.SemanticInfo.Defs[inst.Name]; !ok {
			sa.CollectSymbols(inst)
		}

		if receiverType != nil {
			if sym, ok := sa.SemanticInfo.Defs[inst.Name]; ok && sym != nil {
				if sa.SemanticInfo.MethodSymbols[receiverType] == nil {
					sa.SemanticInfo.MethodSymbols[receiverType] = make(map[string]*Symbol)
				}
				sa.SemanticInfo.MethodSymbols[receiverType][fnStmt.Name.Value] = sym
				unwrappedReceiver := types.UnwrapLease(receiverType)
				if pt, ok := unwrappedReceiver.(*types.PointerType); ok {
					unwrappedReceiver = pt.Base
				}
				if st, ok := unwrappedReceiver.(*types.StructType); ok {
					st.Methods[fnStmt.Name.Value] = sym.Type
				} else if sumT, ok := unwrappedReceiver.(*types.SumType); ok {
					if sumT.Methods == nil {
						sumT.Methods = make(map[string]types.NRType)
					}
					sumT.Methods[fnStmt.Name.Value] = sym.Type
				}
			}
		}
		return inst
	}

	// 3. Generate New Instance
	// 1. Explicit type parameters
	mapping := make(map[string]ast.TypeNode)
	typeArgIdx := 0
	for _, tp := range fnStmt.TypeParameters {
		if typeArgIdx < len(typeArgs) {
			mapping[tp.Name.Value] = sa.typeToTypeNode(typeArgs[typeArgIdx])
			typeArgIdx++
		}
	}

	// 2. Implicit type parameters from receiver
	if fnStmt.Receiver != nil && len(recTypeArgs) > 0 {
		baseExp := fnStmt.Receiver.Type
		if pt, ok := baseExp.(*ast.PrefixExpression); ok && (pt.Operator == "&" || pt.Operator == "#" || pt.Operator == "@") {
			if tn, ok := pt.Right.(ast.TypeNode); ok {
				baseExp = tn
			}
		}
		if idxExpr, ok := baseExp.(*ast.IndexExpression); ok {
			for i, idx := range idxExpr.Indices {
				if ident, ok := idx.(*ast.Identifier); ok {
					if i < len(recTypeArgs) {
						mapping[ident.Value] = sa.typeToTypeNode(recTypeArgs[i])
					}
				}
			}
		}
	}

	effectiveKey := "_" + types.GetHashSuffix(fnStmt.Name.Value, typeArgs)

	specializedFn := sa.instantiateFunction(fnStmt, mapping, effectiveKey)
	sa.SemanticInfo.Instances[fnStmt][typeKey] = specializedFn

	// 4. Analyze the specialized function
	prevScope := sa.CurrentScope
	prevFn := sa.CurrentFunction
	prevLambda := sa.CurrentLambda
	sa.CurrentFunction = nil
	sa.CurrentLambda = nil

	if !specializedFn.IsGenericTemplate {
		// Concrete monomorphization: Define in the generic template's original definition scope to resolve package-level types correctly.
		pkgScope := sa.FuncScopes[fnStmt]
		if pkgScope == nil {
			pkgScope = sa.CurrentScope
			for pkgScope != nil && pkgScope.Kind != ScopePackage {
				pkgScope = pkgScope.Parent
			}
		}

		// [FIX] Use a temporary scope for monomorphized type arguments to avoid polluting the package scope
		// and causing collisions between different specializations.
		if pkgScope != nil {
			sa.CurrentScope = NewScope(pkgScope, ScopeBlock)
		}

		// Nora: Define monomorphized type arguments in this temporary scope
		// so nested generic structs can auto-specialize during Analyze.
		// 1. Explicit type parameters
		typeArgIdx := 0
		for _, tp := range fnStmt.TypeParameters {
			if typeArgIdx < len(typeArgs) {
				sa.CurrentScope.Define(tp.Name.Value, typeArgs[typeArgIdx], SymType, tp)
				concreteType := types.UnwrapLease(typeArgs[typeArgIdx])
				if concreteType != nil {
					sa.CurrentScope.Define(concreteType.Name(), concreteType, SymType, tp)
				}
				typeArgIdx++
			}
		}
		// 2. Implicit type parameters from receiver
		if fnStmt.Receiver != nil && len(recTypeArgs) > 0 {
			baseExp := fnStmt.Receiver.Type
			if pt, ok := baseExp.(*ast.PrefixExpression); ok && (pt.Operator == "&" || pt.Operator == "#" || pt.Operator == "@") {
				if tn, ok := pt.Right.(ast.TypeNode); ok {
					baseExp = tn
				}
			}
			if idxExpr, ok := baseExp.(*ast.IndexExpression); ok {
				for i, idx := range idxExpr.Indices {
					if ident, ok := idx.(*ast.Identifier); ok {
						if i < len(recTypeArgs) {
							sa.CurrentScope.Define(ident.Value, recTypeArgs[i], SymType, ident)
							concreteType := types.UnwrapLease(recTypeArgs[i])
							if concreteType != nil {
								sa.CurrentScope.Define(concreteType.Name(), concreteType, SymType, ident)
							}
						}
					}
				}
			}
		}

		sa.CollectSymbols(specializedFn)
		sa.Analyze(specializedFn)
	} else {
		// Generic template: Define in the generic template's original definition scope to resolve package-level types correctly,
		// keep current scope (e.g. for resolving T) and skip deep analysis.
		pkgScope := sa.FuncScopes[fnStmt]
		if pkgScope == nil {
			pkgScope = sa.CurrentScope
			for pkgScope != nil && pkgScope.Kind != ScopePackage {
				pkgScope = pkgScope.Parent
			}
		}

		prevScope := sa.CurrentScope
		if pkgScope != nil {
			sa.CurrentScope = NewScope(pkgScope, ScopeBlock)
		}

		// Define monomorphized type arguments in this temporary scope
		// 1. Explicit type parameters
		typeArgIdx := 0
		for _, tp := range fnStmt.TypeParameters {
			if typeArgIdx < len(typeArgs) {
				sa.CurrentScope.Define(tp.Name.Value, typeArgs[typeArgIdx], SymType, tp)
				concreteType := types.UnwrapLease(typeArgs[typeArgIdx])
				if concreteType != nil {
					sa.CurrentScope.Define(concreteType.Name(), concreteType, SymType, tp)
				}
				typeArgIdx++
			}
		}
		// 2. Implicit type parameters from receiver
		if fnStmt.Receiver != nil && len(recTypeArgs) > 0 {
			baseExp := fnStmt.Receiver.Type
			if pt, ok := baseExp.(*ast.PrefixExpression); ok && (pt.Operator == "&" || pt.Operator == "#" || pt.Operator == "@") {
				if tn, ok := pt.Right.(ast.TypeNode); ok {
					baseExp = tn
				}
			}
			if idxExpr, ok := baseExp.(*ast.IndexExpression); ok {
				for i, idx := range idxExpr.Indices {
					if ident, ok := idx.(*ast.Identifier); ok {
						if i < len(recTypeArgs) {
							sa.CurrentScope.Define(ident.Value, recTypeArgs[i], SymType, ident)
							concreteType := types.UnwrapLease(recTypeArgs[i])
							if concreteType != nil {
								sa.CurrentScope.Define(concreteType.Name(), concreteType, SymType, ident)
							}
						}
					}
				}
			}
		}

		sa.CollectSymbols(specializedFn)
		sa.CurrentScope = prevScope
	}
	sa.CurrentScope = prevScope
	sa.CurrentFunction = prevFn
	sa.CurrentLambda = prevLambda

	// 5. Register in MethodSymbols if it's a method
	if receiverType != nil {
		if sym, ok := sa.SemanticInfo.Defs[specializedFn.Name]; ok && sym != nil {
			if sa.SemanticInfo.MethodSymbols[receiverType] == nil {
				sa.SemanticInfo.MethodSymbols[receiverType] = make(map[string]*Symbol)
			}
			// Register under the ORIGINAL method name so the selector resolution can find it.
			sa.SemanticInfo.MethodSymbols[receiverType][fnStmt.Name.Value] = sym

			// Also update the Methods map on the struct type itself
			unwrappedReceiver := types.UnwrapLease(receiverType)
			if pt, ok := unwrappedReceiver.(*types.PointerType); ok {
				unwrappedReceiver = pt.Base
			}
			if st, ok := unwrappedReceiver.(*types.StructType); ok {
				st.Methods[fnStmt.Name.Value] = sym.Type
			} else if sumT, ok := unwrappedReceiver.(*types.SumType); ok {
				if sumT.Methods == nil {
					sumT.Methods = make(map[string]types.NRType)
				}
				sumT.Methods[fnStmt.Name.Value] = sym.Type
			}
		}
	}

	return specializedFn
}

func (sa *SemanticAnalyzer) instantiateFunction(fn *ast.FunctionStatement, mapping map[string]ast.TypeNode, key string) *ast.FunctionStatement {

	if sa.DebugMode {
		fmt.Printf("[DEBUG instantiateFunction] mapping keys: ")
		for k, v := range mapping {
			fmt.Printf("%s -> %#v, ", k, v)
		}
		fmt.Printf("\n")
	}

	// Deep Clone and Substitute
	cloned := sa.cloneAndSubstitute(fn, mapping).(*ast.FunctionStatement)

	// Check if this instantiation contains generic parameters (e.g. T passed to T)
	hasGenericArgs := false
	for _, ta := range mapping {
		if sa.hasGeneric(sa.resolveTypeNode(ta)) {
			hasGenericArgs = true
			break
		}
	}

	cloned.IsGenericTemplate = hasGenericArgs
	cloned.TypeParameters = nil // Always clear so CollectSymbols analyzes it

	// Rename to specialized name
	cloned.Name = &ast.Identifier{
		Token: fn.Name.Token,
		Value: sanitizeCIdentifier(fn.Name.Value + key),
	}
	// Return the cloned function

	return cloned
}

func (sa *SemanticAnalyzer) cloneAndSubstitute(node ast.Node, mapping map[string]ast.TypeNode) ast.Node {
	if ast.IsNil(node) {
		return nil
	}

	sa.depth++
	defer func() { sa.depth-- }()
	if sa.depth > 1000 {
		panic("infinite recursion detected in cloneAndSubstitute")
	}

	cloned := sa.cloneAndSubstituteHelper(node, mapping)
	if cloned != nil && node != nil && mapping == nil {
		if t, ok := sa.SemanticInfo.Types[node]; ok && t != nil {
			sa.SemanticInfo.Types[cloned] = t
		}
	}
	return cloned
}

func (sa *SemanticAnalyzer) cloneAndSubstituteHelper(node ast.Node, mapping map[string]ast.TypeNode) ast.Node {
	switch n := node.(type) {
	case *ast.Identifier:
		if concrete, ok := mapping[n.Value]; ok {
			// Deep clone the concrete replacement to avoid cross-linkage of AST nodes
			return sa.cloneAndSubstitute(concrete, nil)
		}
		return &ast.Identifier{Token: n.Token, Value: n.Value}

	case *ast.FunctionStatement:
		res := &ast.FunctionStatement{
			Token:             n.Token,
			Name:              &ast.Identifier{Token: n.Name.Token, Value: n.Name.Value},
			IsGenericTemplate: n.IsGenericTemplate,
			IsExtern:          n.IsExtern,
			IsExport:          n.IsExport,
			Attributes:        n.Attributes,
		}
		if n.Receiver != nil {
			res.Receiver = &ast.Parameter{
				Token:     n.Receiver.Token,
				Name:      &ast.Identifier{Token: n.Receiver.Name.Token, Value: n.Receiver.Name.Value},
				Type:      sa.cloneAndSubstitute(n.Receiver.Type, mapping).(ast.TypeNode),
				LeaseKind: n.Receiver.LeaseKind,
			}
		}
		if n.ReturnType != nil {
			res.ReturnType = sa.cloneAndSubstitute(n.ReturnType, mapping).(ast.TypeNode)
		}
		for _, tp := range n.TypeParameters {
			res.TypeParameters = append(res.TypeParameters, sa.cloneAndSubstitute(tp, mapping).(*ast.TypeParameter))
		}
		for _, p := range n.Parameters {
			res.Parameters = append(res.Parameters, &ast.Parameter{
				Token:     p.Token,
				Name:      &ast.Identifier{Token: p.Name.Token, Value: p.Name.Value},
				Type:      sa.cloneAndSubstitute(p.Type, mapping).(ast.TypeNode),
				LeaseKind: p.LeaseKind,
			})
		}
		if n.Body != nil {
			res.Body = sa.cloneAndSubstitute(n.Body, mapping).(*ast.BlockStatement)
		}
		return res

	case *ast.TypeParameter:
		res := &ast.TypeParameter{
			Token: n.Token,
			Name:  &ast.Identifier{Token: n.Name.Token, Value: n.Name.Value},
		}
		if !ast.IsNil(n.Constraint) {
			res.Constraint = sa.cloneAndSubstitute(n.Constraint, mapping).(ast.TypeNode)
		}
		return res

	case *ast.BlockStatement:
		res := &ast.BlockStatement{Token: n.Token}
		for _, s := range n.Statements {
			cloned := sa.cloneAndSubstitute(s, mapping)
			if cloned == nil {
				panic(fmt.Sprintf("cloneAndSubstitute returned nil for statement of type %T (IsNil: %v)", s, ast.IsNil(s)))
			}
			res.Statements = append(res.Statements, cloned.(ast.Statement))
		}
		return res

	case *ast.ExpressionStatement:
		return &ast.ExpressionStatement{
			Token:      n.Token,
			Expression: sa.cloneAndSubstitute(n.Expression, mapping).(ast.Expression),
		}

	case *ast.ReturnStatement:
		res := &ast.ReturnStatement{
			Token: n.Token,
		}
		if !ast.IsNil(n.ReturnValue) {
			res.ReturnValue = sa.cloneAndSubstitute(n.ReturnValue, mapping).(ast.Expression)
		}
		return res

	case *ast.VarStatement:
		stmt := &ast.VarStatement{
			Token: n.Token,
			Name:  &ast.Identifier{Token: n.Name.Token, Value: n.Name.Value},
		}
		if !ast.IsNil(n.Type) {
			stmt.Type = sa.cloneAndSubstitute(n.Type, mapping).(ast.TypeNode)
		}
		if !ast.IsNil(n.Value) {
			stmt.Value = sa.cloneAndSubstitute(n.Value, mapping).(ast.Expression)
		}
		return stmt

	case *ast.CallExpression:
		res := &ast.CallExpression{
			Token:    n.Token,
			Function: sa.cloneAndSubstitute(n.Function, mapping).(ast.Expression),
		}
		for _, ta := range n.TypeArguments {
			res.TypeArguments = append(res.TypeArguments, sa.cloneAndSubstitute(ta, mapping).(ast.TypeNode))
		}
		for _, arg := range n.Arguments {
			res.Arguments = append(res.Arguments, &ast.ArgumentsExpression{
				Token: arg.Token,
				Value: sa.cloneAndSubstitute(arg.Value, mapping).(ast.Expression),
			})
		}
		return res

	case *ast.PrefixExpression:
		return &ast.PrefixExpression{
			Token:    n.Token,
			Operator: n.Operator,
			Right:    sa.cloneAndSubstitute(n.Right, mapping).(ast.Expression),
		}

	case *ast.GroupedExpression:
		return &ast.GroupedExpression{
			Token:      n.Token,
			Expression: sa.cloneAndSubstitute(n.Expression, mapping).(ast.Expression),
		}

	case *ast.InfixExpression:
		return &ast.InfixExpression{
			Token:    n.Token,
			Left:     sa.cloneAndSubstitute(n.Left, mapping).(ast.Expression),
			Operator: n.Operator,
			Right:    sa.cloneAndSubstitute(n.Right, mapping).(ast.Expression),
		}

	case *ast.IndexExpression:
		res := &ast.IndexExpression{
			Token: n.Token,
			Left:  sa.cloneAndSubstitute(n.Left, mapping).(ast.Expression),
		}
		for _, idx := range n.Indices {
			res.Indices = append(res.Indices, sa.cloneAndSubstitute(idx, mapping).(ast.Expression))
		}
		return res

	case *ast.IntegerLiteral, *ast.StringLiteral, *ast.Boolean, *ast.FloatLiteral, *ast.RuneLiteral:
		return n // Literals are immutable and leaf nodes

	case *ast.NoneLiteral:
		return n

	case *ast.AllocExpression:
		return &ast.AllocExpression{
			Token: n.Token,
			Value: sa.cloneAndSubstitute(n.Value, mapping).(ast.Expression),
		}

	case *ast.StructLiteral:
		res := &ast.StructLiteral{
			Token: n.Token,
		}
		if n.Name != nil {
			res.Name = sa.cloneAndSubstitute(n.Name, mapping).(ast.Expression)
		}
		for _, f := range n.Fields {
			newField := &ast.FieldDefinition{
				Token: f.Token,
				Name:  f.Name,
			}
			if f.Value != nil {
				newField.Value = sa.cloneAndSubstitute(f.Value, mapping).(ast.Expression)
			}
			if f.Type != nil {
				newField.Type = sa.cloneAndSubstitute(f.Type, mapping).(ast.TypeNode)
			}
			res.Fields = append(res.Fields, newField)
		}
		return res

	case *ast.SelectorExpression:
		return &ast.SelectorExpression{
			Token: n.Token,
			Left:  sa.cloneAndSubstitute(n.Left, mapping).(ast.Expression),
			Field: &ast.Identifier{Token: n.Field.Token, Value: n.Field.Value},
		}

	case *ast.FunctionType:
		res := &ast.FunctionType{
			Token: n.Token,
		}
		for _, p := range n.Parameters {
			res.Parameters = append(res.Parameters, sa.cloneAndSubstitute(p, mapping).(ast.TypeNode))
		}
		if n.ReturnType != nil {
			res.ReturnType = sa.cloneAndSubstitute(n.ReturnType, mapping).(ast.TypeNode)
		}
		return res

	case *ast.AssignmentStatement:
		stmt := &ast.AssignmentStatement{
			Token: n.Token,
			Left:  sa.cloneAndSubstitute(n.Left, mapping).(ast.Expression),
			Value: sa.cloneAndSubstitute(n.Value, mapping).(ast.Expression),
		}
		return stmt

	case *ast.WhileStatement:
		return &ast.WhileStatement{
			Token:     n.Token,
			Condition: sa.cloneAndSubstitute(n.Condition, mapping).(ast.Expression),
			Body:      sa.cloneAndSubstitute(n.Body, mapping).(*ast.BlockStatement),
		}

	case *ast.IfExpression:
		res := &ast.IfExpression{
			Token:       n.Token,
			Condition:   sa.cloneAndSubstitute(n.Condition, mapping).(ast.Expression),
			Consequence: sa.cloneAndSubstitute(n.Consequence, mapping).(*ast.BlockStatement),
		}
		if !ast.IsNil(n.Alternative) {
			res.Alternative = sa.cloneAndSubstitute(n.Alternative, mapping).(ast.Expression)
		}
		return res

	case *ast.ForStatement:
		res := &ast.ForStatement{
			Token: n.Token,
		}

		if !ast.IsNil(n.Iterable) {
			res.Iterable = sa.cloneAndSubstitute(n.Iterable, mapping).(ast.Expression)
		}
		if !ast.IsNil(n.Value) {
			res.Value = &ast.Identifier{Token: n.Value.Token, Value: n.Value.Value}
		}
		if !ast.IsNil(n.Key) {
			res.Key = &ast.Identifier{Token: n.Key.Token, Value: n.Key.Value}
		}
		res.Body = sa.cloneAndSubstitute(n.Body, mapping).(*ast.BlockStatement)
		return res

	case *ast.BreakStatement:
		return n

	case *ast.ContinueStatement:
		return n

	case *ast.SpawnExpression:
		res := &ast.SpawnExpression{
			Token: n.Token,
		}
		if n.MonitorChannel != nil {
			res.MonitorChannel = sa.cloneAndSubstitute(n.MonitorChannel, mapping).(ast.Expression)
		}
		if n.Call != nil {
			res.Call = sa.cloneAndSubstitute(n.Call, mapping).(*ast.CallExpression)
		}
		if n.Body != nil {
			res.Body = sa.cloneAndSubstitute(n.Body, mapping).(*ast.BlockStatement)
		}
		return res

	case *ast.DeferStatement:
		res := &ast.DeferStatement{
			Token: n.Token,
		}
		if n.Call != nil {
			res.Call = sa.cloneAndSubstitute(n.Call, mapping).(*ast.CallExpression)
		}
		return res

	case *ast.MatchExpression:
		res := &ast.MatchExpression{
			Token:  n.Token,
			Target: sa.cloneAndSubstitute(n.Target, mapping).(ast.Expression),
			Type:   n.Type,
		}
		for _, c := range n.Cases {
			newCase := &ast.MatchCase{
				Pattern: sa.cloneAndSubstitute(c.Pattern, mapping).(ast.Expression),
			}
			if c.Body != nil {
				newCase.Body = sa.cloneAndSubstitute(c.Body, mapping).(*ast.BlockStatement)
			}
			res.Cases = append(res.Cases, newCase)
		}
		return res

	case *ast.ArrayLiteral:
		res := &ast.ArrayLiteral{
			Token: n.Token,
		}
		for _, el := range n.Elements {
			res.Elements = append(res.Elements, sa.cloneAndSubstitute(el, mapping).(ast.Expression))
		}
		return res

	case *ast.MapLiteral:
		res := &ast.MapLiteral{
			Token: n.Token,
			Pairs: make(map[ast.Expression]ast.Expression),
		}
		for k, v := range n.Pairs {
			newK := sa.cloneAndSubstitute(k, mapping).(ast.Expression)
			newV := sa.cloneAndSubstitute(v, mapping).(ast.Expression)
			res.Pairs[newK] = newV
		}
		return res

	case *ast.InterpolatedString:
		res := &ast.InterpolatedString{
			Token: n.Token,
		}
		for _, p := range n.Parts {
			res.Parts = append(res.Parts, sa.cloneAndSubstitute(p, mapping).(ast.Expression))
		}
		return res

	case *ast.SendExpression:
		return &ast.SendExpression{
			Token: n.Token,
			Left:  sa.cloneAndSubstitute(n.Left, mapping).(ast.Expression),
			Right: sa.cloneAndSubstitute(n.Right, mapping).(ast.Expression),
		}

	case *ast.ReceiveExpression:
		return &ast.ReceiveExpression{
			Token: n.Token,
			Value: sa.cloneAndSubstitute(n.Value, mapping).(ast.Expression),
		}

	case *ast.LambdaExpression:
		res := &ast.LambdaExpression{
			Token: n.Token,
		}
		for _, p := range n.Parameters {
			res.Parameters = append(res.Parameters, &ast.Parameter{
				Token:     p.Token,
				Name:      &ast.Identifier{Token: p.Name.Token, Value: p.Name.Value},
				Type:      sa.cloneAndSubstitute(p.Type, mapping).(ast.TypeNode),
				LeaseKind: p.LeaseKind,
			})
		}
		if n.ReturnType != nil {
			res.ReturnType = sa.cloneAndSubstitute(n.ReturnType, mapping).(ast.TypeNode)
		}
		if n.Body != nil {
			res.Body = sa.cloneAndSubstitute(n.Body, mapping).(*ast.BlockStatement)
		}
		return res

	case *ast.ScopeExpression:
		res := &ast.ScopeExpression{
			Token: n.Token,
		}
		if n.Body != nil {
			res.Body = sa.cloneAndSubstitute(n.Body, mapping).(*ast.BlockStatement)
		}
		return res

	case *ast.ParallelExpression:
		res := &ast.ParallelExpression{
			Token: n.Token,
			Type:  n.Type,
		}
		if n.Body != nil {
			res.Body = sa.cloneAndSubstitute(n.Body, mapping).(*ast.BlockStatement)
		}
		return res

	default:
		return n // Fallback
	}
}

func (sa *SemanticAnalyzer) implements(concrete types.NRType, protocol *types.ProtocolType) bool {
	baseConcrete := concrete
	for {
		if pt, ok := baseConcrete.(*types.PointerType); ok {
			baseConcrete = pt.Base
		} else {
			break
		}
	}

	if gt, ok := baseConcrete.(*types.GenericType); ok {
		if gt.Constraint != nil {
			if sa.implements(gt.Constraint, protocol) {
				return true
			}
		}
	}

	if pt, ok := baseConcrete.(*types.ProtocolType); ok {
		if types.Equals(pt, protocol) {
			return true
		}
	}

	for methodName, protoMethod := range protocol.Methods {
		var fnType *types.FunctionType

		// 1. Try to find as a method on the struct/type itself
		if st, ok := baseConcrete.(*types.StructType); ok {
			if m, ok := st.Methods[methodName].(*types.FunctionType); ok {
				fnType = m
			}
		} else if sumT, ok := baseConcrete.(*types.SumType); ok {
			if mType, ok := sumT.Methods[methodName]; ok {
				if m, ok := mType.(*types.FunctionType); ok {
					fnType = m
				}
			}
		} else if primT, ok := baseConcrete.(*types.PrimitiveType); ok {
			if mType, ok := primT.Methods[methodName]; ok {
				if m, ok := mType.(*types.FunctionType); ok {
					fnType = m
				}
			}
		}

		// 2. Try to find as a global function (structural)
		if fnType == nil {
			sym, exists := sa.CurrentScope.Resolve(methodName)
			if exists && sym.Kind == SymFunc {
				if ft, ok := sym.Type.(*types.FunctionType); ok {
					fnType = ft
				}
			}
		}

		if fnType == nil {
			return false
		}

		// Distinguish Self and Parameters
		var actualSelf types.NRType
		var actualMethodParams []types.NRType

		if fnType.Receiver != nil {
			actualSelf = fnType.Receiver
			actualMethodParams = fnType.Params
		} else {
			if len(fnType.Params) == 0 {
				return false
			}
			actualSelf = fnType.Params[0]
			actualMethodParams = fnType.Params[1:]
		}

		// Check parameter counts (excluding self)
		if len(actualMethodParams) != len(protoMethod.Params) {
			return false
		}

		// 1. Verify 'self' matches the concrete type
		actualBase := actualSelf
		for {
			if pt, ok := actualBase.(*types.PointerType); ok {
				actualBase = pt.Base
			} else {
				break
			}
		}
		concreteBase := concrete
		for {
			if pt, ok := concreteBase.(*types.PointerType); ok {
				concreteBase = pt.Base
			} else {
				break
			}
		}
		if !types.Equals(actualBase, concreteBase) {
			if strings.Contains(concreteBase.Name(), "8f98b6a7") {
				if sa.DebugMode {
					fmt.Printf("DEBUG 8f98b6a7: method %s self mismatch: actualBase %s (%T), concreteBase %s (%T)\n", methodName, actualBase.Name(), actualBase, concreteBase.Name(), concreteBase)
				}
			}
			return false
		}

		// 2. Check the rest of the parameters
		for i := 0; i < len(actualMethodParams); i++ {
			if !types.Equals(actualMethodParams[i], protoMethod.Params[i]) {
				if strings.Contains(concreteBase.Name(), "8f98b6a7") {
					if sa.DebugMode {
						fmt.Printf("DEBUG 8f98b6a7: method %s param %d mismatch: expected %s (%T), got %s (%T)\n", methodName, i, protoMethod.Params[i].Name(), protoMethod.Params[i], actualMethodParams[i].Name(), actualMethodParams[i])
					}
				}
				return false
			}
		}

		// 3. Check return type
		if !types.Equals(fnType.Return, protoMethod.Return) {
			if strings.Contains(concreteBase.Name(), "8f98b6a7") {
				if sa.DebugMode {
					fmt.Printf("DEBUG 8f98b6a7: method %s return mismatch: expected %s (%T), got %s (%T)\n", methodName, protoMethod.Return.Name(), protoMethod.Return, fnType.Return.Name(), fnType.Return)
				}
			}
			return false
		}
	}
	return true
}

func (sa *SemanticAnalyzer) checkImplicitMoveLoad(valExpr ast.Expression, target types.NRType) {
	if valExpr == nil || target == nil {
		return
	}
	valType := sa.SemanticInfo.Types[valExpr]
	if pt, ok := valType.(*types.PointerType); ok && pt.Kind == types.LeaseMove {
		// Only block if we are implicitly unwrapping (target is not a pointer)
		if _, targetIsPtr := target.(*types.PointerType); !targetIsPtr {
			if types.Equals(types.UnwrapLease(pt.Base), types.UnwrapLease(target)) {
				// If it's not an interior pointer move (PrefixExpression), block it
				if _, isPrefix := valExpr.(*ast.PrefixExpression); !isPrefix {
					sa.AddError(valExpr.Pos(), "cannot implicitly load value from move-lease. You must manually unpack the fields if you intend to copy it, or ensure the pointer is freed.")
				}
			}
		}
	}
}

func (sa *SemanticAnalyzer) checkInterfaceCompatibility(valExpr ast.Expression, target types.NRType) {
	sa.checkImplicitMoveLoad(valExpr, target)
	valType := sa.SemanticInfo.Types[valExpr]
	if valType == nil || valType == types.ErrorType {
		return
	}

	if types.IsAssignable(target, valType) {
		return
	}

	expectedName := target.Name()
	actualName := valType.Name()

	if proto, ok := target.(*types.ProtocolType); ok {
		if !sa.implements(valType, proto) {
			sa.AddError(valExpr.Pos(), "type '%s' does not implement interface '%s'",
				actualName, expectedName)
		} else {
			// Interface cast requires vtable to be built.
			// The concrete methods must be monomorphized to be available to the code generator.
			baseType := valType
			if pt, ok := baseType.(*types.PointerType); ok {
				baseType = pt.Base
			}
			sa.ensureMethodsSpecialized(baseType)
		}
	} else {
		sa.AddError(valExpr.Pos(), "type mismatch: expected %s, got %s", expectedName, actualName)
	}
}

func (sa *SemanticAnalyzer) sortedFieldNames(fields map[string]types.NRType) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (sa *SemanticAnalyzer) inferTypeArguments(fnStmt *ast.FunctionStatement, call *ast.CallExpression) []types.NRType {
	inferred := make(map[string]types.NRType)

	// 1. Infer from Receiver (if it's a method call)
	if sel, ok := call.Function.(*ast.SelectorExpression); ok && fnStmt.Receiver != nil {
		receiverType := sa.SemanticInfo.Types[sel.Left]
		if receiverType != nil {
			sa.matchType(fnStmt.Receiver.Type, receiverType, fnStmt.TypeParameters, inferred, false)
		}
	}

	// 2. Infer from Arguments
	for i, arg := range call.Arguments {
		if i >= len(fnStmt.Parameters) {
			break
		}
		// Analyze argument to get its type
		sa.Analyze(arg.Value)
		argType := sa.SemanticInfo.Types[arg.Value]

		if argType != nil {
			// fmt.Printf("[DEBUG] infer from arg %d: pattern = %s (%T), actual = %s (%T)\n", i, fnStmt.Parameters[i].Type.String(), fnStmt.Parameters[i].Type, argType.Name(), argType)
		} else {
			// fmt.Printf("[DEBUG] infer from arg %d: pattern = %s (%T), actual = nil\n", i, fnStmt.Parameters[i].Type.String(), fnStmt.Parameters[i].Type)
		}

		// Match against parameter type pattern
		sa.matchType(fnStmt.Parameters[i].Type, argType, fnStmt.TypeParameters, inferred, true)
	}

	// Build the final slice
	result := []types.NRType{}
	for _, tp := range fnStmt.TypeParameters {
		t, ok := inferred[tp.Name.Value]
		if !ok {
			sa.AddError(call.Pos(), "could not infer type for generic parameter %s", tp.Name.Value)
			return nil
		}
		result = append(result, t)
	}
	return result
}

func (sa *SemanticAnalyzer) matchType(pattern ast.Node, actual types.NRType, typeParams []*ast.TypeParameter, inferred map[string]types.NRType, stripLease bool) {
	if pattern == nil || actual == nil {
		return
	}

	switch pt := pattern.(type) {
	case *ast.Identifier:
		for _, tp := range typeParams {
			if pt.Value == tp.Name.Value {
				val := actual
				if stripLease {
					for {
						if pt, ok := val.(*types.PointerType); ok && pt.Leased && !pt.IsArray {
							if pt.Kind == types.LeaseRead || pt.Kind == types.LeaseWrite || pt.Kind == types.LeaseMove {
								val = pt.Base
							} else {
								break
							}
						} else {
							break
						}
					}
				}
				if existing, ok := inferred[pt.Value]; ok {
					if ptExisting, ok := existing.(*types.PointerType); ok && ptExisting.Leased {
						if !types.IsPointerLike(val) {
							return
						}
					}
					return
				}
				inferred[pt.Value] = val
				return
			}
		}
	case *ast.PrefixExpression:
		// Handle pointer types in patterns: *T matched against *i32
		if pt.Operator == "*" || pt.Operator == "#" || pt.Operator == "&" || pt.Operator == "@" {
			if at, ok := actual.(*types.PointerType); ok {
				sa.matchType(pt.Right, at.Base, typeParams, inferred, stripLease)
			}
		}
	case *ast.IndexExpression:
		// Handle List[T], Map[K, V], Box[T]
		baseNode, ok := pt.Left.(ast.TypeNode)
		if !ok {
			return
		}
		patternBaseType := sa.resolveTypeNode(baseNode)

		if lt, ok := actual.(*types.ListType); ok && patternBaseType.Name() == "List" && len(pt.Indices) == 1 {
			sa.matchType(pt.Indices[0], lt.ElementType, typeParams, inferred, false)
		} else if mt, ok := actual.(*types.MapType); ok && patternBaseType.Name() == "Map" && len(pt.Indices) == 2 {
			sa.matchType(pt.Indices[0], mt.Key, typeParams, inferred, false)
			sa.matchType(pt.Indices[1], mt.Value, typeParams, inferred, false)
		} else if st, ok := actual.(*types.StructType); ok {
			if st.BaseType == patternBaseType && len(st.TypeArgs) == len(pt.Indices) {
				for i, idx := range pt.Indices {
					sa.matchType(idx, st.TypeArgs[i], typeParams, inferred, false)
				}
			}
		} else if st, ok := actual.(*types.SumType); ok {
			if st.BaseType == patternBaseType && len(st.TypeArgs) == len(pt.Indices) {
				for i, idx := range pt.Indices {
					sa.matchType(idx, st.TypeArgs[i], typeParams, inferred, false)
				}
			}
		}
	case *ast.FunctionType:
		if ft, ok := actual.(*types.FunctionType); ok {
			if len(pt.Parameters) == len(ft.Params) {
				for i, p := range pt.Parameters {
					sa.matchType(p, ft.Params[i], typeParams, inferred, stripLease)
				}
			}
			if pt.ReturnType != nil {
				sa.matchType(pt.ReturnType, ft.Return, typeParams, inferred, false)
			}
		}
	}
}

func (sa *SemanticAnalyzer) typeToTypeNode(t types.NRType) ast.TypeNode {
	if t == nil {
		return nil
	}
	var res ast.TypeNode
	switch kt := t.(type) {
	case *types.PrimitiveType:
		res = &ast.Identifier{Value: kt.Name()}
	case *types.ListType:
		res = &ast.IndexExpression{
			Left:    &ast.Identifier{Value: "List"},
			Indices: []ast.Expression{sa.typeToTypeNode(kt.ElementType).(ast.Expression)},
		}
	case *types.MapType:
		res = &ast.IndexExpression{
			Left: &ast.Identifier{Value: "Map"},
			Indices: []ast.Expression{
				sa.typeToTypeNode(kt.Key).(ast.Expression),
				sa.typeToTypeNode(kt.Value).(ast.Expression),
			},
		}
	case *types.PointerType:
		op := "@"
		var tokType token.TokenType = token.MOVE
		if kt.Kind == types.LeaseRead {
			op = "#"
			tokType = token.HASH
		} else if kt.Kind == types.LeaseWrite {
			op = "&"
			tokType = token.AND
		}
		res = &ast.PrefixExpression{
			Token:    token.Token{Type: tokType, Literal: op},
			Operator: op,
			Right:    sa.typeToTypeNode(kt.Base).(ast.Expression),
		}
	default:
		// Fallback for named types (Structs, SumTypes)
		res = &ast.Identifier{Value: kt.Name()}
	}

	if res != nil {
		sa.SemanticInfo.Types[res] = t
	}
	return res
}
func (sa *SemanticAnalyzer) analyzeAllocExpression(n *ast.AllocExpression) {
	// Case 1: Array Allocation (e.g., alloc i32[10] or alloc @Entry[K, V][10])
	target := n.Value
	var arrayIdx *ast.IndexExpression
	var typeNode ast.TypeNode

	if idx, ok := target.(*ast.IndexExpression); ok {
		arrayIdx = idx
		if tn, ok := idx.Left.(ast.TypeNode); ok {
			typeNode = tn
		}
	} else if pref, ok := target.(*ast.PrefixExpression); ok {
		if idx, ok := pref.Right.(*ast.IndexExpression); ok {
			arrayIdx = idx
			// Construct a virtual TypeNode that includes the lease: @Node
			typeNode = &ast.PrefixExpression{
				Token:    pref.Token,
				Operator: pref.Operator,
				Right:    idx.Left,
			}
		}
	}

	if arrayIdx != nil {
		baseType := sa.resolveTypeNode(typeNode)
		if baseType != types.ErrorType {
			if len(arrayIdx.Indices) > 0 {
				sizeExpr := arrayIdx.Indices[0]
				sa.Analyze(sizeExpr)
				sizeType := sa.SemanticInfo.Types[sizeExpr]
				if sizeType != types.Int && sizeType != types.I32 {
					sa.AddError(sizeExpr.Pos(), "array size must be an integer, got %s", sizeType.Name())
				}
			}

			n.Type = &types.PointerType{Base: baseType, IsArray: true, Leased: false, Kind: types.LeaseMove}
			sa.SemanticInfo.Types[n] = n.Type
			return
		}
	}

	// Case 2: Individual Value Allocation (e.g., alloc User{...})
	sa.Analyze(n.Value)
	valType := sa.SemanticInfo.Types[n.Value]
	if pref, ok := n.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
		valType = sa.SemanticInfo.Types[pref.Right]
	}

	if valType == nil || valType == types.ErrorType {
		sa.SemanticInfo.Types[n] = types.ErrorType
		return
	}

	// Pointer to the allocated type
	n.Type = &types.PointerType{Base: valType, Leased: false, Kind: types.LeaseMove}
	sa.SemanticInfo.Types[n] = n.Type
}

func (sa *SemanticAnalyzer) checkConstraint(t types.NRType, constraint types.NRType, pos token.Position) bool {
	if sa.CollectingSymbols {
		return true // Skip checking during collection to avoid false positives on incomplete methods
	}
	if constraint == nil || constraint == types.Void || constraint == types.ErrorType {
		return true
	}

	if proto, ok := constraint.(*types.ProtocolType); ok {
		// INHERIT CONSTRAINT FOR IMPLICIT GENERICS
		if gt, ok := t.(*types.GenericType); ok {
			if gt.Constraint == nil || gt.Constraint == types.Any {
				gt.Constraint = proto
				return true
			}
		}

		if !sa.implements(t, proto) {
			if sa.DebugMode {
				println("[DEBUG] checkConstraint FAIL. t:", t.Name(), "kind:", t.GetKind(), "constraint:", proto.Name())
				println("  t dynamic type:", fmt.Sprintf("%T", t))
				if gt, ok := t.(*types.GenericType); ok {
					println("  t is GenericType. Constraint:", gt.Constraint)
					if gt.Constraint != nil {
						println("  t.Constraint.Name():", gt.Constraint.Name(), "kind:", gt.Constraint.GetKind())
					}
				}
			}
			sa.AddError(pos, "type '%s' does not satisfy constraint '%s'", t.Name(), proto.Name())
			return false
		}
	} else {
		// Simple type equality constraint (e.g. [T: i32])
		if !types.Equals(t, constraint) {
			sa.AddError(pos, "type mismatch for constraint: expected '%s', got '%s'", constraint.Name(), t.Name())
			return false
		}
	}
	return true
}

func (sa *SemanticAnalyzer) unwrapToCollection(t types.NRType) types.NRType {
	for {
		if pt, ok := t.(*types.PointerType); ok {
			if pt.IsArray {
				return pt
			}
			if pt.Leased {
				t = pt.Base
				continue
			}
		}
		break
	}
	return t
}

func (sa *SemanticAnalyzer) hasUnsafeAttr() bool {
	if sa.CurrentFunction != nil {
		return ast.GetAttribute(sa.CurrentFunction.Attributes, "unsafe") != nil
	}
	return false
}

func (sa *SemanticAnalyzer) isUnsafeAllowed(pos token.Position) bool {
	isAllowed := sa.AllowUnsafe
	if !isAllowed {
		filename := pos.Filename
		cleanFilename := filepath.Clean(filename)
		if runtime.GOOS == "windows" {
			cleanFilename = strings.ToLower(cleanFilename)
		}
		for _, dir := range sa.AllowedUnsafeDirs {
			if strings.HasPrefix(cleanFilename, dir) {
				isAllowed = true
				break
			}
		}
	}
	return isAllowed
}

func (sa *SemanticAnalyzer) checkWritePermission(expr ast.Expression) bool {
	if sa.isUnsafeAllowed(expr.Pos()) && sa.hasUnsafeAttr() {
		return true
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		sym := sa.SemanticInfo.Uses[e]
		if sym != nil {
			return sym.WritePerm
		}
		return true
	case *ast.SelectorExpression:
		if !sa.checkWritePermission(e.Left) {
			return false
		}
		baseType := sa.SemanticInfo.Types[e.Left]
		for {
			if pt, ok := baseType.(*types.PointerType); ok {
				if pt.Leased && pt.Kind == types.LeaseRead {
					return false
				}
				baseType = pt.Base
			} else {
				break
			}
		}
		return true
	case *ast.IndexExpression:
		if !sa.checkWritePermission(e.Left) {
			return false
		}
		baseType := sa.SemanticInfo.Types[e.Left]
		for {
			if pt, ok := baseType.(*types.PointerType); ok {
				if pt.IsArray {
					if pt.Leased && pt.Kind == types.LeaseRead {
						return false
					}
					baseType = pt.Base
					continue
				}
				if pt.Leased {
					if pt.Kind == types.LeaseRead {
						return false
					}
					baseType = pt.Base
					continue
				}
			}
			break
		}
		return true
	default:
		return true
	}
}

func isIntegerType(t types.NRType) bool {
	if t == nil {
		return false
	}
	name := t.Name()
	return name == "int" || name == "i32" || name == "i64" || name == "u64" || name == "i8" || name == "byte"
}

func (sa *SemanticAnalyzer) isHeapAllocated(t types.NRType) bool {
	if t == nil {
		return false
	}
	if _, ok := t.(*types.PointerType); ok {
		return true
	}
	if _, ok := t.(*types.ListType); ok {
		return true
	}
	if _, ok := t.(*types.MapType); ok {
		return true
	}
	if _, ok := t.(*types.ChanType); ok {
		return true
	}
	if t.Name() == "str" {
		return true
	}
	return false
}

func normalizePath(path string) string {
	p := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(p)
	}
	return p
}

func (sa *SemanticAnalyzer) checkUnusedSymbolsInScope(scope *Scope) {
	if scope == nil {
		return
	}
	for name, sym := range scope.Symbols {
		// Only check local variables and parameters
		if sym.Kind != SymVar && sym.Kind != SymParam {
			continue
		}
		// Skip variables that start with an underscore (escape hatch)
		if strings.HasPrefix(name, "_") {
			continue
		}
		// If it's not used, report a compiler warning!
		if !sym.IsUsed {
			var pos token.Position
			if sym.DefNode != nil {
				pos = sym.DefNode.Pos()
			} else {
				pos = token.Position{Line: 1, Column: 1, Filename: sa.GlobalScope.Symbols["print"].DefNode.Pos().Filename}
			}

			sa.Diagnostics.Add(diag.Diagnostic{
				Range: diag.Range{
					Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
					End:   diag.Position{Line: pos.Line, Column: pos.Column + len(name), Offset: pos.Offset + len(name)},
				},
				Severity: diag.Warning,
				Message:  fmt.Sprintf("unused variable '%s'", name),
				Source:   "Semantic",
				File:     pos.Filename,
				Hint:     fmt.Sprintf("if this is intentional, prefix the name with an underscore: '_%s'", name),
			})
		}
	}
}

func (sa *SemanticAnalyzer) AnalyzeFileTypes(file *ast.File) {
	if file == nil || file.Name == "" {
		return
	}
	packageName := sa.GetPackageName(file)
	prevScope := sa.CurrentScope
	sa.CurrentScope = sa.GetPackageScope(packageName)

	if strings.Contains(file.Name, "map.nr") {
		if sa.DebugMode {
			println("[DEBUG AnalyzeFileTypes] file:", file.Name)
			println("[DEBUG AnalyzeFileTypes] symbols in PackageScope:", packageName)
			for k, sym := range sa.CurrentScope.Symbols {
				println("  -", k, "kind:", sym.Kind, "type:", sym.Type.Name())
			}
		}
	}

	for _, stmt := range file.Statements {
		if ts, ok := stmt.(*ast.TypeStatement); ok {
			sa.Analyze(ts)
		}
	}

	sa.CurrentScope = prevScope
}

func (sa *SemanticAnalyzer) ensureMethodsSpecialized(t types.NRType) {
	if st, ok := t.(*types.SumType); ok && st.BaseType != nil {
		hasGenericArgs := false
		for _, arg := range st.TypeArgs {
			if sa.hasGeneric(arg) {
				hasGenericArgs = true
				break
			}
		}
		for mName, mType := range st.BaseType.Methods {
			if _, exists := st.Methods[mName]; !exists {
				if st.Methods == nil {
					st.Methods = make(map[string]types.NRType)
				}
				st.Methods[mName] = mType
			}
			if !hasGenericArgs {
				if methodSyms, ok := sa.SemanticInfo.MethodSymbols[st.BaseType]; ok {
					if methodSym, ok := methodSyms[mName]; ok {
						if methodFn, ok := methodSym.DefNode.(*ast.FunctionStatement); ok {
							if len(methodFn.TypeParameters) <= len(st.TypeArgs) {
								sa.Monomorphize(methodFn, st.TypeArgs, nil, st)
							}
						}
					}
				}
			}
		}
	} else if st, ok := t.(*types.StructType); ok && st.BaseType != nil {
		hasGenericArgs := false
		for _, arg := range st.TypeArgs {
			if sa.hasGeneric(arg) {
				hasGenericArgs = true
				break
			}
		}
		for mName, mType := range st.BaseType.Methods {
			if _, exists := st.Methods[mName]; !exists {
				if st.Methods == nil {
					st.Methods = make(map[string]types.NRType)
				}
				st.Methods[mName] = mType
			}
			if !hasGenericArgs {
				if methodSyms, ok := sa.SemanticInfo.MethodSymbols[st.BaseType]; ok {
					if methodSym, ok := methodSyms[mName]; ok {
						if methodFn, ok := methodSym.DefNode.(*ast.FunctionStatement); ok {
							if len(methodFn.TypeParameters) <= len(st.TypeArgs) {
								sa.Monomorphize(methodFn, st.TypeArgs, nil, st)
							}
						}
					}
				}
			}
		}
	}
}

func isStdOrCoreFile(filename string) bool {
	norm := filepath.ToSlash(filename)
	return strings.Contains(norm, "/std/") || strings.HasSuffix(norm, "/std") || strings.Contains(norm, "/core/") || strings.HasSuffix(norm, "/core")
}

func (sa *SemanticAnalyzer) checkUnusedProgramSymbols(prog *ast.Program) {
	if prog == nil || len(prog.Files) == 0 {
		return
	}
	entryPackage := sa.GetPackageName(prog.Files[0])

	for _, f := range prog.Files {
		if f == nil {
			continue
		}
		if isStdOrCoreFile(f.Name) {
			continue
		}
		packageName := sa.GetPackageName(f)
		if packageName != entryPackage {
			continue
		}
		scope := sa.GetPackageScope(packageName)
		if scope == nil {
			continue
		}

		for _, stmt := range f.Statements {
			if stmt == nil {
				continue
			}

			switch n := stmt.(type) {
			case *ast.ImportStatement:
				pkgPath := n.PathValue()
				parts := strings.Split(pkgPath, "/")
				name := parts[len(parts)-1]
				if n.Alias != nil {
					name = n.Alias.Value
				}

				if strings.HasPrefix(name, "_") {
					continue
				}

				if sym, found := scope.Symbols[name]; found {
					if sym.Kind == SymPackage && !sym.IsUsed {
						pos := n.Pos()
						sa.Diagnostics.Add(diag.Diagnostic{
							Range: diag.Range{
								Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
								End:   diag.Position{Line: pos.Line, Column: pos.Column + 6, Offset: pos.Offset + 6},
							},
							Severity: diag.Warning,
							Message:  fmt.Sprintf("unused import '%s'", name),
							Source:   "Semantic",
							File:     pos.Filename,
							Hint:     "remove the unused import",
						})
					}
				}

			case *ast.TypeStatement:
				if n.Name == nil || strings.HasPrefix(n.Name.Value, "_") {
					continue
				}

				if sym, found := scope.Symbols[n.Name.Value]; found {
					if sym.Kind == SymType && sym.Visible == Private && !sym.IsUsed {
						pos := n.Name.Pos()
						sa.Diagnostics.Add(diag.Diagnostic{
							Range: diag.Range{
								Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
								End:   diag.Position{Line: pos.Line, Column: pos.Column + len(n.Name.Value), Offset: pos.Offset + len(n.Name.Value)},
							},
							Severity: diag.Warning,
							Message:  fmt.Sprintf("unused type '%s'", n.Name.Value),
							Source:   "Semantic",
							File:     pos.Filename,
							Hint:     fmt.Sprintf("if this is intentional, prefix the name with an underscore: '_%s'", n.Name.Value),
						})
					}
				}

			case *ast.FunctionStatement:
				if n.Name == nil || strings.HasPrefix(n.Name.Value, "_") || n.Name.Value == "main" {
					continue
				}

				if sym, found := scope.Symbols[n.Name.Value]; found {
					if sym.Kind == SymFunc && sym.Visible == Private && !sym.IsUsed {
						pos := n.Name.Pos()
						sa.Diagnostics.Add(diag.Diagnostic{
							Range: diag.Range{
								Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
								End:   diag.Position{Line: pos.Line, Column: pos.Column + len(n.Name.Value), Offset: pos.Offset + len(n.Name.Value)},
							},
							Severity: diag.Warning,
							Message:  fmt.Sprintf("unused function '%s'", n.Name.Value),
							Source:   "Semantic",
							File:     pos.Filename,
							Hint:     fmt.Sprintf("if this is intentional, prefix the name with an underscore: '_%s'", n.Name.Value),
						})
					}
				}
			}
		}
	}
}

func sanitizeCIdentifier(name string) string {
	name = strings.ReplaceAll(name, "@", "")
	name = strings.ReplaceAll(name, "#", "")
	name = strings.ReplaceAll(name, "&", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, "[", "")
	name = strings.ReplaceAll(name, "]", "")
	name = strings.ReplaceAll(name, "*", "")
	return name
}

func (sa *SemanticAnalyzer) hasGeneric(t types.NRType) bool {
	if t == nil {
		return false
	}
	switch kt := t.(type) {
	case *types.GenericType:
		return true
	case *types.PointerType:
		return sa.hasGeneric(kt.Base)
	case *types.ListType:
		return sa.hasGeneric(kt.ElementType)
	case *types.MapType:
		return sa.hasGeneric(kt.Key) || sa.hasGeneric(kt.Value)
	case *types.StructType:
		if len(kt.TypeParams) > 0 {
			return true
		}
		for _, arg := range kt.TypeArgs {
			if sa.hasGeneric(arg) {
				return true
			}
		}
	case *types.SumType:
		if len(kt.TypeParams) > 0 {
			return true
		}
		for _, arg := range kt.TypeArgs {
			if sa.hasGeneric(arg) {
				return true
			}
		}
	case *types.ProtocolType:
		if len(kt.TypeParams) > 0 {
			return true
		}
		for _, arg := range kt.TypeArgs {
			if sa.hasGeneric(arg) {
				return true
			}
		}
	case *types.FunctionType:
		for _, param := range kt.Params {
			if sa.hasGeneric(param) {
				return true
			}
		}
		return sa.hasGeneric(kt.Return)
	}
	return t.GetKind() == types.KindGeneric
}

func (sa *SemanticAnalyzer) getMangledSerializationFuncName(funcName string, t types.NRType) string {
	t = types.UnwrapLease(t)
	for {
		if pt, ok := t.(*types.PointerType); ok {
			t = pt.Base
		} else {
			break
		}
	}

	tName := t.Name()
	st, ok := t.(*types.StructType)
	if !ok {
		// Primitives are in the serialize package
		return sanitizeCIdentifier("serialize_" + funcName + "_" + tName)
	}

	pkgName := ""
	found := false

	// Search sa.PackageScopes directly to find the original package where the type is defined
	for name, scope := range sa.PackageScopes {
		if sym, exists := scope.Symbols[st.TypeName]; exists {
			if sym.Kind == SymType && (sym.Type == st || (st.BaseType != nil && sym.Type == st.BaseType)) {
				pkgName = name
				found = true
				break
			}
		}
	}

	if !found {
		for name, scope := range sa.PackageScopes {
			if _, exists := scope.Symbols[st.TypeName]; exists {
				pkgName = name
				found = true
				break
			}
		}
	}

	if !found {
		if sym, exists := sa.CurrentScope.Resolve(st.TypeName); exists {
			if sym.DefScope != nil {
				s := sym.DefScope
				for s != nil {
					if s.Kind == ScopePackage {
						pkgName = s.PackageName
						found = true
						break
					}
					s = s.Parent
				}
			}
		}
	}

	targetFnName := funcName + "_" + tName
	if pkgName != "" && pkgName != "main" {
		safePkg := strings.ReplaceAll(pkgName, "/", "_")
		safePkg = strings.ReplaceAll(safePkg, ".", "_")
		targetFnName = safePkg + "_" + targetFnName
	}
	resName := sanitizeCIdentifier(targetFnName)
	if sa.DebugMode {
		if sa.DebugMode {
			fmt.Printf("[DEBUG-SERIALIZE] funcName=%s, type=%s, pkgName=%s, found=%v, resName=%s\n", funcName, tName, pkgName, found, resName)
		}
	}
	return resName
}

func (sa *SemanticAnalyzer) hasLease(t types.NRType, visited map[types.NRType]bool) bool {
	if t == nil {
		return false
	}
	if visited == nil {
		visited = make(map[types.NRType]bool)
	}
	if visited[t] {
		return false // prevent infinite loops on recursive types
	}
	visited[t] = true

	if pt, ok := t.(*types.PointerType); ok {
		if pt.Leased && pt.Kind != types.LeaseMove {
			return true
		}
		return sa.hasLease(pt.Base, visited)
	}
	if st, ok := t.(*types.StructType); ok {
		for _, ft := range st.Fields {
			if sa.hasLease(ft, visited) {
				return true
			}
		}
	}
	if sum, ok := t.(*types.SumType); ok {
		for _, vt := range sum.Variants {
			for _, vtt := range vt.Fields {
				if sa.hasLease(vtt, visited) {
					return true
				}
			}
		}
	}
	if lt, ok := t.(*types.ListType); ok {
		return sa.hasLease(lt.ElementType, visited)
	}
	if mt, ok := t.(*types.MapType); ok {
		return sa.hasLease(mt.Key, visited) || sa.hasLease(mt.Value, visited)
	}
	if ct, ok := t.(*types.ChanType); ok {
		return sa.hasLease(ct.Elem, visited)
	}
	return false
}

func (sa *SemanticAnalyzer) containsOwnedLease(t types.NRType, visited map[types.NRType]bool) bool {
	if t == nil {
		return false
	}
	if visited == nil {
		visited = make(map[types.NRType]bool)
	}
	if visited[t] {
		return false // prevent infinite loops on recursive types
	}
	visited[t] = true

	if pt, ok := t.(*types.PointerType); ok {
		if pt.Leased && pt.Kind == types.LeaseMove {
			return true
		}
		return sa.containsOwnedLease(pt.Base, visited)
	}
	if st, ok := t.(*types.StructType); ok {
		for _, ft := range st.Fields {
			if sa.containsOwnedLease(ft, visited) {
				return true
			}
		}
	}
	if sum, ok := t.(*types.SumType); ok {
		for _, vt := range sum.Variants {
			for _, vtt := range vt.Fields {
				if sa.containsOwnedLease(vtt, visited) {
					return true
				}
			}
		}
	}
	if lt, ok := t.(*types.ListType); ok {
		return sa.containsOwnedLease(lt.ElementType, visited)
	}
	if mt, ok := t.(*types.MapType); ok {
		return sa.containsOwnedLease(mt.Key, visited) || sa.containsOwnedLease(mt.Value, visited)
	}
	if ct, ok := t.(*types.ChanType); ok {
		return sa.containsOwnedLease(ct.Elem, visited)
	}
	return false
}

func (sa *SemanticAnalyzer) hasRestrictedClosure(t types.NRType, visited map[types.NRType]bool) bool {
	if t == nil {
		return false
	}
	if visited == nil {
		visited = make(map[types.NRType]bool)
	}
	if visited[t] {
		return false
	}
	visited[t] = true

	if ft, ok := t.(*types.FunctionType); ok {
		if ft.CapturesLease {
			return true
		}
	}
	if pt, ok := t.(*types.PointerType); ok {
		return sa.hasRestrictedClosure(pt.Base, visited)
	}
	if st, ok := t.(*types.StructType); ok {
		for _, ft := range st.Fields {
			if sa.hasRestrictedClosure(ft, visited) {
				return true
			}
		}
	}
	if sum, ok := t.(*types.SumType); ok {
		for _, vt := range sum.Variants {
			for _, vtt := range vt.Fields {
				if sa.hasRestrictedClosure(vtt, visited) {
					return true
				}
			}
		}
	}
	if lt, ok := t.(*types.ListType); ok {
		return sa.hasRestrictedClosure(lt.ElementType, visited)
	}
	if mt, ok := t.(*types.MapType); ok {
		return sa.hasRestrictedClosure(mt.Key, visited) || sa.hasRestrictedClosure(mt.Value, visited)
	}
	if ct, ok := t.(*types.ChanType); ok {
		return sa.hasRestrictedClosure(ct.Elem, visited)
	}
	return false
}

func (sa *SemanticAnalyzer) forceAnalyzeType(sym *Symbol) types.NRType {
	if sym == nil {
		return types.Void
	}
	if sym.Type != types.Void && sym.Type != nil {
		return sym.Type
	}
	if sym.DefNode == nil {
		return types.Void
	}

	if ts, ok := sym.DefNode.(*ast.TypeStatement); ok {
		if !sa.analyzingTypes[sym] {
			sa.analyzingTypes[sym] = true
			prevScope := sa.CurrentScope
			if sym.DefScope != nil {
				sa.CurrentScope = sym.DefScope
			}
			sa.Analyze(ts)
			sa.CurrentScope = prevScope
			delete(sa.analyzingTypes, sym)
		} else {
			return types.ErrorType
		}
	}
	return sym.Type
}

