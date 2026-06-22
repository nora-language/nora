# Linear Type System Moves Primitives (Zeroing Memory)

## Problem

When building `nora_physics`, we encountered a bizarre issue where passing a structural primitive field (e.g., `s.radius` of type `T` which is a `f64`) into a generic function argument (e.g., `norm_dir.MulScalar[T](s.radius)`) caused the memory of the original struct to be silently corrupted or evaluated as `0.0`. 

Worse, subsequent usages of the same struct field triggered topology errors: `Error: use of partially moved value 'sphere1' (field 'sphere1.radius' was moved)`.

## Reproduction

```nora
pub fn (s: #Sphere[T]) Support[T](dir: #vector.Vector3[T]) @vector.Vector3[T] {
    var norm_dir = dir.Normalize[T]()
    
    // BUG: Passing s.radius (primitive T) to a generic function argument
    // moves it, zeroing out the memory of s.radius in the original struct!
    var offset = norm_dir.MulScalar[T](s.radius) 
    
    // ... Any subsequent access of s.radius or the original sphere fails
}
```

## Root Cause

Nora's strictly linear type system treats ALL generic type parameters `T` as linear by default. It does not automatically derive `Copy` for primitives when they are passed as generic arguments. 

Therefore, when `s.radius` is passed into `MulScalar[T](s: T)`, the compiler executes a **Move**. In Nora, a move transfers ownership and **zeroes out the original memory location** to prevent use-after-free and aliasing, even if it's just a `f64`.

Because `s` was a borrow (`#Sphere[T]`), the compiler incorrectly allowed the move to proceed without emitting a `cannot move out of borrowed context` error (which is an underlying bug in the Lease Solver). Instead, it silently zeroed out the primitive field inside the borrowed struct, leading to catastrophic logic bugs.

## Fix / Workaround

To prevent primitive values from being moved, they must be explicitly cloned using binary operators (which do not consume their operands).

```nora
pub fn (s: #Sphere[T]) Support[T](dir: #vector.Vector3[T]) @vector.Vector3[T] {
    var norm_dir = dir.Normalize[T]()
    
    // Workaround: Clone the primitive using binary math operators
    var zero = s.radius - s.radius
    var cloned_radius = s.radius - zero
    
    var offset = norm_dir.MulScalar[T](cloned_radius) // cloned_radius is moved instead
}
```

By performing `s.radius - zero`, the result is a fresh value of `T`. When this fresh value is passed to `MulScalar[T]`, it is safely moved without affecting `s.radius`.

## Validation

Applied the workaround to `sphere.nr`, and the GJK intersection logic executed flawlessly with `f64` primitives, preventing memory corruption and allowing successful detection between `Sphere-Sphere` and `Sphere-Box`.
