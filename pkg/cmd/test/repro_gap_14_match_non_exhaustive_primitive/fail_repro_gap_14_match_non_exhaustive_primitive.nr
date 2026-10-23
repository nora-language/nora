// Gap 2.2b: Non-exhaustive match expression on primitive i32 (without wildcard _) is accepted
// by pkg/semantic/analyzer.go:2585, leaving the result uninitialized in C if no branch matches.
package main

fn classify(x: i32) i32 {
    var res: i32 = match x {
        1 => { 10 }
        2 => { 20 }
    }
    return res
}

fn main() i32 {
    return classify(3)
}
