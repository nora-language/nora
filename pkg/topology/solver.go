package topology

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"math"

	"github.com/nora-language/nora/pkg/diag"
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/token"
	"github.com/nora-language/nora/pkg/types"
)

const AnchorEndOfFunction = math.MaxInt32

type DropInfo struct {
	Symbol *semantic.Symbol
	Field  *ast.SelectorExpression
	Index  *ast.IndexExpression
	Lambda *ast.LambdaExpression
	Expr   ast.Expression
}

type Lifecycle struct {
	Symbol               *semantic.Symbol
	DefinedAt            int
	LastUsedAt           int
	IsMoved              bool
	IsConditionallyMoved bool
	MovedBy              ast.Node         // The node that caused the move
	AliasOf              *semantic.Symbol // The origin symbol if this is an alias (borrow)
	FieldMoves           map[string]bool  // Tracked field moves for this symbol
	IsExempt             bool             // If true, IsMoved is only for RAII (preventing double-drops), not for reporting violations
	IsMovedByTerminal    bool             // If true, the move was caused by a terminal statement (return, break, continue)
}

type Solver struct {
	SemanticInfo   *semantic.SemanticInfo
	Diagnostics    *diag.Collection
	Drops          map[*ast.BlockStatement]map[int][]DropInfo
	PreDrops       map[*ast.BlockStatement]map[int][]DropInfo
	AssignDrops    map[ast.Node]bool
	TryDrops       map[*ast.TryExpression][]DropInfo
	Dependencies   map[*semantic.Symbol][]*semantic.Symbol // provider -> dependents
	Providers      map[*semantic.Symbol][]*semantic.Symbol // dependent -> providers
	Moves          map[ast.Node]bool
	InSecondPass   bool // Track if we are in the second pass of a loop analysis
	DebugMode      bool
	LoopParentVars []map[*semantic.Symbol]bool
	CurrentFunction *types.FunctionType
}

func NewSolver(sem *semantic.SemanticInfo) *Solver {
	return &Solver{
		SemanticInfo:   sem,
		Diagnostics:    &diag.Collection{},
		Drops:          make(map[*ast.BlockStatement]map[int][]DropInfo),
		PreDrops:       make(map[*ast.BlockStatement]map[int][]DropInfo),
		AssignDrops:    make(map[ast.Node]bool),
		TryDrops:       make(map[*ast.TryExpression][]DropInfo),
		Dependencies:   make(map[*semantic.Symbol][]*semantic.Symbol),
		Providers:      make(map[*semantic.Symbol][]*semantic.Symbol),
		Moves:          make(map[ast.Node]bool),
		DebugMode:      false,
		LoopParentVars: []map[*semantic.Symbol]bool{},
	}
}

func (s *Solver) ReportError(pos token.Position, format string, args ...interface{}) {
	s.ReportErrorWithNotes(pos, fmt.Sprintf(format, args...), nil)
}

func (s *Solver) ReportErrorWithNotes(pos token.Position, msg string, notes []string) {
	if s.Diagnostics == nil {
		return
	}
	s.Diagnostics.Add(diag.Diagnostic{
		Range: diag.Range{
			Start: diag.Position{Line: pos.Line, Column: pos.Column, Offset: pos.Offset},
			End:   diag.Position{Line: pos.Line, Column: pos.Column + 1, Offset: pos.Offset + 1},
		},
		Severity: diag.Error,
		Message:  msg,
		Notes:    notes,
		Source:   "Topology",
		File:     pos.Filename,
	})
}

func (s *Solver) debug(format string, args ...interface{}) {
	if s.DebugMode {
		fmt.Fprintf(os.Stderr, "[Topology] "+format+"\n", args...)
	}
}

func isNilNode(node ast.Node) bool {
	if node == nil {
		return true
	}
	val := reflect.ValueOf(node)
	return val.Kind() == reflect.Ptr && val.IsNil()
}

func (s *Solver) Solve(node ast.Node) {
	if isNilNode(node) {
		return
	}
	switch n := node.(type) {
	case *ast.Program:
		for _, file := range n.Files {
			s.Solve(file)
		}
		for _, instances := range s.SemanticInfo.Instances {
			for _, inst := range instances {
				s.Solve(inst)
			}
		}
	case *ast.File:
		for _, stmt := range n.Statements {
			s.Solve(stmt)
		}
	case *ast.FunctionStatement:
		if len(n.TypeParameters) > 0 {
			return
		}
		if n.Body != nil {
			oldFn := s.CurrentFunction
			if sym := s.SemanticInfo.Defs[n.Name]; sym != nil {
				if fnT, ok := sym.Type.(*types.FunctionType); ok {
					s.CurrentFunction = fnT
				}
			}
			s.debug(">>> STARTING FUNCTION: %s", n.Name.Value)
			params := make(map[*semantic.Symbol]*Lifecycle)
			if n.Receiver != nil {
				if sym := s.SemanticInfo.Defs[n.Receiver.Name]; sym != nil {
					// CRITICAL: Do NOT drop the receiver in its own drop method,
					// as it is called by the RAII system which will handle the final free.
					// Handle monomorphized names like drop_i32
					// Handle monomorphized names like drop_i32
					if strings.HasPrefix(n.Name.Value, "drop") {
						// CRITICAL: Mark as 'IsMoved' so we don't register a recursive drop call,
						// AND 'IsExempt' so we allow usage inside the drop method itself.
						params[sym] = &Lifecycle{Symbol: sym, DefinedAt: -1, LastUsedAt: -1, IsMoved: true, IsExempt: true}
					} else {
						params[sym] = &Lifecycle{Symbol: sym, DefinedAt: -1, LastUsedAt: -1}
					}
				}
			}
			var isIntoRaw bool
			if fnSym, ok := s.SemanticInfo.Defs[n.Name]; ok && fnSym != nil {
				if fnSym.DefScope != nil && fnSym.DefScope.PackageName == "ffi" {
					if strings.HasPrefix(fnSym.Name, "IntoRaw") {
						isIntoRaw = true
					}
				}
			}

			for _, p := range n.Parameters {
				if sym := s.SemanticInfo.Defs[p.Name]; sym != nil {
					// Initialize to 0 so they are dropped at the start if unused,
					// or at the end if the block is empty. Exempt IntoRaw parameters.
					if isIntoRaw {
						params[sym] = &Lifecycle{Symbol: sym, DefinedAt: -1, LastUsedAt: -1, IsMoved: true, IsExempt: true}
					} else {
						params[sym] = &Lifecycle{Symbol: sym, DefinedAt: -1, LastUsedAt: 0}
					}
				}
			}
			s.analyzeBlock(n.Body, params)

			// Register drops for parameters in the function's main block if not moved
			for sym, lc := range params {
				if !lc.IsMoved && sym.Kind == semantic.SymParam {
					if s.isOwned(sym) {
						// Ensure it's dropped at least at the end of the block
						dropIdx := lc.LastUsedAt + 1
						if dropIdx > len(n.Body.Statements) {
							dropIdx = len(n.Body.Statements)
						}
						s.registerDrop(n.Body, dropIdx, DropInfo{Symbol: sym}, "PARAM")
					}
				}
			}
			s.CurrentFunction = oldFn
		}
	case *ast.LambdaExpression:
		if n.Body != nil {
			s.debug(">>> STARTING LAMBDA")
			params := make(map[*semantic.Symbol]*Lifecycle)
			for _, p := range n.Parameters {
				if sym := s.SemanticInfo.Defs[p.Name]; sym != nil {
					params[sym] = &Lifecycle{Symbol: sym, DefinedAt: -1, LastUsedAt: 0}
				}
			}
			s.analyzeBlock(n.Body, params)

			// Register drops for parameters in the lambda's body if not moved
			for sym, lc := range params {
				if !lc.IsMoved && sym.Kind == semantic.SymParam {
					if s.isOwned(sym) {
						dropIdx := lc.LastUsedAt + 1
						if dropIdx > len(n.Body.Statements) {
							dropIdx = len(n.Body.Statements)
						}
						s.registerDrop(n.Body, dropIdx, DropInfo{Symbol: sym}, "PARAM")
					}
				}
			}
		}
	}
}

func (s *Solver) isOwned(sym *semantic.Symbol) bool {
	if sym == nil {
		return false
	}
	// Check if the type itself is a lease (borrow)
	if pt, ok := sym.Type.(*types.PointerType); ok && pt.Leased {
		if pt.Kind == types.LeaseRead || pt.Kind == types.LeaseWrite {
			return false
		}
	}
	owned := types.IsOwnedType(sym.Type)
	return owned
}

func (s *Solver) registerDrop(block *ast.BlockStatement, index int, info DropInfo, reason string) {
	if s.InSecondPass {
		return
	}
	if s.Drops[block] == nil {
		s.Drops[block] = make(map[int][]DropInfo)
	}
	s.debug("      RAII: Registering DROP at index %d for %s (%s)", index, info.Symbol.Name, reason)
	s.Drops[block][index] = append(s.Drops[block][index], info)
}

func (s *Solver) registerPreDrop(block *ast.BlockStatement, index int, info DropInfo, reason string) {
	if s.InSecondPass {
		return
	}
	if s.PreDrops[block] == nil {
		s.PreDrops[block] = make(map[int][]DropInfo)
	}
	s.PreDrops[block][index] = append(s.PreDrops[block][index], info)
	if info.Symbol != nil {
		s.debug("      RAII: Registering PRE-DROP for %s at index %d (%s)", info.Symbol.Name, index, reason)
	} else if info.Field != nil {
		s.debug("      RAII: Registering FIELD PRE-DROP at index %d (%s)", index, reason)
	}
}

func (s *Solver) registerTryDrops(try *ast.TryExpression, visible map[*semantic.Symbol]*Lifecycle) {
	if s.InSecondPass {
		return
	}
	for sym, lc := range visible {
		if !lc.IsMoved && (sym.Kind == semantic.SymVar || sym.Kind == semantic.SymParam) {
			if s.isOwned(sym) {
				s.TryDrops[try] = append(s.TryDrops[try], DropInfo{Symbol: sym})
			}
		}
	}
}

func (s *Solver) analyzeBlock(block *ast.BlockStatement, trackedLifecycles map[*semantic.Symbol]*Lifecycle) {
	if block == nil {
		return
	}

	parentVars := make(map[*semantic.Symbol]bool)
	for sym := range trackedLifecycles {
		parentVars[sym] = true
	}

	localLifecycles := make(map[*semantic.Symbol]*Lifecycle)
	allVisible := trackedLifecycles

	for i, stmt := range block.Statements {
		if isNilNode(stmt) {
			continue
		}

		// A. VARIABLE BIRTH
		if varStmt, ok := stmt.(*ast.VarStatement); ok && !isNilNode(varStmt.Name) {
			sym := s.SemanticInfo.Defs[varStmt.Name]
			if sym != nil {
				lc := &Lifecycle{Symbol: sym, DefinedAt: i, LastUsedAt: i}

				// If the variable is initialized from indexing into a LEASED array,
				// the element is a borrowed reference — the variable does NOT own it.
				// Mark it as moved+exempt so the solver won't schedule a drop for it.
				if idx, ok := varStmt.Value.(*ast.IndexExpression); ok {
					srcType := s.SemanticInfo.Types[idx.Left]
					if pt, ok := srcType.(*types.PointerType); ok && pt.IsArray && pt.Leased && pt.Kind != types.LeaseMove {
						lc.IsMoved = true
						lc.IsExempt = true
						s.debug("      Variable Birth (BORROWED from leased array): %s at Index %d", sym.Name, i)
					}
				}

				// Rule 2: Receiver Lease Propagation for Method Returns
				if call, ok := varStmt.Value.(*ast.CallExpression); ok {
					fnTypeObj := s.SemanticInfo.Types[call.Function]
					if fn, ok := fnTypeObj.(*types.FunctionType); ok && fn.IsMethod {
						isReadBorrowReceiver := false
						if fn.ReceiverLease == types.LeaseRead {
							isReadBorrowReceiver = true
						} else if sel, ok := call.Function.(*ast.SelectorExpression); ok {
							recType := s.SemanticInfo.Types[sel.Left]
							if pt, ok := recType.(*types.PointerType); ok && pt.Leased && pt.Kind == types.LeaseRead {
								isReadBorrowReceiver = true
							}
						}
						isCollectionGet := false
						if sel, ok := call.Function.(*ast.SelectorExpression); ok {
							recType := s.SemanticInfo.Types[sel.Left]
							if recType != nil {
								ut := types.UnwrapLease(recType)
								if pt, ok := ut.(*types.PointerType); ok {
									ut = pt.Base
								}
								if st, ok := ut.(*types.StructType); ok && st.CoreIntrinsic == "Collection" {
									isCollectionGet = true
								}
							}
						}
						if isReadBorrowReceiver && sym.Type.Name() == "str" && isCollectionGet {
							lc.IsMoved = true
							lc.IsExempt = true
							s.debug("      Variable Birth (PROPAGATED BORROW from read-borrowed receiver): %s at Index %d", sym.Name, i)
						}
					}
				}

				if !lc.IsMoved {
					s.debug("      Variable Birth: %s at Index %d", sym.Name, i)
				}
				localLifecycles[sym] = lc
				allVisible[sym] = lc
			}
		}

		// B. PIN HANDLING (ANCHORING)
		if pinStmt, ok := stmt.(*ast.PinStatement); ok {
			for _, target := range pinStmt.Targets {
				targetSym := s.SemanticInfo.Uses[target]
				if targetSym == nil {
					for sym := range allVisible {
						if sym.Name == target.Value {
							targetSym = sym
							break
						}
					}
				}

				if targetSym != nil {
					if lc, exists := allVisible[targetSym]; exists {
						s.debug("      PIN detected for %s. Anchoring to end of block", targetSym.Name)
						lc.LastUsedAt = len(block.Statements) - 1
					}
				}
			}
		}

		// C. LEASE DEPENDENCY TRACKING (Initialization & Assignment)
		var assign *ast.AssignmentStatement
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			if a, ok := exprStmt.Expression.(*ast.AssignmentStatement); ok {
				assign = a
			}
		} else if a, ok := stmt.(*ast.AssignmentStatement); ok {
			assign = a
		}

		// [NEW] C. TRY EXPRESSION HANDLING
		ast.Inspect(stmt, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			switch n := node.(type) {
			case *ast.BlockStatement, *ast.IfExpression, *ast.MatchExpression, *ast.WhileStatement, *ast.ForStatement, *ast.ParallelExpression, *ast.SpawnExpression:
				return false
			case *ast.TryExpression:
				s.registerTryDrops(n, allVisible)
			}
			return true
		})

		// [NEW] D. GENERAL USAGE & PROPAGATION (Check BEFORE moves are recorded for this statement)
		usages := s.findUsagesInStatement(stmt)

		// Filter out LHS of assignment/var to prevent incorrect lifecycle extension
		var filteredUsages []*ast.Identifier
		for _, ident := range usages {
			isLHS := false
			if assign != nil {
				if id, ok := assign.Left.(*ast.Identifier); ok && id == ident {
					isLHS = true
				}
			}
			if varStmt, ok := stmt.(*ast.VarStatement); ok {
				if varStmt.Name == ident {
					isLHS = true
				}
			}
			if !isLHS {
				filteredUsages = append(filteredUsages, ident)
			}
		}

		for _, ident := range filteredUsages {
			sym := s.SemanticInfo.Uses[ident]
			if sym == nil {
				continue
			}

			// Check for use-after-move violation
			// Skip if in second pass and it's a return (can't happen twice)
			isRet := false
			if _, ok := stmt.(*ast.ReturnStatement); ok {
				isRet = true
			}

			isWriteLHS := false
			if assign != nil {
				if baseId := s.getBaseIdentifier(assign.Left); baseId != nil && baseId == ident {
					isWriteLHS = true
				}
			}

			if !isWriteLHS {
				if lc, exists := allVisible[sym]; exists && !lc.IsExempt {
					msg := ""
					notes := []string{}

					if lc.IsMoved || lc.IsConditionallyMoved {
						msg = fmt.Sprintf("use of moved value '%s'", ident.Value)
						if lc.MovedBy != nil {
							pos := lc.MovedBy.Pos()
							notes = append(notes, fmt.Sprintf("value moved here at %s:%d:%d", pos.Filename, pos.Line, pos.Column))
						}
					} else if lc.AliasOf != nil {
						if olc, ok := allVisible[lc.AliasOf]; ok && olc.IsMoved {
							msg = fmt.Sprintf("use of borrow '%s' whose origin '%s' was moved", ident.Value, lc.AliasOf.Name)
							if olc.MovedBy != nil {
								pos := olc.MovedBy.Pos()
								notes = append(notes, fmt.Sprintf("origin '%s' moved here at %s:%d:%d", lc.AliasOf.Name, pos.Filename, pos.Line, pos.Column))
							}
						}
					}

					// Check for partial moves (if not already fully moved)
					if msg == "" && s.isOwned(sym) {
						for fieldMove := range lc.FieldMoves {
							if strings.HasPrefix(fieldMove, sym.Name+".") {
								msg = fmt.Sprintf("use of partially moved value '%s' (field '%s' was moved)", sym.Name, fieldMove)
								break
							}
						}
					}

					if msg != "" {
						if s.InSecondPass && isRet {
							// Skip
						} else {
							s.ReportErrorWithNotes(ident.Pos(), msg, notes)
						}
					}
				}
			}

			visited := make(map[*semantic.Symbol]bool)
			s.updateLifecycle(sym, i, allVisible, visited)
		}

		// C. MOVE / CONSUMPTION
		// We MUST check moves BEFORE re-assignment drops to handle 'x = f(x)' correctly.
		// Skip for complex statements that handle their own internal moves via recursive analysis.
		isComplex := false
		if _, ok := stmt.(*ast.WhileStatement); ok {
			isComplex = true
		}
		if _, ok := stmt.(*ast.ForStatement); ok {
			isComplex = true
		}
		if es, ok := stmt.(*ast.ExpressionStatement); ok {
			if _, ok := es.Expression.(*ast.IfExpression); ok {
				isComplex = true
			}
			if _, ok := es.Expression.(*ast.MatchExpression); ok {
				isComplex = true
			}
		}
		if _, ok := stmt.(*ast.SelectStatement); ok {
			isComplex = true
		}
		if _, ok := stmt.(*ast.DeferStatement); ok {
			isComplex = true
		}

		if !isComplex {
			allIdents := s.findAllIdentsInStatement(stmt)
			for sym, ident := range allIdents {
				if s.isMoveOperation(stmt, ident) {
					if lc, exists := allVisible[sym]; exists {
						lc.IsMoved = true
						lc.MovedBy = stmt // Record the statement that caused the move
						s.debug("      MOVE DETECTED: %s is now consumed by stmt at line %d", sym.Name, ident.Pos().Line)
					} else {
						s.debug("      MOVE DETECTED BUT NOT IN VISIBLE: %s", sym.Name)
					}
				}
			}
		}

		// [NEW] Record all moves for codegen nullification
		s.recordMovesInStatement(stmt)

		// [NEW] Check for field moves in the entire statement
		ast.Inspect(stmt, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpression); ok {
				if s.isMoveOperationForSelector(stmt, sel) {
					selStr := s.stringifySelector(sel)
					rootSym := s.getRootSymbol(sel)
					if rootSym != nil {
						if isBorrowType(rootSym.Type) {
							s.ReportError(sel.Pos(), "cannot move out of borrowed context '%s'", rootSym.Name)
						} else if lc, exists := allVisible[rootSym]; exists {
							if lc.FieldMoves == nil {
								lc.FieldMoves = make(map[string]bool)
							}
							lc.FieldMoves[selStr] = true
							s.debug("      FIELD MOVE: %s is now consumed", selStr)
						}
					}
				}
			}
			return true
		})

		// B. MUTATION / ASSIGNMENT / BIRTH
		if varStmt, ok := stmt.(*ast.VarStatement); ok {
			sym := s.SemanticInfo.Defs[varStmt.Name]
			if sym != nil {
				s.trackDependencies(sym, varStmt.Value)
				// Alias Tracking
				if lc, exists := allVisible[sym]; exists {
					if origin := s.getAliasOrigin(varStmt.Value); origin != nil {
						lc.AliasOf = origin
						s.debug("      ALIAS: %s is now an alias of %s", sym.Name, origin.Name)
					}
				}
			}
		} else if assign != nil {
			if id, ok := assign.Left.(*ast.Identifier); ok {
				sym := s.SemanticInfo.Uses[id]
				if sym != nil {
					// RAII: If it's an owned variable and NOT moved, we must drop old value before overwrite
					if lc, exists := allVisible[sym]; exists {
						if !lc.IsMoved && s.isOwned(sym) {
							s.AssignDrops[assign] = true
						}
						s.debug("      RE-ASSIGNMENT: %s resetting IsMoved from %v to false", sym.Name, lc.IsMoved)
						lc.IsMoved = false                    // Reset move status for the new value
						lc.FieldMoves = make(map[string]bool) // Clear all field moves on parent overwrite
						lc.LastUsedAt = i

						// Alias Tracking
						lc.AliasOf = s.getAliasOrigin(assign.Value)
						if lc.AliasOf != nil {
							s.debug("      ALIAS: %s is now an alias of %s", sym.Name, lc.AliasOf.Name)
						}
					}

					// Clear existing providers for this symbol to handle re-assignment
					delete(s.Providers, sym)
					s.trackDependencies(sym, assign.Value)
					s.debug("      RE-ASSIGNMENT: %s dependencies cleared and updated", sym.Name)
				}
			} else if sel, ok := assign.Left.(*ast.SelectorExpression); ok {
				selStr := s.stringifySelector(sel)
				rootSym := s.getRootSymbol(sel)
				// RAII: Handle field re-assignment for owned types only
				t := s.SemanticInfo.Types[sel]
				if s.isImplicitMoveType(t) {
					if rootSym != nil {
						if lc, ok := allVisible[rootSym]; ok {
							if lc.FieldMoves == nil || !lc.FieldMoves[selStr] {
								s.AssignDrops[assign] = true
							}
						}
					}
				}
				// Reset moved state after re-assignment for any type
				if rootSym != nil {
					if lc, ok := allVisible[rootSym]; ok {
						if lc.FieldMoves != nil {
							delete(lc.FieldMoves, selStr)
						}
					}
				}
			} else if idx, ok := assign.Left.(*ast.IndexExpression); ok {
				// RAII: Handle array element re-assignment
				t := s.SemanticInfo.Types[idx]
				if s.isImplicitMoveType(t) {
					s.AssignDrops[assign] = true
				}
			}
		}

		// D. GENERAL USAGE & PROPAGATION
		// (Moved up)

		// E. TERMINAL HANDLING (Return, Break, Continue)
		if ret, ok := stmt.(*ast.ReturnStatement); ok {
			s.debug("      RETURN detected at index %d. Registering scope drops.", i)
			var exclude *semantic.Symbol
			if ret.ReturnValue != nil {
				// If we are returning an identifier directly, exclude it from drops
				if id, ok := ret.ReturnValue.(*ast.Identifier); ok {
					exclude = s.SemanticInfo.Uses[id]
				}
			}
			s.registerScopeDrops(block, i, allVisible, "TERMINAL (RETURN)", exclude, nil)
		} else if _, ok := stmt.(*ast.BreakStatement); ok {
			s.debug("      BREAK detected at index %d. Registering scope drops.", i)
			excludeParents := parentVars
			if len(s.LoopParentVars) > 0 {
				excludeParents = s.LoopParentVars[len(s.LoopParentVars)-1]
			}
			s.registerScopeDrops(block, i, allVisible, "TERMINAL (BREAK)", nil, excludeParents)
		} else if _, ok := stmt.(*ast.ContinueStatement); ok {
			s.debug("      CONTINUE detected at index %d. Registering scope drops.", i)
			excludeParents := parentVars
			if len(s.LoopParentVars) > 0 {
				excludeParents = s.LoopParentVars[len(s.LoopParentVars)-1]
			}
			s.registerScopeDrops(block, i, allVisible, "TERMINAL (CONTINUE)", nil, excludeParents)
		} else if branch, ok := stmt.(*ast.BranchStatement); ok {
			s.debug("      %s detected at index %d. Registering scope drops.", branch.Token.Literal, i)
			excludeParents := parentVars
			if len(s.LoopParentVars) > 0 {
				excludeParents = s.LoopParentVars[len(s.LoopParentVars)-1]
			}
			s.registerScopeDrops(block, i, allVisible, "TERMINAL ("+branch.Token.Literal+")", nil, excludeParents)
		}

		// [NEW] E. ANONYMOUS R-VALUE DROPS (Closures & Temporaries)
		if !s.InSecondPass {
			unconsumed := s.findUnconsumedRValues(stmt)
			for _, expr := range unconsumed {
				if s.Drops[block] == nil {
					s.Drops[block] = make(map[int][]DropInfo)
				}
				if lambda, ok := expr.(*ast.LambdaExpression); ok {
					s.Drops[block][i+1] = append(s.Drops[block][i+1], DropInfo{Lambda: lambda, Expr: lambda})
				} else {
					s.Drops[block][i+1] = append(s.Drops[block][i+1], DropInfo{Expr: expr})
				}
				s.debug("      Scheduled temporary drop for inline RValue at Index %d (%T)", i, expr)
			}
		}

		// F. RECURSIVE BLOCKS (If, While, For, Pin, etc.)
		ast.Inspect(stmt, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.LambdaExpression:
				s.Solve(n)
				return false

			case *ast.IfExpression:
				// Branch 1: Consequence
				branch1Lifecycles := s.cloneLifecycles(allVisible)
				s.analyzeBlock(n.Consequence, branch1Lifecycles)
				b1Term := s.isTerminalBlock(n.Consequence)

				if n.Alternative != nil {
					// Branch 2: Alternative
					branch2Lifecycles := s.cloneLifecycles(allVisible)
					var b2Term bool

					if altBlock, ok := n.Alternative.(*ast.BlockStatement); ok {
						s.analyzeBlock(altBlock, branch2Lifecycles)
						b2Term = s.isTerminalBlock(altBlock)
					} else if altIf, ok := n.Alternative.(*ast.IfExpression); ok {
						virtualBlock := &ast.BlockStatement{Statements: []ast.Statement{&ast.ExpressionStatement{Expression: altIf}}}
						s.analyzeBlock(virtualBlock, branch2Lifecycles)
						b2Term = s.isTerminalBlock(virtualBlock)
					}

					s.mergeBranchLifecycles(allVisible, []map[*semantic.Symbol]*Lifecycle{branch1Lifecycles, branch2Lifecycles}, []bool{b1Term, b2Term})
				} else {
					// No alternative: alternative branch is empty (non-terminal)
					branch2Lifecycles := s.cloneLifecycles(allVisible)
					s.mergeBranchLifecycles(allVisible, []map[*semantic.Symbol]*Lifecycle{branch1Lifecycles, branch2Lifecycles}, []bool{b1Term, false})
				}
				return false

			case *ast.MatchExpression:
				// Record moves in target
				ast.Inspect(n.Target, func(cn ast.Node) bool {
					if cn == nil {
						return false
					}
					if pref, ok := cn.(*ast.PrefixExpression); ok && pref.Operator == "@" {
						s.Moves[pref.Right] = true
						return true
					}
					if call, ok := cn.(*ast.CallExpression); ok {
						fnTypeObj := s.SemanticInfo.Types[call.Function]
						if fn, ok := fnTypeObj.(*types.FunctionType); ok {
							for i, arg := range call.Arguments {
								isMove := false
								if i < len(fn.ParamLeases) {
									if fn.ParamLeases[i] == types.LeaseMove {
										isMove = true
									}
								} else if i < len(fn.Params) && s.isImplicitMoveType(fn.Params[i]) && !s.isExternCall(call) {
									isMove = true
								}
								if isMove {
									if s.isMoveCandidate(arg.Value) {
										s.Moves[arg.Value] = true
									}
								}
							}
						}
					}
					return true
				})
				// General usage update
				for _, id := range s.findUsagesInNode(n.Target) {
					if sym := s.SemanticInfo.Uses[id]; sym != nil {
						s.updateLifecycle(sym, i, allVisible, make(map[*semantic.Symbol]bool))
					}
				}

				var branchMaps []map[*semantic.Symbol]*Lifecycle
				var isTerminal []bool
				for _, cas := range n.Cases {
					branchLifecycles := s.cloneLifecycles(allVisible)
					patternSyms := s.analyzePattern(cas.Pattern, branchLifecycles, i)
					s.analyzeBlock(cas.Body, branchLifecycles)

					// Register drops for pattern variables
					for _, sym := range patternSyms {
						lc := branchLifecycles[sym]
						if lc != nil && !lc.IsMoved && s.isOwned(sym) {
							dropIdx := lc.LastUsedAt + 1
							if cas.Body != nil && dropIdx > len(cas.Body.Statements) {
								dropIdx = len(cas.Body.Statements)
							}
							if cas.Body != nil {
								s.registerDrop(cas.Body, dropIdx, DropInfo{Symbol: sym}, "MATCH PATTERN VAR")
							}
						}
					}

					branchMaps = append(branchMaps, branchLifecycles)
					isTerminal = append(isTerminal, s.isTerminalBlock(cas.Body))
				}
				s.mergeBranchLifecycles(allVisible, branchMaps, isTerminal)
				return false

			case *ast.SelectStatement:
				s.debug("      SELECT statement detected at index %d", i)
				var branchMaps []map[*semantic.Symbol]*Lifecycle
				var isTerminal []bool
				for _, cas := range n.Cases {
					branchLifecycles := s.cloneLifecycles(allVisible)

					// Analyze Condition (Send/Receive)
					if cas.Condition != nil {
						usages := s.findUsagesInNode(cas.Condition)
						for _, id := range usages {
							if sym := s.SemanticInfo.Uses[id]; sym != nil {
								s.updateLifecycle(sym, i, branchLifecycles, make(map[*semantic.Symbol]bool))
							}
						}
						// Record moves in condition
						idents := s.findAllIdentsInStatement(cas.Condition)
						for sym, ident := range idents {
							if s.isMoveOperation(cas.Condition, ident) {
								if lc, exists := branchLifecycles[sym]; exists {
									lc.IsMoved = true
									lc.MovedBy = cas.Condition
								}
							}
						}
					}

					s.analyzeBlock(cas.Body, branchLifecycles)

					branchMaps = append(branchMaps, branchLifecycles)
					isTerminal = append(isTerminal, s.isTerminalBlock(cas.Body))
				}
				s.mergeBranchLifecycles(allVisible, branchMaps, isTerminal)
				return false

			case *ast.DeferStatement:
				s.debug("      DEFER statement detected at index %d", i)
				// All variables used in the defer call MUST stay alive until the end of the function.
				usages := s.findUsagesInNode(n.Call)
				for _, id := range usages {
					if sym := s.SemanticInfo.Uses[id]; sym != nil {
						if lc, exists := allVisible[sym]; exists {
							s.debug("      DEFER: Anchoring %s to end of scope", sym.Name)
							lc.LastUsedAt = AnchorEndOfFunction // Special value for 'end of function'
						}
					}
				}
				return false

			case *ast.WhileStatement:
				s.debug("      WHILE loop detected at index %d", i)
				// Extend lifecycles of variables used in condition
				conditionUsages := s.findUsagesInNode(n.Condition)
				for _, id := range conditionUsages {
					if sym := s.SemanticInfo.Uses[id]; sym != nil {
						visited := make(map[*semantic.Symbol]bool)
						s.updateLifecycle(sym, i, allVisible, visited)
					}
				}
				// Analyze body (Two passes to catch cross-iteration moves)
				bodyLifecycles := s.cloneLifecycles(allVisible)

				// Push loop parent variables to stack
				loopVars := make(map[*semantic.Symbol]bool)
				for sym := range allVisible {
					loopVars[sym] = true
				}
				s.LoopParentVars = append(s.LoopParentVars, loopVars)

				s.analyzeBlock(n.Body, bodyLifecycles) // Pass 1: Collect moves

				// Pass 2: Detect violations (with InSecondPass flag)
				s.InSecondPass = true
				s.analyzeBlock(n.Body, bodyLifecycles)
				s.InSecondPass = false

				// Pop loop parent variables from stack
				s.LoopParentVars = s.LoopParentVars[:len(s.LoopParentVars)-1]

				// Merge back: Loop can execute 0 or N times.
				// If moved in body in a NON-TERMINAL path, it's moved in parent.
				for sym, lc := range bodyLifecycles {
					if parentLc, exists := allVisible[sym]; exists {
						if lc.IsMoved && !lc.IsMovedByTerminal {
							parentLc.IsMoved = true
						}
						// Merge FieldMoves from loop body
						for fk, fv := range lc.FieldMoves {
							if fv {
								if parentLc.FieldMoves == nil {
									parentLc.FieldMoves = make(map[string]bool)
								}
								parentLc.FieldMoves[fk] = true
							}
						}
						if lc.LastUsedAt >= AnchorEndOfFunction {
							parentLc.LastUsedAt = AnchorEndOfFunction
						} else if lc.LastUsedAt >= 0 {
							if i > parentLc.LastUsedAt {
								parentLc.LastUsedAt = i
							}
						}
					}
				}
				return false

			case *ast.ForStatement:
				s.debug("      FOR loop detected at index %d", i)

				loopUsages := s.findUsagesInNode(n.Iterable)
				for _, id := range loopUsages {
					if sym := s.SemanticInfo.Uses[id]; sym != nil {
						visited := make(map[*semantic.Symbol]bool)
						s.updateLifecycle(sym, i, allVisible, visited)
					}
				}
				// Analyze body (Two passes to catch cross-iteration moves)
				bodyLifecycles := s.cloneLifecycles(allVisible)

				// Push loop parent variables to stack
				loopVars := make(map[*semantic.Symbol]bool)
				for sym := range allVisible {
					loopVars[sym] = true
				}
				s.LoopParentVars = append(s.LoopParentVars, loopVars)

				s.analyzeBlock(n.Body, bodyLifecycles) // Pass 1
				s.InSecondPass = true
				s.analyzeBlock(n.Body, bodyLifecycles) // Pass 2
				s.InSecondPass = false

				// Pop loop parent variables from stack
				s.LoopParentVars = s.LoopParentVars[:len(s.LoopParentVars)-1]
				for sym, lc := range bodyLifecycles {
					if parentLc, exists := allVisible[sym]; exists {
						if lc.IsMoved && !lc.IsMovedByTerminal {
							parentLc.IsMoved = true
						}
						// Merge FieldMoves from loop body
						for fk, fv := range lc.FieldMoves {
							if fv {
								if parentLc.FieldMoves == nil {
									parentLc.FieldMoves = make(map[string]bool)
								}
								parentLc.FieldMoves[fk] = true
							}
						}
						if lc.LastUsedAt >= AnchorEndOfFunction {
							parentLc.LastUsedAt = AnchorEndOfFunction
						} else if lc.LastUsedAt >= 0 {
							if i > parentLc.LastUsedAt {
								parentLc.LastUsedAt = i
							}
						}
					}
				}
				return false

			case *ast.ParallelExpression:
				s.debug("      PARALLEL block detected at index %d", i)
				captured := s.findCapturedVariables(n.Body)
				for _, sym := range captured {
					if lc, exists := allVisible[sym]; exists {
						s.debug("      CAPTURE: %s is captured by parallel block. Anchoring to current index", sym.Name)
						lc.LastUsedAt = i
					}
				}
				s.analyzeBlock(n.Body, s.cloneLifecycles(allVisible))
				return false

			case *ast.SpawnExpression:
				s.debug("      SPAWN expression detected at index %d", i)
				if n.Body != nil {
					captured := s.findCapturedVariables(n.Body)
					for _, sym := range captured {
						if lc, exists := allVisible[sym]; exists {
							s.debug("      CAPTURE: %s is captured by spawn block. Anchoring to end of function", sym.Name)
							lc.LastUsedAt = AnchorEndOfFunction
						}
					}
					s.analyzeBlock(n.Body, s.cloneLifecycles(allVisible))
				}
				if n.Call != nil {
					// We must still analyze the call arguments for dependencies/usages
					// Since we return false, we do it here manually.
					for _, id := range s.findUsagesInNode(n.Call) {
						if sym := s.SemanticInfo.Uses[id]; sym != nil {
							if lc, exists := allVisible[sym]; exists {
								s.debug("      CAPTURE: %s is captured by spawn call. Anchoring to end of function", sym.Name)
								lc.LastUsedAt = AnchorEndOfFunction
							}
							s.updateLifecycle(sym, i, allVisible, make(map[*semantic.Symbol]bool))
						}
					}
				}
				return false

			case *ast.BlockStatement:
				if node != block {
					s.analyzeBlock(node.(*ast.BlockStatement), s.cloneLifecycles(allVisible))
				}
				return false
			}
			return true
		})
	}
	s.refineLifecyclesWithNLL(block, allVisible)

	for _, lc := range localLifecycles {
		if lc.Symbol.Kind == semantic.SymVar || lc.Symbol.Kind == semantic.SymParam {
			if (!lc.IsMoved || lc.IsConditionallyMoved) && s.isOwned(lc.Symbol) {
				dropIdx := lc.LastUsedAt + 1
				if dropIdx > len(block.Statements) {
					dropIdx = len(block.Statements)
				}
				s.registerDrop(block, dropIdx, DropInfo{Symbol: lc.Symbol}, "SCOPE END")
			}
		}
	}

}

func (s *Solver) refineLifecyclesWithNLL(block *ast.BlockStatement, visible map[*semantic.Symbol]*Lifecycle) {
	if block == nil || len(block.Statements) == 0 {
		return
	}

	// 1. Build CFG for the block
	builder := NewCFGBuilder(s.SemanticInfo)
	start, exit := builder.Build(block)

	// 2. Collect all nodes in CFG
	var allNodes []*CFGNode
	visited := make(map[int]bool)
	var collect func(node *CFGNode)
	collect = func(node *CFGNode) {
		if node == nil || visited[node.ID] {
			return
		}
		visited[node.ID] = true
		if node.ID != -1 {
			allNodes = append(allNodes, node)
		}
		for _, succ := range node.Succs {
			collect(succ)
		}
	}
	collect(start)

	// 3. Solve liveness
	SolveLiveness(allNodes, exit)

	// 4. Map each statement to its index in the block
	stmtToIndex := make(map[ast.Statement]int)
	for i, stmt := range block.Statements {
		stmtToIndex[stmt] = i
	}

	// 5. Update LastUsedAt for all visible variables based on liveness!
	for _, node := range allNodes {
		if node.Stmt == nil {
			continue
		}
		idx, exists := stmtToIndex[node.Stmt]
		if !exists {
			continue
		}

		// For each symbol in the node's LiveIn set, update LastUsedAt
		for sym, live := range node.LiveIn {
			if live {
				if lc, ok := visible[sym]; ok {
					if idx > lc.LastUsedAt && lc.LastUsedAt < AnchorEndOfFunction {
						lc.LastUsedAt = idx
					}
				}
			}
		}
	}
}

func (s *Solver) isTerminalBlock(block *ast.BlockStatement) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Statements {
		switch stmt.(type) {
		case *ast.ReturnStatement, *ast.BreakStatement, *ast.ContinueStatement:
			return true
		}
	}
	return false
}

func (s *Solver) registerScopeDrops(block *ast.BlockStatement, index int, visible map[*semantic.Symbol]*Lifecycle, reason string, exclude *semantic.Symbol, excludeParents map[*semantic.Symbol]bool) {
	excluded := make(map[*semantic.Symbol]bool)
	if exclude != nil {
		excluded[exclude] = true
	}

	for sym, lc := range visible {
		s.debug("      Checking visibility: %s (kind=%v, moved=%v)", sym.Name, sym.Kind, lc.IsMoved)
		if excluded[sym] || (excludeParents != nil && excludeParents[sym]) {
			continue
		}
		if sym.Kind == semantic.SymVar || sym.Kind == semantic.SymParam {
			if !lc.IsMoved && s.isOwned(sym) {
				s.registerPreDrop(block, index, DropInfo{Symbol: sym}, reason)
			}
			lc.IsMoved = true           // Prevent double-drop statically in subsequent unreachable code
			lc.IsMovedByTerminal = true // Indicate this move is due to a return/break/continue
		}
	}
}

func (s *Solver) collectProviders(sym *semantic.Symbol, result map[*semantic.Symbol]bool) {
	if result[sym] {
		return
	}
	result[sym] = true
	for _, provider := range s.Providers[sym] {
		s.collectProviders(provider, result)
	}
}

func (s *Solver) trackDependencies(sym *semantic.Symbol, value ast.Expression) {
	ast.Inspect(value, func(node ast.Node) bool {
		if id, ok := node.(*ast.Identifier); ok {
			provider := s.SemanticInfo.Uses[id]
			if provider != nil && provider != sym {
				s.Dependencies[provider] = append(s.Dependencies[provider], sym)
				s.Providers[sym] = append(s.Providers[sym], provider)
				s.debug("      DEPENDENCY: %s now depends on %s", sym.Name, provider.Name)
			}
		}
		return true
	})
}

func (s *Solver) updateLifecycle(sym *semantic.Symbol, index int, visible map[*semantic.Symbol]*Lifecycle, visited map[*semantic.Symbol]bool) {
	lc, exists := visible[sym]
	if !exists || visited[sym] {
		return
	}
	visited[sym] = true

	if index > lc.LastUsedAt {
		s.debug("      Lifecycle Update: %s extended to index %d", sym.Name, index)
		lc.LastUsedAt = index

		for _, provider := range s.Providers[sym] {
			s.updateLifecycle(provider, index, visible, visited)
		}
	}
}

func (s *Solver) findUsagesInStatement(stmt ast.Statement) []*ast.Identifier {
	var ids []*ast.Identifier
	ast.Inspect(stmt, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		if id, ok := n.(*ast.Identifier); ok {
			ids = append(ids, id)
		}
		return true
	})
	return ids
}

func (s *Solver) cloneLifecycles(orig map[*semantic.Symbol]*Lifecycle) map[*semantic.Symbol]*Lifecycle {
	clone := make(map[*semantic.Symbol]*Lifecycle)
	for k, v := range orig {
		clone[k] = &Lifecycle{
			Symbol:               v.Symbol,
			DefinedAt:            v.DefinedAt,
			LastUsedAt:           v.LastUsedAt,
			IsMoved:              v.IsMoved,
			IsConditionallyMoved: v.IsConditionallyMoved,
			IsExempt:             v.IsExempt,
			IsMovedByTerminal:    v.IsMovedByTerminal,
			MovedBy:              v.MovedBy,
			AliasOf:              v.AliasOf,
			FieldMoves:           make(map[string]bool),
		}
		for fk, fv := range v.FieldMoves {
			clone[k].FieldMoves[fk] = fv
		}
	}
	return clone
}

func (s *Solver) mergeBranchLifecycles(parent map[*semantic.Symbol]*Lifecycle, branches []map[*semantic.Symbol]*Lifecycle, isTerminal []bool) {
	var activeBranches []map[*semantic.Symbol]*Lifecycle
	for idx, b := range branches {
		if !isTerminal[idx] {
			activeBranches = append(activeBranches, b)
		}
	}

	if len(activeBranches) == 0 {
		return
	}

	for sym, parentLc := range parent {
		moveCount := 0
		condMoveCount := 0
		var lastMovedBy ast.Node
		var lastFieldMoves []map[string]bool
		lastUsedAt := parentLc.LastUsedAt

		for _, b := range activeBranches {
			if lc, exists := b[sym]; exists {
				if lc.IsMoved {
					moveCount++
					lastMovedBy = lc.MovedBy
				} else if lc.IsConditionallyMoved {
					condMoveCount++
				}
				if len(lc.FieldMoves) > 0 {
					lastFieldMoves = append(lastFieldMoves, lc.FieldMoves)
				}
				if lc.LastUsedAt > lastUsedAt {
					lastUsedAt = lc.LastUsedAt
				}
			}
		}

		if moveCount == len(activeBranches) {
			parentLc.IsMoved = true
			parentLc.IsConditionallyMoved = false
			parentLc.MovedBy = lastMovedBy
		} else if moveCount > 0 || condMoveCount > 0 {
			parentLc.IsMoved = false
			parentLc.IsConditionallyMoved = true
		} else {
			parentLc.IsMoved = false
			parentLc.IsConditionallyMoved = false
		}

		for _, fm := range lastFieldMoves {
			for fk, fv := range fm {
				if fv {
					if parentLc.FieldMoves == nil {
						parentLc.FieldMoves = make(map[string]bool)
					}
					parentLc.FieldMoves[fk] = true
				}
			}
		}

		parentLc.LastUsedAt = lastUsedAt
	}
}

func (s *Solver) findAllIdentsInStatement(stmt ast.Statement) map[*semantic.Symbol]*ast.Identifier {
	idents := make(map[*semantic.Symbol]*ast.Identifier)
	ast.Inspect(stmt, func(n ast.Node) bool {
		if id, ok := n.(*ast.Identifier); ok {
			if sym := s.SemanticInfo.Uses[id]; sym != nil {
				idents[sym] = id
			}
		}
		return true
	})
	return idents
}

func (s *Solver) stringifySelector(sel *ast.SelectorExpression) string {
	var left string
	if id, ok := sel.Left.(*ast.Identifier); ok {
		left = id.Value
	} else if subSel, ok := sel.Left.(*ast.SelectorExpression); ok {
		left = s.stringifySelector(subSel)
	}
	return left + "." + sel.Field.Value
}

func (s *Solver) isMoveOperationForSelector(stmt ast.Statement, target *ast.SelectorExpression) bool {
	isMove := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if isMove || n == nil {
			return false
		}
		if _, ok := n.(*ast.BlockStatement); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.IfExpression); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.MatchExpression); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.WhileStatement); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.ForStatement); ok && n != stmt {
			return false
		}

		// 1. Function Calls
		if call, ok := n.(*ast.CallExpression); ok {
			fnTypeObj := s.SemanticInfo.Types[call.Function]
			if fn, ok := fnTypeObj.(*types.FunctionType); ok {
				for i, arg := range call.Arguments {
					if pref, ok := arg.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
						if s.isSameSelector(pref.Right, target) {
							isMove = true
							return false
						}
					}
					if s.isSameSelector(arg.Value, target) {
						isArgMove := false
						if i < len(fn.ParamLeases) && fn.ParamLeases[i] == types.LeaseMove {
							isArgMove = true
						} else if i < len(fn.Params) && s.isImplicitMoveType(fn.Params[i]) && !s.isExternCall(call) {
							isArgMove = true
						}
						if isArgMove {
							isMove = true
							return false
						}
					}
				}
				// Handle method receiver move
				if fn.IsMethod && (fn.ReceiverLease == types.LeaseMove || s.isImplicitMoveReceiver(fn.Receiver)) {
					if sel, ok := call.Function.(*ast.SelectorExpression); ok {
						if s.isSameSelector(sel.Left, target) {
							isMove = true
							return false
						}
					}
				}
			} else {
				// Special case: unchecked_set consumes its second argument
				if sel, ok := call.Function.(*ast.SelectorExpression); ok && sel.Field.Value == "unchecked_set" {
					if len(call.Arguments) == 2 {
						if s.isSameSelector(call.Arguments[1].Value, target) {
							isMove = true
							return false
						}
					}
				}
			}
		}

		// 2. Variable Declarations
		if varStmt, ok := n.(*ast.VarStatement); ok {
			if pref, ok := varStmt.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if s.isSameSelector(pref.Right, target) {
					isMove = true
					return false
				}
			}
			if s.isSameSelector(varStmt.Value, target) {
				sym := s.SemanticInfo.Defs[varStmt.Name]
				if sym != nil && types.IsOwnedType(sym.Type) && (!sym.Type.IsLeased() || sym.Type.(*types.PointerType).Kind == types.LeaseMove) {
					isMove = true
					return false
				}
			}
		}

		// 3. Assignments
		if assign, ok := n.(*ast.AssignmentStatement); ok {
			if pref, ok := assign.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if s.isSameSelector(pref.Right, target) {
					isMove = true
					return false
				}
			}
			if s.isSameSelector(assign.Value, target) {
				targetType := s.SemanticInfo.Types[assign.Left]
				sourceType := s.SemanticInfo.Types[assign.Value]
				if s.isImplicitMoveType(targetType) && s.isImplicitMoveType(sourceType) {
					isMove = true
					return false
				}
			}
		}

		// 4. Return Statement
		if ret, ok := n.(*ast.ReturnStatement); ok && ret.ReturnValue != nil {
			if pref, ok := ret.ReturnValue.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if s.isSameSelector(pref.Right, target) {
					isMove = true
					return false
				}
			}
			if s.isSameSelector(ret.ReturnValue, target) {
				if s.CurrentFunction != nil && s.CurrentFunction.Return != nil && s.CurrentFunction.Return.IsLeased() {
					// Function returns a lease, do not treat as move
				} else {
					sourceType := s.SemanticInfo.Types[ret.ReturnValue]
					if s.isImplicitMoveType(sourceType) {
						isMove = true
						return false
					}
				}
			}
		}

		// 5. Send Expression: c <- @x
		if send, ok := n.(*ast.SendExpression); ok {
			if pref, ok := send.Right.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if s.isSameSelector(pref.Right, target) {
					isMove = true
					return false
				}
			}
			// Implicit move if right side is owned
			if s.isSameSelector(send.Right, target) {
				t := s.SemanticInfo.Types[target]
				if s.isImplicitMoveType(t) {
					isMove = true
					return false
				}
			}
		}

		return true
	})
	return isMove
}

func (s *Solver) recordMovesInStatement(stmt ast.Statement) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// 1. Explicit move operator: @x
		if pref, ok := n.(*ast.PrefixExpression); ok && pref.Operator == "@" {
			s.Moves[pref.Right] = true
			return true
		}

		// Implicit moves in struct instantiation: JsonProperty{ name: key }
		if structLit, ok := n.(*ast.StructLiteral); ok {
			structTypeObj := s.SemanticInfo.Types[structLit]
			if st, ok := structTypeObj.(*types.StructType); ok {
				for _, field := range structLit.Fields {
					fName := field.Name.Value
					fExpr := field.Value
					if fExpr != nil {
						fType := st.Fields[fName]
						if s.isImplicitMoveType(fType) {
							if s.isMoveCandidate(fExpr) {
								s.Moves[fExpr] = true
							}
						}
					}
				}
			}
		}

		// Implicit moves in array literal elements: [x, y]
		if arrLit, ok := n.(*ast.ArrayLiteral); ok {
			arrTypeObj := s.SemanticInfo.Types[arrLit]
			if arrType, ok := arrTypeObj.(*types.ListType); ok {
				if s.isImplicitMoveType(arrType.ElementType) {
					for _, elem := range arrLit.Elements {
						if s.isMoveCandidate(elem) {
							s.Moves[elem] = true
						}
					}
				}
			}
		}

		// Implicit moves in map literal keys/values: { k: v }
		if mapLit, ok := n.(*ast.MapLiteral); ok {
			mapTypeObj := s.SemanticInfo.Types[mapLit]
			if mapType, ok := mapTypeObj.(*types.MapType); ok {
				isKeyOwned := s.isImplicitMoveType(mapType.Key)
				isValueOwned := s.isImplicitMoveType(mapType.Value)
				for k, v := range mapLit.Pairs {
					if isKeyOwned && s.isMoveCandidate(k) {
						s.Moves[k] = true
					}
					if isValueOwned && s.isMoveCandidate(v) {
						s.Moves[v] = true
					}
				}
			}
		}

		// 2. Implicit moves in assignment: x = y
		if assign, ok := n.(*ast.AssignmentStatement); ok {
			targetType := s.SemanticInfo.Types[assign.Left]
			if targetType == nil {
				if id, ok := assign.Left.(*ast.Identifier); ok {
					if sym := s.SemanticInfo.Uses[id]; sym != nil {
						targetType = sym.Type
					} else if sym := s.SemanticInfo.Defs[id]; sym != nil {
						targetType = sym.Type
					}
				}
			}
			sourceType := s.SemanticInfo.Types[assign.Value]
			if s.isImplicitMoveType(targetType) && s.isImplicitMoveType(sourceType) {
				if s.isMoveCandidate(assign.Value) {
					s.Moves[assign.Value] = true
				}
			}
		}

		// 3. Implicit moves in var declaration: var x = y
		if varStmt, ok := n.(*ast.VarStatement); ok && varStmt.Value != nil {
			sym := s.SemanticInfo.Defs[varStmt.Name]
			sourceType := s.SemanticInfo.Types[varStmt.Value]
			if sym != nil && s.isImplicitMoveType(sym.Type) && s.isImplicitMoveType(sourceType) {
				if s.isMoveCandidate(varStmt.Value) {
					s.Moves[varStmt.Value] = true
				}
			}
		}

		// 4. Implicit moves in call arguments: f(y)
		if call, ok := n.(*ast.CallExpression); ok {
			fnTypeObj := s.SemanticInfo.Types[call.Function]
			if fn, ok := fnTypeObj.(*types.FunctionType); ok {
				for i, arg := range call.Arguments {
					isMove := false
					if i < len(fn.ParamLeases) {
						if fn.ParamLeases[i] == types.LeaseMove {
							isMove = true
						}
					} else if i < len(fn.Params) && s.isImplicitMoveType(fn.Params[i]) && !s.isExternCall(call) {
						isMove = true
					}
					if isMove {
						if s.isMoveCandidate(arg.Value) {
							s.Moves[arg.Value] = true
						}
					}
				}
				// Handle method receiver move
				if fn.IsMethod && (fn.ReceiverLease == types.LeaseMove || s.isImplicitMoveReceiver(fn.Receiver)) {
					if sel, ok := call.Function.(*ast.SelectorExpression); ok {
						if s.isMoveCandidate(sel.Left) {
							s.Moves[sel.Left] = true
						}
					}
				}
			} else {
				// Special case: unchecked_set consumes its second argument
				if sel, ok := call.Function.(*ast.SelectorExpression); ok && sel.Field.Value == "unchecked_set" {
					if len(call.Arguments) == 2 {
						if s.isMoveCandidate(call.Arguments[1].Value) {
							s.Moves[call.Arguments[1].Value] = true
						}
					}
				}
			}
		}

		// 5. Implicit moves in return: return y
		if ret, ok := n.(*ast.ReturnStatement); ok && ret.ReturnValue != nil {
			if s.CurrentFunction != nil && s.CurrentFunction.Return != nil && s.CurrentFunction.Return.IsLeased() {
				// Borrow, not a move
			} else {
				sourceType := s.SemanticInfo.Types[ret.ReturnValue]
				if s.isImplicitMoveType(sourceType) {
					if s.isMoveCandidate(ret.ReturnValue) {
						s.Moves[ret.ReturnValue] = true
					}
				}
			}
		}

		// 6. Spawn
		if spawn, ok := n.(*ast.SpawnExpression); ok {
			for _, arg := range spawn.Call.Arguments {
				t := s.SemanticInfo.Types[arg.Value]
				if s.isImplicitMoveType(t) && s.isMoveCandidate(arg.Value) {
					s.Moves[arg.Value] = true
				}
			}
			// Receiver
			fnTypeObj := s.SemanticInfo.Types[spawn.Call.Function]
			if fn, ok := fnTypeObj.(*types.FunctionType); ok && fn.IsMethod {
				if sel, ok := spawn.Call.Function.(*ast.SelectorExpression); ok {
					if s.isMoveCandidate(sel.Left) {
						s.Moves[sel.Left] = true
					}
				}
			}
		}

		// 7. Send: c <- @x
		if send, ok := n.(*ast.SendExpression); ok {
			if pref, ok := send.Right.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				s.Moves[pref.Right] = true
			} else if s.isImplicitMoveType(s.SemanticInfo.Types[send.Right]) {
				if s.isMoveCandidate(send.Right) {
					s.Moves[send.Right] = true
				}
			}
		}

		return true
	})
}

func (s *Solver) isMoveCandidate(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.Identifier, *ast.SelectorExpression:
		return true
	case *ast.PrefixExpression:
		if e.Operator == "@" {
			return true
		}
	}
	return false
}

func (s *Solver) isSelectorUsedIn(expr ast.Expression, target *ast.SelectorExpression) bool {
	used := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if used || n == nil {
			return false
		}
		if s.isSameSelector(n, target) {
			used = true
			return false
		}
		return true
	})
	return used
}

func (s *Solver) isSelectorMovedIn(expr ast.Expression, target *ast.SelectorExpression) bool {
	isMove := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if isMove || n == nil {
			return false
		}

		// Check if 'n' is a move of 'target'
		if call, ok := n.(*ast.CallExpression); ok {
			fnTypeObj := s.SemanticInfo.Types[call.Function]
			if fn, ok := fnTypeObj.(*types.FunctionType); ok {
				for i, arg := range call.Arguments {
					// Explicit move operator @curr.next
					if pref, ok := arg.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
						if s.isSameSelector(pref.Right, target) {
							isMove = true
							return false
						}
					}
					// Implicit move by parameter lease
					if s.isSameSelector(arg.Value, target) {
						isArgMove := false
						if i < len(fn.ParamLeases) {
							if fn.ParamLeases[i] == types.LeaseMove {
								isArgMove = true
							}
						} else if i < len(fn.Params) && s.isImplicitMoveType(fn.Params[i]) && !s.isExternCall(call) {
							isArgMove = true
						}
						if isArgMove {
							isMove = true
							return false
						}
					}
				}
			} else {
				// Special case: unchecked_set consumes its second argument
				if sel, ok := call.Function.(*ast.SelectorExpression); ok && sel.Field.Value == "unchecked_set" {
					if len(call.Arguments) == 2 {
						if s.isSameSelector(call.Arguments[1].Value, target) {
							isMove = true
							return false
						}
					}
				}
			}
		}

		return true
	})
	return isMove
}

func (s *Solver) getAliasOrigin(expr ast.Expression) *semantic.Symbol {
	if pref, ok := expr.(*ast.PrefixExpression); ok && (pref.Operator == "#" || pref.Operator == "&") {
		if id, ok := pref.Right.(*ast.Identifier); ok {
			return s.SemanticInfo.Uses[id]
		}
		// Handle nested aliases: var c = #b (where b is alias of a)
		if sel, ok := pref.Right.(*ast.SelectorExpression); ok {
			// For now, only track root symbol as origin
			root := sel.Left
			for {
				if sub, ok := root.(*ast.SelectorExpression); ok {
					root = sub.Left
				} else {
					break
				}
			}
			if id, ok := root.(*ast.Identifier); ok {
				return s.SemanticInfo.Uses[id]
			}
		}
	}
	// Handle simple alias without operator (if allowed by type system)
	if id, ok := expr.(*ast.Identifier); ok {
		t := s.SemanticInfo.Types[id]
		if pt, ok := t.(*types.PointerType); ok && pt.Leased {
			return s.SemanticInfo.Uses[id]
		}
	}
	return nil
}

func (s *Solver) isSameSelector(e1, e2 ast.Node) bool {
	// Base Case: Identifier
	if id1, ok1 := e1.(*ast.Identifier); ok1 {
		if id2, ok2 := e2.(*ast.Identifier); ok2 {
			return id1.Value == id2.Value
		}
	}

	// Recursive Case: Selector
	s1, ok1 := e1.(*ast.SelectorExpression)
	s2, ok2 := e2.(*ast.SelectorExpression)
	if !ok1 || !ok2 {
		return false
	}
	if s1.Field.Value != s2.Field.Value {
		return false
	}
	return s.isSameSelector(s1.Left, s2.Left)
}

func (s *Solver) isExternCall(call *ast.CallExpression) bool {
	if id, ok := call.Function.(*ast.Identifier); ok {
		if sym := s.SemanticInfo.Uses[id]; sym != nil {
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok && fnStmt.IsExtern {
				return true
			}
		}
	}
	return false
}

func (s *Solver) isImplicitMoveType(t types.NRType) bool {
	if t == nil {
		return false
	}
	return types.IsOwnedType(t)
}

func (s *Solver) needsDrop(t types.NRType) bool {
	if t == nil {
		return false
	}
	ut := types.UnwrapLease(t)
	if ut == nil {
		return false
	}
	if ut.Name() == "str" {
		return true
	}
	if st, ok := ut.(*types.StructType); ok {
		for _, fType := range st.Fields {
			if s.needsDrop(fType) {
				return true
			}
		}
		return false
	}
	if sum, ok := ut.(*types.SumType); ok {
		for _, variant := range sum.Variants {
			for _, fType := range variant.Fields {
				if s.needsDrop(fType) {
					return true
				}
			}
		}
		return false
	}
	k := ut.GetKind()
	return k == types.KindList || k == types.KindMap || k == types.KindChan
}

func (s *Solver) isImplicitMoveReceiver(t types.NRType) bool {
	if t == nil {
		return false
	}
	if t.IsLeased() {
		return false
	}
	if pt, ok := t.(*types.PointerType); ok && pt.Leased {
		return false
	}
	ut := types.UnwrapLease(t)
	if ut == nil {
		return false
	}
	if ut.GetKind() == types.KindGeneric || ut.GetKind() == types.KindProtocol {
		return false
	}
	return s.needsDrop(t)
}

func (s *Solver) isMoveOperation(stmt ast.Statement, target *ast.Identifier) bool {
	// if target.Value == "buf" {
	// 	fmt.Fprintf(os.Stderr, "[DEBUG Solver] isMoveOperation for buf on stmt: %T\n", stmt)
	// }
	isMove := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if isMove || n == nil {
			return false
		}
		if _, ok := n.(*ast.BlockStatement); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.IfExpression); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.MatchExpression); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.WhileStatement); ok && n != stmt {
			return false
		}
		if _, ok := n.(*ast.ForStatement); ok && n != stmt {
			return false
		}

		isTarget := func(expr ast.Expression) bool {
			if id, ok := expr.(*ast.Identifier); ok {
				return id.Value == target.Value
			}
			return false
		}

		if pref, ok := n.(*ast.PrefixExpression); ok && pref.Operator == "@" {
			if id, ok := pref.Right.(*ast.Identifier); ok && id.Value == target.Value {
				isMove = true
				return false
			}
		}

		// Struct Instantiation
		// Struct Instantiation
		if structLit, ok := n.(*ast.StructLiteral); ok {
			structTypeObj := s.SemanticInfo.Types[structLit]
			if st, ok := structTypeObj.(*types.StructType); ok {
				for _, field := range structLit.Fields {
					fName := field.Name.Value
					fExpr := field.Value
					if fExpr != nil && isTarget(fExpr) {
						fType := st.Fields[fName]
						if s.isImplicitMoveType(fType) {
							s.debug("        Move detected: Struct Instantiation owned field")
							isMove = true
							return false
						}
					}
				}
			}
		}

		// Array Literal
		if arrLit, ok := n.(*ast.ArrayLiteral); ok {
			arrTypeObj := s.SemanticInfo.Types[arrLit]
			if arrType, ok := arrTypeObj.(*types.ListType); ok {
				if s.isImplicitMoveType(arrType.ElementType) {
					for _, elem := range arrLit.Elements {
						if isTarget(elem) {
							s.debug("        Move detected: Array Literal owned element")
							isMove = true
							return false
						}
					}
				}
			}
		}

		// Map Literal
		if mapLit, ok := n.(*ast.MapLiteral); ok {
			mapTypeObj := s.SemanticInfo.Types[mapLit]
			if mapType, ok := mapTypeObj.(*types.MapType); ok {
				isKeyOwned := s.isImplicitMoveType(mapType.Key)
				isValueOwned := s.isImplicitMoveType(mapType.Value)
				for k, v := range mapLit.Pairs {
					if isKeyOwned && isTarget(k) {
						s.debug("        Move detected: Map Literal owned key")
						isMove = true
						return false
					}
					if isValueOwned && isTarget(v) {
						s.debug("        Move detected: Map Literal owned value")
						isMove = true
						return false
					}
				}
			}
		}

		// 1. Function Calls
		if call, ok := n.(*ast.CallExpression); ok {
			funcExpr := call.Function
			for {
				if ae, ok := funcExpr.(*ast.ArgumentsExpression); ok {
					funcExpr = ae.Value
				} else {
					break
				}
			}
			fnTypeObj := s.SemanticInfo.Types[call.Function]
			if fnTypeObj == nil {
				fnTypeObj = s.SemanticInfo.Types[funcExpr]
			}
			if fnTypeObj == nil {
				if ident, ok := funcExpr.(*ast.Identifier); ok {
					if sym := s.SemanticInfo.Uses[ident]; sym != nil {
						fnTypeObj = sym.Type
					} else if sym := s.SemanticInfo.Defs[ident]; sym != nil {
						fnTypeObj = sym.Type
					}
				}
			}
			// fmt.Fprintf(os.Stderr, "[DEBUG Solver] Call to %s, resolved type: %T\n", funcExpr.String(), fnTypeObj)
			// if fnTypeObj != nil {
			// 	fmt.Fprintf(os.Stderr, "[DEBUG Solver]   type name: %s\n", fnTypeObj.Name())
			// }
			if fn, ok := fnTypeObj.(*types.FunctionType); ok {
				for i, arg := range call.Arguments {
					// Explicit move operator @u
					if pref, ok := arg.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
						if isTarget(pref.Right) {
							s.debug("        Move detected: Explicit @ on target in call")
							isMove = true
							return false
						}
					}
					// Function signature lease (move)
					if isTarget(arg.Value) {
						isArgMove := false
						if i < len(fn.ParamLeases) {
							if fn.ParamLeases[i] == types.LeaseMove {
								isArgMove = true
							}
						} else if i < len(fn.Params) && s.isImplicitMoveType(fn.Params[i]) && !s.isExternCall(call) {
							isArgMove = true
						}
						if isArgMove {
							s.debug("        Move detected: Implicit move lease or owned param type in call")
							isMove = true
							return false
						}
					}
				}
				// Handle method receiver move
				if fn.IsMethod && (fn.ReceiverLease == types.LeaseMove || s.isImplicitMoveReceiver(fn.Receiver)) {
					if sel, ok := call.Function.(*ast.SelectorExpression); ok {
						if isTarget(sel.Left) {
							isMove = true
							return false
						}
					}
				}
			} else {
				// Special case: unchecked_set consumes its second argument
				if sel, ok := call.Function.(*ast.SelectorExpression); ok && sel.Field.Value == "unchecked_set" {
					if len(call.Arguments) == 2 {
						if isTarget(call.Arguments[1].Value) {
							s.debug("        Move detected: unchecked_set value argument")
							isMove = true
							return false
						}
					}
				}

				var sumType *types.SumType
				var ok bool
				if ft, ok2 := fnTypeObj.(*types.FunctionType); ok2 {
					sumType, ok = ft.Return.(*types.SumType)
				} else if fnTypeObj != nil {
					sumType, ok = fnTypeObj.(*types.SumType)
				}
				if ok && sumType != nil {
					var vName string
					if ident, ok := funcExpr.(*ast.Identifier); ok {
						vName = ident.Value
					} else if sel, ok := funcExpr.(*ast.SelectorExpression); ok {
						vName = sel.Field.Value
					}
					var matchedVariant *types.Variant
					for _, variant := range sumType.Variants {
						if variant.Name == vName {
							matchedVariant = variant
							break
						}
					}
					if matchedVariant != nil {
						for i, arg := range call.Arguments {
							if isTarget(arg.Value) {
								if i < len(matchedVariant.FieldNames) {
									fName := matchedVariant.FieldNames[i]
									fType := matchedVariant.Fields[fName]
									isOwned := s.isImplicitMoveType(fType)
									if isOwned {
										s.debug("        Move detected: Passing owned type to variant constructor")
										isMove = true
										return false
									}
								}
							}
						}
					}
				}
			}
		}

		if pref, ok := n.(*ast.PrefixExpression); ok && pref.Operator == "@" {
			if id, ok := pref.Right.(*ast.Identifier); ok && id.Value == target.Value {
				isMove = true
				return false
			}
		}

		// isTarget is already defined above

		// 2. Variable Declarations: var x = y
		if varStmt, ok := n.(*ast.VarStatement); ok {
			if isTarget(varStmt.Value) {
				targetSym := s.SemanticInfo.Defs[varStmt.Name]
				if targetSym != nil && types.IsOwnedType(targetSym.Type) && (!targetSym.Type.IsLeased() || targetSym.Type.(*types.PointerType).Kind == types.LeaseMove) {
					isMove = true
					return false
				}
			}
			// Explicit move: var x = @y
			if pref, ok := varStmt.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if isTarget(pref.Right) {
					isMove = true
					return false
				}
			}
		}

		// 4. Spawn
		if spawn, ok := n.(*ast.SpawnExpression); ok {
			for _, arg := range spawn.Call.Arguments {
				if isTarget(arg.Value) {
					t := s.SemanticInfo.Types[arg.Value]
					if s.isImplicitMoveType(t) {
						isMove = true
						return false
					}
				}
			}
			// Receiver
			fnTypeObj := s.SemanticInfo.Types[spawn.Call.Function]
			if fn, ok := fnTypeObj.(*types.FunctionType); ok && fn.IsMethod {
				if sel, ok := spawn.Call.Function.(*ast.SelectorExpression); ok {
					if isTarget(sel.Left) {
						isMove = true
						return false
					}
				}
			}
		}

		// 3. Assignments: x = y
		if assign, ok := n.(*ast.AssignmentStatement); ok {
			if isTarget(assign.Value) {
				// Assume move if source is owned, regardless of target type
				sourceType := s.SemanticInfo.Types[assign.Value]
				if s.isImplicitMoveType(sourceType) {
					isMove = true
					return false
				}
			}
			// Explicit move: x = @y
			if pref, ok := assign.Value.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if isTarget(pref.Right) {
					isMove = true
					return false
				}
			}
		}

		// 4. Return
		if ret, ok := n.(*ast.ReturnStatement); ok && ret.ReturnValue != nil {
			if isTarget(ret.ReturnValue) {
				if s.CurrentFunction != nil && s.CurrentFunction.Return != nil && s.CurrentFunction.Return.IsLeased() {
					// Function returns a lease, do not treat as move
				} else {
					t := s.SemanticInfo.Types[ret.ReturnValue]
					if s.isImplicitMoveType(t) {
						s.debug("        Move detected: Return of owned target")
						isMove = true
						return false
					}
				}
			}
		}

		// 6. Send: c <- @x
		if send, ok := n.(*ast.SendExpression); ok {
			if isTarget(send.Right) {
				t := s.SemanticInfo.Types[send.Right]
				if s.isImplicitMoveType(t) {
					isMove = true
					return false
				}
			}
			if pref, ok := send.Right.(*ast.PrefixExpression); ok && pref.Operator == "@" {
				if isTarget(pref.Right) {
					isMove = true
					return false
				}
			}
		}

		return true
	})
	return isMove
}

func (s *Solver) findUsagesInNode(node ast.Node) []*ast.Identifier {
	var ids []*ast.Identifier
	if node == nil {
		return ids
	}
	ast.Inspect(node, func(n ast.Node) bool {
		if id, ok := n.(*ast.Identifier); ok {
			ids = append(ids, id)
		}
		return true
	})
	return ids
}

func (s *Solver) findCapturedVariables(block *ast.BlockStatement) []*semantic.Symbol {
	captured := make(map[*semantic.Symbol]bool)
	definedInBlock := make(map[*semantic.Symbol]bool)

	ast.Inspect(block, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch node := n.(type) {
		case *ast.VarStatement:
			if sym := s.SemanticInfo.Defs[node.Name]; sym != nil {
				definedInBlock[sym] = true
			}
		case *ast.Identifier:
			if sym := s.SemanticInfo.Uses[node]; sym != nil {
				if sym.Kind == semantic.SymVar || sym.Kind == semantic.SymParam {
					if !definedInBlock[sym] {
						captured[sym] = true
					}
				}
			}
		}
		return true
	})

	var result []*semantic.Symbol
	for sym := range captured {
		result = append(result, sym)
	}
	return result
}
func (s *Solver) getRootSymbol(sel *ast.SelectorExpression) *semantic.Symbol {
	root := sel.Left
	for {
		if sub, ok := root.(*ast.SelectorExpression); ok {
			root = sub.Left
		} else {
			break
		}
	}
	if id, ok := root.(*ast.Identifier); ok {
		return s.SemanticInfo.Uses[id]
	}
	return nil
}
func (s *Solver) getBaseIdentifier(expr ast.Expression) *ast.Identifier {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e
	case *ast.SelectorExpression:
		return s.getBaseIdentifier(e.Left)
	case *ast.IndexExpression:
		return s.getBaseIdentifier(e.Left)
	}
	return nil
}
func (s *Solver) analyzePattern(pattern ast.Expression, visible map[*semantic.Symbol]*Lifecycle, index int) []*semantic.Symbol {
	if pattern == nil {
		return nil
	}
	var born []*semantic.Symbol
	ast.Inspect(pattern, func(node ast.Node) bool {
		if id, ok := node.(*ast.Identifier); ok {
			sym := s.SemanticInfo.Defs[id]
			if sym != nil {
				if _, exists := visible[sym]; !exists {
					lc := &Lifecycle{Symbol: sym, DefinedAt: index, LastUsedAt: index}
					visible[sym] = lc
					born = append(born, sym)
					s.debug("      Pattern Variable Birth: %s at Index %d", sym.Name, index)
				}
			}
		}
		return true
	})
	return born
}

func isBorrowType(t types.NRType) bool {
	if pt, ok := t.(*types.PointerType); ok && pt.Leased {
		return pt.Kind == types.LeaseRead
	}
	return false
}

func (s *Solver) isOwnedRValueType(t types.NRType) bool {
	if t == nil {
		return false
	}
	if pt, ok := t.(*types.PointerType); ok && pt.Leased {
		if pt.Kind == types.LeaseRead || pt.Kind == types.LeaseWrite {
			return false
		}
	}
	return types.IsOwnedType(t)
}

func (s *Solver) findUnconsumedLambdas(stmt ast.Statement) []*ast.LambdaExpression {
	var lambdas []*ast.LambdaExpression
	for _, expr := range s.findUnconsumedRValues(stmt) {
		if l, ok := expr.(*ast.LambdaExpression); ok {
			lambdas = append(lambdas, l)
		}
	}
	return lambdas
}

func (s *Solver) findUnconsumedRValues(stmt ast.Statement) []ast.Expression {
	var rvalues []ast.Expression
	s.walkUnconsumedRValues(stmt, false, &rvalues)
	return rvalues
}

func (s *Solver) walkUnconsumedRValues(node ast.Node, isConsumed bool, out *[]ast.Expression) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.LambdaExpression:
		if !isConsumed {
			*out = append(*out, n)
		}
		return
	case *ast.ReturnStatement:
		s.walkUnconsumedRValues(n.ReturnValue, true, out)
	case *ast.VarStatement:
		s.walkUnconsumedRValues(n.Value, true, out)
	case *ast.AssignmentStatement:
		s.walkUnconsumedRValues(n.Left, false, out)
		s.walkUnconsumedRValues(n.Value, true, out)
	case *ast.CallExpression:
		if !isConsumed {
			if t := s.SemanticInfo.Types[n]; t != nil && s.isOwnedRValueType(t) {
				*out = append(*out, n)
			}
		}
		recvConsumed := false
		fnTypeObj := s.SemanticInfo.Types[n.Function]
		if ft, ok := fnTypeObj.(*types.FunctionType); ok && ft.IsMethod {
			if ft.ReceiverLease == types.LeaseMove || s.isImplicitMoveReceiver(ft.Receiver) {
				recvConsumed = true
			}
		}
		s.walkUnconsumedRValues(n.Function, recvConsumed, out)
		if ft, ok := fnTypeObj.(*types.FunctionType); ok {
			for i, arg := range n.Arguments {
				argConsumed := false
				if i < len(ft.ParamLeases) {
					if ft.ParamLeases[i] == types.LeaseMove {
						argConsumed = true
					}
				} else if i < len(ft.Params) {
					if s.isImplicitMoveType(ft.Params[i]) {
						argConsumed = true
					}
				}
				s.walkUnconsumedRValues(arg, argConsumed, out)
			}
		} else {
			for _, arg := range n.Arguments {
				s.walkUnconsumedRValues(arg, false, out)
			}
		}
	case *ast.SpawnExpression:
		s.walkUnconsumedRValues(n.Call, true, out)
	case *ast.StructLiteral:
		if !isConsumed {
			if t := s.SemanticInfo.Types[n]; t != nil && s.isOwnedRValueType(t) {
				*out = append(*out, n)
			}
		}
		for _, arg := range n.Fields {
			s.walkUnconsumedRValues(arg, true, out)
		}
	case *ast.FieldDefinition:
		if n.Value != nil {
			s.walkUnconsumedRValues(n.Value, isConsumed, out)
		}
	case *ast.ArrayLiteral:
		if !isConsumed {
			if t := s.SemanticInfo.Types[n]; t != nil && s.isOwnedRValueType(t) {
				*out = append(*out, n)
			}
		}
		for _, arg := range n.Elements {
			s.walkUnconsumedRValues(arg, true, out)
		}
	case *ast.IndexExpression:
		s.walkUnconsumedRValues(n.Left, isConsumed, out)
		for _, idx := range n.Indices {
			s.walkUnconsumedRValues(idx, false, out)
		}
	case *ast.SelectorExpression:
		s.walkUnconsumedRValues(n.Left, isConsumed, out)
	case *ast.IfExpression:
		s.walkUnconsumedRValues(n.Condition, false, out)
		s.walkUnconsumedRValues(n.Consequence, isConsumed, out)
		s.walkUnconsumedRValues(n.Alternative, isConsumed, out)
	case *ast.MatchExpression:
		s.walkUnconsumedRValues(n.Target, false, out)
	case *ast.TryExpression:
		s.walkUnconsumedRValues(n.Value, isConsumed, out)
	case *ast.PrefixExpression:
		s.walkUnconsumedRValues(n.Right, isConsumed, out)
	case *ast.InfixExpression:
		s.walkUnconsumedRValues(n.Left, isConsumed, out)
		s.walkUnconsumedRValues(n.Right, isConsumed, out)
	case *ast.ExpressionStatement:
		s.walkUnconsumedRValues(n.Expression, isConsumed, out)
	case *ast.ArgumentsExpression:
		s.walkUnconsumedRValues(n.Value, isConsumed, out)
	}
}
