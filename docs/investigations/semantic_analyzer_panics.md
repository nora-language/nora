# Semantic Analyzer Panics: Generic Inference Cascading Failure

**Status:** Resolved (Fixed directly in Compiler `pkg/semantic`)

## Problem
During the implementation of Phase 25 (Shader Cross-Compilation & Reflection) in `nora_wgpu`, the Nora compiler's semantic analyzer (`github.com/nora-language/nora/pkg/semantic.(*SemanticAnalyzer).Analyze`) repeatedly panicked with `invalid memory address or nil pointer dereference` when compiling `examples/shader_reflection/main.nr`.

Previous investigations incorrectly assumed these panics were caused by:
1. Instantiating a `collections.Vector[T]` where `T` is a struct implementing a `drop()` method.
2. Taking the address of a stack-allocated struct (`ptr(bg_entry)`) and assigning it directly to a pointer field within an initialization block.

However, these were **red herrings**. The panic was actually a cascading failure that masked the true underlying semantic errors.

## Root Cause
The true root cause of the panic lay in a missing error fallback inside the compiler's generic type inference engine (`pkg/semantic/analyzer.go`).

1. The compiler encountered a legitimate semantic error (e.g., trying to use `ptr(bg_entry)` on an incompatible type, or failing to infer `T` from `parse_res.IsErr()`).
2. Type inference for a generic function call (`IsErr[T, E]()`) failed, causing `sa.inferTypeArguments` to return `nil`.
3. The `handleGenericCall` function checked if `typeArgs == nil` and simply returned early **without** explicitly assigning `types.ErrorType` to the AST node.
4. As a result, the type of the expression node became implicitly `nil`.
5. When this `nil`-typed expression was evaluated as a condition inside an `IfExpression`, the compiler attempted to call `.Name()` on the `nil` type interface, resulting in a hard crash (`nil pointer dereference`).

Because the compiler crashed at this stage, it completely hid all preceding semantic errors from the developer, leading to the assumption that the `ptr()` casts or generic drops were the direct cause of the panic.

## Fix
The immediate fix was applied directly to the compiler frontend (`pkg/semantic/analyzer.go`).

Inside `handleGenericCall`, a fallback was added to explicitly assign `types.ErrorType` to the node whenever type inference fails:
```go
	if len(n.TypeArguments) == 0 {
		typeArgs = sa.inferTypeArguments(fnStmt, n)
		if typeArgs == nil {
            sa.SemanticInfo.Types[n] = types.ErrorType // <--- FIX
			return // Inference failed
		}
	}
```

## Validation
By applying this compiler patch, the panic is fully mitigated. 
An isolated negative test case has been added to the compiler's integration suite at `pkg/cmd/test/fail_semantic_analyzer_panics/fail_semantic_analyzer_panics.nr` to guarantee that generic inference failures gracefully emit `Error: could not infer type for generic parameter T` instead of panicking.

## Next Steps for `nora_wgpu`
With the compiler now correctly reporting semantic errors instead of panicking, the original `nora_wgpu` workarounds (`ffi.BorrowToRaw`) can be re-evaluated. The compiler will now accurately flag the true semantic violations (such as `Error: cannot cast WGPUBindGroupEntry to ptr`), allowing for correct API usage to be implemented based on accurate compiler feedback.
