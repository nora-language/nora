package main

import "io"

fn main() i32 {
    var n: i32 = 0
    var k: i32 = 0
    io.Scan(#n, #k)

    var count = 0
    var k_score = 0
    var i = 0

    while (i < n) {
        var score = 0
        io.Scan(#score)

        if (i < k) {
            if (score > 0) {
                count++
                if (i == k - 1) {
                    k_score = score
                }
            }
        } else {
            // scores are non-increasing, so score >= k_score implies score == k_score or better
            // but since i >= k, score <= k_score.
            // thus score >= k_score is only possible if score == k_score.
            if (score == k_score && score > 0) {
                count++
            }
        }
        i++
    }

    io.PrintLn(count)
    return 0
}
