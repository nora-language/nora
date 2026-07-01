# Compiler Investigation: Multi-Parameter Generic Return Types

Status: Completed

## Problem & Overview
In the Nora Programming Language compiler, writing a generic function that returns a multi-parameter generic struct (such as `NewQuery2[A, B]` returning `@Query2[A, B]`) was believed to fail or was avoided in the codebase (returning a raw `ptr` and casting instead). We investigated the compiler's frontend (parser, semantic analyzer) and code generator to identify where the production gap lies.

## Root Cause & Technical Details

1. **Parser Capability**:
   The Recursive Descent Parser (`pkg/parser/parser.go`) fully supports parsing multi-parameter generic return types. The syntax `@Query2[A, B]` is parsed as a PrefixExpression (lease operator `@`) containing an `IndexExpression` with `Indices: [A, B]`. It successfully handles comma separators inside `parseTypeSuffix`.

2. **Semantic Type Resolution**:
   The Semantic Analyzer (`pkg/semantic/analyzer.go`) is fully capable of resolving these type nodes. When it encounters `Query2[A, B]`, it successfully specialize it via `specializeStructType` as a nominal `types.StructType` with nominal type arguments.

3. **The Real Production Gap: C Codegen Declaration Order & Monomorphization**:
   The root cause is not in the parser or type solver, but in **C Codegen's Forward Declaration Order** for monomorphized methods:
   - When generic methods (like `InitColumn[T]`) are instantiated inside other generic scopes or lambda wrappers (like `NewComponentMeta[T]`), the Nora compiler generates monomorphized symbols (e.g. `gecs_InitColumn_127f6aa5`).
   - Because these specialized signatures are created dynamically during the code generation phase, their prototypes are not collected during the early `emitPrototypes()` pass.
   - This results in Clang compilation failures because the generated C function body is emitted or called before its forward declaration prototype exists in the output `.c` file.

To bypass this monomorphization prototype order bug, GECS had to return a raw `ptr` and cast manually using `IntoRaw[Query2[A, B]]`, hiding the true type structure from early signature tracking.

## Solution Alternatives

### Alternative A: Pre-pass Prototype Harvesting (Recommended)
Add a pre-generation pass in `Generator.collectDefinitions()` that recursively traverses all expression nodes and lambdas to collect and monomorphize every referenced symbol before generating any C structure or prototype. This guarantees that all required generic signatures (including multi-parameter generic functions) are registered and written to the C header prototypes cleanly.

### Alternative B: Desugaring to Raw Pointers during Codegen
desugar generic types inside function return positions to raw `ptr` at the C codegen level while keeping them fully type-safe in the Nora frontend.

## Chosen Solution

We choose **Alternative A (Pre-pass Prototype Harvesting)** because:
1. It maintains the premium zero-cost, type-safe abstraction philosophy of Nora.
2. It fully resolves prototype declaration ordering issues for all generic functions and methods, not just return types.
3. It keeps the generated C11 code readable and clean.

We have fully implemented the Top-down Prototype Generation pass.
