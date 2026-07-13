package fail_conditional_compilation

pub fn main() i32 {
    linux_only_func() // This should be undefined on windows
    return 0
}
