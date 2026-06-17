import "io"

pub fn main() i32 {
    var a = alloc i32[5]
    a[0] = 10
    a[1] = 20
    var b: &i32 = &a.unchecked_get(1)
    b = 30
    if a[1] != 30 {
        return 1
    }
    return 0
}
