package main

import "io"

fn switchkey() {
    var a = Some[i32](10)
    // suppose can handle literal, user defind, not only sumtype
    match a {
        Some(v) => {
            io.PrintLn("Value: ${v}")
        }
        _ => {
            io.PrintLn("Default")
        }
    }
}
