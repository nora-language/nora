package main

type Box[T] = struct {
    val: T
}

// Method generic parameter T shadows the struct's generic parameter T.
// This should either be rejected by the compiler, or correctly scoped without crashing.
pub fn (b: #Box[T]) map[T](other: T) {
}

fn main() i32 {
    var b = alloc Box[i32] { val: 5 }
    // We pass str, shadowing the i32 T parameter?
    b.map[str]("hello")
    return 0
}
