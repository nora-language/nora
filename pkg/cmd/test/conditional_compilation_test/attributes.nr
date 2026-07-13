package conditional_compilation_test

[cfg("os=windows")]
pub fn call_windows_specific() {
}

[cfg("os=linux")]
pub fn call_windows_specific() {
}
