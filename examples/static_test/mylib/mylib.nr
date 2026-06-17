package mylib

extern fn my_c_function()

pub fn say_hello() {
    my_c_function()
}
