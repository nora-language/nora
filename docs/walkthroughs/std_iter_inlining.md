# Walkthrough: Standard Library Zero-Cost Iterators

## Goal
The goal was to investigate and implement the foundational `std/iter` package for the Nora programming language, specifically proving that iterator adaptors like `TakeIter` can be implemented as zero-cost abstractions through explicit compiler inlining.

## Implementation Details

### The Challenge
Historically, iterator wrappers in generic programming languages often incur overhead due to interface boxing and vtable dynamic dispatch. In Nora, achieving "Zero-Cost Iterators" (similar to Rust) requires ensuring that iterator methods can be statically dispatched and entirely inlined, bypassing any runtime overhead from pointer indirections or interface usage.

### The Solution: Constrained Generics and the `[inline]` Pass
Rather than relying on `interface` implementations that trigger dynamic dispatch, we designed the iterator combinators using **Constrained Generics**.
We created a bounded interface (`IntIterator`) which enforces the static duck-typing rule (`Next() Option[T]`), allowing the generic parameter `I` to be type-checked prior to monomorphization.

```nora
pub type IntIterator = interface {
    fn Next() Option[i32]
}

pub type TakeIter[I: IntIterator] = struct {
    iter: I
    limit: i32
    current: i32
}
```

The core iteration functions are annotated with `[inline]`:
```nora
[inline]
pub fn (t: &TakeIter[I]) Next[I: IntIterator]() Option[i32] {
    if t.current >= t.limit {
        return None[i32]
    }
    t.current = t.current + 1
    return t.iter.Next()
}
```

### Compiler Optimizer Inlining
When the `Next` method is called, the compiler's newly introduced optimization passes (which include HIR variable cloning to prevent scope shadowing) completely collapse the iterator layer.
During AST-to-HIR lowering, the `Call` to `TakeIter.Next` is replaced directly with the condition checks and the inner `t.iter.Next()` call. Because `t.iter` is structurally passed around and fully transparent to the compiler, the result is equivalent to handwritten `while` loops over pointers or arrays.

## Verification
We verified the exact inline behavior with the integration test suite in `pkg/cmd/test/iter_test/iter_test.nr`:
- A basic heap-allocated slice iterator (`ArrayIter`)
- The wrapper (`TakeIter[&ArrayIter]`) bounded by `[I: IntIterator]`
- Repeated invocations to explicitly consume iteration chunks and verify capacity constraints.

The compiler safely checks constraints, monomorphizes `TakeIter` for the specific `&ArrayIter` pointer struct, and correctly unrolls and inlines the logic without emitting costly generic interfaces.

## Future Considerations
- The current implementation serves as a functional proof-of-concept. As Nora's support for generic interfaces and method parameters matures, `TakeIter` can be expanded to dynamically bind to `Iterator[T]` instead of being hard-coded to integer streams.
- The `std/iter` package will be extended to include `Filter`, `Map`, `StepBy`, and `Zip`, all utilizing the `[inline]` attributes to ensure chained iterator patterns fold completely away at compile-time.
