import "io"

pub fn main() i32 {
    var a = alloc i32[5]
    a[0] = 10
    a[1] = 20
    var b = a.unchecked_get(1)
    if b != 20 {
        return 1
    }
    return 0
}
