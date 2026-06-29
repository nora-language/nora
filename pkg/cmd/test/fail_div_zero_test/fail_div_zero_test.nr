package main

pub type Fixed64 = struct {
    val: i64,
}

pub fn (self: #Fixed64) cmp(other: #Fixed64) i32 {
    if (self.val < other.val) { return -1 }
    if (self.val > other.val) { return 1 }
    return 0
}

pub fn (self: #Fixed64) div(other: #Fixed64) Fixed64 {
    return Fixed64 { val: (self.val * 65536) / other.val }
}

pub fn main() {
    var zero = Fixed64 { val: 0 }
    var a = Fixed64 { val: 0 }
    
    if (a == zero) {
        var x = a.div(#zero)
    }
}
