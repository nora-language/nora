package main

import "io"

type Trait = interface {
    fn do()
}

type Impl = struct {
    data: @str
}

fn (i: #Impl) do() {
    io.PrintLn("doing:", i.data)
}

fn main() i32 {
    // We allocate a struct containing an owned string (RAII resource)
    var obj = alloc Impl { data: "leak me?" }
    
    // We put it into an interface existential box.
    // When `iface` goes out of scope, does the compiler generate a call to the virtual drop function
    // to free the inner `Impl.data` string? If it leaks, we found a bug!
    var iface: Trait = obj
    
    iface.do()
    
    return 0
}
