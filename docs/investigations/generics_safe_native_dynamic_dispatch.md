# Compiler Investigation: Safe Native Dynamic Dispatch to Replace Raw Pointer Type-Erasure

Status: Completed

## Problem
In the Nora Programming Language, memory safety is enforced by a static **Topological Lease Solver**. However, the current implementation of the Entity Component System framework (GECS) under `examples/port_gecs` relies on raw pointer type-erasure (`ptr`). This is done using helper utilities in `examples/port_gecs/gecs/src/lib.nr` such as:
- `IntoRaw[T](val: @T) ptr`
- `IntoRawVector[T](val: @collections.Vector[T]) ptr`
- `CopyPtr(p: #ptr) ptr`

These functions completely bypass the lease solver, permit unsafe access, and degrade developer ergonomics (DX) by introducing verbose boilerplate. To establish Nora as a fully safe, consistent, and premium systems language, we need to replace these raw pointer workarounds with **safe native dynamic dispatch** (interfaces and the existential type `any`) and cleanly remove `lib.nr`.

---

## Current Setup & Usage of `lib.nr`
Currently, raw `ptr` type-erasure is used across GECS in the following critical areas:
1. **Dynamic Columns (`world.nr` / `a_types.nr`)**:
   `columns: @collections.HashMap[i32, ptr]` maps component IDs to type-erased vectors of components.
2. **Metadata & Lifecycle Functions (`world.nr` / `a_types.nr`)**:
   `ComponentMeta` is stored as `ptr` and contains function pointers to compile-time generated type-erased operations (e.g., initialization and moves).
3. **Observers (`world.nr` / `a_types.nr`)**:
   `observers: @collections.HashMap[i32, ptr]` maps component IDs to type-erased `ObserverTrigger` structures.
4. **Systems Scheduler (`scheduler.nr`)**:
   `systems: @collections.Vector[ptr]` stores systems as raw pointers, requiring unsafe casting to `&System` when executed or verified for scheduling conflicts.

---

## Design Alternatives Considered

### Option 1: Existential Type Erasure (`any`)
Nora provides a built-in empty interface / existential type `any` (represented as `Any` in the compiler type system). Instead of using raw `ptr`, we can use `any` or references to `any` (`#any`, `&any`).

- **Syntax**:
  ```nora
  var queue: &EventQueue = EventQueue(queue_opt)
  ```
- **Pros**:
  - **Memory Safety**: The Topological Lease Solver can track the lifecycle of `any` containers, inserting drops and preventing invalid accesses automatically.
  - **DX**: Eliminates the need for helpers like `IntoRaw`. Users perform direct safe casts (e.g., `MyType(val_any)`).
  - **Uniformity**: Follows Nora's core language design.
- **Cons**:
  - Incurs minimal runtime checking overhead during casting back to concrete types.

### Option 2: First-Class Protocols/Interfaces
For behavioral dispatch (such as systems), we can define explicit interface structures.

- **Syntax**:
  ```nora
  pub type System = interface {
      fn Run(w: &World, cb: &CommandBuffer) void
      fn Access() &Access
  }
  ```
- **Pros**:
  - **Strong Typing**: Compiler guarantees that only types implementing the method signatures can be added.
  - **No Casts**: Directly call `sys.Run(w, cb)` without type casting.
- **Cons**:
  - Does not solve dynamic storage for heterogeneous component arrays (like `Vector[T]`) because components do not share a common set of methods.

---

## Recommended Solution

For the Nora programming language, the optimal solution is a **hybrid approach** utilizing both **first-class interfaces** for behavioral dynamic dispatch and the **existential `any`** type for heterogeneous data storage.

### 1. Behavioral Dynamic Dispatch: System Interface
Define a clean `System` interface instead of using raw `ptr` and custom update function pointers:
```nora
pub type System = interface {
    fn Run(w: &World, cb: &CommandBuffer) void
    fn Access() &Access
}
```
This allows the scheduler to hold a `collections.Vector[System]` natively and safely execute them concurrently without unsafe pointer dereferences.

### 2. Heterogeneous Data Storage: Existential `any`
For storage maps (like `columns`, `observers`, and `eventMgr`), use `any` instead of `ptr`:
```nora
pub type World = struct {
    ...
    observers: @collections.HashMap[i32, any],
    archetypes: @collections.HashMap[i64, any]
}
```
This enables Nora's compiler to track type safety, run type-checking at compile-time, and handle cleanups.

### 3. Removal of `lib.nr`
By shifting to `any` and `interface`, the compiler can natively assign concrete types to interfaces and perform existential type coercion. The entire `lib.nr` file is rendered obsolete and can be completely removed.

---

## Implementation Plan for GECS Refactoring
1. **Define Interfaces**: Create the `System` interface in `a_types.nr`.
2. **Update Signatures**: Replace all instances of `ptr` in `World`, `Scheduler`, `Archetype`, and `EventManager` structures with `any` or explicit interface/reference types.
3. **Refactor Code**: Update method calls and casts in `world.nr`, `scheduler.nr`, and `event.nr` to use safe cast syntax (`Type(val_any)`) and remove all imports and references to `gecs.IntoRaw` or `gecs.CopyPtr`.
4. **Remove `lib.nr`**: Delete the file `examples/port_gecs/gecs/src/lib.nr` from the workspace.
