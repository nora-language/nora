import "io"

[inline]
pub fn add_five(x: i32) i32 {
    return x + 5
}

[inline]
pub fn multiply_by_two(x: i32) i32 {
    return x * 2
}

pub fn main() {
    var initial = 10
    var added = add_five(initial)
    var multiplied = multiply_by_two(added)
    
    // 10 + 5 = 15. 15 * 2 = 30.
    if multiplied == 30 {
        io.PrintLn("success")
    } else {
        io.PrintLn("failed")
    }
}
