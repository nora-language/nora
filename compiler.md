# Nora Compiler Bug Repros

Bugs discovered during `nora_physics` → `nora_math` migration.

---

## Bug 1 — RETURN TYPE IS NIL INTERFACE Panic

**Status:** Confirmed ✅  
**Repro:** [fail_repro_missing_import_return](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_repro_missing_import_return/)

### Error
```
goroutine 1 [running]:
pkg/semantic.(*SemanticAnalyzer).Analyze(...)
    analyzer.go:2802: interface conversion: types.NRType is nil
```

### Root Cause
When a function's return type references a type from an unresolved package import, the return type resolves to `nil`. The compiler then panics with a nil interface dereference in `analyzer.go:2802` instead of emitting a graceful diagnostic error.

### Trigger
```nora
import "missing_pkg"                // import fails silently
pub fn foo() missing_pkg.SomeType { // return type resolves to nil → PANIC
    ...
}
```

---

## Bug 2 — Codegen: Primitive Generic Return Passed as `void*` in String Interpolation

**Status:** Confirmed ✅  
**Repro:** [fail_repro_codegen_generic_method_to_str](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_repro_codegen_generic_method_to_str/)  
**Also seen:** [fail_repro_cross_pkg_generic_method_arg](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_repro_cross_pkg_generic_method_arg/)

### Error
```
out_pkg_main.c:79: error: passing 'float' to parameter of incompatible type 'void *'
  { printf("%s%s", nr_to_str(d), ...); }
```

### Root Cause
When a cross-package generic method (e.g. `Dot[T: Copy]`) returns a primitive type (`f32`), and that return value is used in a string interpolation (`"result = ${d}"`), the C codegen emits `nr_to_str(d)` where `nr_to_str` expects `void*` but `d` is a C `float`. The type-erased monomorphization incorrectly treats the return as a pointer type.

### Trigger
```nora
// pkg_vec/pkg_vec.nr
pub fn (self: #Vec2[T]) Dot[T: Copy](other: #Vec2[T]) T { ... }

// main.nr
var a = pkg_vec.NewVec2[f32](...)
var b = pkg_vec.NewVec2[f32](...)
var d = a.Dot[f32](#b)
io.PrintLn("dot = ${d}\n")  // ← BUG: d is float but nr_to_str expects void*
```

---

## Bug 3 — Cross-Package Generic Return Type Name Mismatch (Nested Generics)

**Status:** Could not isolate with simple repro (needs deeper investigation)  
**Observed in:** `nora_physics/src/collision/aabb.nr`

### Error
```
cannot return value of type AABB from function returning aabb_AABB_8f45746c
  --> src/collision/aabb.nr:11
   11 |     return AABB[T] { min: min, max: max }
```

### Root Cause
When a generic struct `AABB[T]` is defined in package `aabb` and contains `Vector3[T]` fields (another generic from a separate package), the compiler assigns two different names to the same type in the same function body:
- Inside the function: `AABB` (local name)
- As the return type from the call site's perspective: `aabb_AABB_8f45746c` (cross-package monomorphized name)

The compiler fails to unify these. Workaround: convert to value-type returns (remove `@`/`alloc`) reduces some cases, but the semantic-level unification is the real fix needed.

---

## Bug 4 — Cross-Package Generic Struct: `expected #Vector2, got #vector_Vector2_XXXX`

**Status:** Confirmed (observed in `nora_physics`)  
**Observed in:** `nora_math/src/math/noise.nr`

### Error
```
type mismatch: expected #Vector2, got #vector_Vector2_f1109c62
  --> nora_math/src/math/noise.nr:109
  109 |     var n00 = g00.Dot(#v00)
```

### Root Cause
When a value of type `Vector2[f32]` is created via a constructor in the same file (`noise.nr`), the compiler assigns it the cross-package monomorphized name `vector_Vector2_f1109c62`. When this is passed to `Dot(other: #Vector2[T])` whose parameter uses the un-prefixed local name `#Vector2`, the compiler fails to match them.

**Workaround applied:** Inline the dot product arithmetic (`x*x + y*y`) to bypass the method call.

### Trigger
```nora
// Inside noise.nr (package noise), after importing vector
var v00 = vector.NewVector2[f32](pf_x, pf_y)  // type: vector_Vector2_f1109c62
var n00 = g00.Dot(#v00)  // Bug: Dot expects #Vector2, gets #vector_Vector2_f1109c62
```
