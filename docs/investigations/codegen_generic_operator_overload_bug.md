# Investigation: Generic Operator Overloading Codegen Bug

**Status:** Resolved
**Component:** `pkg/codegen` (HIR to C Lowering)

## Context
While implementing the `nora_math` generic math library (`Vector2[T]`, `Vector3[T]`, etc.), an attempt was made to overload standard mathematical operators (`+`, `-`, `*`, `/`) for generic structs using Nora's standard operator overload naming convention (`add`, `sub`, `mul`, `div`).

The code successfully passed the frontend parsing, semantic analysis, and topology resolution phases. However, the compilation failed during the C generation step via GCC/Clang.

## Problem
When the AST lowers a generic operator overload (e.g., `v2_a + v2_b`) down to HIR and subsequently to C, the C code generator drops the type-erasure hash suffix (e.g., `_aa15546d`) for the method call. 

This results in the generated C code attempting to call an undeclared, un-hashed base function name, which the C compiler immediately rejects.

## Reproduction
1. Define a generic struct and overload an operator:
```nora
pub type Vector2[T] = struct { x: T, y: T }

pub fn (self: #Vector2[T]) add[T](other: #Vector2[T]) Vector2[T] {
    return Vector2[T] { x: self.x + other.x, y: self.y + other.y }
}
```
2. Call the overloaded operator in code:
```nora
var a = vector.NewVector2[f32](1.0, 1.0)
var b = vector.NewVector2[f32](2.0, 2.0)
var c = a + b // Fails here during C generation
```

**Compiler Output / Error:**
```text
E:\Project\Project Nora\nora\build\debug\out_pkg_main.c:104:14: error: call to undeclared function 'vector_Vector2_f1109c62_add'; ISO C99 and later do not support implicit function declarations [-Wimplicit-function-declaration]
  104 |     v2_add = vector_Vector2_f1109c62_add(NULL, &v2_a, &v2_b);
      |              ^
E:\Project\Project Nora\nora\build\debug\out_pkg_main.c:104:14: note: did you mean 'vector_Vector2_f1109c62_add_aa15546d'?
E:\Project\Project Nora\nora\build\debug\out.h:107:25: note: 'vector_Vector2_f1109c62_add_aa15546d' declared here
  107 | vector_Vector2_f1109c62 vector_Vector2_f1109c62_add_aa15546d(void* _env_ptr, vector_Vector2_f1109c62* self, vector_Vector2_f1109c62* other);
```

## Root Cause
Nora's Type-Erased Shared Monomorphization (`pkg/codegen`) handles appending unique identifier hashes for generic type combinations. However, the specific HIR lowering logic for BinOps mapped to operator overloads seems to bypass the method name resolution step that attaches this hash. 

The C generator correctly hashes the definition in `out.h` (`vector_Vector2_f1109c62_add_aa15546d`), but the caller in `out_pkg_main.c` is hardcoded to emit the base symbol (`vector_Vector2_f1109c62_add`) without evaluating the generic type arguments of the operands.

## Workaround / Fix
**Immediate Workaround:**
Do not use the lowercase operator overload syntax (`add`, `sub`, `mul`, `div`) for generic structs. Instead, rename the methods to explicit capitalized function names (`Add`, `Sub`, `Mul`, `Div`) and call them directly using method syntax:

```nora
// Definition
pub fn (self: #Vector2[T]) Add[T](other: #Vector2[T]) Vector2[T] { ... }

// Usage
var c = a.Add[f32](#b) 
```
*(Note: Be careful with move semantics when passing structs to explicitly generic methods. It is recommended to pass `other` as a read-only borrow `#other` to avoid consuming the right-hand operand).*

**Compiler Fix Required:**
Modify `pkg/codegen` (specifically `hir_codegen.go` or the BinOp generation node) to query the Semantic Analyzer's resolved method identifier, ensuring that the monomorphization hash is appended to operator overloads exactly as it is for standard method calls.

## Validation
Validation will require implementing the compiler fix and ensuring that generic math vectors can be seamlessly added using standard mathematical notation (`+`, `-`) in `nora_math` without triggering implicit C function declarations.

**Result:** Fix implemented and tested. Added an integration test `test_generic_operator_overload_test/main.nr` inside the test suite `pkg/cmd/test`. The C codegen phase successfully resolves the monomorphized hashed name `vector_Vector2_hash_add_hash` by dynamically querying `SemanticInfo.MethodSymbols` and appending `types.GetHashSuffix` directly in `expressions.go` during AST `InfixExpression` code generation.
