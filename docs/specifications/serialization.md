# Attribute-Driven Serialization

## Title & Overview
This specification details the introduction of **compile-time attribute-driven serialization** to the Nora Programming Language. By leveraging the `[serialize]` compiler attribute, Nora automatically generates serialization and deserialization routines for structs, exposing a clean, high-performance API under `std/serialize`.

## Motivation
Nora does not support runtime reflection due to its emphasis on zero runtime overhead and predictable memory layouts. However, applications like ECS frameworks (e.g. GECS), web services, and configuration parsers require the ability to save/load state. 

Instead of writing verbose manual boilerplate or relying on slow reflection, this proposal introduces compile-time code generation driven by attributes, producing type-safe serialization methods at zero runtime cost.

## Syntax
Structs targeting serialization are marked with the `[serialize]` attribute:

```nora
[serialize]
pub type Position = struct {
    x: i32,
    y: i32
}
```

Fields can be renamed in the serialization output (e.g. JSON keys) using the `[rename]` attribute:

```nora
[serialize]
pub type User = struct {
    id: i32,
    [rename("full_name")]
    name: str,
    email: str
}
```

## Semantics

### Compile-Time Code Generation
When the parser encounters a `TypeStatement` containing the `[serialize]` attribute, the compiler generates companion functions for the struct:
1. `nr_serialize_json_T(val: #T) str` (encodes the struct to a JSON string)
2. `nr_deserialize_json_T(data: str) @T` (allocates and reconstructs the struct)
3. `nr_deserialize_json_T_from_val(v: #json.JsonValue) @T` (reconstructs the struct from a parsed json value tree)

These functions are parsed and appended to the package-scoped split backend output during transpilation.

### The `std/serialize` Library
The standard library exposes a uniform API wrapping these compiler-generated hooks:

```nora
package serialize

// Compiler-implemented intrinsic hooks
extern fn nr_serialize_json[T](val: #T) str
extern fn nr_deserialize_json[T](data: str) @T

pub fn ToJSON[T](val: #T) str {
    return nr_serialize_json[T](val)
}

pub fn FromJSON[T](data: str) @T {
    return nr_deserialize_json[T](data)
}
```

## Type Rules
1. Only structs containing fields whose types also implement serialization (primitives, strings, arrays, or other `[serialize]` structs) can be decorated with `[serialize]`.
2. Sum types and Protocols are not currently supported by the compiler-generated `[serialize]` attribute.

## Lease Rules
* `ToJSON[T](val: #T)` takes a read-only lease (`#`) to prevent copies and ensure thread-safety during serialization.
* `FromJSON[T](data: str)` returns an owned value (`@T`), transferring ownership of the newly allocated memory to the caller.

## Examples

### ECS Snapshot Integration
Integrating compile-time serialization into the GECS engine:

```nora
package gecs

import "serialize"

pub fn (w: &World) SerializeComponent[T](col_ptr: ptr, row: i32) str {
    var vec: &collections.Vector[T] = col_ptr
    return serialize.ToJSON[T](#vec.data[row])
}
```

## Edge Cases
* **Cyclic Data Structures**: Nora fully supports cyclic references in memory (e.g., `ListNode` with `@next` and `#prev`). However, serializing them to hierarchical formats like JSON leads to infinite recursion. To prevent this, the serializer by default ignores read-only/mutable borrowed fields (`#` and `&`) during serialization, only traversing owned paths (`@`).
* **Format Errors**: `FromJSON` will return a default value of the struct if the input is corrupted or invalid.
