package main

import "io"
import "./pkg_vec"

// Repro: Codegen bug variant — calling a cross-package generic method (Dot)
// and passing the result to string interpolation causes a C compile error.
// The semantic analysis passes correctly, but C codegen emits:
//   nr_to_str(d)  where d is a 'float' (f32), not a void*
//
// This is the SAME codegen bug as fail_repro_codegen_generic_method_to_str.
// Keeping this folder to show the original discovery context.

pub fn main() {
    var a = pkg_vec.NewVec2[f32](f32(1.0), f32(2.0))
    var b = pkg_vec.NewVec2[f32](f32(3.0), f32(4.0))
    // Semantic analysis passes. Codegen fails when the result is interpolated:
    var d = a.Dot[f32](#b)
    io.PrintLn("dot = ${d}\n")
}
