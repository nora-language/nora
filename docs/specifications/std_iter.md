# Iterators

## Overview

Nora provides a zero-cost, type-safe Iterator abstraction located in `std/iter/iter.nr`. The `Iterator[T]` interface powers lazy sequence evaluation and functional-style data pipelines. 

## The `Iterator[T]` Interface

Any struct that implements a `Next()` method returning an `Option[T]` satisfies the `Iterator[T]` interface:

```nora
pub type Iterator[T] = interface {
    fn Next() Option[T]
}
```

## Combinators

Nora implements a suite of lazy combinator structs (`TakeIter`, `MapIter`, `FilterIter`, `StepByIter`, `SkipIter`). Because these combinators are implemented natively as structs holding generic boundaries, they allocate zero heap memory and can be aggressively inlined by the C11 compiler.

These methods are implemented as extension methods on specific iterators (like `prelude.Range`):

### 1. `Take(limit: i32)`
Yields exactly the first `limit` elements of the underlying iterator before returning `None`.
```nora
var r = 0..10
var first_five = r.Take(5)
```

### 2. `Filter(predicate: fn(#T) bool)`
Yields only the elements for which the `predicate` closure returns `true`. The predicate is passed a read-only lease (`#T`) to prevent unintended mutations during filtering.
```nora
var evens = (0..10).Filter(fn(val: #i32) bool {
    return val % 2 == 0
})
```

### 3. `Map(mapper: fn(T) U)`
Transforms each element `T` yielded by the iterator into type `U` by applying the `mapper` closure.
```nora
var doubled = (0..5).Map(fn(val: i32) i32 {
    return val * 2
})
```

### 4. `StepBy(step: i32)`
Yields the first element, then skips `step - 1` elements before yielding the next element. 
```nora
var stepping = (0..10).StepBy(2) // Yields 0, 2, 4, 6, 8
```

### 5. `Skip(skip: i32)`
Discards the first `skip` elements of the iterator, yielding all elements that follow.
```nora
var skipped = (0..10).Skip(5) // Yields 5, 6, 7, 8, 9
```
