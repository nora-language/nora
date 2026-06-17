# Implementation Plan: Compile-Time Attribute-Driven Serialization

* **Status**: Completed
* **Date**: 2026-06-03

## Goal
Implement compile-time attribute-driven serialization support for the Nora Programming Language without reflection.

## Affected Components
* **Parser (`pkg/parser/`)**: Support `[serialize]` type annotations and generate AST method statements.
* **Semantic Analyzer (`pkg/semantic/`)**: Type-check lease-move allocations and resolve generic serialization calls.
* **Codegen (`pkg/codegen/`)**: Transpile assignments of temporary heap-allocated structures without memory leaks.
* **Standard Library (`std/json/`, `std/serialize/`)**: Format-specific helper wrappers.

## Implementation Checklist
- [x] Parse `[serialize]` attributes on structs.
- [x] Implement field renaming mapping `[rename("json_key")]`.
- [x] Write code generation function mapping to AST representation.
- [x] Fix variant type allocation double-pointer type checks (`alloc @val`).
- [x] Automatically clean up temporary heap struct pointers returned by value in C codegen.
- [x] Avoid cyclic borrow recursion by excluding `#` and `&` fields.
- [x] Duplicate deserialized strings via `"" + ` to avoid use-after-free corruption.

## Test Plan
- Run serialization test case `pkg/cmd/test/serialization_test/serialize.nr`.
- Ensure structural deserialization is correct and verify zero memory leaks in the C allocator via `--debug-memory`.
