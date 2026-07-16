package main

import "os"
import "io"
import "strconv"
import "math"

fn eval_A(i: i32, j: i32) f64 {
    var fi = f64(i)
    var fj = f64(j)
    return 1.0 / (((fi + fj) * (fi + fj + 1.0) / 2.0) + fi + 1.0)
}

fn eval_A_times_u(N: i32, u: #f64[], Au: &f64[]) {
    var i = 0
    while (i < N) {
        Au[i] = 0.0
        var j = 0
        while (j < N) {
            Au[i] = Au[i] + eval_A(i, j) * u[j]
            j = j + 1
        }
        i = i + 1
    }
}

fn eval_At_times_u(N: i32, u: #f64[], Au: &f64[]) {
    var i = 0
    while (i < N) {
        Au[i] = 0.0
        var j = 0
        while (j < N) {
            Au[i] = Au[i] + eval_A(j, i) * u[j]
            j = j + 1
        }
        i = i + 1
    }
}

fn eval_AtA_times_u(N: i32, u: #f64[], AtAu: &f64[]) {
    var v = alloc f64[N]
    eval_A_times_u(N, u, &v)
    eval_At_times_u(N, #v, AtAu)
}

fn main() i32 {
    var n = 5500
    var args = os.Args()
    if (args.Len[str]() > 1) {
        var parsed = strconv.Atoi(#args.Get[str](1))
        if (!parsed.IsErr()) {
            n = parsed.Unwrap()
        }
    }

    var u = alloc f64[n]
    var v = alloc f64[n]
    
    var i = 0
    while (i < n) {
        u[i] = 1.0
        i = i + 1
    }

    var k = 0
    while (k < 10) {
        eval_AtA_times_u(n, #u, &v)
        eval_AtA_times_u(n, #v, &u)
        k = k + 1
    }

    var vBv = 0.0
    var vv = 0.0
    var j = 0
    while (j < n) {
        vBv = vBv + u[j] * v[j]
        vv = vv + v[j] * v[j]
        j = j + 1
    }

    var res = math.Sqrt(vBv / vv)
    var res_str = strconv.FormatFloat(res, "f", 9)
    io.PrintLn(res_str)

    return 0
}
