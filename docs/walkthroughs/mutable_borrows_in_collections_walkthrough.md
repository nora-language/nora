# Walkthrough: Direct Mutable Borrows into Collections

We have successfully implemented first-class support for returning mutable leases (`&T`) from functions and methods in the Nora compiler and updated GECS.

## Changes Made

### 1. Codegen Support for LHS Dereferencing
- Modified the code generator in [hir_codegen.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/hir_codegen.go) within the `*hir.Store` and `*hir.Assign` statement handling.
- Added a check: if the LHS destination is a pointer in C (`isOperandPointerInC`), but the RHS value is not, the compiler prepends `*` to the destination (`*destStr`), ensuring assignments write directly to the referenced memory slot instead of attempting to overwrite the pointer variable.

### 2. Standard Library Collections
- Added the `GetMut[T](index: i32) &T` method to [vector.nr](file:///e:/Project/Project%20Chronos/second/std/collections/vector.nr). This returns a mutable lease pointing directly to the slice index.

### 3. GECS ECS Framework
- Added `GetMut[T](w: &World, e: Entity, cid: i32) &T` to [world.nr](file:///e:/Project/Project%20Chronos/second/examples/port_gecs/gecs/src/world.nr) to return a mutable lease to components inside storage columns.
- Refactored parent-child relationship handling in [relation.nr](file:///e:/Project/Project%20Chronos/second/examples/port_gecs/gecs/src/relation.nr) to use `GetMut`, completely removing the verbose `CopyPtr` stack-copy workarounds.

---

## Verification Results

We verified the implementation using the GECS premium integration test basic example:
```bash
go run pkg/cmd/nora/main.go run --example basic
```
This successfully compiled and executed:
```text
================================================
--- GECS Premium Complete Port Integration ---
================================================
1. World initialized.
2. Registering component observers:
3. Queueing deferred mutations in CommandBuffer:
   Playback queued deferred commands...
   [Observer: Add] Entity 1: Position added {x: 10, y: 20}
   [Observer: Add] Entity 2: Position added {x: 100, y: 200}
4. Linking parent-child hierarchies:
   Child ID 2 parent is confirmed: ID 1
5. Allocating spatial components:
6. Initializing Scheduler & systems:
   Running Scheduler (Transform propagation)...
   Child propagated GlobalTransform: {x: 5, y: 5, z: 0} (Parent local + Child local)
7. Broadcasting & reading custom events:
   Read HitEvent: damage = 99
   Read HitEvent: damage = 150
   Clearing event buffers...
================================================
GECS Full Port Test Successful!
================================================
```
