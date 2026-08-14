# Compiler Investigation: Implicit Semicolon Insertion in Multiline Arithmetic

## Status
Completed

## Problem
The `generic_math_demo` in `nora_math` panicked during execution when running the `SolveLinearSystem3x3F32` solver. The computed solutions for standard linear systems were wildly incorrect, yet the arithmetic expressions appeared structurally correct in the `.nr` source code. 

## Reproduction
The problem occurs when writing multiline arithmetic statements where the previous line ends with a closed parenthesis `)` and the subsequent line begins with an operator such as `-` or `+`.

```nr
var det = m.m00 * (m.m11 * m.m22 - m.m12 * m.m21)
        - m.m01 * (m.m10 * m.m22 - m.m12 * m.m20)
        + m.m02 * (m.m10 * m.m21 - m.m11 * m.m20)
```

## Root Cause
Nora's compiler implements Automatic Semicolon Insertion (ASI). When the parser encounters a newline after a valid closing token like `)`, it implicitly inserts a semicolon, terminating the expression. 

Because of this, the compiler parses the single arithmetic expression as three distinct statements. It assigns only the first term to the variable, and parses the remaining lines as separate, dead-code evaluations:

```nr
// How the parser interprets the code:
var det = m.m00 * (m.m11 * m.m22 - m.m12 * m.m21); // Assignment terminates here
(- m.m01 * (m.m10 * m.m22 - m.m12 * m.m20));       // Dead code statement
(+ m.m02 * (m.m10 * m.m21 - m.m11 * m.m20));       // Dead code statement
```

This caused `SolveLinearSystem3x3F32` to silently compute only the first sub-determinant term of Cramer's rule, completely invalidating the linear solver.

## Fix
The immediate fix applied to the codebase was to re-format the expressions so that the operators (`-`, `+`) remain at the end of the preceding line. This prevents the ASI pass from erroneously terminating the expression.

```nr
var det = m.m00 * (m.m11 * m.m22 - m.m12 * m.m21) -
          m.m01 * (m.m10 * m.m22 - m.m12 * m.m20) +
          m.m02 * (m.m10 * m.m21 - m.m11 * m.m20)
```

For the long term, the compiler parser (`pkg/parser`) should be updated. A potential fix could be looking ahead to the next token on the newline before applying ASI; if the token is a binary operator that expects a left-hand operand, the parser should not insert a semicolon.

## Validation
After adjusting the trailing operators in `nora_math/src/math/numerical.nr`, the transpiled C code correctly merged all terms into single assignment statements. Running `..\nora.exe run -debug-memory --example generic_math_demo` verified that the `SolveLinearSystem3x3F32` function computed the determinants accurately, passing all mathematical parity tests.
