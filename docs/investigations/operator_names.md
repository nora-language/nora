# Investigation: Operator Overloading Hardcoded Names

**Status**: Completed
**Context**: Nora currently hardcodes common names like `add` and `sub` for operator overloading.

## The Problem with Duck Typing Operators
Currently, in `analyzer.go`, the `+` operator looks for a method named `add`. 
This creates several architectural problems:
1. **Accidental Overloading**: If a user creates a `Database` struct and adds an `add(user: User)` method, they can suddenly write `db + user`. This is semantically incorrect and confusing.
2. **Namespace Pollution**: Words like `add`, `sub`, and `mul` are extremely common English verbs. Reserving them exclusively for math operators restricts API design.
3. **Readability**: When reading a struct definition, `fn add(...)` looks like a normal method. It is not obvious that it hooks into language-level syntax (`+`).

## Proposed Solutions

### Option A: Explicit Operator Syntax (C++ Style)
Instead of using standard identifiers, introduce a special keyword or syntax for operators.
```nora
pub fn (self: &Vector) operator+(other: Vector) Vector { ... }
// or
pub fn (self: &Vector) op_add(other: Vector) Vector { ... }
```
**Pros**: 
* Completely unambiguous. 
* Solves namespace pollution (you can still have a normal `add` method).
* Instantly readable intent.

### Option B: Magic Method Names (Python Style)
Change the hardcoded names from common English words to reserved magic names, such as `__add__` or `__sub__`.
```nora
pub fn (self: &Vector) __add__(other: Vector) Vector { ... }
```
**Pros**: Requires zero parser changes.
**Cons**: Looks slightly ugly ("dunder" methods).

## Conclusion
The user is correct: using common names like `add` for operator overloading is a bad practice for a strictly typed language. Moving to an explicit `operator+` or a reserved `op_add` naming convention is highly recommended.
