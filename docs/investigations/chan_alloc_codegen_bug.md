# Investigation: chan[@T] Codegen Bug — Sends Pointer-to-Pointer Instead of Struct Value

## Status
**OPEN** — Workaround applied in `examples/multithreaded/main.nr`

## Problem

When sending an `@T` (alloc'd, owned struct) value through a `chan[T]` channel, the compiler generates incorrect C code. The channel receive produces a struct whose fields contain the **heap address** of the original struct, rather than the **contents** of that struct.

## Reproduction

```nora
pub type CommandBuffer = struct { handle: ptr }
pub fn Finish() @CommandBuffer {
    return alloc CommandBuffer { handle: raw_wgpu_handle }
}

var ch = make(chan[CommandBuffer], 1)
var cmd = Finish()
ch <- @cmd  // BUG: sends ptr-to-ptr

var received = <-ch
// received.handle == address of CommandBuffer on heap
// expected: received.handle == raw_wgpu_handle
```

## Root Cause

In the generated C, when an `@T` value is sent via a channel:

```c
// Broken generated code:
core_CommandBuffer* _send_val = cmd_buf;    // heap pointer to the struct
channel_send(_c, &_send_val);               // sends ptr** (8 bytes = heap address)

// Receive side:
core_CommandBuffer _res;
channel_recv(ch, &_res);                    // reads 8 bytes into _res
// _res.handle = heap address, NOT wgpu handle!
```

The channel is created with `channel_make(1, sizeof(core_CommandBuffer))` — element size = 8 bytes (correct). But `channel_send` is given `&_send_val` (a `core_CommandBuffer**`), so it copies the **pointer value** (the heap address) rather than the **struct contents**.

The correct generated code should dereference the pointer before sending:
```c
channel_send(_c, cmd_buf);   // cmd_buf IS the pointer to struct, so this sends the struct contents
```

## Fix (Required in Compiler)

In `pkg/codegen/hir_codegen.go` (or wherever channel send is lowered), when the sent value is of an `alloc`'d (heap-pointer) type `@T`, the codegen must dereference the pointer before passing it to `channel_send`:
- **Wrong**: `channel_send(ch, &heap_ptr)`
- **Correct**: `channel_send(ch, heap_ptr)` — since `heap_ptr` already points to the struct

## Workaround

Use `chan[ptr]` instead of `chan[T]` and send/receive the raw field values manually:

```nora
// Worker:
var results: chan[ptr]
...
results <- cmd_buf.handle   // send the raw ptr field
cmd_buf.handle = none       // prevent drop() double-release

// Main:
var h = <-results
var buf = core.CommandBuffer { handle: h }
queue.Submit(&buf)
```

## Validation

The workaround is applied in `nora_wgpu/examples/multithreaded/main.nr` and the example runs stably with no wgpu validation errors.

## Affected Files
- `pkg/codegen/hir_codegen.go` — channel send lowering for `@T` values
- `pkg/hir/` — possibly the `Send` HIR instruction

## Future Considerations
- Once fixed, `chan[core.CommandBuffer]` can be restored in the multithreaded example
- The fix must ensure `drop()` is NOT called on the sender's variable after the `@` move send (this part is already correct — only the `channel_send` argument is wrong)
