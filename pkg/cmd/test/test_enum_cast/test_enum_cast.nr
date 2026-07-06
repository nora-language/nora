package main

[repr("i32")]
pub type MyEnum = enum {
    VariantA = 10,
    VariantB = 20,
}

pub fn main() {
    var int_val: i32 = 20
    var enum_val: MyEnum = MyEnum(int_val) // Should cast i32 to MyEnum
}
