# Inliner Name Collision (`inliner.hirFuncs`) Fix Plan

## Status
Completed

## Metadata
- **Author:** Antigravity
- **Date:** 2026-07-18
- **Components:** `pkg/hir/optimize`
- **Related Investigations:** `docs/investigations/inliner_name_collision.md`, `docs/investigations/simd_compiler_collision.md`

## Goal
To resolve the critical symbol name collision bug in the High-level Intermediate Representation (HIR) inliner pass (`pkg/hir/optimize/inliner.go`). Currently, the inliner stores target functions in a map (`inliner.hirFuncs`) keyed purely by the short unmangled function name (`hf.FuncSymbol.Name` or `hf.Name`, e.g., `"Sqrt"` or `"Set1"`). When multiple packages (or functions across different modules) declare functions sharing the exact same short name—such as `math.Sqrt` and `simd.Sqrt`—they collide and overwrite each other in `inliner.hirFuncs`. When the inliner encounters a call to `math.Sqrt`, querying `inliner.hirFuncs["Sqrt"]` returns the HIR body of `simd.Sqrt` and incorrectly inlines it into the `math.Sqrt` callsite.

We will eliminate this collision completely by keying inline functions primarily by their exact `*semantic.Symbol` pointer while retaining name-based fallbacks for synthetic or symbol-less functions.

## Affected Compiler Components
- `pkg/hir/optimize/inliner.go`:
  - `Inliner` struct definition (`hirFuncs` map).
  - `runInlinePass` where functions are registered into `Inliner`.
  - `processBlock` (`case *hir.Call`) where functions are looked up and inlined.

## Implementation Checklist
### 1. Update `Inliner` Struct & Function Registration (`inliner.go`)
- [x] Add `hirBySymbol map[*semantic.Symbol]*hir.Function` alongside `hirByName map[string]*hir.Function` inside the `Inliner` struct.
- [x] In `runInlinePass()`, register every function `hf` into `inliner.hirBySymbol[hf.FuncSymbol] = hf` whenever `hf.FuncSymbol != nil`.
- [x] Retain registering `hf` into `inliner.hirByName[name]` as a safe fallback for any functions where `FuncSymbol` is nil.

### 2. Update Function Lookup in `processBlock()` (`inliner.go`)
- [x] In `processBlock()` under `case *hir.Call:`, first attempt to look up `targetFunc` by `inl.hirBySymbol[i.FuncSymbol]` when `i.FuncSymbol != nil`.
- [x] If not found via symbol lookup (or if `i.FuncSymbol == nil`), fall back to `inl.hirByName[targetName]`.
- [x] Verify that `targetFunc.FuncSymbol.IsInline` is checked accurately before cloning and lowering parameters.

### 3. Verification & Regression Testing
- [x] Create a dedicated reproduction test suite in `pkg/cmd/test/inliner_collision_test/inliner_collision_test.nr` along with a secondary module/file or local functions that share identical short names between `[inline]` and non-inline functions.
- [x] Verify that before the fix (or demonstrating the exact scenario), `inliner.go` caused symbol substitution errors.
- [x] Verify that after the fix, `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder/pkg/cmd/test/inliner_collision_test` passes cleanly without any name collision or incorrect function body substitution.
- [x] Run the full integration test suite (`go test ./pkg/cmd/nora/...`) to ensure zero regressions across all packages.

## Test Plan
1. **New Integration Test:** `pkg/cmd/test/inliner_collision_test/inliner_collision_test.nr` will import or define multiple functions sharing the same short name (`Foo` or `Calc`), where one variant (`pkg_a.Calc`) is marked `[inline]` returning value `A` and another variant (`pkg_b.Calc` or `local.Calc`) is not inlined (or `[inline]` with a different formula) returning value `B`. The test will assert that calling each function returns the exact correct return value without cross-inlining substitution.
2. **Regression Suite:** Run `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder` across existing benchmarks and SIMD suites (`simd_test`, `spectralnorm`, `stdlib_math_test`, etc.).

## Risks
- None. Keying by `*semantic.Symbol` pointer is guaranteed to be 100% unique per function declaration within the AST and semantic analyzer (`pkg/semantic/symbol.go`), and falling back to `string` preserves compatibility with any synthetic symbol-less functions.

## Completion Criteria
- Plan created and documented under `docs/plans/inliner_name_collision_plan.md`.
- Test cases created under `pkg/cmd/test/inliner_collision_test/`.
- `inliner.go` updated and formatted.
- All integration and compiler tests pass without errors.
