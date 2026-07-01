# Investigation: Value-Type Assignment Segfaults & Pass-by-Value Memory Leak

Status: Completed

## Problem
Two related memory-management issues surfaced during test runs (`repro_leak_struct_by_val`, `test_array_header_corruption`, and `repro_gecs_serialization_leak`):
1. **Segmentation Faults (Segfaults):** The C compiler emitted binaries that crashed when returning or assigning a struct/value-type into an uninitialized stack pointer.
2. **14-byte Memory Leak:** Deserializing a JSON payload into a heap-allocated struct pointer and subsequently moving it into a value-type parameter (`store_component(@val)`) resulted in a 14-byte memory leak. The topological solver believed the local function still owned the pointer (generating a drop/free at the end of the block), but the codegen had nullified the pointer during the move, rendering the final free a no-op.

## Reproduction
**Segfault:**
```nr
pub fn get_component() MyComponent {
    return MyComponent { x: 10 }
}

pub fn main() i32 {
    let comp = get_component() // Depending on codegen, comp might not have valid block-scoped memory.
    // Reading comp.x could trigger a segfault.
}
```

**Memory Leak (`repro_gecs_serialization_leak`):**
```nr
let val = nr_deserialize_json_MyComponent(...) // val is a heap-allocated pointer
store_component(@val) // store_component takes `comp: MyComponent` (by-value)
```

## Root Cause
1. **Segfault:** In `pkg/codegen/hir_codegen.go`, when translating an HIR `Assign` or `Store` from a value-type right-hand side (RHS) to a pointer destination, the generated C code essentially attempted to directly assign the value without first ensuring the pointer had valid memory allocated on the C stack.
2. **Memory Leak:** 
    - The Nora compiler's `AddressOf` (`@`) operation implementation inside `hir_codegen.go` generated a C statement expression that automatically nullified heap pointers (`val = NULL;`) to prevent double frees. 
    - At the same time, the `cleanMovedHeapPointers` function was responsible for implicitly freeing heap pointers after they were moved. However, `cleanMovedHeapPointers` checked if the *C-level* parameter type was a pointer (`g.isPointerTypeInC`). Because `store_component` accepted a value-type, it correctly copied the data but did *not* take ownership of the pointer itself.
    - As a result, `cleanMovedHeapPointers` emitted an `nr_free(val); val = NULL;`. But because `val` was *already* nullified inside the `@` instruction, the `nr_free` call silently did nothing, leaving the original 14-byte heap allocation orphaned.

## Fix
1. **Segfault Fix:** Updated `Assign` and `Store` logic in `hir_codegen.go` to explicitly wrap value-type RHS assignments inside C compound array literals (`((Type[]){ val })`) when the destination is a pointer. This ensures the value is immediately backed by block-scoped stack memory, making it safe to dereference or access fields.
2. **Memory Leak Fix:**
    - Removed the unconditional `val = NULL` nullification from the `AddressOf` logic for `VarOperand`s in `hir_codegen.go`. Nora already utilizes runtime drop flags (`_df_val = false`) to accurately prevent double-frees, so manual nullification was redundant and harmful.
    - Updated `cleanMovedHeapPointers` and `processStore` to use `!types.IsPointerLike(paramType)` instead of `g.isPointerTypeInC()`. This ensures that when an allocated pointer is moved into a *value-type parameter*, the caller properly retains responsibility for freeing the heap allocation immediately after the value copy is performed.

## Validation
1. `test_array_header_corruption` and `repro_leak_struct_by_val` pass without segmentation faults.
2. `repro_gecs_serialization_leak` passes completely without memory leaks.
3. The full Nora integration test suite (`go test -v ./pkg/cmd/nora`) runs and passes completely, verifying that no codegen regressions were introduced.
