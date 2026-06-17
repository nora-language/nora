# Walkthrough: Removing Magic Semantic Bypasses

We have successfully removed the "God Mode" bypasses in the Nora compiler and fully replaced them with strict, attribute-based opt-ins!

## What Was Changed

### 1. Replaced String Mutation Bypasses with `[unsafe]` Attribute
We completely removed the hardcoded write-permission bypasses (like `strings.Contains(filename, "std/")`) in `pkg/semantic/analyzer.go`.

Instead, we introduced a formalized `[unsafe]` attribute. The compiler now explicitly checks whether a mutating function carries the `[unsafe]` attribute before allowing normally illegal pointer and string manipulations:
- In `analyzer.go`, `checkWritePermission` now explicitly checks the `[unsafe]` attribute and verifies the context.
- We support opting-in to `[unsafe]` capabilities either via a project-wide manifest setting (`allow_unsafe: true` in `nora.yaml`) or explicitly via the `--allow-unsafe` CLI flag, keeping standard code safe by default without penalizing users consuming standard libraries.

### 2. Standard Library Safely Opts-In
We meticulously traced the entire standard library (`std/`) and applied `[unsafe]` to methods that genuinely required raw mutation:
- **Collections**: `list.nr`, `vector.nr`, and `map.nr` mutating methods like `Push`, `GetMut`, `Insert`, and `Pop`.
- **String Operations**: `string.nr` methods that build strings natively (`ToUpper`, `ToLower`, `Substring`, `Join`, `Replace`).
- **IO Utilities**: `Scanner.Next()` and `NextLine()` in `io.nr`.
- **Testing**: We also applied `[unsafe]` to relevant standard library test cases in `pkg/cmd/test/stdlib_udp_test` that directly assigned to string indexes.

### 3. Eliminated Hardcoded Core Type Checks
We successfully eradicated the hardcoded type name checks (`"Vector"`, `"List"`, `"Result"`, `"Option"`) from the compiler's semantic checking and topology solving logic:
- Designed and implemented a `[core_intrinsic("...")]` attribute parsed natively during symbol resolution.
- Annotated types in `std/prelude.nr` (`Result`, `Option`) and `std/collections/` (`Vector`, `LinkedList`, `HashMap`) with their respective `core_intrinsic` markers.
- Refactored `analyzer.go`, `topology/solver.go`, and `codegen/expressions.go` to depend completely on `Type.CoreIntrinsic` metadata rather than doing unreliable `strings.Contains` checks on the file path or type name.

## Validation Results

The full Nora integration test suite (`go test ./pkg/cmd/nora -v`) has been completed! All standard library tests pass successfully under the new strict enforcement structure, and the legacy bypasses have been completely eradicated without sacrificing performance or system capabilities.

> [!TIP]
> The language now natively supports attribute tagging via `[unsafe]` and `[core_intrinsic(...)]` which paves the way for further extensions like `[no_mangle]` or `[inline]` in the future!
