package main
import "io"

pub type Inner = struct { val: i32 }
pub type Outer = struct { p: ptr }

fn take_ptr(p: ptr) {}

fn return_ptr() ptr {
    return alloc Inner { val: 77 }
}

fn main() {
    // Scenario 1: Assignment to unowned simple identifier
    var p: ptr
    p = alloc Inner { val: 42 }

    // Scenario 2: Array assignment
    var arr = alloc ptr[2]
    arr[0] = alloc Inner { val: 43 }

    // Scenario 3: Struct literal
    var o = alloc Outer {
        p: alloc Inner { val: 44 }
    }

    // Scenario 4: Struct field assignment
    o.p = alloc Inner { val: 45 }

    // Scenario 5: Function call argument
    take_ptr(alloc Inner { val: 46 })

    // Scenario 6: Return statement
    var ret = return_ptr()

    io.PrintLn("PASS")
}
