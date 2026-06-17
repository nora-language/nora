# Bounds Check Elimination (BCE)

## Title & Overview
**Bounds Check Elimination (BCE)** is an automated compiler optimization in Nora that safely eliminates runtime array bounds checking `array_bounds_check()` during array or vector access without compromising the compile-time or runtime memory safety guarantees of the language.

BCE is driven by static range analysis performed during the semantic analysis pass. By tracking the loop index variables and evaluating their boundaries, the compiler can omit the boundary verification when an index is proven to be strictly bounded by the `len()` of the structure.

## Motivation
Standard element access operations on arrays and slices (`arr[i]`) carry a slight runtime overhead because the generated C11 code natively instruments these with a macro `array_bounds_check()`. This checks that `i` is within `[0, len(arr))`. In performance-critical tight loops (like mathematical vector transformations, serialization routines, and memory shuffling), this repeated verification incurs unnecessary overhead.

The goal of BCE is to achieve "zero-cost abstractions". By inferring the bounds of variables mathematically at compile time, we maintain the high performance of C while preserving memory safety.

## Syntax
There is no new syntax introduced by BCE. This is a purely internal static compiler optimization applied implicitly to standard language constructs.

## Semantics
The semantic analyzer tracks mathematical bounds `[MinBound, MaxSymbol]` of local integer variables within scopes (`VarBounds`).
When a loop is established, the semantic analyzer intercepts the initialization, condition, and increment rules:
1. **Seeding:** In a loop like `for var i = 0; i < len(arr); i++`, `i` is assigned a `VarBounds` with `MinBound = 0` and `MaxSymbol = len(arr)`.
2. **Evaluation:** When the index expression `arr[i]` is encountered, the analyzer determines if the upper bound of `i` matches `len(arr)`.
3. **Invalidation:** If `i` is explicitly mutated inside the loop block (e.g., `i = 5` or `i = i - 1`), the compiler immediately invalidates the bounds, falling back to safe runtime bound checks.

When the conditions match, the AST Node `ast.IndexExpression` is marked with `NoBoundsCheck = true`. The C-codegen backend respects this flag and directly outputs zero-cost C array access.

## Type Rules
- BCE is strictly applied to integer types used as indexers (e.g. `i32`, `i64`).
- It applies to arrays, dynamically sized arrays (slices), and memory structures with `.data` buffers where length bounds are well-defined.

## Lease Rules
BCE has no explicit interaction with lease lifetimes. It simply optimizes the generated code. Any underlying moves or borrows inside the array index expressions will still behave as standard Nora semantics.

## Examples
### Standard BCE
```nora
pub fn sum(arr: #i32[]) i32 {
    var total = 0
    // The compiler proves `i` ranges from 0 to len(arr) - 1.
    // Therefore, no runtime bounds checks are generated for `arr[i]`.
    for var i = 0; i < len(arr); i++ {
        total = total + arr[i]
    }
    return total
}
```

## Edge Cases
1. **Manual Index Mutability:** If the developer mutates the loop index inside the body, the range bounds are invalidated and standard runtime checking applies:
   ```nora
   for var i = 0; i < len(arr); i++ {
       if (i == 5) {
           i = 6 // Compiler invalidates bounds
       }
       arr[i] = 10 // Safe: bound checks injected!
   }
   ```
2. **Compound/Complex Conditions:** Currently, only simple bounds conditions (`i < len(arr)`) are optimized. Complex logical operators (`&&`, `||`) are intentionally ignored by the BCE analyzer to prevent conservative edge-case bugs.

## Errors & Diagnostics
BCE emits no new errors or warnings, as it is a transparent optimization. Failures to prove safe bounds silently result in the generation of standard checked array accesses (`array_bounds_check`).

## Future Considerations
- Expand Range Analysis to support variable offsets like `i < len(arr) - 2`.
- Enable cross-block range tracking across `if` boundaries to prove limits beyond just `ForStatement` conditions.
- Include data-flow analysis to optimize bounds when `len(arr)` is evaluated via generic method receiver interfaces.
