package main

import "non_existent_package"

pub fn ReturnsMissingType() non_existent_package.SomeType {
    return non_existent_package.SomeType{}
}

pub fn main() {
    ReturnsMissingType()
}
