# Full HIR Codegen Migration Plan

## Status
Not Started

## Metadata
* **Author:** Antigravity
* **Date:** July 2026
* **Component:** Compiler Backend (`pkg/codegen`, `pkg/hir`)

## Goal
Completely migrate the Nora compiler backend to exclusively use the High-level Intermediate Representation (HIR) for C11 code generation. This involves eliminating the legacy AST-based codegen fallback (over 10,000 lines of code in `expressions.go` and `statements.go`) currently used for generic monomorphizations and complex expressions.

## Affected Compiler Components
* `pkg/hir/lower.go`: Needs to fully lower `Try` expressions and overloaded operators to native HIR nodes, removing the `ASTExpr` fallback mechanism.
* `pkg/hir/optimize/inliner.go`: Needs an exported `OptimizeFunction` method to optimize dynamically generated HIR instances.
* `pkg/codegen/generator.go`: Needs to dynamically invoke the HIR lowerer for type-erased monomorphized generic instances instead of falling back to legacy `genFunction`.
* `pkg/codegen/hir_codegen.go`: Must be decoupled from the legacy `exprToString` function.
* `pkg/codegen/expressions.go` & `pkg/codegen/statements.go`: Marked for complete deletion.

## Implementation Checklist

### Phase 1: Complete HIR Lowering (Eliminate `ASTExpr` Fallbacks)
To decouple the compiler from legacy codegen, we must build native HIR lowering for every AST node that currently falls back to `hir.ASTExpr` or is completely unhandled by `lower.go`.

**Missing Statements (Currently ignored or crashing in HIR):**
- [x] Lower `ast.DeferStatement` (Currently missing native HIR statement support).
- [ ] Lower `ast.MatchExpression` (Complex pattern matching, heavily relies on legacy codegen `genMatchExpression`).
- [x] Lower `ast.SelectStatement` channels (Needs verification if fully lowered or relies on fallback).
- [ ] Lower `ast.PinStatement` (if applicable to HIR lowering).

**Control Flow & Operators:**
- [x] Lower `ast.TryExpression` into HIR control flow blocks (checking variants and branching).
- [ ] Lower Struct/Sum-Type Operator Overloads (prefix `!`, `-`, `~` and infix `+`, `-`, `==`, `!=`) into HIR `Call` or native instructions.
- [ ] Lower `ast.RangeExpression` (`a..b`) into proper struct initialization.

**Data Structures & Literals:**
- [x] Lower `ast.StructLiteral` (currently falls back in `lower.go:1396`).
- [x] Lower `ast.ArrayLiteral` (currently falls back in `lower.go:1416`).
- [x] Lower `ast.MapLiteral` (currently caught by the `default` fallback).
- [ ] Lower `ast.InterfaceLiteral` (if applicable).
- [x] Lower `ast.SliceExpression` (if not natively handled by bounds checking yet).
- [x] Lower `ast.IndexExpression` (including `IndexAccess`, array bounds checking, and struct overloads).

**Advanced Expressions:**
- [ ] Lower `ast.InterpolatedString` (Currently missing, relies on `genInterpolatedString`).
- [ ] Lower `ast.ParallelExpression` (If applicable, needs HIR representation).
- [ ] Lower Lambda Captures and Environment Struct Initialization (currently falls back in `lower.go:1369`).
- [ ] Lower Type Cast Expressions (`ast.CastExpression`, currently caught by fallback).
- [ ] Final Audit: Remove the `default` switch cases in `lower.go` (`lowerExpression` and `lowerStatement`) to guarantee no hidden fallbacks or silent ignores exist.
- [ ] Delete `exprToString` dependencies completely from `pkg/codegen/hir_codegen.go`.

### Phase 2: Dynamic Lowering of Monomorphizations
- [ ] Rename `lowerFunction` to `LowerFunction` (exported) in `pkg/hir/lower.go`.
- [ ] Add `OptimizeFunction(hf *hir.Function) *hir.Function` to `pkg/hir/optimize/inliner.go`.
- [ ] Modify `generator.go` (specifically lines ~1915 and ~2429 where `g.genFunction` is used as a fallback). Change the logic to dynamically invoke `LowerFunction` and `OptimizeFunction` for generic instances on-the-fly, then emit them using `genHIRFunction`.

### Phase 3: Legacy Codegen Deletion
- [ ] Delete `pkg/codegen/expressions.go`.
- [ ] Delete `pkg/codegen/statements.go`.
- [ ] Remove the `genFunction` method from `pkg/codegen/generator.go`.
- [ ] Remove `exprToString` and other unused legacy formatting utilities from the codebase.

## Test Plan
- Run the full integration test suite via `nora test`.
- Manually inspect the generated C11 code (`nora build --verbose`) for generic monomorphizations to ensure type erasure is still functioning perfectly and optimized.
- Validate that RAII drop calls (`nr_free`, `nr_drop`) are still correctly inserted by the Topological Lease Solver around the newly lowered HIR nodes, without introducing memory leaks.

## Risks
- **Control Flow Alterations:** Lowering `try` expressions manually into HIR blocks might alter the control flow graph, which could surface edge cases in the Topological Lease Solver.
- **Evaluation Order:** Expanding complex AST expressions into linear HIR assignments might slightly change the order of evaluation if not carefully implemented.

## Completion Criteria
- The compiler successfully compiles the standard library and all examples without invoking any legacy codegen paths.
- `expressions.go` and `statements.go` are completely removed from the repository.
- `nora test` passes with zero regressions or infrastructure memory leaks.
