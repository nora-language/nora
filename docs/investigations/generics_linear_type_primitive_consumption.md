# Investigation 6 (Phase 5): Generics Linear Type Primitive Consumption

## Status
Completed

## Problem
During the implementation of Phase 3 constraints (Hinge and Slider) in `nora_physics`, we encountered recurring `use of moved value` compilation errors when performing arithmetic operations (e.g. `-`, `+`, `/`) or assignments (`var b = a`) on primitive values like floating-point numbers (`f64`). 

Interestingly, these errors occurred selectively:
- Variables of type `T` (generic type parameter) threw linear move errors upon basic assignment (`var b = a` consumes `a`).
- Explicit primitives like `var a: f64 = 1.0` did **not** throw move errors on assignment or arithmetic.

This caused significant friction in the math-heavy constraint solver code, leading to bizarre workarounds like redefining zeroes `(dt-dt)` to avoid consuming local scalar variables.

## Reproduction
To isolate the bug, we created two identical test scripts.

**Test 1: Explicit Primitives (`test_linear.nr`)**
```nora
package main
fn main() {
    var a = 1.0 // Inferred as f64
    var b = a   // DOES NOT consume a
    var c = a + 2.0 // Success
}
```
*Result*: Compiles successfully.

**Test 2: Generics (`test_generic.nr`)**
```nora
package main
fn test[T](a: T) {
    var b = a   // Consumes a
    var c = a   // ERROR: use of moved value 'a'
}
fn main() {
    test[f64](1.0)
}
```
*Result*: Fails compilation with `use of moved value 'a'`.

## Root Cause
This is **not a bug**, but an intended architectural consequence of Nora's compiler design:
1. **Semantic Phase Separation**: Nora runs its semantic analysis and linear lifetime checks *before* monomorphization.
2. **Conservative Generic Lifetimes**: During the semantic pass, the compiler encounters the generic type `T`. Because `T` could theoretically be an inherently linear struct (like a `Box[T]`), the semantic analyzer must conservatively assume `T` is strictly linear. 
3. **Primitive Inference**: Conversely, for concrete primitives (like `f64`), the semantic analyzer explicitly knows they are trivially copyable (register-sized values without lifecycle hooks) and thus permits implicit copying without consumption.

Because our physics engine is entirely generic (`body.Body[T]`, `vector.Vector3[T]`, `dt: T`), any local variable mapped to `T` acts as a strict linear type.

## Fix & Workarounds
Since this is an intentional compiler design feature representing conservative generic bounds, we must adhere to it in generic libraries:
1. **Pass-By-Reference for Read-Only Uses**: Always use `#` (read-only borrow) for generic operands if they are not meant to be consumed. (e.g., `v_rel.Dot[T](#t1)`).
2. **Recalculation over Caching**: For generic primitives that cannot be borrowed easily in complex arithmetic, recalculating cheap inline operations (like dot products or `dt-dt`) is preferred over caching them in variables.
3. **Future Consideration**: The Nora language specification could introduce generic bounds (e.g., `[T: Copy]`) to instruct the semantic analyzer that `T` will only ever be instantiated with trivially copyable primitives, removing this strict move limitation for generic math algorithms.

## Validation
By rigorously avoiding variable caching and heavily leveraging `#` read-only borrows and dynamic derivation `(dt-dt)`, the generic constraints compiled successfully without violating the strict linear lifetime topology.
