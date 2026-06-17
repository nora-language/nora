# Standard Library: Strconv

## Overview

The `std/strconv` package implements conversions to and from string representations of basic data types. It provides fast, non-allocating formatting routines and robust parsing algorithms for numeric, boolean, and floating-point data.

## Core Functions

### String to Primitive (Parsing)
*   `ParseBool(str: str) (bool, Option[Error])`: Returns the boolean value represented by the string. It accepts "1", "t", "T", "true", "TRUE", "True", "0", "f", "F", "false", "FALSE", "False".
*   `ParseInt(s: str, base: i32, bitSize: i32) (i64, Option[Error])`: Interprets a string `s` in the given base (0, 2 to 36) and bit size (0 to 64) and returns the corresponding value.
*   `ParseUint(s: str, base: i32, bitSize: i32) (u64, Option[Error])`: Like `ParseInt` but for unsigned numbers.
*   `ParseFloat(s: str, bitSize: i32) (f64, Option[Error])`: Converts the string `s` to a floating-point number with the precision specified by `bitSize` (32 or 64).

### Primitive to String (Formatting)
*   `FormatBool(b: bool) str`: Returns "true" or "false" according to the value of `b`.
*   `FormatInt(i: i64, base: i32) str`: Returns the string representation of `i` in the given base.
*   `FormatUint(i: u64, base: i32) str`: Returns the string representation of `i` in the given base.
*   `FormatFloat(f: f64, fmt: u8, prec: i32, bitSize: i32) str`: Formats floating-point numbers.

## Convenience Wrappers
*   `Atoi(s: str) (i32, Option[Error])`: Equivalent to `ParseInt(s, 10, 0)`, converted to `i32`.
*   `Itoa(i: i32) str`: Equivalent to `FormatInt(i64(i), 10)`.

## Memory Considerations
Functions like `Itoa` and `FormatInt` allocate dynamically to return a newly formatted `str`. For high-performance formatting without heap allocations, use the `Append*` family of functions (e.g., `AppendInt(dst: &[]u8, i: i64, base: i32) []u8`), which format data directly into pre-allocated slice leases.
