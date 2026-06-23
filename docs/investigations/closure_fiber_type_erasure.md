# Investigation: Type-Erased Closures Across Fiber Boundaries

## Problem
The `ParMap` method for parallel vector iterations uses the `spawn` primitive to run closures concurrently across fibers. During compilation, the C code generation produced two distinct warnings (which the test suite considers errors):
1. `warning: incompatible pointer types passing 'void **' to parameter of type 'char **'`
2. `warning: incompatible pointer types passing 'char **' to parameter of type 'void **'`

These warnings appeared in `__spawn_wrapper_%d` functions and type-erased closure invocations (`collections_run_par_map_worker_ptr`). Because primitive pointers like `int*` (for `i32`) were not type erased, whereas `char**` (for `str`) was type erased to `void**`, the generated C argument types misaligned severely between the caller and the function signature.

## Reproduction
When attempting to `spawn` a type-erased closure function containing a pointer-like type:
```nora
[unsafe]
fn run_par_map_worker[T](val: &T, f: #fn(&T)) {
    f(val)
}

// ...
spawn run_par_map_worker[T](ref, #f)
```
Invoking the generic test `TestCompilerWithTestFolder/par_map_test` (which uses `Vector[str]`) or `TestCompilerWithTestFolder/collections_vector_test` (which uses `Vector[i32]`) would fail with Clang pointer mismatch errors.

## Root Cause
There were three intertwined root causes:

1. **Over-Erasure of Function Closures:**
   In `pkg/codegen/generator.go`, the `IsPointerLike` function returns `true` for `FunctionType` (closures). The `eraseType` function would encounter the closure parameter and forcibly overwrite its erased shape to a raw `types.Ptr` (`void*`). This destroyed the native `nr_closure_t` struct layout, breaking C backend generation entirely.

2. **Missing Casts in Closure Function Pointers (`genCallExpression`):**
   When invoking the type-erased closure via `_c.fn_ptr`, the function signature retrieved via `getCFunctionPointerType(ft)` required the explicit unerased types (e.g. `char**`). However, inside the erased worker body, the argument was represented as `void**`. C requires an explicit cast when converting `void**` to `char**`.

3. **Missing Casts in Fiber Wrappers (`genSpawnExpression`):**
   The generated M:N scheduler wrappers (`__spawn_wrapper_%d`) accept arguments inside a packed `args` struct. When expanding these arguments into the generic target function, the argument expressions were passed as strictly typed `arg%d`. Because the target function could be dynamically type-erased (`void**`) OR monomorphized (`int*`), the argument structs required uniform flexible casting.

## Fix
1. **Preserve `nr_closure_t` Type Shape**: Updated `eraseType` within `generator.go` (around line 652) to check `pt.Base.(*types.FunctionType)` and preserve the closure pointer type, preventing fallback to a raw `types.Ptr`.
2. **Explicit Casting in Closure Calls**: Modified `genCallExpression` in `expressions.go` to explicitly retrieve the `paramCType` and inject an explicit cast (e.g. `(char**)args->arg0`) when resolving variable pointer assignments.
3. **Generic `(void*)` Casts for Spawn Worker Structs**: Modified `genSpawnExpression` in `expressions.go` to intercept mismatched pointer arguments. Since C permits implicit casting from `void*` into ANY pointer type without warnings, we inject `(void*)` around the argument expression when the target parameter is a pointer. This robustly satisfied both type-erased (`void**`) and monomorphized (`int*`) function calls transparently.

## Validation
Executing `go test -v ./pkg/cmd/nora` against the integration suites:
- `par_map_test.nr` (`str` erased types) passes flawlessly.
- `collections_vector_test.nr` (`i32` primitive monomorphizations) passes flawlessly.
- No `void**` vs `char**` warnings remain in C emission.

Status: **Resolved**
