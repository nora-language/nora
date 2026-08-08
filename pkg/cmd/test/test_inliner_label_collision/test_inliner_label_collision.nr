package main

pub fn has_early_return(flag: bool) i32 {
    if flag {
        return 1
    }
    return 0
}

pub fn main() {
    // Calling the function twice triggers multiple inlinings
    // Without resetting the cloner maps, the second inline will reuse the
    // generated labels from the first inline, causing a C compilation error
    // "redefinition of label".
    var a = has_early_return(true)
    var b = has_early_return(false)
    var c = a + b
}
