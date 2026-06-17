# Walkthrough - Compiler Fixes: Pattern Matching, Monomorphization Type-Erasure, and Test Suite Crash Detection

## Goals
1. Fix a compile-time type mismatch error (`cannot return value of type #T from function returning T`) when pattern matching on unprefixed function parameters or method receivers (e.g. `self: Option[T]` in `prelude.nr`'s `Unwrap`).
2. Fix a C compiler error (`member reference base type 'void *' is not a structure or union`) during generic struct monomorphization (e.g. `generic_hashmap_test.nr` calling `resize_hashmap`).
3. Add hardware/OS-level crash (segmentation fault / access violation) detection to the integration test runner to prevent failing runs from silently passing or timing out.

## Root Causes & Solutions

### 1. Pattern Matching Lease Detection
- **Root Cause**: Unprefixed function parameters and receivers default to `types.LeaseRead` internally. The pattern matcher checked `sym.LeaseKind == LeaseRead` to determine if a target was leased, causing owned parameters like `self` to be treated as leased. This wrapped matched fields in read-leases (`#T`), causing generic return type errors.
- **Solution**: Modified `analyzePattern` target check in [analyzer.go](file:///e:/Project/Project%20Chronos/second/pkg/semantic/analyzer.go#L1619-L1625) to check whether `sym.Type` is a leased type (`sym.Type.IsLeased()`) rather than looking at `sym.LeaseKind`.

### 2. Monomorphization Type-Erasure base type mismatch
- **Root Cause**: During monomorphization type-erasure, the generator overrode all pointer-like parameters and return types (such as `&HashMap[K, V]` which is a pointer to the type-erased structure `HashMap_ptr`) to `types.Ptr` (`void*` in C). This converted the struct lease type to `void*`, and when compiling its lease kind (`LeaseWrite`) it added another pointer layer resulting in `void** self` in the C signature. Inside the function body, the fields of `self` (`self->buckets`, etc.) were accessed as a struct pointer, causing type errors in C.
- **Solution**: Updated `collectDefinitions` in [generator.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/generator.go#L599-L610) to preserve pointers to struct or sum types (`&HashMap_ptr`, `&SumType_ptr`) when erasing pointer-like parameters and return types rather than forcing them to `types.Ptr`.

### 3. Test Runner Crash Detection
- **Root Cause**: The integration test runner checked for non-zero exit codes to verify expected panic/deadlock tests. However, it had no distinct detection for hardware/OS crashes (access violations / segfaults / stack overflows), which would allow actual C-level memory corruption bugs to pass if expected to exit non-zero, or fail with vague runtime errors.
- **Solution**: Updated [compiler_test.go](file:///e:/Project/Project%20Chronos/second/pkg/cmd/nora/compiler_test.go#L268-L278) to explicitly intercept common crash codes on Windows (`0xC0000005`, `0xC000001D`, `0xC0000094`, `0xC00000FD`) and signaled termination on Unix, outputting a clear `SEGMENTATION FAULT / CRASH` failure error.

## Verification
- Wrote regression test [repro_prelude_unwrap.nr](file:///e:/Project/Project%20Chronos/second/pkg/cmd/test/repro_prelude_unwrap/repro_prelude_unwrap.nr) which now runs successfully.
- Verified that `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder/generic_hashmap_test` compiles, runs, and passes cleanly.
- Verified that all Go unit/integration test suites under `./pkg/semantic/...`, `./pkg/lsp/...`, and `./pkg/cmd/nora/...` pass fully.
