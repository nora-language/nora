package main

pub fn main() {
    // REAL UNRESOLVED LIMITATION #5: No Inner-Function Statement-Level [cfg] Conditional Compilation
    // What you want to write:
    // Attaching [cfg(target = "windows")] to individual statements inside a function body.
    //
    // The Reality / Compiler Error:
    // Parser error / Syntax error on inner [cfg].
    // Nora's [cfg] attribute only works on top-level fn and type declarations.

    [cfg(target = "windows")] // FAILS
    var x = 10
}
