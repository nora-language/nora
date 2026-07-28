# Investigation: Generic Struct Operator Overloading Codegen Failure

**Status**: Open
**Date**: 2026-07-28
**Investigator**: Antigravity

## Problem
During the implementation of Phase 38 (Fixed-Point Deterministic Math), we instantiated the generic Vector3[T: Copy] struct with the Fixed64 struct type. The compilation failed during the C11 codegen phase with the following error:
`error: invalid operands to binary expression ('fixed_Fixed64' (aka 'struct fixed_Fixed64') and 'fixed_Fixed64')`
This error occurred anywhere Vector3 methods used comparison operators (<, <=, >, >=) on the generic type T. Because Fixed64 is a struct, the C11 compiler rejected standard binary operators between them, expecting a method call instead.

## Reproduction
1. Define a struct MyStruct with a cmp(other: MyStruct) i32 method.
2. Instantiate a generic struct that performs comparisons on its generic type: var v = Vector3[MyStruct]{...}.
3. Call a method on that generic struct that internally uses < or <= on the type T, such as v.Equals(v, epsilon).
4. Run nora build or nora run. 
5. The semantic analyzer succeeds, but C compilation fails in out_globals.c complaining about invalid operands for struct types.

## Root Cause
The root cause involves the interaction between the Nora compiler's generic AST lowering and the C11 Code Generator:

1. **AST Lowering (pkg/hir/lower.go)**: 
   When the AST for Vector3[T] is lowered into the High-Level Intermediate Representation (HIR), the type of T is still generic (*types.GenericType). The isStructEq and isStructOp flags in lower.go depend on whether the operand's type is known to be a struct (*types.StructType). Since T is generic, these checks fail, and the operators are lowered as basic *hir.BinOp instructions instead of *hir.ASTExpr (which would correctly trigger overloaded method generation).
2. **Type-Erased Monomorphization**:
   The compiler monomorphizes the generic HIR by replacing the generic types with their concrete types. However, the instruction itself remains a *hir.BinOp.
3. **C11 Codegen (pkg/codegen/hir_codegen.go)**:
   The *hir.BinOp branch in the code generator blindly translates all binary operators directly into their C equivalents (e.g. (left < right)). Because the monomorphized type is a C struct, emitting standard operators produces invalid C11 code.

An earlier attempt to patch hir_codegen.go by intercepting struct types within *hir.BinOp and generating calls to _cmp or _operator_eq failed because:
- The mangled name of the method must include the generic instance hash (e.g., _f2c063a2), which isn't easily accessible at the BinOp emission stage without evaluating SemanticInfo.MonomorphizedNames.
- It accidentally intercepted pointer equality checks (such as those generated within auto-dereferenced operator!=), causing undeclared function errors for operator_eq on types that don't define one.

## Fix
A proper fix requires modifying the AST lowering phase or the monomorphization pass to correctly handle overloaded operators for generic types that resolve to structs.
- **Option A (Relowering)**: Ensure that when generic methods are monomorphized, any BinOp instructions involving structs are re-lowered into method calls (ASTExpr or Call instructions).
- **Option B (Codegen Interface)**: Update hir_codegen.go to look up the correct overloaded method symbol and emit a proper function call, correctly handling pointer semantics and generic mangling hashes. 

This fix is deferred to a separate scope to maintain the stability of the compiler while Phase 38 development completes using an alternative workaround or in a subsequent patch.

## Validation
A successful fix will allow generic_math_demo.nr to instantiate and execute Vector3[Fixed64] without C11 compilation errors, specifically successfully running operations like v3_fixed.Equals(...) and v3_fixed + v3_fixed.
