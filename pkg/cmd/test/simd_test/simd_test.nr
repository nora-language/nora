package main

import "std/simd"
import "std/io"



var g_dummy: ptr

fn black_hole(p: ptr) {
    g_dummy = p
}

fn main() i32 {
    // 1. Test Vec4d (f64)
    var v0 = simd.VecSet4(1.0, 2.0, 3.0, 4.0)
    var v1 = simd.VecSet1(10.0)
    var v_add = v0 + v1
    var v_mul = v0 * v1
    
    // Allocate space and store the result to print it
    var arr = alloc f64[4]
    simd.VecStore(simd.VecSlicePtr(arr, 0), v_add)
    io.PrintLn("Vec4d Add Result: ${arr[0]}, ${arr[1]}, ${arr[2]}, ${arr[3]}")
    
    black_hole(&v_add)
    black_hole(&v_mul)
    
    // 2. Test Vec8f (f32)
    var v8_f0 = simd.Set1_8f(f32(1.5))
    var v8_f1 = simd.Set1_8f(f32(2.5))
    var v8_add = v8_f0 + v8_f1
    var v8_mul = v8_f0 * v8_f1
    black_hole(&v8_add)
    black_hole(&v8_mul)

    // 3. Test Vec4f (f32)
    var v4_f0 = simd.Set1_4f(f32(100.0))
    var v4_f1 = simd.Set1_4f(f32(20.0))
    var v4_div = v4_f0 / v4_f1
    black_hole(&v4_div)

    // 4. Test Vec8i (i32)
    var v8_i0 = simd.Set1_8i(5)
    var v8_i1 = simd.Set1_8i(10)
    var v8_i_add = v8_i0 + v8_i1
    black_hole(&v8_i_add)

    // 5. Test Vec4i (i32)
    var v4_i0 = simd.Set1_4i(50)
    var v4_i1 = simd.Set1_4i(2)
    var v4_i_mul = v4_i0 * v4_i1
    black_hole(&v4_i_mul)

    // 6. Test Vec32i8 (i8 256-bit)
    var v32_i8_0 = simd.Set1_32i8(i8(5))
    var v32_i8_1 = simd.Set1_32i8(i8(10))
    var v32_i8_add = v32_i8_0 + v32_i8_1
    black_hole(&v32_i8_add)

    // 7. Test Vec8i16 (i16 128-bit)
    var v8_i16_0 = simd.Set1_8i16(i16(100))
    var v8_i16_1 = simd.Set1_8i16(i16(50))
    var v8_i16_sub = v8_i16_0 - v8_i16_1
    black_hole(&v8_i16_sub)

    // 8. Test Vec4u32 (u32 128-bit)
    var v4_u32_0 = simd.Set1_4u32(u32(300))
    var v4_u32_1 = simd.Set1_4u32(u32(200))
    var v4_u32_add = v4_u32_0 + v4_u32_1
    black_hole(&v4_u32_add)

    // 9. Test Vec2i64 (i64 128-bit)
    var v2_i64_0 = simd.Set1_2i64(i64(5000))
    var v2_i64_1 = simd.Set1_2i64(i64(2))
    var v2_i64_mul = v2_i64_0 * v2_i64_1
    black_hole(&v2_i64_mul)

    // 10. Test smaller types
    var v16_i8 = simd.Set1_16i8(i8(1))
    black_hole(&v16_i8)

    // Phase 3: Comparisons & Blending
    var cmp_a = simd.VecSet4(1.0, 5.0, 3.0, 8.0)
    var cmp_b = simd.VecSet4(2.0, 5.0, 1.0, 8.0)
    var mask_eq = simd.CmpEq4d(cmp_a, cmp_b)
    var mask_lt = simd.CmpLt4d(cmp_a, cmp_b)
    var blended = simd.Blend4d(mask_eq, cmp_a, simd.VecSet1(0.0)) // Keep elements of A if equal, else 0
    
    var arr2 = alloc f64[4]
    simd.VecStore(simd.VecSlicePtr(arr2, 0), blended)
    io.PrintLn("Vec4d Blend (A == B ? A : 0.0): ${arr2[0]}, ${arr2[1]}, ${arr2[2]}, ${arr2[3]}")

    // Phase 4: Memory Operations (Gather & Scatter)
    var mem_data = alloc f32[8]
    mem_data[0] = f32(10.0); mem_data[1] = f32(20.0); mem_data[2] = f32(30.0); mem_data[3] = f32(40.0); 
    mem_data[4] = f32(50.0); mem_data[5] = f32(60.0); mem_data[6] = f32(70.0); mem_data[7] = f32(80.0);

    var v_indices = simd.VecSet8i(1, 3, 5, 7, 0, 2, 4, 6)
    var mem_ptr = @ptr(mem_data)
    var gathered = simd.Gather8f(mem_ptr, v_indices)

    var out_arr = alloc f32[8]
    simd.StoreUnaligned8f(@ptr(out_arr), gathered)
    io.PrintLn("Gather8f Result: ${f64(out_arr[0])}, ${f64(out_arr[1])}, ${f64(out_arr[2])}, ${f64(out_arr[3])}")

    var v_scatter_indices = simd.VecSet8i(0, 2, 4, 6, 1, 3, 5, 7)
    simd.Scatter8f(mem_ptr, v_scatter_indices, gathered)
    
    io.PrintLn("Scatter8f verification:")
    io.PrintLn(" mem_data[0]=${f64(mem_data[0])}, mem_data[2]=${f64(mem_data[2])}, mem_data[4]=${f64(mem_data[4])}, mem_data[6]=${f64(mem_data[6])}")

    // Phase 5: Shuffles, Swizzles & Permutations
    var perm_in = simd.VecSet8f(f32(1.0), f32(2.0), f32(3.0), f32(4.0), f32(5.0), f32(6.0), f32(7.0), f32(8.0))
    var perm_idx = simd.VecSet8i(7, 6, 5, 4, 3, 2, 1, 0)
    var perm_out = simd.Permute8f(perm_in, perm_idx)
    var out_perm_arr = alloc f32[8]
    simd.StoreUnaligned8f(@ptr(out_perm_arr), perm_out)
    io.PrintLn("Permute8f (Reverse): ${f64(out_perm_arr[0])}, ${f64(out_perm_arr[1])}, ${f64(out_perm_arr[2])}, ${f64(out_perm_arr[7])}")

    var unpack_a = simd.VecSet8f(f32(1.0), f32(1.0), f32(1.0), f32(1.0), f32(5.0), f32(5.0), f32(5.0), f32(5.0))
    var unpack_b = simd.VecSet8f(f32(2.0), f32(2.0), f32(2.0), f32(2.0), f32(6.0), f32(6.0), f32(6.0), f32(6.0))
    var unpack_res = simd.UnpackLo8f(unpack_a, unpack_b)
    var out_upk_arr = alloc f32[8]
    simd.StoreUnaligned8f(@ptr(out_upk_arr), unpack_res)
    io.PrintLn("UnpackLo8f: ${f64(out_upk_arr[0])}, ${f64(out_upk_arr[1])}, ${f64(out_upk_arr[2])}, ${f64(out_upk_arr[3])}")


    // Phase 6: Conversions & Horizontal Reductions
    var float_vec = simd.VecSet8f(f32(1.5), f32(2.7), f32(-3.2), f32(4.9), f32(5.1), f32(6.8), f32(7.3), f32(-8.9))
    var int_vec = simd.Cvt8fTo8i(float_vec)
    var out_int_arr = alloc i32[8]
    simd.Store8i(@ptr(out_int_arr), int_vec)
    io.PrintLn("Cvt8fTo8i: ${out_int_arr[0]}, ${out_int_arr[1]}, ${out_int_arr[2]}, ${out_int_arr[3]}, ${out_int_arr[4]}, ${out_int_arr[5]}, ${out_int_arr[6]}, ${out_int_arr[7]}")

    var reduce_add = simd.ReduceAdd8f(float_vec)
    var reduce_max = simd.ReduceMax8f(float_vec)
    io.PrintLn("ReduceAdd8f: ${f64(reduce_add)}")
    io.PrintLn("ReduceMax8f: ${f64(reduce_max)}")
    
    var f4_vec = simd.Set1_4f(f32(1.23))
    var d4_vec = simd.Cvt4fTo4d(f4_vec)
    var reduce_d4 = simd.ReduceAdd4d(d4_vec)
    io.PrintLn("Cvt4fTo4d & ReduceAdd4d: ${reduce_d4}")


    // Phase 7: Transcendental Math
    var math_in = simd.VecSet8f(f32(0.0), f32(1.0), f32(3.14159265), f32(1.57079632), f32(2.7182818), f32(10.0), f32(-1.0), f32(0.5))
    var math_sin = simd.Sin8f(math_in)
    var math_cos = simd.Cos8f(math_in)
    var math_exp = simd.Exp8f(math_in)
    var math_log = simd.Log8f(math_in)

    var out_sin_arr = alloc f32[8]
    simd.StoreUnaligned8f(@ptr(out_sin_arr), math_sin)
    io.PrintLn("Sin8f[PI/2]: ${f64(out_sin_arr[3])}, Sin8f[PI]: ${f64(out_sin_arr[2])}")
    
    var out_cos_arr = alloc f32[8]
    simd.StoreUnaligned8f(@ptr(out_cos_arr), math_cos)
    io.PrintLn("Cos8f[0]: ${f64(out_cos_arr[0])}, Cos8f[PI/2]: ${f64(out_cos_arr[3])}")

    var out_exp_arr = alloc f32[8]
    simd.StoreUnaligned8f(@ptr(out_exp_arr), math_exp)
    io.PrintLn("Exp8f[1]: ${f64(out_exp_arr[1])}, Exp8f[10]: ${f64(out_exp_arr[5])}")

    var out_log_arr = alloc f32[8]
    simd.StoreUnaligned8f(@ptr(out_log_arr), math_log)
    io.PrintLn("Log8f[1]: ${f64(out_log_arr[1])}, Log8f[E]: ${f64(out_log_arr[4])}")

    var _mask_lt = mask_lt // Silence unused variable warning
    return 0
}

