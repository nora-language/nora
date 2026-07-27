# Investigation: Generic Overload Resolution Bug

**Status:** Resolved
**Date:** July 27, 2026

## Problem
Generic struct types (like `Box[T]` or `Vector[T]`) failed to properly resolve and generate C codegen for overloaded binary operators (e.g., `operator+`). 
When a generic struct had multiple overloads (e.g., `operator+(other: Box[T])` and `operator+(scalar: T)`), the compiler would correctly select the right overload index during the semantic pass, but would then inexplicably emit raw `(b1 + 5)` C-code instead of the required mangled C function call. This resulted in C compilation failures: `error: invalid operands to binary expression`.

## Reproduction
```nora
pub type Box[T] = struct { val: T }

pub fn (_self: #Box[T]) operator+(other: Box[T]) Box[T] { return Box[T] { val: other.val } }
pub fn (_self: #Box[T]) operator+(scalar: T) Box[T] { return Box[T] { val: scalar } }

pub fn main() {
    var b1 = Box[i32] { val: 10 }
    var b2 = Box[i32] { val: 20 }
    
    var b3 = b1 + b2 // Worked correctly, instantiated Box+Box
    var b4 = b1 + 5  // Failed: Emitted (b1 + 5) in C instead of Box_operator_plus(...)
}
```

## Root Cause
The failure was the culmination of three intertwined bugs in the compiler frontend:

1. **Lexical Scope Desynchronization in `CollectSymbols`**: 
   When `CollectSymbols` encountered methods belonging to generic structs, it created a temporary `ScopeFunction` to hold the generic parameter `T`. It then defined the method symbol (e.g., `"Box_operator+"`) *inside* this temporary scope. Because each overloaded method was processed in a different temporary scope, the semantic analyzer failed to detect the symbol collision in the package scope. As a result, the methods were never merged into a `SymOverloadGroup`, and instead, the last parsed method silently overwrote the previous ones in the struct's `MethodSymbols` map.

2. **Destructive Registration in `Monomorphize`**:
   When `Box+Box` was evaluated, the compiler correctly instantiated the concrete specialized function. However, `Monomorphize` then improperly registered the newly generated concrete symbol back into the struct's `MethodSymbols` map under the **original** un-mangled method name (`"operator+"`). This destructive action overwrote the generic `SymOverloadGroup`, leaving only the monomorphized `Box+Box` function in the cache.

3. **Flawed Generic Method Fallback Logic**:
   When the compiler subsequently evaluated `Box+5`, it attempted to look up `"operator+"` in the concrete struct's `MethodSymbols` map. The lookup failed because `Monomorphize` had overwritten it with the mangled name. The compiler was supposed to fall back to the generic template's `st.BaseType` method map. However, the fallback condition strictly checked if the concrete struct's map was `nil`. Because the map itself had been initialized during the first monomorphization pass, the compiler bypassed the fallback entirely. Without a valid `targetSym`, the HIR lowerer bypassed the overloaded function lowering (flagging `isStructOp = false`), and pushed a raw `BinOp` instruction, generating raw C addition `(b1 + 5)`.

## Fix
1. **Scope Binding:** `CollectSymbols` was modified to crawl up the scope chain and enforce that all method symbols are defined in the `ScopePackage`. This ensures overloads collide correctly and are successfully merged into a single `SymOverloadGroup`.
2. **Preserve Original Names:** Removed the destructive dictionary overwrite inside `Monomorphize`. Monomorphized functions are now explicitly bound to their mangled C-names, preserving the generic `SymOverloadGroup` under the original un-mangled name.
3. **Correct Fallback Handling:** Corrected the generic method fallback logic in `resolveOverload` to ensure it falls back to the generic `st.BaseType` whenever the *specific method key* is absent, rather than short-circuiting just because the dictionary itself has been allocated.

## Validation
Integration tests (`pkg/cmd/test/generic_overload_bug/main.nr`) containing sequential, interleaved calls to `Box+Box` and `Box+T` compile correctly. C code generation correctly outputs mangled C function calls for all generic struct overloaded operators.
