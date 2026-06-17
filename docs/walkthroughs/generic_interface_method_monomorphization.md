# Generic Interface Method Monomorphization Walkthrough

## Goal
Fix a compilation bug where methods of a generic struct implementing an interface fail to compile/link, causing undefined reference errors during code generation (e.g. `undeclared function 'Pair_cb8e8945_Map_a870e67c'`).

## Root Cause
When monomorphizing a generic method (e.g. `impl3.Map[i32](99)`) belonging to a generic receiver (e.g., `Pair[i32, str]`), the semantic analyzer in `Monomorphize` was incorrectly attempting to bind the receiver's type parameters using the method's own generic `typeArgs`. Since the method only has one type argument (`[i32]` for `V`), the receiver's type arguments (`T` and `U`) were left unbound. This led the code generator to treat the struct methods as generic templates and omit their concrete C generation, leading to compiler failures.

## Solution
We updated `pkg/semantic/analyzer.go` inside the `Monomorphize` function:
1. Dynamically extract the underlying receiver type (unwrapping pointer types).
2. Read the concrete type arguments (`TypeArgs`) from the receiver `StructType` or `SumType`.
3. Bind the receiver's implicit type parameters to its own type arguments (`recTypeArgs`) instead of the method's generic `typeArgs`.

## Validation
- Successfully compiled and ran the new test case `pkg/cmd/test/pass_generic_interface/pass_generic_interface.nr`.
- Verified that all integration tests in `pkg/cmd/test/...` pass successfully using the Go test runner `go test ./pkg/cmd/nora -v`.
