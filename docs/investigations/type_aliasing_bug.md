# Investigation Report: Package Type Aliasing Bug

**Status**: Open
**Date**: 2026-07-03
**Component**: `pkg/semantic` (Type Resolution)

## Problem

When the Nora compiler processes any `TypeAliasStatement` (e.g., `pub type MyDummy = my_pkg.Dummy`, `pub type MyInt = i32`, or `pub type MyHandle = ptr`), it fails to properly register or resolve the alias. Any subsequent usage of the alias—whether as a parameter in a local function, or as a parameter/return type in an `extern fn`—results in an `undefined type` error.

This unified issue covers multiple previous symptoms that were initially misdiagnosed:
- **Duplicate Import Bug**: We originally thought importing the same package in multiple files caused it, because `wgpu_native` tried to alias `sys.Adapter` to `Adapter`.
- **Extern Fn Type Bug**: We originally thought `extern fn` signatures strictly rejected type aliases, but it turns out the aliases themselves were fundamentally broken in the compiler.

## Reproduction

A unified minimal reproduction test has been added to the test suite at:
`pkg/cmd/test/violations/fail_aliasing_bug/`

The test consists of:
1. `my_pkg/types.nr` (exports a type `Dummy`)
2. `main.nr` (attempts to alias external types, primitives, and pointers)

```nora
package main
import "pkg/cmd/test/violations/fail_aliasing_bug/my_pkg"

pub type MyDummy = my_pkg.Dummy
pub type MyInt = i32
pub type MyHandle = ptr

// EXPECTED SEMANTIC ERROR: undefined type 'MyDummy' and 'MyInt'
pub fn foo(d: MyDummy, i: MyInt) { 
}

// EXPECTED SEMANTIC ERROR: undefined type 'MyHandle'
pub extern fn do_something(handle: MyHandle) MyHandle

fn main() i32 { return 0 }
```

When building this directory, the compiler output includes:
```text
Error: undefined type: 'MyDummy'
Error: undefined type: 'MyInt'
Error: undefined type: 'MyHandle'
```

## Root Cause

The Nora compiler's semantic analyzer (`pkg/semantic/analyzer.go`) fails to properly evaluate the right-hand side of a type alias. Whether the RHS is an external package selector expression (like `my_pkg.Dummy`), a primitive type (like `i32`), or a pointer (`ptr`), the compiler never successfully registers the new alias in the local symbol table. This leads to cascading `undefined type` errors whenever the alias is referenced.

## Workaround

Until the compiler frontend is patched, developers must avoid using `TypeAliasStatement` entirely. 

- Instead of aliasing external types, use the fully qualified package type directly in function signatures (e.g., `sys.Device`), or use a struct wrapper (`pub type Device = struct { handle: sys.Device }`).
- Instead of aliasing pointers for FFI (e.g., `pub type Instance = ptr`), `extern fn` signatures must use the primitive `ptr` directly.

## Validation / Required Fix

The `pkg/semantic` type resolver must be updated to correctly resolve all nodes on the right side of a `TypeAliasStatement`, including `SelectorExpression`, `Ident` (for primitives/pointers), and other type nodes.

Once patched, running `go run pkg/cmd/nora/main.go build pkg/cmd/test/violations/fail_aliasing_bug/` must compile successfully without emitting any `undefined type` errors.
