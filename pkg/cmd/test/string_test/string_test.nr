package main

import "io"
import "string"

fn main() i32 {
    var s = "  Hello Nora  "
    io.PrintLn("Original: [${s}]")

    var trimmed = string.Trim(s)
    io.PrintLn("Trimmed:  [${trimmed}]")

    var upper = string.ToUpper(trimmed)
    var lower = string.ToLower(trimmed)
    io.PrintLn("Upper:    [${upper}]")
    io.PrintLn("Lower:    [${lower}]")

    if (string.Contains(trimmed, "Hello")) {
        io.PrintLn("Contains 'Hello': yes")
    }

    if (string.HasPrefix(trimmed, "Hell")) {
        io.PrintLn("Has prefix 'Hell': yes")
    }

    if (string.HasSuffix(trimmed, "nos")) {
        io.PrintLn("Has suffix 'nos': yes")
    }

    var eq = string.Equals("abc", "abc")
    if (eq) {
        io.PrintLn("equals('abc','abc'): yes")
    }

    io.PrintLn("All string tests completed.")
    return 0
}
