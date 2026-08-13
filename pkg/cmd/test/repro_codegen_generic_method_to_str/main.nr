package main

import "io"
import "./pkg_vec"

// Repro: Codegen bug — string interpolating the result of a cross-package
// generic method that returns a primitive type (e.g. f32) causes a C
// compilation error: "passing 'float' to parameter of incompatible type 'void *'"
//
// Expected: io.PrintLn("dot = ${d}\n") compiles and prints the f32 result.
// Actual:   C codegen generates:
//   nr_to_str(d)  -- where d is a float, but nr_to_str expects void*
//
// Root cause: The type-erased monomorphization uses the wrong type erasure
// for the return value of Dot[T](). When T=f32, the result is a primitive float,
// but the codegen treats it as a pointer type due to wrong type erasure.
//
// The same method called without string interpolation works fine.

pub fn main() {
    var a = pkg_vec.NewVec2[f32](f32(1.0), f32(2.0))
    var b = pkg_vec.NewVec2[f32](f32(3.0), f32(4.0))
    var d = a.Dot[f32](#b)
    // Bug triggers here: 'float' passed to nr_to_str(void*)
    io.PrintLn("dot = ${d}\n")
}
