# Investigation: Type-Erased Closures Capturing Generic Primitive Structs

## Problem
When a closure captures a generic struct containing primitive types (e.g., `&Vector3[T]`) and is passed to a type-erased higher-order function (e.g., `ParMap[BodyPtr[T]]`), the C code generation fails with:
`error: unknown type name 'Vector3'`
`error: unknown type name 'BodyPtr'`

The compiler attempts to generate a single, shared C environment struct (`nr_lambda_..._env_t`) and C function prototype for the lambda because `ParMap` is erased. However, the generator mistakenly emits the literal unmangled string `"Vector3*"` and `"BodyPtr*"` instead of properly erased types (like `void*`) or fully monomorphized concrete types (like `vector_Vector3_f64*`).

## Reproduction
Test case added to `pkg/cmd/test/pass_closure_capture_generic/main.nr`.
```nora
pub fn Step[T](wrapper: &BodyPtr[T], gravity: &Vector3[T]) {
    wrapper.DoMap[T](fn(b_wrapper: &BodyPtr[T]) {
        var g = #gravity
        b_wrapper.ptr.position.x = g.x
    })
}
```
If `T=f64`, `BodyPtr[f64]` is pointer-like, causing `DoMap` to be type-erased. The lambda is generated globally.

## Root Cause
1. **Generic Lambda Generation**: The lambda environment and prototype are generated in `generator.go` inside `genHIRFunction` by iterating over all lowered HIR functions globally.
2. **Missing Erasure of Lambda Captures & Params**: When emitting the environment struct:
   ```go
   envDef.WriteString(fmt.Sprintf("    %s %s;\n", g.cType(cap.Type), g.mangleName(cap)))
   ```
   and the parameters:
   ```go
   fnDef.WriteString(g.cParamType(p, lease))
   ```
   `g.cType` is called on the generic `cap.Type` (e.g., `&Vector3[TypeParam]`). Because this is a generic type, `mangledTypeName` fails to find it in `g.Structs` (which only stores monomorphized concrete types) and falls back to `t.Name()`, printing the literal string `"Vector3*"`.
3. **No Explicit Cast Back to Concrete Types**: Inside the lambda body, if the lambda captures are erased to `void*`, they would need to be safely cast back to the specific concrete type (e.g., `(vector_Vector3_f64*)_env->gravity`) for the operations inside the lambda body to succeed.

## Proposed Solution (See Implementation Plan)
The lambda must be emitted as an erased function where all pointer-like arguments and captures are properly passed through `g.eraseType()`. When compiling the body of the lambda, the generator must inject explicit casts to allow the erased `void*` fields to be correctly utilized as their intended concrete types. Alternatively, the compiler needs to clone lambdas alongside their monomorphized parent functions.
