# Investigation: Dangling Pointer on Inline Alloc in Struct Initialization

## Status
Completed (Compiler Bugs Fixed)

## Problem
In `nora_wgpu`, the triangle example was experiencing a periodic memory leak on the C backend (`wgpu-native`). A previous investigation incorrectly concluded that the issue was a "dangling pointer" caused by the Topological Lease Solver eagerly dropping a temporary inline allocation. However, further testing with `-debug-memory` revealed that Nora was actually suffering from a persistent 4-byte memory leak at the site of `alloc` expressions. The memory was not being freed early—it was never being freed at all.

## Reproduction
Inside the render loop in `examples/triangle/main.nr`, a `sys.RenderPassDescriptor` was initialized with an inline allocation for its `colorAttachments` field:

```nora
var pass_desc = alloc sys.RenderPassDescriptor {
    nextInChain: none,
    label: sys.StringView { data: none, length: 0 },
    colorAttachmentCount: 1,
    colorAttachments: alloc sys.RenderPassColorAttachment {
        view: view.handle,
        resolveTarget: none,
        loadOp: sys.WGPULoadOp.Clear,
        storeOp: sys.WGPUStoreOp.Store,
        clearValue: sys.Color { r: 0.1, g: 0.2, b: 0.3, a: 1.0 },
        depthSlice: -1
    },
    // ...
}
```

This caused the program's memory footprint to continuously increase.

## True Root Cause
The root cause was a combination of three distinct bugs in the compiler's Semantic Analyzer and Topological Lease Solver, which collectively caused unconsumed `alloc` temporaries to bypass RAII drop insertion, resulting in memory leaks across multiple scenarios.

### Scenario 1: Unconsumed `AllocExpression` Ignored
The `walkUnconsumedRValues` function in `pkg/topology/solver.go` lacked a handler for `*ast.AllocExpression`. Even when parent AST nodes correctly determined that the allocation was not consumed (e.g., evaluating it purely for its side effects or discarding the result), the solver silently skipped the node and failed to append it to the block's `Drops` list.

### Scenario 2: Assignments to Primitive/Unowned Types (`p = alloc Inner`)
When assigning a heap allocation to an unowned variable (e.g., `var p: ptr; p = alloc Inner`), the solver is supposed to recognize that `ptr` does not take ownership and therefore must insert a drop for the temporary RHS. 
However, `analyzeAssignmentStatement` in `pkg/semantic/analyzer.go` explicitly skipped analyzing the left-hand-side identifier to prevent false "use-after-move" errors during variable revival. Because it skipped analysis, `sa.SemanticInfo.Types[ident]` was never populated. When the Topological Lease Solver checked the LHS type, it received `<nil>`. The solver conservatively treated `<nil>` as an owned type, causing it to incorrectly assume the RHS `alloc` was consumed and skipping the drop.

### Scenario 3: Inline Struct Field Initialization (`StructLiteral`)
When initializing a struct with an inline allocation for an unowned field (e.g., `Outer { p: alloc Inner }` where `p` is `ptr`), the solver failed to verify the expected type of the struct field. It naively assumed all values passed into a struct literal were consumed. This caused the temporary allocation passed to the `ptr` field to be leaked, which was the exact scenario observed in the `sys.RenderPassDescriptor` initialization in `nora_wgpu`.

## Fix
The compiler frontend was patched to properly enforce RAII semantics for unconsumed allocations across all contexts:
1. **`pkg/topology/solver.go`**: Added `*ast.AllocExpression` traversal. If the allocation is an owned R-Value and is not consumed, it is now correctly added to the drops list.
2. **`pkg/topology/solver.go`**: Updated `StructLiteral` traversal to look up the struct's field definitions and accurately evaluate whether each field's type expects ownership (`isOwnedRValueType`), correctly marking `valConsumed` as `false` for `ptr` and `#T` fields.
3. **`pkg/semantic/analyzer.go`**: Added `sa.SemanticInfo.Types[ident] = sym.Type` inside the assignment identifier bypass block. This guarantees that the Topological Lease Solver has the correct LHS type information to properly evaluate whether ownership is being transferred during assignments.

## Validation
Integration tests (e.g., `codegen_dangling_pointer_inline_alloc_struct`) were added to the `pkg/cmd/test` suite. Running `Nora build -debug-memory` and executing the test confirms a strict 0-byte memory leak, validating that temporaries are now reliably dropped at the end of their parent statements.
