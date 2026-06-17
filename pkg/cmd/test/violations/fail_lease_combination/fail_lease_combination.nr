package main

type User = struct {
    name: str
}

// 1. Two-level invalid combinations
fn invalid_borrow_owner(u: @#User) {}

fn invalid_mut_owner(u: @&User) {}

fn invalid_double_owner(u: @@User) {}

// 2. Three-level invalid combinations
fn invalid_triple_owner(u: @@@User) {}

fn invalid_read_double_owner(u: #@@User) {}

fn invalid_mut_read_owner(u: &@#User) {}

fn main() i32 {
    return 0
}
