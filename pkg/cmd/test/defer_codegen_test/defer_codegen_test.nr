package main

import "io"

pub type MockResource = struct {
    id: i32
    closed: bool
}

pub fn (r: &MockResource) Close() {
    io.PrintLn("Closing resource id: ${r.id}")
    r.closed = true
}

// 1. Basic LIFO execution ordering
fn test_defer_lifo() {
    io.PrintLn("--- test_defer_lifo ---")
    io.PrintLn("First")
    defer io.PrintLn("Defer 1")
    defer io.PrintLn("Defer 2")
    defer io.PrintLn("Defer 3")
    io.PrintLn("Last")
}

// 2. Early return vs late return cleanup
fn test_defer_early_exit(cond: bool) i32 {
    io.PrintLn("--- test_defer_early_exit(${cond}) ---")
    defer io.PrintLn("Always cleaned up")
    if cond {
        io.PrintLn("Early return")
        return 10
    }
    io.PrintLn("Late return")
    return 20
}

// 3. Defer with RAII allocated struct method call (Canonical resource management)
fn test_defer_struct_method() {
    io.PrintLn("--- test_defer_struct_method ---")
    var res = alloc MockResource { id: 101, closed: false }
    defer res.Close()
    io.PrintLn("Using resource id: ${res.id}")
}

// 4. Defer inside a loop accumulating calls until function exit
fn test_defer_in_loop() {
    io.PrintLn("--- test_defer_in_loop ---")
    var i = 0
    while i < 3 {
        defer io.PrintLn("Loop defer step")
        i = i + 1
    }
    io.PrintLn("Loop finished, exiting function")
}

// 5. Defer inside nested conditional blocks
fn test_defer_nested_blocks(flag: bool) {
    io.PrintLn("--- test_defer_nested_blocks(${flag}) ---")
    if flag {
        defer io.PrintLn("Nested if branch defer")
        io.PrintLn("Inside if branch")
    }
    io.PrintLn("Exiting nested blocks function")
}

// 6. Defer observing variable state changes before function exit
fn test_defer_state_mutation() {
    io.PrintLn("--- test_defer_state_mutation ---")
    var counter = 1
    defer io.PrintLn("Deferred counter check: ${counter}")
    counter = counter + 99
    io.PrintLn("Counter mutated to: ${counter}")
}

fn main() i32 {
    test_defer_lifo()
    var r1 = test_defer_early_exit(true)
    var r2 = test_defer_early_exit(false)
    test_defer_struct_method()
    test_defer_in_loop()
    test_defer_nested_blocks(true)
    test_defer_state_mutation()

    if r1 == 10 && r2 == 20 {
        io.PrintLn("ALL DEFER SCENARIOS PASSED")
    }
    return 0
}
