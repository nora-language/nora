# Investigation: Generic Operator Interfaces vs Overloading

**Status**: Completed
**Context**: We investigated how to support `vector + scalar` alongside `vector + vector` in Nora. Rust solves this using generic traits (`Add<Rhs>`). Can Nora do the same?

## The Structural Typing Problem

Nora uses **Go-style Structural Typing** for interfaces (as seen in `interface_demo.nr`). This means methods are attached directly to the struct, and interfaces are satisfied implicitly if the struct happens to have matching methods.

Because methods belong directly to the struct, they share a single namespace (`st.Methods` is a `map[string]NRType`). 

If we try to use a generic interface for the `+` operator in Nora:
```nora
pub type Add[Rhs] = interface {
    fn add(other: Rhs) Self
}
```

To make `Vector` satisfy both `Add[Vector]` and `Add[i32]`, the developer would have to write:
```nora
pub fn (self: &Vector) add(other: Vector) Vector { ... }
pub fn (self: &Vector) add(other: i32) Vector { ... }
```

**This results in a name collision.** The second `add` method would overwrite the first in the compiler's symbol table, or throw a "duplicate symbol" error. 

In Rust, this collision doesn't happen because Rust uses **Nominal Typing with Explicit Impl Blocks**. The methods are namespaced by the trait (`<Vector as Add<i32>>::add`), not just the struct.

## Conclusion & Available Solutions

Because Nora uses structural typing, we **cannot** use generic interfaces to solve operator overloading without fundamentally changing how interfaces work. 

To support `vector + scalar` and `vector + vector` using the `+` operator, Nora MUST choose one of the following paths:

### Option A: The `[overload]` Attribute (Original Proposal)
Allow methods on the same struct to share a name if marked with `[overload]`.
* **Pros**: Solves the problem immediately. Keeps structural typing intact.
* **Cons**: Complicates the `SemanticAnalyzer` which now has to resolve method calls based on argument types.

### Option B: Explicit Trait `impl` Blocks (The Rust Path)
Introduce `impl Protocol for Struct` syntax, allowing methods to be namespaced by the protocol rather than the struct.
* **Pros**: Cleanest semantic model for operators. Avoids ad-hoc function overloading entirely.
* **Cons**: Massive language change. Moves Nora away from simple Go-style interfaces towards complex Rust-style traits.

### Option C: Type-Erased / Sum Type Operators
Pass a `SumType` to the `add` method.
```nora
pub type ScalarOrVec = enum { Vec(Vector), Scalar(i32) }
pub fn (self: &Vector) add(other: ScalarOrVec) Vector { ... }
```
* **Pros**: No language changes needed.
* **Cons**: Creates runtime overhead (tag checking) and annoying developer ergonomics (having to wrap arguments).

## Recommendation
If the goal is to keep Nora's Go-style structural interfaces but still allow ergonomic math (`vec + 5`), then implementing the **`[overload]` attribute for struct methods** is actually the most pragmatic choice. 
