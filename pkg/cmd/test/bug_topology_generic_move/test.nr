pub type Box[T] = struct { val: T }

pub fn (b: #Box[T]) UseValue(t: T) T {
    var x = t // Moves t
    return t  // Topology error: use of moved value `t` (because T is treated as owned)
}

pub fn main() {
    var b = Box[i32]{ val: 5 }
    var x = b.UseValue(10)
}
