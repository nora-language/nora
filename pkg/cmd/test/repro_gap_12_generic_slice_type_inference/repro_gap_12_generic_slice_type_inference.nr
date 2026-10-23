// Gap 2.1d: matchType (pkg/semantic/analyzer.go:7620) fails to infer type parameter T from a slice/list parameter T[].
package main

import "os"

fn head_or_zero[T](xs: #T[], fallback: T) T {
    if len(xs) > 0 {
        return xs[0]
    }
    return fallback
}

fn main() i32 {
    var arr: i32[] = [10, 20, 30]
    // Calling head_or_zero without explicit [i32] should infer T = i32 from #T[]
    var v = head_or_zero(#arr, 0)
    if v != 10 {
        os.Exit(1)
    }
    return 0
}
