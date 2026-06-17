package main

import "std/io"
import "core/iter"

pub fn is_even(x: #i32) bool {
    return x % 2 == 0
}

pub fn double_it(x: i32) i32 {
    return x * 2
}


pub fn main() {
    io.PrintLn("--- Iter Combinator Test ---")
    
    // Take
    io.PrintLn("Take 3:")
    var sum_take = 0
    for i in (1..10).Take(3) {
        sum_take = sum_take + i
    }
    if sum_take != 6 {
        panic("Take failed")
    }

    // Filter
    io.PrintLn("Filter evens 1..5:")
    var sum_filter = 0
    for i in (1..6).Filter(is_even) {
        sum_filter = sum_filter + i
    }
    if sum_filter != 6 {
        panic("Filter failed")
    }

    // Map
    io.PrintLn("Map double 1..3:")
    var sum_map = 0
    for i in (1..4).Map(double_it) {
        sum_map = sum_map + i
    }
    if sum_map != 12 {
        panic("Map failed")
    }

    // StepBy
    io.PrintLn("StepBy 2 from 1..6:")
    var sum_step = 0
    for i in (1..6).StepBy(2) {
        sum_step = sum_step + i
    }
    if sum_step != 9 {
        panic("StepBy failed")
    }
    
    // Skip
    io.PrintLn("Skip 2 from 1..5:")
    var sum_skip = 0
    for i in (1..6).Skip(2) {
        sum_skip = sum_skip + i
    }
    // 3 + 4 + 5 = 12
    if sum_skip != 12 {
        panic("Skip failed")
    }

    io.PrintLn("SUCCESS")
}
