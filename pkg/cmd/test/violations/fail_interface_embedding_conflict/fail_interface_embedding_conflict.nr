package main

type Reader = interface {
    fn read(self: #Reader) str
}

type Scanner = interface {
    fn read(self: #Scanner) i32
}

type Combined = interface {
    Reader
    Scanner // <-- EXPECTED SEMANTIC ERROR: interface embedding conflict: method 'read' defined in embedded interfaces has incompatible signatures
}

fn main() i32 {
    return 0
}
