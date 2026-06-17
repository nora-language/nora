package main

import "io"

pub type PrimitiveHasher = interface {
    fn Hash() i32
}

pub fn (self: #i32) Hash() i32 {
    return self + 100
}

pub fn (self: #str) Hash() i32 {
    var sum = 0
    // Simplified hash for test
    if (self == "hello") {
        return 999
    }
    return 0
}

pub fn GenericHash[T: PrimitiveHasher](val: #T) i32 {
    return val.Hash()
}

pub fn main() {
    var x = 5
    var h = x.Hash()
    io.PrintLn("i32 Hash: " + io.FormatInt(h))

    if (h != 105) {
        panic("failed i32 hash")
    }

    var s = "hello"
    var h2 = s.Hash()
    io.PrintLn("str Hash: " + io.FormatInt(h2))
    
    if (h2 != 999) {
        panic("failed str hash")
    }

    var h3 = GenericHash[i32](#x)
    if (h3 != 105) {
        panic("failed generic i32 hash")
    }

    io.PrintLn("Success!")
}
