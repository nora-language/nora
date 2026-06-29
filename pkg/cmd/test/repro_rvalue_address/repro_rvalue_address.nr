package main

pub type Dummy = struct {
    val: i64
}

pub fn (s: Dummy) eq(other: Dummy) bool {
    return s.val == other.val
}

pub fn (s: Dummy) sub(other: Dummy) Dummy {
    return Dummy { val: s.val - other.val }
}

pub fn Eq[T](a: #T, b: #T) bool {
    return false
}

pub fn Sub[T](a: #T, b: #T) T {
    panic("dummy")
}

pub fn main() i32 {
    var a = Dummy { val: 10 }
    var dt = Dummy { val: 5 }
    
    var clos = fn() Dummy {
        return dt - dt
    }
    
    // This triggers the bug: 
    // Eq[T](#a, #Sub[T](#dt, #dt)) 
    // The C compiler will fail to take the address of the temporary struct returned by Sub.
    if a == clos() {
        return 1
    }
    
    return 0
}
