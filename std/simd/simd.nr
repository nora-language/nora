package simd

[shared]
[vector_size("32")]
pub type Vec4d = struct {
    v0: f64
}

pub fn SetZero() Vec4d {
    return nr_simd_set1(0.0)
}

extern fn nr_simd_load(src: ptr) Vec4d
extern fn nr_simd_store(dst: ptr, src: Vec4d)
extern fn nr_simd_set1(val: f64) Vec4d
extern fn nr_simd_set4(v0: f64, v1: f64, v2: f64, v3: f64) Vec4d
extern fn nr_simd_add(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_sub(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_mul(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_div(a: Vec4d, b: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_hadd(a: Vec4d, b: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_blend_1100(a: Vec4d, b: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_blend_1010(a: Vec4d, b: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_blend_0011(a: Vec4d, b: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_permute_0101(a: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_permute2f128_01(a: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_permute2f128_1(a: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_permute2f128_21(a: Vec4d, b: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_unpacklo(a: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_unpackhi(a: Vec4d) Vec4d
extern fn nr_simd_slice_ptr(slice: #f64[], index: i32) ptr

[inline]
pub fn Load(data: ptr) Vec4d {
    return nr_simd_load(data)
}

[inline]
pub fn SlicePtr(slice: #f64[], index: i32) ptr {
    return nr_simd_slice_ptr(slice, index)
}

[inline]
pub fn Store(data: ptr, vec: Vec4d) {
    nr_simd_store(data, vec)
}

[inline]
pub fn Set1(val: f64) Vec4d {
    return nr_simd_set1(val)
}

[inline]
pub fn Set4(v0: f64, v1: f64, v2: f64, v3: f64) Vec4d {
    return nr_simd_set4(v0, v1, v2, v3)
}

[inline]
pub fn Add(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a + v_b
}

[inline]
pub fn Sub(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a - v_b
}

[inline]
pub fn Mul(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a * v_b
}

[inline]
pub fn Div(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a / v_b
}

[cfg("target_feature=avx")]
[inline]
pub fn Hadd(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_hadd(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn Blend1100(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_blend_1100(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn Blend1010(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_blend_1010(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn Blend0011(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_blend_0011(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn Permute0101(a: Vec4d) Vec4d {
    return nr_simd_permute_0101(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn Permute2F128_01(a: Vec4d) Vec4d {
    return nr_simd_permute2f128_01(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn Permute2F128_1(a: Vec4d) Vec4d {
    return nr_simd_permute2f128_1(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn Permute2F128_21(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_permute2f128_21(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn UnpackLo(a: Vec4d) Vec4d {
    return nr_simd_unpacklo(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn UnpackHi(a: Vec4d) Vec4d {
    return nr_simd_unpackhi(a)
}


