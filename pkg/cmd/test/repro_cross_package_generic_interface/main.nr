package main

import "io"
import "./pkg_a"
import "./pkg_b"

fn render[T: pkg_a.Printable](item: T) str {
    return item.to_string()
}

fn main() i32 {
    var w = pkg_b.Widget { name: "button_widget" }
    var s = render(w)
    
    io.PrintLn("Rendered:", s)
    if (s != "button_widget") {
        return 1
    }
    return 0
}
