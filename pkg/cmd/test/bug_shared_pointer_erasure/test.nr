pub type MyOption[T] = type {
    MySome(val: T),
    MyNone
}

pub type A = struct { val: i32 }
pub type B = struct { val: f32 }

pub fn main() {
    var a = alloc A { val: 5 }
    var _opt_a = MyOption[#A].MySome(a)

    var b = alloc B { val: f32(3.14) }
    var _opt_b = MyOption[#B].MySome(b)
}
