package main

import "os"
import "io"
import "math"
import "strconv"
import "simd"

[unsafe]
[inline]
fn eval_A_vec(v_i: simd.Vec8d, v_j: simd.Vec8d) simd.Vec8d {
    var ONE = simd.Set1_8d(1.0)
    var TWO = simd.Set1_8d(2.0)
    
    var ipj = v_i + v_j
    var ipjp1 = ipj + ONE
    var ip1 = v_i + ONE
    
    var a = (ipj * ipjp1) / TWO + ip1
    
    return a
}

[unsafe]
fn eval_A_times_u_worker(start: i32, end: i32, N: i32, u: #f64[], Au: &f64[]) {
    var i = start
    while (i < end) {
        var vi0 = simd.Set1_8d(f64(i))
        var vi1 = simd.Set1_8d(f64(i+1))
        var vi2 = simd.Set1_8d(f64(i+2))
        var vi3 = simd.Set1_8d(f64(i+3))
        
        var j = 0
        var s0 = simd.Set1_8d(0.0)
        var s1 = simd.Set1_8d(0.0)
        var s2 = simd.Set1_8d(0.0)
        var s3 = simd.Set1_8d(0.0)
        var ONE = simd.Set1_8d(1.0)
        
        while (j < N) {
            var vj = simd.Load8d(simd.VecSlicePtr(u, j))
            
            var j_idx = simd.VecSet8(f64(j), f64(j+1), f64(j+2), f64(j+3), f64(j+4), f64(j+5), f64(j+6), f64(j+7))
            
            var denoms0 = eval_A_vec(vi0, j_idx)
            var denoms1 = eval_A_vec(vi1, j_idx)
            var denoms2 = eval_A_vec(vi2, j_idx)
            var denoms3 = eval_A_vec(vi3, j_idx)
            
            var inv_denoms0 = ONE / denoms0
            var inv_denoms1 = ONE / denoms1
            var inv_denoms2 = ONE / denoms2
            var inv_denoms3 = ONE / denoms3
            
            var terms0 = inv_denoms0 * vj
            var terms1 = inv_denoms1 * vj
            var terms2 = inv_denoms2 * vj
            var terms3 = inv_denoms3 * vj
            
            s0 = s0 + terms0
            s1 = s1 + terms1
            s2 = s2 + terms2
            s3 = s3 + terms3
            
            j = j + 8
        }
        
        // Sum each s[i] horizontally and store
        Au[i] = simd.VecReduceAdd8d(s0)
        Au[i+1] = simd.VecReduceAdd8d(s1)
        Au[i+2] = simd.VecReduceAdd8d(s2)
        Au[i+3] = simd.VecReduceAdd8d(s3)
        
        i = i + 4
    }
}

[unsafe]
fn eval_At_times_u_worker(start: i32, end: i32, N: i32, u: #f64[], Au: &f64[]) {
    var i = start
    while (i < end) {
        var vi0 = simd.Set1_8d(f64(i))
        var vi1 = simd.Set1_8d(f64(i+1))
        var vi2 = simd.Set1_8d(f64(i+2))
        var vi3 = simd.Set1_8d(f64(i+3))
        
        var j = 0
        var s0 = simd.Set1_8d(0.0)
        var s1 = simd.Set1_8d(0.0)
        var s2 = simd.Set1_8d(0.0)
        var s3 = simd.Set1_8d(0.0)
        var ONE = simd.Set1_8d(1.0)
        
        while (j < N) {
            var vj = simd.Load8d(simd.VecSlicePtr(u, j))
            
            var j_idx = simd.VecSet8(f64(j), f64(j+1), f64(j+2), f64(j+3), f64(j+4), f64(j+5), f64(j+6), f64(j+7))
            
            // For At, we swap i and j in eval_A_vec(j, i)
            var denoms0 = eval_A_vec(j_idx, vi0)
            var denoms1 = eval_A_vec(j_idx, vi1)
            var denoms2 = eval_A_vec(j_idx, vi2)
            var denoms3 = eval_A_vec(j_idx, vi3)
            
            var inv_denoms0 = ONE / denoms0
            var inv_denoms1 = ONE / denoms1
            var inv_denoms2 = ONE / denoms2
            var inv_denoms3 = ONE / denoms3
            
            var terms0 = inv_denoms0 * vj
            var terms1 = inv_denoms1 * vj
            var terms2 = inv_denoms2 * vj
            var terms3 = inv_denoms3 * vj
            
            s0 = s0 + terms0
            s1 = s1 + terms1
            s2 = s2 + terms2
            s3 = s3 + terms3
            
            j = j + 8
        }
        
        // Sum each s[i] horizontally and store
        Au[i] = simd.VecReduceAdd8d(s0)
        Au[i+1] = simd.VecReduceAdd8d(s1)
        Au[i+2] = simd.VecReduceAdd8d(s2)
        Au[i+3] = simd.VecReduceAdd8d(s3)
        
        i = i + 4
    }
}

[unsafe]
fn eval_A_times_u(N: i32, u: #f64[], Au: &f64[]) {
    scope {
        var num_fibers = 16
        var chunk = N / num_fibers
        var i = 0
        while (i < num_fibers) {
            var start = i * chunk
            var end = (i + 1) * chunk
            if (i == num_fibers - 1) {
                end = N
            }
            spawn eval_A_times_u_worker(start, end, N, u, Au)
            i = i + 1
        }
    }
}

[unsafe]
fn eval_At_times_u(N: i32, u: #f64[], Au: &f64[]) {
    scope {
        var num_fibers = 16
        var chunk = N / num_fibers
        var i = 0
        while (i < num_fibers) {
            var start = i * chunk
            var end = (i + 1) * chunk
            if (i == num_fibers - 1) {
                end = N
            }
            spawn eval_At_times_u_worker(start, end, N, u, Au)
            i = i + 1
        }
    }
}

[unsafe]
fn eval_AtA_times_u(N: i32, u: #f64[], AtAu: &f64[]) {
    var v = alloc f64[N]
    eval_A_times_u(N, u, &v)
    eval_At_times_u(N, #v, AtAu)
}

fn main() i32 {
    var args = os.Args()
    var n = 5500
    if (args.Len[str]() > 1) {
        var parsed = strconv.Atoi(#args.Get[str](1))
        if (!parsed.IsErr[i32, str]()) {
            n = parsed.Unwrap[i32, str]()
        }
    }

    // Allocate overhang buffer for SIMD to prevent out of bounds loads
    var u = alloc f64[n+8]
    var v = alloc f64[n+8]
    
    var i = 0
    while (i < n) {
        u[i] = 1.0
        i = i + 1
    }

    i = 0
    while (i < 10) {
        eval_AtA_times_u(n, #u, &v)
        eval_AtA_times_u(n, #v, &u)
        i = i + 1
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
