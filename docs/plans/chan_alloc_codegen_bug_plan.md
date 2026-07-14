# Channel Allocation Codegen Bug (`@T` sent to `chan[T]`) Fix Plan

## Status
Completed

## Metadata
- **Author:** Antigravity
- **Date:** 2026-07-14
- **Components:** `pkg/codegen`

## Goal
To fix a critical code-generation bug where sending an allocated/owned struct value (`@T` or `alloc T`) through a `chan[T]` channel passes the address of the pointer variable (`&_send_val` of type `T**`) to `channel_send` instead of passing the pointer value (`_send_val` of type `T*`). When `channel_send` executes, it copies `sizeof(T)` bytes from the address `&_send_val`, which copies the pointer address itself (plus surrounding stack garbage) into the channel buffer instead of copying the struct value. Receiving from the channel then yields a corrupted struct containing the heap address instead of the actual struct fields.

## Affected Compiler Components
- `pkg/codegen/hir_codegen.go`: Handles lowering of `*hir.ChanSend` instructions to C code in `hirInstStr()`.
- `pkg/codegen/expressions.go`: Handles lowering of AST `*ast.SendExpression` in `genSendExpression()`.

## Implementation Checklist
### 1. Update `*hir.ChanSend` Lowering in `hir_codegen.go`
- [x] In `hirInstStr()` (`case *hir.ChanSend:` around line 1410), determine the underlying element type of the channel (`ElemType`) by unwrapping `i.Chan.GetType()`.
- [x] Check whether `ElemType` is a value struct in C (`!g.isPointerTypeInC(ElemType) && !strings.HasSuffix(g.cType(ElemType), "*")`).
- [x] Check whether the sent value (`i.Val.GetType()`) is represented as a pointer in C (`g.isPointerTypeInC(ValType) || strings.HasSuffix(g.cType(ValType), "*")`).
- [x] If sending a pointer (`ValType`) into a struct-value channel (`ElemType`), pass `_send_val` (or `(void*)_send_val`) to `channel_send` instead of `&_send_val` so that `sizeof(T)` bytes are copied from the heap struct rather than from the pointer variable on the stack.

### 2. Update `genSendExpression()` in `expressions.go`
- [x] In `genSendExpression()` (`case *ast.SendExpression` around line 2229), perform the same check between `types.UnwrapLease(g.SemanticInfo.Types[e.Left])` (the channel type) and `g.SemanticInfo.Types[e.Right]` (the value type).
- [x] If the channel element type is a struct value and the value being sent is a C pointer (e.g., from `alloc` or `@` move of an owned struct), emit `channel_send(_c, _send_val);` instead of `channel_send(_c, &_send_val);`.

### 3. Verification & Regression Testing
- [x] Create a reproduction test in `pkg/cmd/test/chan_alloc_codegen/chan_alloc_codegen.nr`.
- [x] Verify that `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder/pkg/cmd/test/chan_alloc_codegen` fails with the old compiler behavior or demonstrates the exact fix after application.
- [x] Verify that existing channel tests (`channel_demo`, `actor_scheduler_test`, `concurrency_bench`, etc.) pass without regressions.

## Test Plan
1. **New Integration Test:** `pkg/cmd/test/chan_alloc_codegen/chan_alloc_codegen.nr` will define a struct with multiple fields (`struct { handle: ptr, id: i32 }`), allocate it on the heap (`alloc`), and send it via `@` move into a `chan[CommandBuffer]` channel. It will verify that the received struct matches the exact field values sent and not a pointer address.
2. **Regression Suite:** Run `TestCompilerWithTestFolder` across all existing concurrency and channel test cases (`channel_demo`, `actor_scheduler_test`, `parallel_test`, etc.).

## Risks
- Misidentifying when an element is a pointer in C vs value in C could cause double dereferencing or wrong copying for other channel element types (`chan[@T]` or `chan[ptr]`). We mitigate this by specifically checking that `ElemType` is NOT a C pointer while `ValType` IS a C pointer before altering the passed argument to `channel_send`.

## Completion Criteria
- Implementation plan created and reviewed.
- Reproduction test added to `pkg/cmd/test/chan_alloc_codegen/chan_alloc_codegen.nr`.
- Codegen fixes implemented cleanly in `hir_codegen.go` and `expressions.go`.
- All integration tests pass.
