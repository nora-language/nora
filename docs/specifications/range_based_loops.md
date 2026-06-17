# Specification: Range-Based For Loops

## Title & Overview
**Range-Based For Loops** introduce a syntactically concise, zero-cost mechanism for numeric iteration over bounded integer sequences in Nora. By utilizing the `..` operator as part of a `for-in` loop's iterable component, developers can clearly express bounded loop conditions without manually managing indices or falling back to C-style `while` loops.

## Motivation
Historically, looping in Nora over a sequential set of integers required explicit `while` loop index management. This boilerplate is verbose, prone to off-by-one bugs, and visually obscures the intent of the loop.
Range-based for loops (`for i in start..end`) standardize this pattern. By moving range iteration directly into the core language syntax, the compiler can apply static safety bounds-checking and compile the loop into an optimized `while` structure with zero runtime allocations or iterator interface overheads.

## Syntax
A numeric range is constructed using the `..` infix operator. It is primarily used within the iterable clause of a `for-in` loop.

```bnf
RangeExpression ::= Expression ".." Expression
ForRangeStatement ::= "for" Identifier "in" RangeExpression BlockStatement
```

### Constraints:
- Range syntax is only parsed as part of expressions; within loops it intercepts the `for x in y` pattern.
- Unlike traditional arrays, a range loop binds only a single variable (`value`), and prohibits the declaration of a secondary `key`/`index` variable (e.g. `for i, v in 0..10` is illegal).

## Semantics
The expression `start..end` evaluates bounds that run from `start` (inclusive) to `end` (exclusive).

When lowered to HIR (High-level Intermediate Representation), a range loop expands to a simple `<` comparison check equivalent to:

```nora
// Syntax
for i in start..end {
    // Body
}

// Lowers conceptually to:
var _end_tmp = end
var _i = start
while _i < _end_tmp {
    var i = _i
    // Body
    _i = _i + 1
}
```

## Type Rules
1. **Primitive Integral Constraints**: The expressions used for `start` and `end` must evaluate to a primitive integral type (e.g., `i32`, `u64`, `i8`). Floats, structs, and strings are statically rejected.
2. **Type Equivalence**: The type of `start` and `end` must match exactly. The Nora compiler enforces strict matching to avoid implicit widening conversions. If `start` is `i32` and `end` is `u64`, the compiler will emit a semantic error.
3. **Inferred Variable Type**: The loop control variable inherits its type from the `start` bound automatically.

## Lease Rules
- The bounds expressions (`start` and `end`) are evaluated once at the beginning of the loop execution. Consequently, any leases or ownerships taken during their evaluation expire or persist based on normal expression rules prior to entering the loop body.
- The iteration variable generated inside the block is an independent value assignment (`Initialized`) and operates locally within the `for` loop's lexical scope.

## Examples

### Basic Iteration
```nora
var sum = 0
for i in 0..10 {
    sum += i // Sums numbers 0 through 9
}
```

### Using Variables as Bounds
```nora
fn process_slice(start: i32, end: i32) {
    for i in start..end {
        io.print_i32(i)
    }
}
```

## Edge Cases
- **Negative Bounds**: Iterating from negative to positive (e.g., `-5..5`) correctly steps positively.
- **Empty Ranges**: If `start >= end`, the loop body evaluates zero times and exits immediately.
- **Reverse Iteration**: Iterating backwards (e.g., `10..0`) will currently evaluate as an empty range since the generated HIR uses `<` for its loop condition and `+1` for stepping. (Reverse loops are out of scope for the initial implementation).

## Errors & Diagnostics
- `range bounds must be integers, got <Type>`: Thrown when a non-integer is supplied to either side of the `..` operator.
- `range bounds must have exactly the same type, got <Type1> and <Type2>`: Thrown when the user attempts an implicit narrowing/widening conversion between bounds.
- `range loop cannot have both key and value variables, just one`: Thrown if the user attempts to bind two variables (e.g., `for k, v in 0..10 { ... }`).

## Future Considerations
- **Inclusive Range Operator**: Introducing `..=` to include the final boundary element.
- **Reverse Iteration/Stepping**: Allowing `step` definitions or negative iteration (e.g. `.step(-1)`) via trait extensions or standard library integration.
- **Standalone Range Objects**: Currently, `RangeExpression` is primarily analyzed securely inside `ast.ForStatement`. Future implementations could elevate `start..end` to instantiate an actual `Range[T]` struct passed around as first-class variables.
