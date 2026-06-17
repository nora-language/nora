package main

extern fn printf(fmt: ptr, ...) i32
fn main() {
  var s = "Hello"
  for i, ch in s {
    printf(ptr("%c\n"), ch)
  }
}
