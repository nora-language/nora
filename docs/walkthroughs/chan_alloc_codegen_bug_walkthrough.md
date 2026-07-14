# Walkthrough: Channel Allocation Codegen Bug (`@T` sent to `chan[T]`)

## Overview
We resolved a critical code-generation bug where sending an `alloc`'d owned struct (`@T`) into a value struct channel (`chan[T]`) passed the address of the pointer variable (`&_send_val`) to `channel_send` instead of passing the pointer value (`_send_val`). Because `channel_send` copies `sizeof(T)` bytes from the supplied address, passing `&_send_val` copied the stack pointer address (plus adjacent stack memory) into the channel buffer instead of copying the struct fields. Furthermore, after copying the struct fields out of the heap pointer into the value channel buffer, the outer container allocation (`_send_val`) was previously leaked. Both issues are now resolved.

## Changes Made

### 1. Type Helper (`pkg/codegen/types.go`)
* Added `getChanElemType(t types.NRType) types.NRType` to recursively unwrap leases (`@`, `#`, `&`) and pointers to retrieve the inner `types.ChanType.Elem` type of any channel expression.

### 2. HIR Channel Send Lowering (`pkg/codegen/hir_codegen.go`)
* In `hirInstStr()` (`case *hir.ChanSend:`), inspected `ElemType` using `getChanElemType()`.
* When `ElemType` is a value struct in C (`!g.isPointerTypeInC(ElemType) && !strings.HasSuffix(g.cType(ElemType), "*")`) and the sent operand `i.Val` is a C pointer (`g.isPointerTypeInC(t) || strings.HasSuffix(g.cType(t), "*")`):
  * Passed `(void*)_send_val` to `channel_send` rather than `&_send_val`.
  * Appended `nr_free(_send_val)` right after `channel_send` to free the outer container heap allocation once its contents have been copied by value into the channel buffer.

### 3. AST Channel Send Lowering (`pkg/codegen/expressions.go`)
* Updated `genSendExpression()` (`case *ast.SendExpression`) with identical logic, verifying `ElemType` against `e.Right`'s type (`t`) to pass `(void*)_send_val` and emit `nr_free(_send_val)` when transferring a C pointer into a C struct value channel.

## Verification Results

### Test Case
Created `pkg/cmd/test/chan_alloc_codegen/chan_alloc_codegen.nr`:
```nora
package main

import "io"

type CommandBuffer = struct {
    handle: u64,
    id: i32
}

fn Finish(id: i32) @CommandBuffer {
    return alloc CommandBuffer {
        handle: 0x12345678,
        id: id
    }
}

fn worker(ch: #chan[CommandBuffer]) {
    var cmd = Finish(42)
    ch <- @cmd
}

fn main() i32 {
    var ch = make(chan[CommandBuffer], 1)

    scope {
        spawn worker(#ch)

        var received = <-ch
        io.PrintLn("Received id: ${received.id}")
        if (received.id != 42) {
            panic("chan_alloc_codegen_bug reproduction: received struct field 'id' corrupted")
        }
        if (received.handle != 0x12345678) {
            panic("chan_alloc_codegen_bug reproduction: received struct field 'handle' corrupted")
        }
    }

    io.PrintLn("PASS: chan_alloc_codegen test passed")
    return 0
}
```

### Before vs After
* **Before Fix**: Executing `go test` reported a corrupted received struct (`Received id: 316278320`) and panicked at runtime.
* **After Fix**: Executing `go test` completes cleanly (`Received id: 42`), with memory leak reports (`nr_mem_report()`) showing 0 active allocations and 0 leaked bytes (`PASS: TestCompilerWithTestFolder/pkg/cmd/test/chan_alloc_codegen/chan_alloc_codegen.nr (1.25s)`).
* **Regression Testing**: Verified that all existing channel and concurrency integration suites (`channel_demo`, `actor_scheduler_test`, `concurrency_bench`, `parallel_demo`) pass without regressions.
