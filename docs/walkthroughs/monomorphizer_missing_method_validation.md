# Walkthrough: Monomorphizer Missing Method Eager Validation

## Overview
When a generic struct or sum type is instantiated and a method call is attempted on an undefined method (e.g., `v1.Clone[T]()` on `Vector3[T]` where `Clone` is not defined), the compiler previously risked silent panics or messy internal diagnostic output during downstream monomorphization.

This walkthrough documents the compiler-side implementation that eagerly validates method and field existence during the semantic analysis pass (`pkg/semantic`), formatting clean user-facing diagnostics and preventing invalid AST nodes from reaching the monomorphizer.

## Summary of Changes

### 1. Semantic Analyzer (`pkg/semantic/analyzer.go`)
- **Clean Type Names in Diagnosics**: Updated `SelectorExpression` resolution for `StructType` and `SumType` so that when a field or method lookup fails, the emitted error uses the clean type name (`t.Name()` or `t.BaseType.Name()`) rather than internal debug pointers (`(ptr: 0x...)`).
- **Protocol Constraint Validation**: Updated generic type parameter method resolution (`case *types.GenericType:`). If a method lookup on a bounded protocol constraint fails, the analyzer now immediately emits a diagnostic error (`interface constraint '%s' for type parameter '%s' has no method '%s'`) instead of silently returning.
- **Specialization Guarding**: Added checks in `Monomorphize()` and `ensureMethodsSpecialized()` to verify `!sa.Diagnostics.HasErrors()`. If semantic checking has already produced errors, specialization is bypassed to prevent redundant error cascades and attempting to monomorphize malformed ASTs.

### 2. Regression Testing (`pkg/cmd/test/`)
- Added a negative integration test: `pkg/cmd/test/fail_missing_generic_method/fail_missing_generic_method.nr`.
- Verified that running `nora build` or `nora test` on this file produces clean semantic error diagnostics without crashing or panicking.

## Verification Results
- Ran `go test ./pkg/cmd/nora -run TestCompilerWithTestFolder/fail_missing_generic_method -v` -> **PASS**.
- Ran `go test ./pkg/semantic` -> **PASS**.
