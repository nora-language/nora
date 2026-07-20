# Investigation: Nora Physics Examples Migration & Compiler Edge Cases

**Status**: Completed
**Date**: 2026-07-20

## Problem
During the process of validating and running the `nora_physics` examples suite (Phase 2 through Phase 9), multiple compiler errors and C-codegen failures prevented successful execution. The most critical failure was a C compiler error (`use of undeclared identifier 'nr_df_g1'`) in `phase5_collisions`, `phase5_hair`, and `phase6_fluid`. Additionally, the Topological Lease Solver emitted severe "move out of borrowed context" and "cannot implicitly load value from move-lease" errors across several character controller and collision detection examples.

## Reproduction
1. Run `nora.exe run -debug-memory --example phase5_collisions` (or `phase5_hair`, `phase6_fluid`).
   - **Result**: Fails during the C compilation phase with `exit status 1` due to undeclared `nr_df_` boolean defer flags in `out_globals.c`.
2. Run `nora.exe run -debug-memory --example phase8_determinism`.
   - **Result**: Fails Topological Lease Solving due to reading a generic `T` field from a borrowed context (`var c_depth = cp_ref.depth`).
3. Run `nora.exe run -debug-memory --example phase4_kinematic_char`.
   - **Result**: Fails Semantic Analysis due to `alloc Vector3[T]` being assigned to a value-type field `ground_normal`.

## Root Cause

### 1. Shared Globals Codegen Bug (`nr_df_` missing)
When the compiler performs type-erased shared monomorphization for generics (generating C code in `out_globals.c`), it must insert boolean "defer flags" (e.g., `bool nr_df_g1 = false;`) to track heap allocations (like `vector.Cross()` which uses `alloc`) that need to be freed at the end of the scope.
However, a bug in the C-backend causes it to **fail to declare these boolean flags** at the top level of the C function if the allocation occurs inside a nested block (such as a `while` loop). The generated C code attempts to set the undeclared variable (`nr_df_g1 = true;`), resulting in a fatal Clang compilation error.

### 2. Generic Borrow Limitation (Topological Lease Solver)
Because a generic type `T` might represent a complex heap-allocated structure requiring deep moves, the Topological Lease Solver conservatively forbids assigning a generic field from a borrowed context (`#cp`) directly to a new variable. It treats the assignment as a strict move, triggering a "cannot move out of borrowed context" error.

### 3. Strict Value vs. Lease Initialization
Nora enforces a strict boundary between stack/inline value types and heap-allocated leases (`@`). Assigning an `alloc` statement (which returns an owned lease) to a struct field expecting a raw value type triggers a "cannot implicitly load value from move-lease" error, intentionally preventing silent deep-copies.

## Fix

### 1. Codegen Bug Workaround
To bypass the `nr_df_` codegen bug in `out_globals.c`, the heap-allocating math logic inside the `while` loop in `src/softbody/solver.nr` (specifically the pressure correction math using `Cross`) was extracted into a top-level helper function: `ApplyPressureCorrection[T]`. 
By isolating the allocations to a function without nested loops, the C-backend correctly emitted the `nr_df_` declarations at the top of that function, successfully compiling the shared generic implementation.

### 2. Arithmetic Generic Copy (`+ zero`)
To safely extract the generic `T` field `depth` from a borrow without triggering a move, the arithmetic `+ zero` trick was used:
```nora
var zero = posA.x - posA.x
var c_depth = cp_ref.depth + zero
```
This forces the compiler to invoke the addition operator, which produces a safe, brand-new copied value instead of attempting an ownership transfer.

### 3. Direct Value Initialization
In `kinematic_char.nr` and `dynamic_char.nr`, `alloc vector.Vector3[T] {...}` was replaced with direct value instantiation `vector.Vector3[T] {...}` to align with the expected field types, resolving the lease-to-value semantic errors.

## Validation
After applying the fixes, the entire `nora_physics` examples suite (`phase2` through `phase9`) was executed via a recursive `nora.exe run -debug-memory` script. All examples compiled successfully and generated the expected physics tick outputs without memory leaks or topological errors.
