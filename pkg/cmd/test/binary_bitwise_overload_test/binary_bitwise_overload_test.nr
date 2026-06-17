package main

import "core/ops"

pub type BitSet = struct {
    val: i32
}

pub fn (a: #BitSet) bitand(b: BitSet) BitSet {
    return BitSet{val: a.val & b.val}
}

pub fn (a: #BitSet) bitor(b: BitSet) BitSet {
    return BitSet{val: a.val | b.val}
}

pub fn (a: #BitSet) bitxor(b: BitSet) BitSet {
    return BitSet{val: a.val ^ b.val}
}

pub fn (a: #BitSet) shl(b: BitSet) BitSet {
    return BitSet{val: a.val << b.val}
}

pub fn (a: #BitSet) shr(b: BitSet) BitSet {
    return BitSet{val: a.val >> b.val}
}

pub fn main() i32 {
    var a = BitSet{val: 5}  // 0101
    var b = BitSet{val: 3}  // 0011

    var c = a & b
    if (c.val != 1) { // 0001
        panic("a & b != 1")
    }

    var d = a | b
    if (d.val != 7) { // 0111
        panic("a | b != 7")
    }

    var e = a ^ b
    if (e.val != 6) { // 0110
        panic("a ^ b != 6")
    }

    var f = BitSet{val: 1}
    var shiftAmt = BitSet{val: 2}

    var g = f << shiftAmt
    if (g.val != 4) {
        panic("f << shiftAmt != 4")
    }

    var h = BitSet{val: 8}
    var i = h >> shiftAmt
    if (i.val != 2) {
        panic("h >> shiftAmt != 2")
    }

    return 0
}
