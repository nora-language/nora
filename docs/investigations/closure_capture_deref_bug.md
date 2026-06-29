# Compiler Investigation: Incompatible Type Assignment in Closure Environment

## Problem

When compiling Nora programs that generate closures capturing multiple parameters—some passed as pointers by optimization (structs) and some passed as explicit leases (`&T`)—the C-compiler failed with incompatible type assignment errors. 

For example, when `phase8_determinism` used `ps.islands.ParMap` with a closure that captured `dt` (a `Fixed64` struct) and `ps` (a `&PhysicsSystem` lease), the compiler emitted errors like:

```text
error: assigning to 'fixed64_Fixed64' (aka 'struct fixed64_Fixed64') from incompatible type 'fixed64_Fixed64 *' (aka 'struct fixed64_Fixed64 *'); dereference with *
_env_local->dt = dt;
```

And upon a naive fix:
```text
error: assigning to 'system_PhysicsSystem_e431148f *' (aka 'struct system_PhysicsSystem_e431148f *') from incompatible type 'system_PhysicsSystem_e431148f' (aka 'struct system_PhysicsSystem_e431148f'); remove *
_env_local->ps = *ps;
```

## Reproduction

The issue is fully reproducible by creating a lambda function that captures an outer scope parameter passed as a struct, as well as a pointer lease. A minimal reproduction was added to the test suite at `pkg/cmd/test/repro_rvalue_address/repro_closure.nr`.

## Root Cause

In Nora, all closures capture their outer scope environment **by value**. However, when generating the C representation of the closure (`_env_local_t`), the C code generator failed to account for Nora's implicit C optimization layer:

1. **The Environment Struct (`_env_t`)**: This C struct defines fields according to the exact Nora AST type (`cap.Type`). For `dt`, the field is `fixed64_Fixed64`. For `ps`, the field is `system_PhysicsSystem_e431148f*`.
2. **The Assigment (`_env_local->cap = cap;`)**: In `hir_codegen.go` and `expressions.go`, the C code string generated the assignment directly (`rhs = name`). 
3. **The Mismatch**: If `dt` was a `SymParam` belonging to the outer function, Nora's C lowering `g.shouldPassByPointer` automatically turns struct arguments into pointers (`fixed64_Fixed64* dt`) to save stack copying overhead. Thus, copying `dt` directly into `_env_local->dt` resulted in a pointer being assigned to a value field.

When we introduced the dereference (`rhs = "*" + rhs`), the previous attempt used `types.UnwrapLease(cap.Type)` and assumed any `StructType` needed dereferencing. This was overly broad because explicit leases (e.g. `&PhysicsSystem`) were *also* unwrapped, identified as structs, and dereferenced, stripping them of their pointer semantics during the environment copy.

## Fix

The fix modifies `hir_codegen.go` and `expressions.go` to inject pointer dereferencing (`*`) dynamically only when the captured parameter `cap` satisfies two constraints:

1. `g.shouldPassByPointer(cap.Type, cap.LeaseKind) == true`: Meaning the compiler lowered the value into a C pointer for the function parameter stack.
2. `!g.isPointerTypeInC(cap.Type)`: Meaning the actual native Nora type (`cap.Type`) is *not* meant to be a pointer type (i.e. it isn't an explicit lease like `&T`, nor a native pointer type like `str` or `ptr`).

**Code Added**:
```go
if rhs == name {
    if cap.Kind == semantic.SymParam {
        if g.shouldPassByPointer(cap.Type, cap.LeaseKind) && !g.isPointerTypeInC(cap.Type) {
            rhs = "*" + rhs
        }
    }
}
```

## Validation

The compiler was rebuilt (`go build -o nora.exe ./pkg/cmd/nora`) and the `phase8_determinism` execution was tested using `..\nora.exe run --example phase8_determinism`. The C compilation warnings/errors disappeared completely, and the deterministic physics lock-step test output `--- Phase 8: Deterministic Lock-Step Test ---` successfully. All test cases in `pkg/cmd/test/` continue to pass.
