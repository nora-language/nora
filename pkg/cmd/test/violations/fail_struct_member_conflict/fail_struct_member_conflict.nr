package main

type Box = struct {
    value: i32 // <-- EXPECTED SEMANTIC ERROR: name conflict in struct 'Box': field 'value' conflicts with method 'value'
}

fn (self: #Box) value() i32 {
    return self.value
}

fn main() i32 {
    return 0
}
