# Standard Library: Math

## Overview

The `std/math` package provides foundational mathematical constants and robust functions for basic computations, trigonometric operations, and floating-point manipulations. It maps closely to C11's `<math.h>` in the generated output to ensure hardware-level performance and precision.

## Constants

*   `E`: Euler's number (~2.71828)
*   `PI`: Archimedes' constant (~3.14159)
*   `PHI`: Golden ratio (~1.61803)
*   `SQRT2`: Square root of 2 (~1.41421)

## Core Functions

### Trigonometry
*   `Sin(x: f64) f64`
*   `Cos(x: f64) f64`
*   `Tan(x: f64) f64`
*   `Asin(x: f64) f64`

### Exponentials & Logarithms
*   `Exp(x: f64) f64`
*   `Log(x: f64) f64` (Natural logarithm)
*   `Log10(x: f64) f64`
*   `Pow(x, y: f64) f64`

### Floating Point Manipulation
*   `Floor(x: f64) f64`
*   `Ceil(x: f64) f64`
*   `Round(x: f64) f64`
*   `Abs(x: f64) f64`
*   `IsNaN(f: f64) bool`
*   `IsInf(f: f64, sign: i32) bool`

## Monomorphization Note

By default, mathematical operations take and return `f64`. Nora's compiler provides generic overloads `math.Abs[T](x: T) T` for other numerical types (`f32`, `i32`, `i64`), automatically generating monomorphized C functions based on the numeric primitive.
