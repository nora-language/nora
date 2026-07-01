# Walkthrough: Topological Solver Anonymous R-Value Drop Tracking

## Overview
We resolved an issue where chaining methods or field accesses on anonymous temporary r-values returning owned heap allocations (e.g., `CreateDummy().Read()`) resulted in memory leaks. The Topological Lease Solver and HIR Lowerer did not previously track or emit drop calls for non-lambda r-value expressions.

## Changes Made

### 1. Topological Lease Solver (`pkg/topology/solver.go`)
* Added `Expr ast.Expression` field to `DropInfo`.
* Added `isOwnedRValueType(t types.NRType) bool` helper to check if an expression returns an owned type (excluding read/write borrows).
* Replaced `findUnconsumedLambdas` with `findUnconsumedRValues(stmt ast.Statement) []ast.Expression`.
* Extended `walkUnconsumedRValues` to detect unconsumed `ast.CallExpression`, `ast.StructInstantiation`, and `ast.ArrayLiteral` nodes that return owned types.
* Updated `analyzeBlock` to schedule `DropInfo{Expr: expr}` at the end of statements for all unconsumed temporary r-values.

### 2. HIR Instructions (`pkg/hir/instruction.go`)
* Added `Expr ast.Expression` field to `hir.Drop`.
* Updated `Drop.String()` method to format `drop expr ...` when debugging instructions.

### 3. HIR Lowerer (`pkg/hir/lower.go`)
* Added `ExprTemps map[ast.Expression]*semantic.Symbol` and `UnconsumedTemps map[ast.Expression]bool` to `Lowerer`.
* In `NewLowerer`, populated `UnconsumedTemps` by scanning all scheduled `Drops` and `PreDrops` in the solver.
* Updated `emitDrop` to emit `&Drop{Symbol: l.ExprTemps[d.Expr], Expr: d.Expr}` when processing scheduled expression drops.
* Wrapped `lowerExpression` to check if an expression is marked in `UnconsumedTemps`. If so, created a temporary stack variable (`_hir_tmp_N`), emitted `Alloca` and `Store` instructions, stored the symbol mapping in `ExprTemps`, and returned a `VarOperand` targeting the temporary.

## Verification Results

### Test Case
Created `pkg/cmd/test/repro_anonymous_temporary_leak/main.nr`:
```nora
import "src/math/vector"

pub type Dummy = struct {
    val: @vector.Vector3[f64]
}

pub fn CreateDummy() @Dummy {
    var v = vector.NewVector3[f64](1.0, 1.0, 1.0)
    return alloc Dummy { val: v }
}

pub fn (d: #Dummy) Read() f64 {
    return d.val.x
}

pub fn main() {
    var val = CreateDummy().Read()
}
```

### Before vs After
* **Before Fix**: Executing `nora run --debug-memory` reported active allocations with 12 bytes leaked.
* **After Fix**: Executing `nora run --debug-memory` reported 0 active allocations and 0 leaked bytes.
