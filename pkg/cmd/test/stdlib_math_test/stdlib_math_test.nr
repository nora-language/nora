package main

import "io"
import "math"

fn assert_approx(val: f64, expected: f64, eps: f64, name: str) {
    var diff = val - expected
    if (diff < 0.0) {
        diff = 0.0 - diff
    }
    if (diff > eps) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        panic("Error: math assertion failed")
    }
}

fn assert_exact(val: f64, expected: f64, name: str) {
    if (val != expected) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        panic("Error: math exact assertion failed")
    }
}

fn main() {
    io.PrintLn("--- Starting std/math Premium Test Suite ---")

    // 1. Verify Constants
    io.PrintLn("Testing Constants...")
    assert_approx(math.Pi(), 3.141592653589793, 0.000000000000001, "Pi")
    assert_approx(math.E(), 2.718281828459045, 0.000000000000001, "E")
    assert_approx(math.Tau(), 6.283185307179586, 0.000000000000001, "Tau")
    assert_approx(math.Sqrt2(), 1.414213562373095, 0.000000000000001, "Sqrt2")
    assert_approx(math.Ln2(), 0.6931471805599453, 0.000000000000001, "Ln2")
    assert_approx(math.Ln10(), 2.302585092994046, 0.000000000000001, "Ln10")

    // 2. Verify Trigonometry
    io.PrintLn("Testing Trigonometry...")
    var half_pi = math.Pi() / 2.0
    assert_approx(math.Sin(half_pi), 1.0, 0.000000000000001, "Sin(Pi/2)")
    assert_approx(math.Cos(math.Pi()), -1.0, 0.000000000000001, "Cos(Pi)")
    assert_approx(math.Tan(math.Pi() / 4.0), 1.0, 0.000000000000001, "Tan(Pi/4)")
    assert_approx(math.Atan2(1.0, 1.0), math.Pi() / 4.0, 0.000000000000001, "Atan2(1.0, 1.0)")
    assert_approx(math.Sinh(0.0), 0.0, 0.000000000000001, "Sinh(0)")
    assert_approx(math.Cosh(0.0), 1.0, 0.000000000000001, "Cosh(0)")

    // 3. Verify Powers & Logarithms
    io.PrintLn("Testing Powers & Logarithms...")
    assert_approx(math.Sqrt(9.0), 3.0, 0.000000000000001, "Sqrt(9)")
    assert_approx(math.Cbrt(27.0), 3.0, 0.000000000000001, "Cbrt(27)")
    assert_approx(math.Pow(2.0, 3.0), 8.0, 0.000000000000001, "Pow(2, 3)")
    assert_approx(math.Log2(8.0), 3.0, 0.000000000000001, "Log2(8)")
    assert_approx(math.Log10(100.0), 2.0, 0.000000000000001, "Log10(100)")
    assert_approx(math.Ln(math.E()), 1.0, 0.000000000000001, "Ln(E)")

    // 4. Verify Rounding & Float Arithmetic
    io.PrintLn("Testing Rounding & Arithmetic...")
    assert_exact(math.Floor(2.8), 2.0, "Floor(2.8)")
    assert_exact(math.Ceil(2.1), 3.0, "Ceil(2.1)")
    assert_exact(math.Round(2.5), 3.0, "Round(2.5)")
    assert_exact(math.Trunc(2.9), 2.0, "Trunc(2.9)")
    assert_exact(math.Abs(-5.5), 5.5, "Abs(-5.5)")
    assert_exact(math.Max(1.2, 3.4), 3.4, "Max")
    assert_exact(math.Min(1.2, 3.4), 1.2, "Min")
    assert_approx(math.Mod(5.5, 2.0), 1.5, 0.000000000000001, "Mod")
    assert_approx(math.Remainder(5.5, 2.0), -0.5, 0.000000000000001, "Remainder")

    io.PrintLn("std/math Premium Test Suite passed successfully!")
}
