# Investigation: Generic Closure Codegen Compilation Failure

## Problem
When compiling a generic function that contains a closure (e.g. `Step[T]` capturing a generic type like `&Vector3[T]`), the generated C code produced an `unknown type name 'Vector3'` error in Clang.
The generated C header for the lambda's environment struct was using the uninstantiated template name (`Vector3`) instead of the monomorphized type-erased name (`Vector3_b296169e`).

## Reproduction
This was reproduced using the positive test `pkg/cmd/test/pass_closure_capture_generic/main.nr` where `Step[f64]` was called.
The generic template `Step[T]` contains a `wrapper.DoMap[T](fn(...) { ... })` inline closure which captures `#gravity` of type `&Vector3[T]`.

## Root Cause
The `pkg/hir/lower.go` lowering pass was not skipping the AST of un-instantiated generic templates. `shouldSkipHIR` was hardcoded to `return false`.
Because of this, the AST for the generic template `Step[T]` was lowered to HIR in addition to the instantiated `Step_f64`.
During lowering, the lambda inside `Step[T]` was added to the `lambdaFuncs` list. Its captured environment had the generic type `Vector3[T]`, which has the name `Vector3`.
Later, when `emitLambdaEnvDefs()` generated the C struct definitions for ALL lambdas in `hirProg`, it generated a C struct for the generic template's lambda:
```c
typedef struct {
    Vector3* gravity;
} nr_lambda_XXXX_env_t;
```
Since `Vector3` is an uninstantiated Nora template, it is never emitted as a concrete C struct, causing Clang to fail compilation.

## Fix
In `pkg/hir/lower.go`, we implemented `shouldSkipHIR` to correctly return `true` for functions with type parameters or those marked as `IsGenericTemplate`.
```go
func (l *Lowerer) shouldSkipHIR(sym *semantic.Symbol, fn *ast.FunctionStatement) bool {
	if fn.IsGenericTemplate || len(fn.TypeParameters) > 0 {
		return true
	}
	return false
}
```
This entirely prevents generic templates from being translated to HIR or reaching the C-codegen phase, leaving only fully specialized/monomorphized instances (like `Step_f64`) in the pipeline, which correctly generate `Vector3_b296169e* gravity`.

## Validation
The compilation error was fully resolved. `nora build` successfully compiled the generated `main.c`.
However, execution revealed a secondary bug: an 8-byte memory leak from the inline closure's environment allocation, which must be addressed separately by tracking r-values in the Topological Lease Solver.
