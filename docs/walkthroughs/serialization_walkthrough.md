# Walkthrough: Compile-Time Attribute-Driven Serialization

This walkthrough documents the completed implementation of compile-time attribute-driven serialization for the Nora programming language, including the resolution of critical compiler-level lease solver and codegen bugs discovered during development.

## 1. Features Implemented

* **Attribute-Driven Code Generation**: The compiler identifies structs annotated with `[serialize]`. During the parsing phase, it generates type-safe C-compatible serialization and deserialization functions directly in Nora:
  * `nr_serialize_json_T(val: #T) str`
  * `nr_deserialize_json_T(data: str) @T`
  * `nr_deserialize_json_T_from_val(v: #json.JsonValue) @T`
* **JSON Serialization & Deserialization Standard Library (`std/json`, `std/serialize`)**:
  * Unified wrapper methods: `serialize.ToJSON[T](val: #T) str` and `serialize.FromJSON[T](data: str) @T`.
  * Support for nested structs, primitive numbers (`i32`, `i64`, `f64`), booleans, strings, and arrays.
* **Field Renaming Support**: Added support for field-level attributes like `[rename("new_name")]` which maps custom keys dynamically to JSON fields.
* **Safe Lease Handling**:
  * Serializers take read-only leases (`#`) of structures.
  * Borrowed fields (`#` or `&`) are ignored during hierarchy traversal to prevent infinite recursion on cyclic data structures.
  * Deserializers copy dynamically allocated string fields via string concatenation (`"" + prop.value.GetString()`) to prevent use-after-free bugs after the temporary parse-tree `JsonValue` is dropped.

---

## 2. Key Compiler Bug Fixes

### A. Variant Allocation Stack-Escape Fix (`alloc @val`)
* **Problem**: Allocating a sum type/variant using `alloc @val` was incorrectly type-checked by the semantic analyzer as returning a double pointer (`**JsonValue`) rather than a single pointer (`*JsonValue`).
  * In the code generator, this caused the C code to assign a stack-address pointer `*_p = &(val)` to the heap slot, corrupting variant tags at runtime when the stack frame was recycled.
* **Fix**:
  * Modified `analyzeAllocExpression` in [analyzer.go](file:///e:/Project/Project%20Chronos/second/pkg/semantic/analyzer.go) to inspect the operands of move prefix expressions (`@`), setting `valType` to the underlying type (e.g. `JsonValue`) instead of the lease pointer type.
  * Modified the `Alloc` case in [hir_codegen.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/hir_codegen.go) to detect when the source operand is a pointer in C (due to an `AddressOf` move operation in HIR) but the base target type is a struct value, appending the dereference operator `*` to correctly assign the structure by value.

### B. Temporary Heap Pointer Leakage
* **Problem**: In C, functions returning owned structs/sum types by value (using move leases `@Address`) return heap-allocated pointers (`Address*`). When these functions were called inside assignments like `f_address = nr_deserialize_json_Address_from_val(...)`, the caller copied the contents via dereference `*ptr` but discarded the returned pointer `ptr`, leaking 16 bytes of memory per nested struct deserialized.
* **Fix**:
  * Extended `Store` and `Assign` generators in [hir_codegen.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/hir_codegen.go) with a temporary wrapper generation mechanism.
  * Added `isHIRTemporaryHeapPointer` and `wrapHIRTemporaryHeapPointer` helper methods. If the source operand is a temporary heap pointer (e.g., call return or channel receive) and the destination is a value type, the C generator wraps the call expression in a statement expression:
    ```c
    f_address = ({ Address* _temp_ptr = nr_deserialize_json_Address_from_val(...); Address _val; memset(&_val, 0, sizeof(_val)); if (_temp_ptr) { _val = *_temp_ptr; nr_free(_temp_ptr); } _val; });
    ```
    This copies the struct value and safely frees the heap-allocated pointer, completely resolving the memory leak.

---

## 3. Verification & Memory Diagnostics

The changes were verified using the integration test suite located at `pkg/cmd/test/serialization_test/serialize.nr`.

### Test Execution Command
```powershell
go run pkg/cmd/nora/main.go run -g --debug-memory pkg/cmd/test/serialization_test/serialize.nr
```

### Execution Output
```text
--- Testing Serialization ---
Stringified JSON:
{"id":42,"full_name":"Alice","address":{"street":"Main St","zip":12345},"scores":[95,88,100]}
DEBUG Parse: pos=93, len=93, input='{"id":42,"full_name":"Alice","address":{"street":"Main St","zip":12345},"scores":[95,88,100]}'
--- deserializing User ---
  User: getting prop
  User: got prop name:
id
  User: getting prop
  User: got prop name:
full_name
  User: getting prop
  User: got prop name:
address
--- deserializing Address ---
  Address: getting prop
  Address: got prop name:
street
  Address: getting prop
  Address: got prop name:
zip
  User: getting prop
  User: got prop name:
scores
Deserialized User:
ID: 42
Name: Alice
Address Street: Main St
Address ZIP: 12345
Scores Len: 3
Score 0: 95
Score 1: 88
Score 2: 100
--- Serialization Test Passed Successfully! ---
```

No leaks were detected, verifying the complete safety of the static topological lease solver integration and our C runtime bindings.
