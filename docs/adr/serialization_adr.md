# ADR: Compile-Time Attribute-Driven Serialization

## Status
Proposed

## Context
Nora lacks runtime reflection and dynamic type introspection by design, maximizing performance and safety. However, this poses a challenge for features like game state serialization (e.g. in GECS snapshots) or web API parsing, which dynamically serialize and deserialize data. 

We need a system to translate Nora structs to/from structured text format (JSON) or binary formats without runtime cost or type safety violations.

## Decision
We choose to implement **compile-time attribute-driven serialization** using:
1. The compiler's attribute parsing pipeline to identify target structs decorated with `[serialize]`.
2. The WASM-based compiler plugin system to intercept type definitions during parser/codegen phase and output serialization/deserialization code.
3. A native standard library module `std/serialize` mapping format-specific conversions.

## Alternatives Considered

### 1. Manual Boilerplate
* **Pros**: Simple, requires no compiler changes.
* **Cons**: Prone to developer error, tedious to maintain as schemas evolve, and reduces the premium developer experience (DX) Nora aims to provide.

### 2. Runtime Reflection Library
* **Pros**: Identical to Go's dynamic serialization.
* **Cons**: Massive binary bloat, requires runtime metadata tables, degrades execution speed, and conflicts with Nora's system language design principles.

## Consequences
* **Pros**:
  * Zero runtime performance overhead (direct, static type conversions).
  * 100% type-safe compilation (layout mismatched JSON structures fail to build).
  * Highly maintainable (adding `[serialize]` is all that is needed).
* **Cons**:
  * Adds compilation phase overhead when compiling code with many serialized structs.
  * Increases generated C output file size due to statically generated serialization functions (monomorphization).
