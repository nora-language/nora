package fail_conditional_compilation

pub fn main() i32 {
    wasm_only_func() // This should be undefined on windows and linux
    return 0
}
