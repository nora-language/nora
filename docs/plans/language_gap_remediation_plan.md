# Title: Language & Compiler Gap Remediation Plan (Gaps 01–30)

## Status
Completed

## Metadata
- **Date**: 2026-10-23
- **Related Assessment**: [`docs/reports/language_gap_assessment.md`](../reports/language_gap_assessment.md)

## Goal
Systematically resolve all 30 reproduced compiler, runtime, and standard library gaps (`pkg/cmd/test/repro_gap_01_*` through `pkg/cmd/test/repro_gap_30_*`) without regressing existing integration tests.

## Affected Compiler Components
1. **AST & Parser (`pkg/parser/`, `pkg/parser/ast/`)**:
   - `walk.go`: Traverse `*ast.ConstStatement` and `SpawnExpression.MonitorChannel`.
   - `parser.go`: Register `token.AND_NOT_ASSIGN` (`&^=`) in `precedences` and `registerInfix`; generalize postfix `++`/`--` to lvalue expressions (`SelectorExpression`, `IndexExpression`, `PrefixExpression`).
2. **Semantic Analyzer (`pkg/semantic/analyzer.go`)**:
   - Fix `AnalyzeLambdaExpression` parameter lease mapping (`#` -> `LeaseRead`, `&` -> `LeaseWrite`, `@` -> `LeaseMove`).
   - Complete `cloneAndSubstituteHelper` for `*ast.TryExpression`, `*ast.ChanType`, `*ast.ConstStatement`, `*ast.SelectStatement`, `*ast.RangeExpression`, `*ast.SliceExpression`, and `*ast.PinStatement`.
   - Complete `substituteType`, `hasGeneric`, `matchType`, and `typeToTypeNode` for `*types.ChanType`, `*types.ArrayType`, and slice/list `T[]` (`IndexExpression` with empty `Indices`).
   - Preserve declaration order (`variant.FieldNames`) instead of alphabetical order (`sortedFieldNames`) in sum-type constructor calls and `match` patterns.
   - Enforce exhaustiveness checking on non-sum types (`i32`, `str`, `bool`, `struct`) in `match`.
   - Support `&^` in infix expressions, `~` on all integer types, `len()` on fixed-size `ArrayType` and `MapType`, and all integer types in `isIntegerType`.
3. **Topological Lease Solver (`pkg/topology/solver.go`, `cfg.go`)**:
   - Track `DeferStatement` variable usages as anchored to the end of the function (`AnchorEndOfFunction`).
   - Track `SelectStatement` cases as independent branches (`cloneLifecycles`) so moves in one case do not falsely trigger use-after-move inside that same case and properly mark conditional moves.
   - Handle closure captures of owned variables so they are not dropped both by the closure environment destructor and the outer scope.
   - Handle interface cast lease compatibility (`#Counter` coerced to interface).
4. **HIR Lowerer & C11 Codegen (`pkg/hir/lower.go`, `pkg/codegen/`)**:
   - Skip `PreDrops` before `*ast.ReturnStatement` in `lowerStatementsWithDrops` (`pkg/hir/lower.go:1561`).
   - Emit active `defer` calls and check dynamic drop flags (`g.hasDropFlag`) in `TryExpression` (`?`) early returns.
   - Emit `& ~` for `&^` in C11 codegen.
   - Support `for` loops over fixed-size arrays (`*types.ArrayType`), `Range` variables, and built-in `Map[K, V]`.
   - Drop unselected non-channel owned values in `hir.Select` codegen (`pkg/codegen/hir_codegen.go:356`).
   - Support `MonitorChannel` inside `genSpawnExpression` (`pkg/codegen/expressions.go:2165`) and prevent `ast.Inspect` from prematurely dropping `mon_ch`.
5. **C11 Runtime & Standard Library (`std/runtime/`, `std/`)**:
   - Fix `g_total_allocations` accounting in `std/runtime/string.c` (`nr_str_concat`, `nr_str_from_cstring`, `nr_strdup`) and `std/runtime/collection.c` (`string_slice`).
   - Fix `std/collections/vector.nr` (`Clear` and `Set` element drops; `Pop`, `Get`, `GetMut`, `Remove`, `Insert` bounds panics).
   - Fix `std/collections/map.nr` (`Remove` double-drop of value / tombstone cleanup).
   - Fix `std/strconv/strconv.nr` (`ParseInt` overflow check for `INT32_MIN`).
   - Fix `std/serialize/serialize.nr` (`nr_serialize_json_i64` 64-bit formatting).
   - Fix `std/context/context.nr` (`Cancel` waking multiple fibers waiting on `Done()`).

## Implementation Checklist
- [x] Batch 1: P0 HIR, Codegen & Lease Solver Soundness (`repro_gap_01` – `repro_gap_08`)
- [x] Batch 2: P1 Generics, Sum Types, Interfaces & Operators (`repro_gap_09` – `repro_gap_19`)
- [x] Batch 3: P2 HIR & C11 Codegen Constructs (`repro_gap_20` – `repro_gap_23`)
- [x] Batch 4: P3 C11 Runtime & Standard Library (`repro_gap_24` – `repro_gap_30`)

## Test Plan
- Verify all 30 `pkg/cmd/test/repro_gap_*` tests pass via `go test -v ./pkg/cmd/nora -run "TestCompilerWithTestFolder/.*repro_gap_"`.
- Run full compiler & integration test suite across `pkg/`.

## Risks
- Modifying closure capture ownership or `DeferStatement` lifetime anchoring could affect existing closure or `defer` tests if read-only borrows vs owned captures are not distinguished carefully.
- Modifying `SelectStatement` branch merging in `solver.go` must set dynamic drop flags when a variable is moved in only a subset of cases.

## Completion Criteria
- All 30 `repro_gap_*` tests pass with zero memory leaks.
- Existing test suite passes without regressions.
- Plan marked `Completed` and walkthrough documented in `docs/walkthroughs/`.
