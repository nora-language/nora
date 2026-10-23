# Language & Compiler Gap Remediation Walkthrough (Gaps 01–30)

## Summary

Following the source-verified [`docs/reports/language_gap_assessment.md`](../reports/language_gap_assessment.md) and [`docs/plans/language_gap_remediation_plan.md`](../plans/language_gap_remediation_plan.md), all 30 identified language, compiler, runtime, and standard library gaps (`pkg/cmd/test/repro_gap_01_*` through `pkg/cmd/test/repro_gap_30_*`) have been reproduced, fixed, and verified under `-DNR_DEBUG_MEM=1`.

---

## Batch 1: P0 HIR, Codegen & Lease Solver Soundness (`repro_gap_01` – `repro_gap_08`)

1. **`repro_gap_01_if_return_double_predrop`** ([`pkg/hir/lower.go`](../../pkg/hir/lower.go)):
   - Guarded pre-drop emission in `lowerStatementsWithDrops` with `if _, isRet := stmt.(*ast.ReturnStatement); !isRet` so `ReturnStatement` does not emit `PreDrops` twice.
2. **`repro_gap_02_closure_capture_owned_double_free`** ([`pkg/topology/solver.go`](../../pkg/topology/solver.go), [`pkg/codegen/generator.go`](../../pkg/codegen/generator.go)):
   - Added `IsCaptured` tracking on `Lifecycle`, anchored captured variables in the outer scope, and prevented `emitLambdaEnvDrops` from double-dropping outer scope variables.
3. **`repro_gap_03_lambda_param_lease_inversion`** ([`pkg/semantic/analyzer.go`](../../pkg/semantic/analyzer.go)):
   - Fixed `ParamLeases` assignment in `AnalyzeLambdaExpression` (`#` -> `LeaseRead`, `&` -> `LeaseWrite`, `@` -> `LeaseMove`).
4. **`repro_gap_04_try_skips_defer` & `repro_gap_05_try_skips_drop_flag`** ([`pkg/hir/lower.go`](../../pkg/hir/lower.go), [`pkg/topology/solver.go`](../../pkg/topology/solver.go), [`pkg/codegen/expressions.go`](../../pkg/codegen/expressions.go)):
   - Captured `ActiveDefers` on `hir.Try`, excluded the LHS variable being initialized in `registerTryDrops`, extended `LastUsedAt` for `TryDrops` variables so conditional moves receive drop flags, and emitted `ActiveDefers`, `g.emitReturnDrops(g.Solver.TryDrops[e])`, and `nr_flush_temps()` in `genTryExpression`.
5. **`repro_gap_06_select_branch_move_double_free`** ([`pkg/topology/solver.go`](../../pkg/topology/solver.go)):
   - Added `*ast.BlockStatement` to `isComplex` and merged nested block lifecycles back into `allVisible`.
6. **`repro_gap_07_defer_borrow_premature_drop`** ([`pkg/topology/solver.go`](../../pkg/topology/solver.go), [`pkg/hir/lower.go`](../../pkg/hir/lower.go)):
   - Propagated `AnchorEndOfFunction` through `updateLifecycle` in `DeferStatement` and emitted `activeDefers` before end-of-function `Drops` on implicit returns in `lowerBlock`.
7. **`repro_gap_08_spawn_monitor_channel_walk`** ([`pkg/parser/ast/walk.go`](../../pkg/parser/ast/walk.go), [`pkg/topology/solver.go`](../../pkg/topology/solver.go)):
   - Added `SpawnExpression.MonitorChannel` and `*ast.ConstStatement` traversal to `ast.Inspect` and `Solver`.

---

## Batch 2: P1 Generics, Sum Types, Interfaces & Operators (`repro_gap_09` – `repro_gap_19`)

8. **`repro_gap_09_generic_try_monomorph_collision`** ([`pkg/semantic/analyzer.go`](../../pkg/semantic/analyzer.go), [`pkg/codegen/expressions.go`](../../pkg/codegen/expressions.go)):
   - Added `*ast.TryExpression` cloning in `cloneAndSubstituteHelper` and dereferenced pointer `e.Value` in `genTryExpression`.
9. **`repro_gap_10_generic_chan_and_range_clone`**, **`repro_gap_11_generic_struct_chan_field`**, **`repro_gap_12_generic_slice_type_inference`** ([`pkg/semantic/analyzer.go`](../../pkg/semantic/analyzer.go)):
   - Completed `cloneAndSubstituteHelper`, `substituteType`, `hasGeneric`, `matchType`, and `typeToTypeNode` for `*ast.ChanType`, `*types.ChanType`, `*types.ArrayType`, `T[]` slice type nodes, `*ast.RangeExpression`, `*ast.SliceExpression`, `*ast.ConstStatement`, `*ast.SelectStatement`, and `*ast.PinStatement`.
10. **`repro_gap_13_sumtype_variant_field_order`** ([`pkg/semantic/analyzer.go`](../../pkg/semantic/analyzer.go)):
    - Preserved declaration order (`variant.FieldNames`) instead of alphabetical order (`sortedFieldNames`) in sum-type variant construction and pattern matching.
11. **`repro_gap_14_match_non_exhaustive_primitive`** ([`pkg/semantic/analyzer.go`](../../pkg/semantic/analyzer.go)):
    - Enforced exhaustiveness checking (`_` wildcard or `true`+`false` for `bool`) on non-SumType `match` expressions.
12. **`fail_repro_gap_15_interface_lease_mismatch`** ([`pkg/semantic/analyzer.go`](../../pkg/semantic/analyzer.go)):
    - Enforced receiver lease compatibility in `sa.implements` so read-only leases (`#T`) cannot satisfy interfaces requiring mutable (`&self`) or owned (`@self`) methods.
13. **`repro_gap_16_operator_bitclear`**, **`repro_gap_17_operator_bitnot_non_i32`**, **`repro_gap_18_postfix_inc_lvalue`**, **`repro_gap_19_array_index_uint_and_len`** ([`pkg/parser/parser.go`](../../pkg/parser/parser.go), [`pkg/semantic/analyzer.go`](../../pkg/semantic/analyzer.go), [`pkg/codegen/`](../../pkg/codegen/)):
    - Added `&^` and `&^=` support across parser, semantic analyzer, and C11 codegen (`& ~`).
    - Allowed bitwise `~` and array indexing on all signed/unsigned integer types (`i8`–`i64`, `u8`–`u64`, `uint`, `byte`).
    - Supported postfix `++`/`--` on general lvalue expressions (`arr[i]++`, `s.field++`).
    - Supported `len()` on fixed-size arrays (`T[N]`) and built-in maps (`Map[K, V]` via `map_count`).

---

## Batch 3: P2 HIR & C11 Codegen Constructs (`repro_gap_20` – `repro_gap_23`)

14. **`repro_gap_20_for_fixed_size_array`** & **`repro_gap_21_for_builtin_map`** ([`pkg/hir/lower.go`](../../pkg/hir/lower.go), [`pkg/codegen/hir_codegen.go`](../../pkg/codegen/hir_codegen.go)):
    - Added HIR lowering and `IndexAccess` codegen for iterating over fixed-size arrays (`*types.ArrayType`) and built-in maps (`*types.MapType`).
15. **`repro_gap_22_select_unselected_send_leak`** ([`pkg/codegen/hir_codegen.go`](../../pkg/codegen/hir_codegen.go)):
    - Dropped unselected owned send temporaries (and restored moved variable drop flags when unselected) after `channel_select`.
16. **`repro_gap_23_scope_spawn_ignores_monitor_chan`** ([`pkg/codegen/expressions.go`](../../pkg/codegen/expressions.go)):
    - Added full `e.MonitorChannel` support to `genSpawnExpression` inside `scope { ... }` blocks and fixed non-erased parameter pointer types in spawn wrappers.

---

## Batch 4: P3 C11 Runtime & Standard Library (`repro_gap_24` – `repro_gap_30`)

17. **`repro_gap_24_string_concat_debug_mem_accounting`** ([`std/runtime/string.c`](../../std/runtime/string.c), [`std/runtime/collection.c`](../../std/runtime/collection.c)):
    - Aligned `g_total_allocations` with `h->count * h->elem_size` for null-terminated strings in `nr_str_concat`, `nr_str_from_cstring`, `nr_strdup`, and `string_slice`.
18. **`repro_gap_25_vector_clear_and_set_leak`** & **`repro_gap_26_vector_oob_fallthrough`** ([`std/collections/vector.nr`](../../std/collections/vector.nr), [`pkg/codegen/generator.go`](../../pkg/codegen/generator.go)):
    - Dropped overwritten/cleared elements in `Vector.Set`, `Vector.Clear`, and `Vector.Resize`, restricted `Vector` auto-drop to `0..self->size`, and replaced out-of-bounds print fallthroughs with `panic(...)`.
19. **`repro_gap_27_hashmap_tombstone_exhaustion`** ([`std/collections/map.nr`](../../std/collections/map.nr), [`pkg/codegen/generator.go`](../../pkg/codegen/generator.go), [`pkg/codegen/statements.go`](../../pkg/codegen/statements.go)):
    - Guarded `MapEntry` auto-drop with `if (!self->occupied) return;`, used `unchecked_set` for unoccupied slots in `HashMap.Insert`, cleared `occupied = false` before moving entries in `rehash`, rehashed on full tombstone probe, and cleared `_target.tag = -1` + source drop flag when destructuring sum-type variants by move.
20. **`repro_gap_28_strconv_int32_min`** ([`std/strconv/strconv.nr`](../../std/strconv/strconv.nr)):
    - Handled `-2147483648` (`INT32_MIN`) in `strconv.ParseInt`.
21. **`repro_gap_29_serialize_i64_truncation`** ([`std/serialize/serialize.nr`](../../std/serialize/serialize.nr)):
    - Used `nr_i64_to_str` instead of `i32(val)` cast in `nr_serialize_json_i64`.
22. **`repro_gap_30_context_cancel_multi_fiber`** ([`std/runtime/channel.c`](../../std/runtime/channel.c), [`std/context/context.nr`](../../std/context/context.nr)):
    - Implemented `channel_close(channel_t* c)` in the C11 channel runtime to broadcast wakeup to all current and future receivers on `ctx.Done()`.
