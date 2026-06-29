# Investigation: C Compiler RValue Reference Error

## Problem
When compiling physics engine code containing mathematical expressions within comparisons (e.g., `if (bA.mass == (dt - dt))`), the Nora compiler generates C code that attempts to take the address of an rvalue struct returned by a function call:
`&fixed64_Fixed64_sub(NULL, dt, dt)`
This causes the backend C compiler (GCC/Clang) to fail with the error:
`error: cannot take the address of an rvalue of type 'fixed64_Fixed64' (aka 'struct fixed64_Fixed64')`

## Reproduction
A minimal reproduction case has been added at `pkg/cmd/test/repro_rvalue_address/repro_rvalue_address.nr`:
```nora
pub type Dummy = struct { val: i64 }
pub fn (s: Dummy) eq(other: Dummy) bool { return s.val == other.val }
pub fn (s: Dummy) sub(other: Dummy) Dummy { return Dummy { val: s.val - other.val } }

pub fn Eq[T](a: #T, b: #T) bool { return a.eq(b) }
pub fn Sub[T](a: #T, b: #T) T { return a.sub(b) }

pub fn main() i32 {
    var a = Dummy { val: 10 }
    var dt = Dummy { val: 5 }
    if a == (dt - dt) { return 1 } // Triggers the bug
    return 0
}
```

## Root Cause
The root cause lies in how generic operator overloading is translated. 
1. `a == (dt - dt)` is lowered by the semantic analyzer to `Eq[T](#a, #Sub[T](#dt, #dt))`.
2. The `Sub[T]` operation returns a struct by value (in C, `struct fixed64_Fixed64`).
3. The `#` operator applied to the result of `Sub[T]` is translated into a `hir.AddressOf` operation in the HIR.
4. When `hir_codegen.go` handles this expression inside `alignCallArgument` or `expressions.go`, it identifies that it needs a pointer to the value. 
5. However, the codegen failed to wrap the function call in a C99 compound literal `((struct ...[]){ ... })` before prepending the `&` operator (or in some paths directly outputs `&fn()`), leading to invalid C.

## Fix
Currently investigating the exact line in `hir_codegen.go` or `lower.go` responsible for the unhandled rvalue reference. The fix involves ensuring that whenever a pointer to an rvalue expression is required (e.g., `#expr`), it is wrapped in a block-scoped C99 compound literal before its address is taken.

## Validation
1. Verify that `nora run` on `repro_rvalue_address.nr` fails on the current compiler.
2. Apply fix.
3. Verify `repro_rvalue_address.nr` compiles and runs successfully.
4. Verify `nora run --example phase8_determinism` successfully compiles the physics engine.
