# Investigation: Generic Method Call Type Inference Failure

## Status
Open

## Problem
When calling methods on instances of generic types (e.g., `Result[T, E]` or `collections.Vector[T]`), the compiler's semantic analyzer fails to automatically infer the generic parameters for the method based on the receiver's type. Developers are forced to redundantly specify the type arguments on every method call.

For example, when operating on a variable `val_res` of type `Result[JsonValue, str]`:
```nora
// Expected to work (inference from receiver type)
if (val_res.IsErr()) { ... }

// Actually required by the current compiler
if (val_res.IsErr[JsonValue, str]()) { ... }
```

Similarly, when operating on an array `arr` of type `collections.Vector[JsonValue]`:
```nora
// Expected to work
var length = arr.Len()
var item = arr.Get(i)

// Actually required
var length = arr.Len[JsonValue]()
var item = arr.Get[JsonValue](i)
```

## Reproduction
1. Define a generic struct and a method for it.
2. Instantiate the generic struct with concrete types.
3. Call the method without explicit generic type arguments.
4. The compiler emits a type inference error (e.g., "cannot infer type parameters").

```nora
import "collections"

pub fn main() {
    var vec = collections.NewVector[i32](10)
    // The following line triggers a compiler error unless explicitly written as vec.Push[i32](5)
    vec.Push(5) 
}
```

## Root Cause
The semantic analyzer (`pkg/semantic`) currently does not propagate the bound generic arguments from a resolved receiver's type into the AST node of the method call during the name resolution and type-checking pass. 

When the parser sees a method call `receiver.Method()`, it creates a `CallExpr` node. The type-checker looks up `Method`, sees that it belongs to a generic definition `Vector[T]`, but it fails to map the `T` from the receiver's concrete type (`Vector[i32]`) into the method's required generic constraints. Thus, it treats the method call as an un-instantiated generic function and demands explicit `[T]` brackets on the call site.

## Fix (Proposed)
The fix needs to be applied in the Semantic Analysis pass (`pkg/semantic`):
1. During `CheckMethodCall` or `CheckCallExpr`, after resolving the type of the `receiver`.
2. If the receiver is a concrete instantiation of a generic type (e.g., `TypeInstance{ Base: Vector, Args: [i32] }`), extract the mapped `Args`.
3. Automatically inject these mapped arguments into the `CallExpr`'s generic arguments list if the user omitted them.
4. Proceed with the standard generic monomorphization/type-checking pass.

## Validation
A future patch to `pkg/semantic` should add the following integration tests:
- **Positive Test:** Ensure `collections.Vector[i32]` can use `.Push(5)`, `.Len()`, and `.Get(0)` without type brackets.
- **Positive Test:** Ensure `Result[str, str]` can use `.IsErr()`, `.Unwrap()`, and `.UnwrapErr()` without type brackets.
- **Regression Test:** Verify that explicit type parameters still work if provided, as long as they match the receiver's instantiated types.
