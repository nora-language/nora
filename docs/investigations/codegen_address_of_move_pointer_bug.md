# Investigation Report: Codegen Address-Of Move Pointer Bug

## Status
**Completed & Validated**

## Problem
During the compilation of generic collections (specifically `std/collections/vector.nr`) for non-pointer-erased types, clang emitted type mismatch errors of the form:
```
../std/collections/vector.nr:49:39: error: assigning to 'gfx_Mesh' (aka 'struct gfx_Mesh') from incompatible type 'gfx_Mesh *' (aka 'struct gfx_Mesh *'); dereference with *
   49 |     ((((gfx_Mesh*)v->data)[v->size])) = *&(val);
```
Here, `val` is a parameter to `Push[T](val: T)`. When instantiated with `T = Mesh`, `val` has the Nora type `Mesh`. However, under C compilation rules, because `Mesh` is a struct, it is passed by pointer under the hood, resulting in the C parameter type `gfx_Mesh * val`. 

Inside `Push`, the move statement `@val` is transpiled. The compiler generated `*&(val)`, which under C type rules evaluates to `gfx_Mesh *`, while the destination buffer expected `gfx_Mesh`.

## Reproduction
1. Define a struct `Mesh` and instantiate `collections.NewVector[Mesh](0)`.
2. Push a mesh to the vector: `vec.Push(@mesh)`.
3. During code generation, `@val` inside `Push[T](val: T)` generates `*&(val)` because:
   - `@val` is translated to an `AddressOf` instruction in HIR.
   - `AddressOf` transpiler generates `&(val)` because it treats `val` as an LValue.
   - The outer `Deref` of the store instruction then wraps it as `*&(val)`.
   - Since `val` is a parameter passed by pointer, it is already `gfx_Mesh *`, so `&(val)` is `gfx_Mesh **` and `*&(val)` is `gfx_Mesh *`.

## Root Cause
Nora's `AddressOf` instruction generator (`pkg/codegen/hir_codegen.go`) didn't account for whether the operand to a move (`@`) operator was already passed by pointer under the C calling convention. Because the type of the move instruction in Nora is the base value type `Mesh` (non-pointer) but the C parameter type is a pointer `Mesh *`, the generator compared the two, found a type mismatch, and fell back to referencing the LValue using `&(val)`, producing a pointer-to-pointer.

## Fix
We modified `pkg/codegen/hir_codegen.go` inside the `*hir.AddressOf` code generation block:
```go
		if i.Operator == "@" {
            ...
			if g.isOperandPointerInC(i.Val) {
				return opStr
			}
		}
```
If the operator is `@` (move) and the operand is already a pointer under the C transpilation convention, the generator returns the operand string (`val`) directly instead of wrapping it in `&(...)`. When the outer `Deref` is applied, it cleanly dereferences the pointer (`*val`), producing the correct base value type `gfx_Mesh` for the assignment.

## Validation
Rebuilt the `nora` compiler and successfully compiled the `gltf_viewer` example target, which now transpiles to clean, valid C code and passes clang compilation.
