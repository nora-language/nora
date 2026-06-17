# Control Flow

## Overview

Nora provides standard imperative control flow structures, ensuring predictability and readability. The primary constructs are `if/else`, `while`, `break`, and `continue`.

*(Note: Pattern matching (`match`) and Range-based loops (`for ... in`) are documented in their own dedicated specifications).*

## Conditional Branching (`if` / `else`)

The `if` statement evaluates a boolean expression. Parentheses around the condition are optional but braces `{}` around the block are **required**.

```nora
var x = 10

if x > 5 {
    io.PrintLn("x is large")
} else if x == 5 {
    io.PrintLn("x is exactly 5")
} else {
    io.PrintLn("x is small")
}
```

*Note: Nora does not support implicit truthiness. You cannot write `if 1 { ... }`; the condition must evaluate to a strict `bool` type.*

## Looping (`while`)

The `while` loop continues executing its block as long as the boolean condition remains `true`.

```nora
var count = 0

while count < 5 {
    io.PrintLn("Count is ${count}")
    count = count + 1
}
```

## Loop Control (`break` and `continue`)

### 1. `break`
The `break` statement immediately terminates the innermost enclosing loop (`while` or `for`). Execution resumes at the next statement following the loop block.

```nora
var i = 0
while true {
    if i == 3 {
        break
    }
    i = i + 1
}
```

### 2. `continue`
The `continue` statement immediately jumps to the next iteration of the innermost enclosing loop, skipping any remaining code in the current iteration block.

```nora
var j = 0
while j < 5 {
    j = j + 1
    if j % 2 == 0 {
        continue // Skip even numbers
    }
    io.PrintLn("Odd: ${j}")
}
```

## Errors & Diagnostics

*   **Non-Boolean Condition:** Supplying a non-`bool` type to an `if` or `while` condition yields a type-check error.
*   **Unreachable Code:** The compiler may emit warnings if it detects code following an unconditional `break` or `return` inside a block.
