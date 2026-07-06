pub fn main() i32 {
    // This is a void-like body but the function returns i32.
    // The compiler should throw a "missing return at end of function" error.
    var x = 10
    var y = x * 2
}

pub fn test_match(x: i32) i32 {
    match x {
        1 => {
            return 1
        }
        _ => {
            // Missing return in this branch
        }
    }
}
