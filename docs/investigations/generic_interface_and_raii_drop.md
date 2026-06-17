# Investigation: Generic Monomorphization Failures in Interfaces and RAII Drops

## Problem
The Nora compiler experienced two major, seemingly distinct failures that shared the same architectural root cause in the semantic analyzer:
1. **Generic Interface Casting Error:** The compiler generated invalid code when casting a monomorphized generic struct (e.g., `MyStruct[i32]`) to an interface (e.g., `MyInterface`). The code generator threw a missing vtable method error during C11 emission. This impacted tests like `pass_generic_interface.nr` and `generic_iterator_repro.nr`.
2. **Memory Leaks in Smart Pointers:** The test `ffi_ownership_test.nr` consistently leaked 4 bytes of heap memory. The RAII `drop()` method defined on the `ffi.Owned[T]` generic struct was silently ignored, preventing cleanup of resources.

## Reproduction
**Reproduction 1 (Interface Cast):**
```nora
pub type Container[T] = struct { val: T }
pub fn (c: &Container[T]) Get() T { return c.val }

pub type Getter = interface { Get() i32 }

// This triggered a missing vtable error during codegen
var c = Container[i32]{val: 10}
var g: Getter = c 
```

**Reproduction 2 (RAII Memory Leak):**
```nora
import "std/ffi"
{
    var raw = ffi.Malloc(4)
    var _owned = ffi.Owned[Resource] { raw: raw }
    // RAII drop() was not called when _owned exited scope!
}
```

## Root Cause
Both failures originated from a deferred method monomorphization logic in the semantic analyzer (`pkg/semantic/analyzer.go`). 

Nora defers the monomorphization of struct methods until they are explicitly called, minimizing compile-time overhead. However, two edge cases were completely missing explicit triggers:
1. **Interface Assignments:** When a generic struct was assigned to an interface, its methods were bound to a dynamic vtable. Because the compiler didn't explicitly call the method in the AST, `sa.Monomorphize` was never triggered, leaving the method completely missing in the target binary.
2. **Implicit RAII Methods (`drop`, `eq`):** The topological lease solver automatically inserts `drop` and `eq` calls directly into the HIR (High-level Intermediate Representation), completely bypassing AST-level explicit function calls. Because `drop` on `ffi.Owned[Resource]` was never explicitly invoked in the syntax tree, it was never monomorphized. Furthermore, `ffi.Owned` had mistakenly defined its drop method with a redundant type parameter (`drop[T]()`), meaning the solver couldn't identify it as the standard parameter-less RAII destructor.

## Fix
1. **Semantic Interception for Interfaces:**
   In `analyzer.go` (`checkInterfaceCompatibility`), when an interface cast occurs, the compiler now immediately invokes `ensureMethodsSpecialized` on the base generic struct, guaranteeing that all struct methods matching the interface are explicitly monomorphized and available to the vtable generator.
   
2. **Implicit Method Monomorphization & Signature Fix:**
   In `analyzer.go` (`specializeStructType`), the compiler now proactively monomorphizes `drop` and `eq` methods the moment a generic struct is instantiated. We also patched `std/ffi/ffi.nr` to fix the `Owned[T].drop` signature from `drop[T]()` to the standardized `drop()`.

## Validation
- `pass_generic_interface.nr` now successfully generates dynamic dispatch vtables.
- `ffi_ownership_test.nr` correctly generates RAII drops and executes custom destructors, yielding `0 bytes leaked`.
- The complete test suite (`go test -v ./pkg/cmd/nora`) was re-run and passed without regressions (`ok github.com/DwiYI/Project-Nora/pkg/cmd/nora 388.411s`).
