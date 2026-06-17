# Compiler Investigation: Mutable Borrows into Collections

Status: Completed

## Problem
In the Nora language, collections (such as `Vector`, `HashMap`, and the Entity Component System storage columns) currently return elements by value. For example, in ECS updates:
```nora
var children_val = Get[Children](w, Entity { id: parent.id }, 4)
var children_ptr = CopyPtr(&children_val)
var children_mut: &Children = children_ptr
children_mut.entities.Push[Entity](Entity { id: child.id })
```
`Get[Children]` returns `Children` by value (copying it). If `Children` were a pure value struct (without a nested heap-allocated pointer like `Vector`), this update would be lost because it modifies a local stack copy rather than the component in the ECS store. 

Nora needs the ability to safely obtain a mutable reference pointing directly to the memory slot inside a collection, rather than forcing value copy-outs.

---

## Design Alternatives Considered

### Option 1: Internal Mutation Closures / Callback-based Mutation
Instead of returning a reference to the inner element, the collection provides a method that accepts a mutation callback.
```nora
pub fn (v: &Vector[T]) Update[T](index: i32, f: fn(&T)) {
    f(&v.data[index])
}
```
**Pros**:
- Safe and trivial to implement in the lease solver since the borrow lifetime is strictly bounded by the scope of the callback execution.

**Cons**:
- Unbelievably verbose for the developer.
- Incurs runtime closure allocation or function pointer overhead in the generated C11 target.
- Violates Nora's design principle of **User Simplicity** and clean, minimal syntax.

---

### Option 2: Explicit Ref Return Types (`&T` / `#T`) with Lifetime Propagation (Recommended)
Allow functions to declare mutable (`&T`) or read-only (`#T`) leases as return types. Accessor methods can return references directly pointing to elements inside their structures.
```nora
pub fn GetMut[T](w: &World, e: Entity, cid: i32) &T {
    var id = e.ID()
    var record = w.entities.data[id]
    var r_row = record.row
    var arch: &Archetype = record.archetype
    var col_ptr = arch.columns.Get[i32, ptr](#cid)
    var vec: &collections.Vector[T] = col_ptr
    return &vec.data[r_row]
}
```

#### Lease & Lifetime Analysis
The static **Topological Lease Solver** (`pkg/topology`) tracks variable lifetimes using a dependency graph. 
When a lease is returned from a function:
```nora
var children_mut: &Children = w.GetMut[Children](parent, 4)
```
The lease solver's `trackDependencies` automatically detects that `w` (the World provider) is used in the RHS expression. It establishes a static dependency: `children_mut` depends on `w`. 
- As long as `children_mut` is alive, the parent `w` remains borrowed.
- Any attempt to mutably access, move, or drop `w` while `children_mut` is active triggers a compile-time lease violation.

This provides absolute memory safety with zero runtime overhead.

#### Codegen
In `pkg/codegen`, returning `&T` compiles directly to a pointer return type `T*` in C11. Standard member access on the returned reference operates directly on the target memory slot.

---

## Recommended Solution

**Option 2 (Explicit Lease Return Types `&T` / `#T`)** is the best solution for the Nora programming language. It is highly consistent with Nora's memory model, avoids runtime overhead, and provides first-class developer ergonomics.

### Implementation Checklist
1. **Parser Support**: Ensure function return types accept prefix operators `&` and `#` (e.g. `fn GetMut() &T`).
2. **Semantic Checking**: Validate that returned leases (`&T` / `#T`) are assignable to variables and that their source lifetimes remain valid.
3. **Lease Solver**: Verify that returned lease dependencies are correctly tracked and propagated to LHS targets.
4. **C11 Codegen**: Compile lease return values into C pointers (`T*`), ensuring proper dereferencing on access.
