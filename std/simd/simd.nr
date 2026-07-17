package simd

[shared]
[vector_size("32")]
pub type Vec4d = struct {
    v0: f64
}

[shared]
[vector_size("32")]
pub type Vec8f = struct {
    v0: f32
}

[shared]
[vector_size("16")]
pub type Vec4f = struct {
    v0: f32
}

[shared]
[vector_size("32")]
pub type Vec8i = struct {
    v0: i32
}

[shared]
[vector_size("16")]
pub type Vec4i = struct {
    v0: i32
}


// --- GENERATED STRUCTS ---
[shared]
[vector_size("16")]
pub type Vec16i8 = struct { v0: i8 }

[shared]
[vector_size("32")]
pub type Vec32i8 = struct { v0: i8 }

[shared]
[vector_size("16")]
pub type Vec16u8 = struct { v0: u8 }

[shared]
[vector_size("32")]
pub type Vec32u8 = struct { v0: u8 }

[shared]
[vector_size("16")]
pub type Vec8i16 = struct { v0: i16 }

[shared]
[vector_size("32")]
pub type Vec16i16 = struct { v0: i16 }

[shared]
[vector_size("16")]
pub type Vec8u16 = struct { v0: u16 }

[shared]
[vector_size("32")]
pub type Vec16u16 = struct { v0: u16 }

[shared]
[vector_size("16")]
pub type Vec4u32 = struct { v0: u32 }

[shared]
[vector_size("32")]
pub type Vec8u32 = struct { v0: u32 }

[shared]
[vector_size("16")]
pub type Vec2i64 = struct { v0: i64 }

[shared]
[vector_size("32")]
pub type Vec4i64 = struct { v0: i64 }

[shared]
[vector_size("16")]
pub type Vec2u64 = struct { v0: u64 }

[shared]
[vector_size("32")]
pub type Vec4u64 = struct { v0: u64 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec16f = struct { v0: f32 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec8d = struct { v0: f64 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec16i32 = struct { v0: i32 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec16u32 = struct { v0: u32 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec8i64 = struct { v0: i64 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec8u64 = struct { v0: u64 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec64i8 = struct { v0: i8 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec64u8 = struct { v0: u8 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec32i16 = struct { v0: i16 }

[cfg("target_feature=avx512f")]
[shared]
[vector_size("64")]
pub type Vec32u16 = struct { v0: u16 }



pub fn VecSetZero() Vec4d {
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
pub fn VecLoad(data: ptr) Vec4d {
    return nr_simd_load(data)
}

[inline]
pub fn VecSlicePtr(slice: #f64[], index: i32) ptr {
    return nr_simd_slice_ptr(slice, index)
}

[inline]
pub fn VecStore(data: ptr, vec: Vec4d) {
    nr_simd_store(data, vec)
}

[inline]
pub fn VecSet1(val: f64) Vec4d {
    return nr_simd_set1(val)
}

[inline]
pub fn VecSet4(v0: f64, v1: f64, v2: f64, v3: f64) Vec4d {
    return nr_simd_set4(v0, v1, v2, v3)
}

[cfg("target_feature=avx512f")] extern fn nr_simd_set8(v0: f64, v1: f64, v2: f64, v3: f64, v4: f64, v5: f64, v6: f64, v7: f64) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_setzero8d() Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_sqrt8d(a: Vec8d) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_min8d(a: Vec8d, b: Vec8d) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_max8d(a: Vec8d, b: Vec8d) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_and8d(a: Vec8d, b: Vec8d) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_or8d(a: Vec8d, b: Vec8d) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_xor8d(a: Vec8d, b: Vec8d) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_andnot8d(a: Vec8d, b: Vec8d) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_reduce_add8d(a: Vec8d) f64

[cfg("target_feature=avx512f")]
[inline]
pub fn VecSet8(v0: f64, v1: f64, v2: f64, v3: f64, v4: f64, v5: f64, v6: f64, v7: f64) Vec8d {
    return nr_simd_set8(v0, v1, v2, v3, v4, v5, v6, v7)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecSetZero8d() Vec8d {
    return nr_simd_setzero8d()
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecSqrt8d(a: Vec8d) Vec8d {
    return nr_simd_sqrt8d(a)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecMin8d(a: Vec8d, b: Vec8d) Vec8d {
    return nr_simd_min8d(a, b)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecMax8d(a: Vec8d, b: Vec8d) Vec8d {
    return nr_simd_max8d(a, b)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecAnd8d(a: Vec8d, b: Vec8d) Vec8d {
    return nr_simd_and8d(a, b)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecOr8d(a: Vec8d, b: Vec8d) Vec8d {
    return nr_simd_or8d(a, b)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecXor8d(a: Vec8d, b: Vec8d) Vec8d {
    return nr_simd_xor8d(a, b)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecAndNot8d(a: Vec8d, b: Vec8d) Vec8d {
    return nr_simd_andnot8d(a, b)
}
[cfg("target_feature=avx512f")]
[inline]
pub fn VecReduceAdd8d(a: Vec8d) f64 {
    return nr_simd_reduce_add8d(a)
}


[inline]
pub fn VecAdd(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a + v_b
}

[inline]
pub fn VecSub(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a - v_b
}

[inline]
pub fn VecMul(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a * v_b
}

[inline]
pub fn VecDiv(v_a: Vec4d, v_b: Vec4d) Vec4d {
    return v_a / v_b
}

[cfg("target_feature=avx")]
[inline]
pub fn VecHadd(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_hadd(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecBlend1100(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_blend_1100(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecBlend1010(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_blend_1010(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecBlend0011(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_blend_0011(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecPermute0101(a: Vec4d) Vec4d {
    return nr_simd_permute_0101(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecPermute2F128_01(a: Vec4d) Vec4d {
    return nr_simd_permute2f128_01(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecPermute2F128_1(a: Vec4d) Vec4d {
    return nr_simd_permute2f128_1(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecPermute2F128_21(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_permute2f128_21(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecUnpackLo(a: Vec4d) Vec4d {
    return nr_simd_unpacklo(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecUnpackHi(a: Vec4d) Vec4d {
    return nr_simd_unpackhi(a)
}

extern fn nr_simd_approx_recip(a: Vec4d) Vec4d
extern fn nr_simd_sqrt(a: Vec4d) Vec4d
extern fn nr_simd_min(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_max(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_and(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_or(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_xor(a: Vec4d, b: Vec4d) Vec4d
extern fn nr_simd_andnot(a: Vec4d, b: Vec4d) Vec4d

[cfg("target_feature=avx")]
[inline]
pub fn VecApproxRecip(a: Vec4d) Vec4d {
    return nr_simd_approx_recip(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecSqrt(a: Vec4d) Vec4d {
    return nr_simd_sqrt(a)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecMin(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_min(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecMax(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_max(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecAnd(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_and(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecOr(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_or(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecXor(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_xor(a, b)
}

[cfg("target_feature=avx")]
[inline]
pub fn VecAndNot(a: Vec4d, b: Vec4d) Vec4d {
    return nr_simd_andnot(a, b)
}

extern fn nr_simd_slice_ptr_f32(slice: #f32[], index: i32) ptr

// Vec8f Functions
extern fn nr_simd_load_f32x8(src: ptr) Vec8f
extern fn nr_simd_store_f32x8(dst: ptr, src: Vec8f)
extern fn nr_simd_set1_f32x8(val: f32) Vec8f

[inline] pub fn SlicePtrF32(slice: #f32[], index: i32) ptr { return nr_simd_slice_ptr_f32(slice, index) }
[inline] pub fn Load8f(data: ptr) Vec8f { return nr_simd_load_f32x8(data) }
[inline] pub fn Store8f(data: ptr, vec: Vec8f) { nr_simd_store_f32x8(data, vec) }
[inline] pub fn Set1_8f(val: f32) Vec8f { return nr_simd_set1_f32x8(val) }
[inline] pub fn VecAdd8f(a: Vec8f, b: Vec8f) Vec8f { return a + b }
[inline] pub fn VecSub8f(a: Vec8f, b: Vec8f) Vec8f { return a - b }
[inline] pub fn VecMul8f(a: Vec8f, b: Vec8f) Vec8f { return a * b }
[inline] pub fn VecDiv8f(a: Vec8f, b: Vec8f) Vec8f { return a / b }

// Vec4f Functions
extern fn nr_simd_load_f32x4(src: ptr) Vec4f
extern fn nr_simd_store_f32x4(dst: ptr, src: Vec4f)
extern fn nr_simd_set1_f32x4(val: f32) Vec4f

[inline] pub fn Load4f(data: ptr) Vec4f { return nr_simd_load_f32x4(data) }
[inline] pub fn Store4f(data: ptr, vec: Vec4f) { nr_simd_store_f32x4(data, vec) }
[inline] pub fn Set1_4f(val: f32) Vec4f { return nr_simd_set1_f32x4(val) }
[inline] pub fn VecAdd4f(a: Vec4f, b: Vec4f) Vec4f { return a + b }
[inline] pub fn VecSub4f(a: Vec4f, b: Vec4f) Vec4f { return a - b }
[inline] pub fn VecMul4f(a: Vec4f, b: Vec4f) Vec4f { return a * b }
[inline] pub fn VecDiv4f(a: Vec4f, b: Vec4f) Vec4f { return a / b }

extern fn nr_simd_slice_ptr_i32(slice: #i32[], index: i32) ptr

// Vec8i Functions
extern fn nr_simd_load_i32x8(src: ptr) Vec8i
extern fn nr_simd_store_i32x8(dst: ptr, src: Vec8i)
extern fn nr_simd_set1_i32x8(val: i32) Vec8i

[inline] pub fn SlicePtrI32(slice: #i32[], index: i32) ptr { return nr_simd_slice_ptr_i32(slice, index) }
[inline] pub fn Load8i(data: ptr) Vec8i { return nr_simd_load_i32x8(data) }
[inline] pub fn Store8i(data: ptr, vec: Vec8i) { nr_simd_store_i32x8(data, vec) }
[inline] pub fn Set1_8i(val: i32) Vec8i { return nr_simd_set1_i32x8(val) }
[inline] pub fn VecAdd8i(a: Vec8i, b: Vec8i) Vec8i { return a + b }
[inline] pub fn VecSub8i(a: Vec8i, b: Vec8i) Vec8i { return a - b }
[inline] pub fn VecMul8i(a: Vec8i, b: Vec8i) Vec8i { return a * b }

// Vec4i Functions
extern fn nr_simd_load_i32x4(src: ptr) Vec4i
extern fn nr_simd_store_i32x4(dst: ptr, src: Vec4i)
extern fn nr_simd_set1_i32x4(val: i32) Vec4i

[inline] pub fn Load4i(data: ptr) Vec4i { return nr_simd_load_i32x4(data) }
[inline] pub fn Store4i(data: ptr, vec: Vec4i) { nr_simd_store_i32x4(data, vec) }
[inline] pub fn Set1_4i(val: i32) Vec4i { return nr_simd_set1_i32x4(val) }
[inline] pub fn VecAdd4i(a: Vec4i, b: Vec4i) Vec4i { return a + b }
[inline] pub fn VecSub4i(a: Vec4i, b: Vec4i) Vec4i { return a - b }
[inline] pub fn VecMul4i(a: Vec4i, b: Vec4i) Vec4i { return a * b }

// --- GENERATED WRAPPERS ---

extern fn nr_simd_load_Vec16i8(src: ptr) Vec16i8
extern fn nr_simd_store_Vec16i8(dst: ptr, src: Vec16i8)
extern fn nr_simd_set1_Vec16i8(val: i8) Vec16i8

[inline] pub fn Load16i8(data: ptr) Vec16i8 { return nr_simd_load_Vec16i8(data) }
[inline] pub fn Store16i8(data: ptr, vec: Vec16i8) { nr_simd_store_Vec16i8(data, vec) }
[inline] pub fn Set1_16i8(val: i8) Vec16i8 { return nr_simd_set1_Vec16i8(val) }
[inline] pub fn VecAdd16i8(a: Vec16i8, b: Vec16i8) Vec16i8 { return a + b }
[inline] pub fn VecSub16i8(a: Vec16i8, b: Vec16i8) Vec16i8 { return a - b }
[inline] pub fn VecMul16i8(a: Vec16i8, b: Vec16i8) Vec16i8 { return a * b }

extern fn nr_simd_load_Vec32i8(src: ptr) Vec32i8
extern fn nr_simd_store_Vec32i8(dst: ptr, src: Vec32i8)
extern fn nr_simd_set1_Vec32i8(val: i8) Vec32i8

[inline] pub fn Load32i8(data: ptr) Vec32i8 { return nr_simd_load_Vec32i8(data) }
[inline] pub fn Store32i8(data: ptr, vec: Vec32i8) { nr_simd_store_Vec32i8(data, vec) }
[inline] pub fn Set1_32i8(val: i8) Vec32i8 { return nr_simd_set1_Vec32i8(val) }
[inline] pub fn VecAdd32i8(a: Vec32i8, b: Vec32i8) Vec32i8 { return a + b }
[inline] pub fn VecSub32i8(a: Vec32i8, b: Vec32i8) Vec32i8 { return a - b }
[inline] pub fn VecMul32i8(a: Vec32i8, b: Vec32i8) Vec32i8 { return a * b }

extern fn nr_simd_load_Vec16u8(src: ptr) Vec16u8
extern fn nr_simd_store_Vec16u8(dst: ptr, src: Vec16u8)
extern fn nr_simd_set1_Vec16u8(val: u8) Vec16u8

[inline] pub fn Load16u8(data: ptr) Vec16u8 { return nr_simd_load_Vec16u8(data) }
[inline] pub fn Store16u8(data: ptr, vec: Vec16u8) { nr_simd_store_Vec16u8(data, vec) }
[inline] pub fn Set1_16u8(val: u8) Vec16u8 { return nr_simd_set1_Vec16u8(val) }
[inline] pub fn VecAdd16u8(a: Vec16u8, b: Vec16u8) Vec16u8 { return a + b }
[inline] pub fn VecSub16u8(a: Vec16u8, b: Vec16u8) Vec16u8 { return a - b }
[inline] pub fn VecMul16u8(a: Vec16u8, b: Vec16u8) Vec16u8 { return a * b }

extern fn nr_simd_load_Vec32u8(src: ptr) Vec32u8
extern fn nr_simd_store_Vec32u8(dst: ptr, src: Vec32u8)
extern fn nr_simd_set1_Vec32u8(val: u8) Vec32u8

[inline] pub fn Load32u8(data: ptr) Vec32u8 { return nr_simd_load_Vec32u8(data) }
[inline] pub fn Store32u8(data: ptr, vec: Vec32u8) { nr_simd_store_Vec32u8(data, vec) }
[inline] pub fn Set1_32u8(val: u8) Vec32u8 { return nr_simd_set1_Vec32u8(val) }
[inline] pub fn VecAdd32u8(a: Vec32u8, b: Vec32u8) Vec32u8 { return a + b }
[inline] pub fn VecSub32u8(a: Vec32u8, b: Vec32u8) Vec32u8 { return a - b }
[inline] pub fn VecMul32u8(a: Vec32u8, b: Vec32u8) Vec32u8 { return a * b }

extern fn nr_simd_load_Vec8i16(src: ptr) Vec8i16
extern fn nr_simd_store_Vec8i16(dst: ptr, src: Vec8i16)
extern fn nr_simd_set1_Vec8i16(val: i16) Vec8i16

[inline] pub fn Load8i16(data: ptr) Vec8i16 { return nr_simd_load_Vec8i16(data) }
[inline] pub fn Store8i16(data: ptr, vec: Vec8i16) { nr_simd_store_Vec8i16(data, vec) }
[inline] pub fn Set1_8i16(val: i16) Vec8i16 { return nr_simd_set1_Vec8i16(val) }
[inline] pub fn VecAdd8i16(a: Vec8i16, b: Vec8i16) Vec8i16 { return a + b }
[inline] pub fn VecSub8i16(a: Vec8i16, b: Vec8i16) Vec8i16 { return a - b }
[inline] pub fn VecMul8i16(a: Vec8i16, b: Vec8i16) Vec8i16 { return a * b }

extern fn nr_simd_load_Vec16i16(src: ptr) Vec16i16
extern fn nr_simd_store_Vec16i16(dst: ptr, src: Vec16i16)
extern fn nr_simd_set1_Vec16i16(val: i16) Vec16i16

[inline] pub fn Load16i16(data: ptr) Vec16i16 { return nr_simd_load_Vec16i16(data) }
[inline] pub fn Store16i16(data: ptr, vec: Vec16i16) { nr_simd_store_Vec16i16(data, vec) }
[inline] pub fn Set1_16i16(val: i16) Vec16i16 { return nr_simd_set1_Vec16i16(val) }
[inline] pub fn VecAdd16i16(a: Vec16i16, b: Vec16i16) Vec16i16 { return a + b }
[inline] pub fn VecSub16i16(a: Vec16i16, b: Vec16i16) Vec16i16 { return a - b }
[inline] pub fn VecMul16i16(a: Vec16i16, b: Vec16i16) Vec16i16 { return a * b }

extern fn nr_simd_load_Vec8u16(src: ptr) Vec8u16
extern fn nr_simd_store_Vec8u16(dst: ptr, src: Vec8u16)
extern fn nr_simd_set1_Vec8u16(val: u16) Vec8u16

[inline] pub fn Load8u16(data: ptr) Vec8u16 { return nr_simd_load_Vec8u16(data) }
[inline] pub fn Store8u16(data: ptr, vec: Vec8u16) { nr_simd_store_Vec8u16(data, vec) }
[inline] pub fn Set1_8u16(val: u16) Vec8u16 { return nr_simd_set1_Vec8u16(val) }
[inline] pub fn VecAdd8u16(a: Vec8u16, b: Vec8u16) Vec8u16 { return a + b }
[inline] pub fn VecSub8u16(a: Vec8u16, b: Vec8u16) Vec8u16 { return a - b }
[inline] pub fn VecMul8u16(a: Vec8u16, b: Vec8u16) Vec8u16 { return a * b }

extern fn nr_simd_load_Vec16u16(src: ptr) Vec16u16
extern fn nr_simd_store_Vec16u16(dst: ptr, src: Vec16u16)
extern fn nr_simd_set1_Vec16u16(val: u16) Vec16u16

[inline] pub fn Load16u16(data: ptr) Vec16u16 { return nr_simd_load_Vec16u16(data) }
[inline] pub fn Store16u16(data: ptr, vec: Vec16u16) { nr_simd_store_Vec16u16(data, vec) }
[inline] pub fn Set1_16u16(val: u16) Vec16u16 { return nr_simd_set1_Vec16u16(val) }
[inline] pub fn VecAdd16u16(a: Vec16u16, b: Vec16u16) Vec16u16 { return a + b }
[inline] pub fn VecSub16u16(a: Vec16u16, b: Vec16u16) Vec16u16 { return a - b }
[inline] pub fn VecMul16u16(a: Vec16u16, b: Vec16u16) Vec16u16 { return a * b }

extern fn nr_simd_load_Vec4u32(src: ptr) Vec4u32
extern fn nr_simd_store_Vec4u32(dst: ptr, src: Vec4u32)
extern fn nr_simd_set1_Vec4u32(val: u32) Vec4u32

[inline] pub fn Load4u32(data: ptr) Vec4u32 { return nr_simd_load_Vec4u32(data) }
[inline] pub fn Store4u32(data: ptr, vec: Vec4u32) { nr_simd_store_Vec4u32(data, vec) }
[inline] pub fn Set1_4u32(val: u32) Vec4u32 { return nr_simd_set1_Vec4u32(val) }
[inline] pub fn VecAdd4u32(a: Vec4u32, b: Vec4u32) Vec4u32 { return a + b }
[inline] pub fn VecSub4u32(a: Vec4u32, b: Vec4u32) Vec4u32 { return a - b }
[inline] pub fn VecMul4u32(a: Vec4u32, b: Vec4u32) Vec4u32 { return a * b }

extern fn nr_simd_load_Vec8u32(src: ptr) Vec8u32
extern fn nr_simd_store_Vec8u32(dst: ptr, src: Vec8u32)
extern fn nr_simd_set1_Vec8u32(val: u32) Vec8u32

[inline] pub fn Load8u32(data: ptr) Vec8u32 { return nr_simd_load_Vec8u32(data) }
[inline] pub fn Store8u32(data: ptr, vec: Vec8u32) { nr_simd_store_Vec8u32(data, vec) }
[inline] pub fn Set1_8u32(val: u32) Vec8u32 { return nr_simd_set1_Vec8u32(val) }
[inline] pub fn VecAdd8u32(a: Vec8u32, b: Vec8u32) Vec8u32 { return a + b }
[inline] pub fn VecSub8u32(a: Vec8u32, b: Vec8u32) Vec8u32 { return a - b }
[inline] pub fn VecMul8u32(a: Vec8u32, b: Vec8u32) Vec8u32 { return a * b }

extern fn nr_simd_load_Vec2i64(src: ptr) Vec2i64
extern fn nr_simd_store_Vec2i64(dst: ptr, src: Vec2i64)
extern fn nr_simd_set1_Vec2i64(val: i64) Vec2i64

[inline] pub fn Load2i64(data: ptr) Vec2i64 { return nr_simd_load_Vec2i64(data) }
[inline] pub fn Store2i64(data: ptr, vec: Vec2i64) { nr_simd_store_Vec2i64(data, vec) }
[inline] pub fn Set1_2i64(val: i64) Vec2i64 { return nr_simd_set1_Vec2i64(val) }
[inline] pub fn VecAdd2i64(a: Vec2i64, b: Vec2i64) Vec2i64 { return a + b }
[inline] pub fn VecSub2i64(a: Vec2i64, b: Vec2i64) Vec2i64 { return a - b }
[inline] pub fn VecMul2i64(a: Vec2i64, b: Vec2i64) Vec2i64 { return a * b }

extern fn nr_simd_load_Vec4i64(src: ptr) Vec4i64
extern fn nr_simd_store_Vec4i64(dst: ptr, src: Vec4i64)
extern fn nr_simd_set1_Vec4i64(val: i64) Vec4i64

[inline] pub fn Load4i64(data: ptr) Vec4i64 { return nr_simd_load_Vec4i64(data) }
[inline] pub fn Store4i64(data: ptr, vec: Vec4i64) { nr_simd_store_Vec4i64(data, vec) }
[inline] pub fn Set1_4i64(val: i64) Vec4i64 { return nr_simd_set1_Vec4i64(val) }
[inline] pub fn VecAdd4i64(a: Vec4i64, b: Vec4i64) Vec4i64 { return a + b }
[inline] pub fn VecSub4i64(a: Vec4i64, b: Vec4i64) Vec4i64 { return a - b }
[inline] pub fn VecMul4i64(a: Vec4i64, b: Vec4i64) Vec4i64 { return a * b }

extern fn nr_simd_load_Vec2u64(src: ptr) Vec2u64
extern fn nr_simd_store_Vec2u64(dst: ptr, src: Vec2u64)
extern fn nr_simd_set1_Vec2u64(val: u64) Vec2u64

[inline] pub fn Load2u64(data: ptr) Vec2u64 { return nr_simd_load_Vec2u64(data) }
[inline] pub fn Store2u64(data: ptr, vec: Vec2u64) { nr_simd_store_Vec2u64(data, vec) }
[inline] pub fn Set1_2u64(val: u64) Vec2u64 { return nr_simd_set1_Vec2u64(val) }
[inline] pub fn VecAdd2u64(a: Vec2u64, b: Vec2u64) Vec2u64 { return a + b }
[inline] pub fn VecSub2u64(a: Vec2u64, b: Vec2u64) Vec2u64 { return a - b }
[inline] pub fn VecMul2u64(a: Vec2u64, b: Vec2u64) Vec2u64 { return a * b }

extern fn nr_simd_load_Vec4u64(src: ptr) Vec4u64
extern fn nr_simd_store_Vec4u64(dst: ptr, src: Vec4u64)
extern fn nr_simd_set1_Vec4u64(val: u64) Vec4u64

[inline] pub fn Load4u64(data: ptr) Vec4u64 { return nr_simd_load_Vec4u64(data) }
[inline] pub fn Store4u64(data: ptr, vec: Vec4u64) { nr_simd_store_Vec4u64(data, vec) }
[inline] pub fn Set1_4u64(val: u64) Vec4u64 { return nr_simd_set1_Vec4u64(val) }
[inline] pub fn VecAdd4u64(a: Vec4u64, b: Vec4u64) Vec4u64 { return a + b }
[inline] pub fn VecSub4u64(a: Vec4u64, b: Vec4u64) Vec4u64 { return a - b }
[inline] pub fn VecMul4u64(a: Vec4u64, b: Vec4u64) Vec4u64 { return a * b }

// --- AVX-512 WRAPPERS ---

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec16f(src: ptr) Vec16f
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec16f(dst: ptr, src: Vec16f)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec16f(val: f32) Vec16f

[cfg("target_feature=avx512f")] [inline] pub fn Load16f(data: ptr) Vec16f { return nr_simd_load_Vec16f(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store16f(data: ptr, vec: Vec16f) { nr_simd_store_Vec16f(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_16f(val: f32) Vec16f { return nr_simd_set1_Vec16f(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd16f(a: Vec16f, b: Vec16f) Vec16f { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub16f(a: Vec16f, b: Vec16f) Vec16f { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul16f(a: Vec16f, b: Vec16f) Vec16f { return a * b }
[cfg("target_feature=avx512f")] [inline] pub fn VecDiv16f(a: Vec16f, b: Vec16f) Vec16f { return a / b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec8d(src: ptr) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec8d(dst: ptr, src: Vec8d)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec8d(val: f64) Vec8d

[cfg("target_feature=avx512f")] [inline] pub fn Load8d(data: ptr) Vec8d { return nr_simd_load_Vec8d(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store8d(data: ptr, vec: Vec8d) { nr_simd_store_Vec8d(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_8d(val: f64) Vec8d { return nr_simd_set1_Vec8d(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd8d(a: Vec8d, b: Vec8d) Vec8d { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub8d(a: Vec8d, b: Vec8d) Vec8d { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul8d(a: Vec8d, b: Vec8d) Vec8d { return a * b }
[cfg("target_feature=avx512f")] [inline] pub fn VecDiv8d(a: Vec8d, b: Vec8d) Vec8d { return a / b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec16i32(src: ptr) Vec16i32
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec16i32(dst: ptr, src: Vec16i32)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec16i32(val: i32) Vec16i32

[cfg("target_feature=avx512f")] [inline] pub fn Load16i32(data: ptr) Vec16i32 { return nr_simd_load_Vec16i32(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store16i32(data: ptr, vec: Vec16i32) { nr_simd_store_Vec16i32(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_16i32(val: i32) Vec16i32 { return nr_simd_set1_Vec16i32(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd16i32(a: Vec16i32, b: Vec16i32) Vec16i32 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub16i32(a: Vec16i32, b: Vec16i32) Vec16i32 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul16i32(a: Vec16i32, b: Vec16i32) Vec16i32 { return a * b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec16u32(src: ptr) Vec16u32
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec16u32(dst: ptr, src: Vec16u32)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec16u32(val: u32) Vec16u32

[cfg("target_feature=avx512f")] [inline] pub fn Load16u32(data: ptr) Vec16u32 { return nr_simd_load_Vec16u32(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store16u32(data: ptr, vec: Vec16u32) { nr_simd_store_Vec16u32(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_16u32(val: u32) Vec16u32 { return nr_simd_set1_Vec16u32(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd16u32(a: Vec16u32, b: Vec16u32) Vec16u32 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub16u32(a: Vec16u32, b: Vec16u32) Vec16u32 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul16u32(a: Vec16u32, b: Vec16u32) Vec16u32 { return a * b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec8i64(src: ptr) Vec8i64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec8i64(dst: ptr, src: Vec8i64)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec8i64(val: i64) Vec8i64

[cfg("target_feature=avx512f")] [inline] pub fn Load8i64(data: ptr) Vec8i64 { return nr_simd_load_Vec8i64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store8i64(data: ptr, vec: Vec8i64) { nr_simd_store_Vec8i64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_8i64(val: i64) Vec8i64 { return nr_simd_set1_Vec8i64(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd8i64(a: Vec8i64, b: Vec8i64) Vec8i64 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub8i64(a: Vec8i64, b: Vec8i64) Vec8i64 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul8i64(a: Vec8i64, b: Vec8i64) Vec8i64 { return a * b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec8u64(src: ptr) Vec8u64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec8u64(dst: ptr, src: Vec8u64)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec8u64(val: u64) Vec8u64

[cfg("target_feature=avx512f")] [inline] pub fn Load8u64(data: ptr) Vec8u64 { return nr_simd_load_Vec8u64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store8u64(data: ptr, vec: Vec8u64) { nr_simd_store_Vec8u64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_8u64(val: u64) Vec8u64 { return nr_simd_set1_Vec8u64(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd8u64(a: Vec8u64, b: Vec8u64) Vec8u64 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub8u64(a: Vec8u64, b: Vec8u64) Vec8u64 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul8u64(a: Vec8u64, b: Vec8u64) Vec8u64 { return a * b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec64i8(src: ptr) Vec64i8
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec64i8(dst: ptr, src: Vec64i8)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec64i8(val: i8) Vec64i8

[cfg("target_feature=avx512f")] [inline] pub fn Load64i8(data: ptr) Vec64i8 { return nr_simd_load_Vec64i8(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store64i8(data: ptr, vec: Vec64i8) { nr_simd_store_Vec64i8(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_64i8(val: i8) Vec64i8 { return nr_simd_set1_Vec64i8(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd64i8(a: Vec64i8, b: Vec64i8) Vec64i8 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub64i8(a: Vec64i8, b: Vec64i8) Vec64i8 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul64i8(a: Vec64i8, b: Vec64i8) Vec64i8 { return a * b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec64u8(src: ptr) Vec64u8
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec64u8(dst: ptr, src: Vec64u8)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec64u8(val: u8) Vec64u8

[cfg("target_feature=avx512f")] [inline] pub fn Load64u8(data: ptr) Vec64u8 { return nr_simd_load_Vec64u8(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store64u8(data: ptr, vec: Vec64u8) { nr_simd_store_Vec64u8(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_64u8(val: u8) Vec64u8 { return nr_simd_set1_Vec64u8(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd64u8(a: Vec64u8, b: Vec64u8) Vec64u8 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub64u8(a: Vec64u8, b: Vec64u8) Vec64u8 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul64u8(a: Vec64u8, b: Vec64u8) Vec64u8 { return a * b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec32i16(src: ptr) Vec32i16
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec32i16(dst: ptr, src: Vec32i16)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec32i16(val: i16) Vec32i16

[cfg("target_feature=avx512f")] [inline] pub fn Load32i16(data: ptr) Vec32i16 { return nr_simd_load_Vec32i16(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store32i16(data: ptr, vec: Vec32i16) { nr_simd_store_Vec32i16(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_32i16(val: i16) Vec32i16 { return nr_simd_set1_Vec32i16(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd32i16(a: Vec32i16, b: Vec32i16) Vec32i16 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub32i16(a: Vec32i16, b: Vec32i16) Vec32i16 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul32i16(a: Vec32i16, b: Vec32i16) Vec32i16 { return a * b }

[cfg("target_feature=avx512f")] extern fn nr_simd_load_Vec32u16(src: ptr) Vec32u16
[cfg("target_feature=avx512f")] extern fn nr_simd_store_Vec32u16(dst: ptr, src: Vec32u16)
[cfg("target_feature=avx512f")] extern fn nr_simd_set1_Vec32u16(val: u16) Vec32u16

[cfg("target_feature=avx512f")] [inline] pub fn Load32u16(data: ptr) Vec32u16 { return nr_simd_load_Vec32u16(data) }
[cfg("target_feature=avx512f")] [inline] pub fn Store32u16(data: ptr, vec: Vec32u16) { nr_simd_store_Vec32u16(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Set1_32u16(val: u16) Vec32u16 { return nr_simd_set1_Vec32u16(val) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAdd32u16(a: Vec32u16, b: Vec32u16) Vec32u16 { return a + b }
[cfg("target_feature=avx512f")] [inline] pub fn VecSub32u16(a: Vec32u16, b: Vec32u16) Vec32u16 { return a - b }
[cfg("target_feature=avx512f")] [inline] pub fn VecMul32u16(a: Vec32u16, b: Vec32u16) Vec32u16 { return a * b }

// --- PHASE 3: COMPARISONS & BLEND ---
extern fn nr_simd_cmpeq_Vec16i8(a: Vec16i8, b: Vec16i8) Vec16i8
[inline] pub fn CmpEq16i8(a: Vec16i8, b: Vec16i8) Vec16i8 { return nr_simd_cmpeq_Vec16i8(a, b) }
extern fn nr_simd_cmpneq_Vec16i8(a: Vec16i8, b: Vec16i8) Vec16i8
[inline] pub fn CmpNeq16i8(a: Vec16i8, b: Vec16i8) Vec16i8 { return nr_simd_cmpneq_Vec16i8(a, b) }
extern fn nr_simd_cmplt_Vec16i8(a: Vec16i8, b: Vec16i8) Vec16i8
[inline] pub fn CmpLt16i8(a: Vec16i8, b: Vec16i8) Vec16i8 { return nr_simd_cmplt_Vec16i8(a, b) }
extern fn nr_simd_cmpgt_Vec16i8(a: Vec16i8, b: Vec16i8) Vec16i8
[inline] pub fn CmpGt16i8(a: Vec16i8, b: Vec16i8) Vec16i8 { return nr_simd_cmpgt_Vec16i8(a, b) }
extern fn nr_simd_blend_Vec16i8(mask: Vec16i8, a: Vec16i8, b: Vec16i8) Vec16i8
[inline] pub fn Blend16i8(mask: Vec16i8, a: Vec16i8, b: Vec16i8) Vec16i8 { return nr_simd_blend_Vec16i8(mask, a, b) }
extern fn nr_simd_cmpeq_Vec32i8(a: Vec32i8, b: Vec32i8) Vec32i8
[inline] pub fn CmpEq32i8(a: Vec32i8, b: Vec32i8) Vec32i8 { return nr_simd_cmpeq_Vec32i8(a, b) }
extern fn nr_simd_cmpneq_Vec32i8(a: Vec32i8, b: Vec32i8) Vec32i8
[inline] pub fn CmpNeq32i8(a: Vec32i8, b: Vec32i8) Vec32i8 { return nr_simd_cmpneq_Vec32i8(a, b) }
extern fn nr_simd_cmplt_Vec32i8(a: Vec32i8, b: Vec32i8) Vec32i8
[inline] pub fn CmpLt32i8(a: Vec32i8, b: Vec32i8) Vec32i8 { return nr_simd_cmplt_Vec32i8(a, b) }
extern fn nr_simd_cmpgt_Vec32i8(a: Vec32i8, b: Vec32i8) Vec32i8
[inline] pub fn CmpGt32i8(a: Vec32i8, b: Vec32i8) Vec32i8 { return nr_simd_cmpgt_Vec32i8(a, b) }
extern fn nr_simd_blend_Vec32i8(mask: Vec32i8, a: Vec32i8, b: Vec32i8) Vec32i8
[inline] pub fn Blend32i8(mask: Vec32i8, a: Vec32i8, b: Vec32i8) Vec32i8 { return nr_simd_blend_Vec32i8(mask, a, b) }
extern fn nr_simd_cmpeq_Vec16u8(a: Vec16u8, b: Vec16u8) Vec16i8
[inline] pub fn CmpEq16u8(a: Vec16u8, b: Vec16u8) Vec16i8 { return nr_simd_cmpeq_Vec16u8(a, b) }
extern fn nr_simd_cmpneq_Vec16u8(a: Vec16u8, b: Vec16u8) Vec16i8
[inline] pub fn CmpNeq16u8(a: Vec16u8, b: Vec16u8) Vec16i8 { return nr_simd_cmpneq_Vec16u8(a, b) }
extern fn nr_simd_cmplt_Vec16u8(a: Vec16u8, b: Vec16u8) Vec16i8
[inline] pub fn CmpLt16u8(a: Vec16u8, b: Vec16u8) Vec16i8 { return nr_simd_cmplt_Vec16u8(a, b) }
extern fn nr_simd_cmpgt_Vec16u8(a: Vec16u8, b: Vec16u8) Vec16i8
[inline] pub fn CmpGt16u8(a: Vec16u8, b: Vec16u8) Vec16i8 { return nr_simd_cmpgt_Vec16u8(a, b) }
extern fn nr_simd_blend_Vec16u8(mask: Vec16i8, a: Vec16u8, b: Vec16u8) Vec16u8
[inline] pub fn Blend16u8(mask: Vec16i8, a: Vec16u8, b: Vec16u8) Vec16u8 { return nr_simd_blend_Vec16u8(mask, a, b) }
extern fn nr_simd_cmpeq_Vec32u8(a: Vec32u8, b: Vec32u8) Vec32i8
[inline] pub fn CmpEq32u8(a: Vec32u8, b: Vec32u8) Vec32i8 { return nr_simd_cmpeq_Vec32u8(a, b) }
extern fn nr_simd_cmpneq_Vec32u8(a: Vec32u8, b: Vec32u8) Vec32i8
[inline] pub fn CmpNeq32u8(a: Vec32u8, b: Vec32u8) Vec32i8 { return nr_simd_cmpneq_Vec32u8(a, b) }
extern fn nr_simd_cmplt_Vec32u8(a: Vec32u8, b: Vec32u8) Vec32i8
[inline] pub fn CmpLt32u8(a: Vec32u8, b: Vec32u8) Vec32i8 { return nr_simd_cmplt_Vec32u8(a, b) }
extern fn nr_simd_cmpgt_Vec32u8(a: Vec32u8, b: Vec32u8) Vec32i8
[inline] pub fn CmpGt32u8(a: Vec32u8, b: Vec32u8) Vec32i8 { return nr_simd_cmpgt_Vec32u8(a, b) }
extern fn nr_simd_blend_Vec32u8(mask: Vec32i8, a: Vec32u8, b: Vec32u8) Vec32u8
[inline] pub fn Blend32u8(mask: Vec32i8, a: Vec32u8, b: Vec32u8) Vec32u8 { return nr_simd_blend_Vec32u8(mask, a, b) }
extern fn nr_simd_cmpeq_Vec8i16(a: Vec8i16, b: Vec8i16) Vec8i16
[inline] pub fn CmpEq8i16(a: Vec8i16, b: Vec8i16) Vec8i16 { return nr_simd_cmpeq_Vec8i16(a, b) }
extern fn nr_simd_cmpneq_Vec8i16(a: Vec8i16, b: Vec8i16) Vec8i16
[inline] pub fn CmpNeq8i16(a: Vec8i16, b: Vec8i16) Vec8i16 { return nr_simd_cmpneq_Vec8i16(a, b) }
extern fn nr_simd_cmplt_Vec8i16(a: Vec8i16, b: Vec8i16) Vec8i16
[inline] pub fn CmpLt8i16(a: Vec8i16, b: Vec8i16) Vec8i16 { return nr_simd_cmplt_Vec8i16(a, b) }
extern fn nr_simd_cmpgt_Vec8i16(a: Vec8i16, b: Vec8i16) Vec8i16
[inline] pub fn CmpGt8i16(a: Vec8i16, b: Vec8i16) Vec8i16 { return nr_simd_cmpgt_Vec8i16(a, b) }
extern fn nr_simd_blend_Vec8i16(mask: Vec8i16, a: Vec8i16, b: Vec8i16) Vec8i16
[inline] pub fn Blend8i16(mask: Vec8i16, a: Vec8i16, b: Vec8i16) Vec8i16 { return nr_simd_blend_Vec8i16(mask, a, b) }
extern fn nr_simd_cmpeq_Vec16i16(a: Vec16i16, b: Vec16i16) Vec16i16
[inline] pub fn CmpEq16i16(a: Vec16i16, b: Vec16i16) Vec16i16 { return nr_simd_cmpeq_Vec16i16(a, b) }
extern fn nr_simd_cmpneq_Vec16i16(a: Vec16i16, b: Vec16i16) Vec16i16
[inline] pub fn CmpNeq16i16(a: Vec16i16, b: Vec16i16) Vec16i16 { return nr_simd_cmpneq_Vec16i16(a, b) }
extern fn nr_simd_cmplt_Vec16i16(a: Vec16i16, b: Vec16i16) Vec16i16
[inline] pub fn CmpLt16i16(a: Vec16i16, b: Vec16i16) Vec16i16 { return nr_simd_cmplt_Vec16i16(a, b) }
extern fn nr_simd_cmpgt_Vec16i16(a: Vec16i16, b: Vec16i16) Vec16i16
[inline] pub fn CmpGt16i16(a: Vec16i16, b: Vec16i16) Vec16i16 { return nr_simd_cmpgt_Vec16i16(a, b) }
extern fn nr_simd_blend_Vec16i16(mask: Vec16i16, a: Vec16i16, b: Vec16i16) Vec16i16
[inline] pub fn Blend16i16(mask: Vec16i16, a: Vec16i16, b: Vec16i16) Vec16i16 { return nr_simd_blend_Vec16i16(mask, a, b) }
extern fn nr_simd_cmpeq_Vec8u16(a: Vec8u16, b: Vec8u16) Vec8i16
[inline] pub fn CmpEq8u16(a: Vec8u16, b: Vec8u16) Vec8i16 { return nr_simd_cmpeq_Vec8u16(a, b) }
extern fn nr_simd_cmpneq_Vec8u16(a: Vec8u16, b: Vec8u16) Vec8i16
[inline] pub fn CmpNeq8u16(a: Vec8u16, b: Vec8u16) Vec8i16 { return nr_simd_cmpneq_Vec8u16(a, b) }
extern fn nr_simd_cmplt_Vec8u16(a: Vec8u16, b: Vec8u16) Vec8i16
[inline] pub fn CmpLt8u16(a: Vec8u16, b: Vec8u16) Vec8i16 { return nr_simd_cmplt_Vec8u16(a, b) }
extern fn nr_simd_cmpgt_Vec8u16(a: Vec8u16, b: Vec8u16) Vec8i16
[inline] pub fn CmpGt8u16(a: Vec8u16, b: Vec8u16) Vec8i16 { return nr_simd_cmpgt_Vec8u16(a, b) }
extern fn nr_simd_blend_Vec8u16(mask: Vec8i16, a: Vec8u16, b: Vec8u16) Vec8u16
[inline] pub fn Blend8u16(mask: Vec8i16, a: Vec8u16, b: Vec8u16) Vec8u16 { return nr_simd_blend_Vec8u16(mask, a, b) }
extern fn nr_simd_cmpeq_Vec16u16(a: Vec16u16, b: Vec16u16) Vec16i16
[inline] pub fn CmpEq16u16(a: Vec16u16, b: Vec16u16) Vec16i16 { return nr_simd_cmpeq_Vec16u16(a, b) }
extern fn nr_simd_cmpneq_Vec16u16(a: Vec16u16, b: Vec16u16) Vec16i16
[inline] pub fn CmpNeq16u16(a: Vec16u16, b: Vec16u16) Vec16i16 { return nr_simd_cmpneq_Vec16u16(a, b) }
extern fn nr_simd_cmplt_Vec16u16(a: Vec16u16, b: Vec16u16) Vec16i16
[inline] pub fn CmpLt16u16(a: Vec16u16, b: Vec16u16) Vec16i16 { return nr_simd_cmplt_Vec16u16(a, b) }
extern fn nr_simd_cmpgt_Vec16u16(a: Vec16u16, b: Vec16u16) Vec16i16
[inline] pub fn CmpGt16u16(a: Vec16u16, b: Vec16u16) Vec16i16 { return nr_simd_cmpgt_Vec16u16(a, b) }
extern fn nr_simd_blend_Vec16u16(mask: Vec16i16, a: Vec16u16, b: Vec16u16) Vec16u16
[inline] pub fn Blend16u16(mask: Vec16i16, a: Vec16u16, b: Vec16u16) Vec16u16 { return nr_simd_blend_Vec16u16(mask, a, b) }
extern fn nr_simd_cmpeq_Vec4u32(a: Vec4u32, b: Vec4u32) Vec4i
[inline] pub fn CmpEq4u32(a: Vec4u32, b: Vec4u32) Vec4i { return nr_simd_cmpeq_Vec4u32(a, b) }
extern fn nr_simd_cmpneq_Vec4u32(a: Vec4u32, b: Vec4u32) Vec4i
[inline] pub fn CmpNeq4u32(a: Vec4u32, b: Vec4u32) Vec4i { return nr_simd_cmpneq_Vec4u32(a, b) }
extern fn nr_simd_cmplt_Vec4u32(a: Vec4u32, b: Vec4u32) Vec4i
[inline] pub fn CmpLt4u32(a: Vec4u32, b: Vec4u32) Vec4i { return nr_simd_cmplt_Vec4u32(a, b) }
extern fn nr_simd_cmpgt_Vec4u32(a: Vec4u32, b: Vec4u32) Vec4i
[inline] pub fn CmpGt4u32(a: Vec4u32, b: Vec4u32) Vec4i { return nr_simd_cmpgt_Vec4u32(a, b) }
extern fn nr_simd_blend_Vec4u32(mask: Vec4i, a: Vec4u32, b: Vec4u32) Vec4u32
[inline] pub fn Blend4u32(mask: Vec4i, a: Vec4u32, b: Vec4u32) Vec4u32 { return nr_simd_blend_Vec4u32(mask, a, b) }
extern fn nr_simd_cmpeq_Vec8u32(a: Vec8u32, b: Vec8u32) Vec8i
[inline] pub fn CmpEq8u32(a: Vec8u32, b: Vec8u32) Vec8i { return nr_simd_cmpeq_Vec8u32(a, b) }
extern fn nr_simd_cmpneq_Vec8u32(a: Vec8u32, b: Vec8u32) Vec8i
[inline] pub fn CmpNeq8u32(a: Vec8u32, b: Vec8u32) Vec8i { return nr_simd_cmpneq_Vec8u32(a, b) }
extern fn nr_simd_cmplt_Vec8u32(a: Vec8u32, b: Vec8u32) Vec8i
[inline] pub fn CmpLt8u32(a: Vec8u32, b: Vec8u32) Vec8i { return nr_simd_cmplt_Vec8u32(a, b) }
extern fn nr_simd_cmpgt_Vec8u32(a: Vec8u32, b: Vec8u32) Vec8i
[inline] pub fn CmpGt8u32(a: Vec8u32, b: Vec8u32) Vec8i { return nr_simd_cmpgt_Vec8u32(a, b) }
extern fn nr_simd_blend_Vec8u32(mask: Vec8i, a: Vec8u32, b: Vec8u32) Vec8u32
[inline] pub fn Blend8u32(mask: Vec8i, a: Vec8u32, b: Vec8u32) Vec8u32 { return nr_simd_blend_Vec8u32(mask, a, b) }
extern fn nr_simd_cmpeq_Vec2i64(a: Vec2i64, b: Vec2i64) Vec2i64
[inline] pub fn CmpEq2i64(a: Vec2i64, b: Vec2i64) Vec2i64 { return nr_simd_cmpeq_Vec2i64(a, b) }
extern fn nr_simd_cmpneq_Vec2i64(a: Vec2i64, b: Vec2i64) Vec2i64
[inline] pub fn CmpNeq2i64(a: Vec2i64, b: Vec2i64) Vec2i64 { return nr_simd_cmpneq_Vec2i64(a, b) }
extern fn nr_simd_cmplt_Vec2i64(a: Vec2i64, b: Vec2i64) Vec2i64
[inline] pub fn CmpLt2i64(a: Vec2i64, b: Vec2i64) Vec2i64 { return nr_simd_cmplt_Vec2i64(a, b) }
extern fn nr_simd_cmpgt_Vec2i64(a: Vec2i64, b: Vec2i64) Vec2i64
[inline] pub fn CmpGt2i64(a: Vec2i64, b: Vec2i64) Vec2i64 { return nr_simd_cmpgt_Vec2i64(a, b) }
extern fn nr_simd_blend_Vec2i64(mask: Vec2i64, a: Vec2i64, b: Vec2i64) Vec2i64
[inline] pub fn Blend2i64(mask: Vec2i64, a: Vec2i64, b: Vec2i64) Vec2i64 { return nr_simd_blend_Vec2i64(mask, a, b) }
extern fn nr_simd_cmpeq_Vec4i64(a: Vec4i64, b: Vec4i64) Vec4i64
[inline] pub fn CmpEq4i64(a: Vec4i64, b: Vec4i64) Vec4i64 { return nr_simd_cmpeq_Vec4i64(a, b) }
extern fn nr_simd_cmpneq_Vec4i64(a: Vec4i64, b: Vec4i64) Vec4i64
[inline] pub fn CmpNeq4i64(a: Vec4i64, b: Vec4i64) Vec4i64 { return nr_simd_cmpneq_Vec4i64(a, b) }
extern fn nr_simd_cmplt_Vec4i64(a: Vec4i64, b: Vec4i64) Vec4i64
[inline] pub fn CmpLt4i64(a: Vec4i64, b: Vec4i64) Vec4i64 { return nr_simd_cmplt_Vec4i64(a, b) }
extern fn nr_simd_cmpgt_Vec4i64(a: Vec4i64, b: Vec4i64) Vec4i64
[inline] pub fn CmpGt4i64(a: Vec4i64, b: Vec4i64) Vec4i64 { return nr_simd_cmpgt_Vec4i64(a, b) }
extern fn nr_simd_blend_Vec4i64(mask: Vec4i64, a: Vec4i64, b: Vec4i64) Vec4i64
[inline] pub fn Blend4i64(mask: Vec4i64, a: Vec4i64, b: Vec4i64) Vec4i64 { return nr_simd_blend_Vec4i64(mask, a, b) }
extern fn nr_simd_cmpeq_Vec2u64(a: Vec2u64, b: Vec2u64) Vec2i64
[inline] pub fn CmpEq2u64(a: Vec2u64, b: Vec2u64) Vec2i64 { return nr_simd_cmpeq_Vec2u64(a, b) }
extern fn nr_simd_cmpneq_Vec2u64(a: Vec2u64, b: Vec2u64) Vec2i64
[inline] pub fn CmpNeq2u64(a: Vec2u64, b: Vec2u64) Vec2i64 { return nr_simd_cmpneq_Vec2u64(a, b) }
extern fn nr_simd_cmplt_Vec2u64(a: Vec2u64, b: Vec2u64) Vec2i64
[inline] pub fn CmpLt2u64(a: Vec2u64, b: Vec2u64) Vec2i64 { return nr_simd_cmplt_Vec2u64(a, b) }
extern fn nr_simd_cmpgt_Vec2u64(a: Vec2u64, b: Vec2u64) Vec2i64
[inline] pub fn CmpGt2u64(a: Vec2u64, b: Vec2u64) Vec2i64 { return nr_simd_cmpgt_Vec2u64(a, b) }
extern fn nr_simd_blend_Vec2u64(mask: Vec2i64, a: Vec2u64, b: Vec2u64) Vec2u64
[inline] pub fn Blend2u64(mask: Vec2i64, a: Vec2u64, b: Vec2u64) Vec2u64 { return nr_simd_blend_Vec2u64(mask, a, b) }
extern fn nr_simd_cmpeq_Vec4u64(a: Vec4u64, b: Vec4u64) Vec4i64
[inline] pub fn CmpEq4u64(a: Vec4u64, b: Vec4u64) Vec4i64 { return nr_simd_cmpeq_Vec4u64(a, b) }
extern fn nr_simd_cmpneq_Vec4u64(a: Vec4u64, b: Vec4u64) Vec4i64
[inline] pub fn CmpNeq4u64(a: Vec4u64, b: Vec4u64) Vec4i64 { return nr_simd_cmpneq_Vec4u64(a, b) }
extern fn nr_simd_cmplt_Vec4u64(a: Vec4u64, b: Vec4u64) Vec4i64
[inline] pub fn CmpLt4u64(a: Vec4u64, b: Vec4u64) Vec4i64 { return nr_simd_cmplt_Vec4u64(a, b) }
extern fn nr_simd_cmpgt_Vec4u64(a: Vec4u64, b: Vec4u64) Vec4i64
[inline] pub fn CmpGt4u64(a: Vec4u64, b: Vec4u64) Vec4i64 { return nr_simd_cmpgt_Vec4u64(a, b) }
extern fn nr_simd_blend_Vec4u64(mask: Vec4i64, a: Vec4u64, b: Vec4u64) Vec4u64
[inline] pub fn Blend4u64(mask: Vec4i64, a: Vec4u64, b: Vec4u64) Vec4u64 { return nr_simd_blend_Vec4u64(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec16f(a: Vec16f, b: Vec16f) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq16f(a: Vec16f, b: Vec16f) Vec16i32 { return nr_simd_cmpeq_Vec16f(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec16f(a: Vec16f, b: Vec16f) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq16f(a: Vec16f, b: Vec16f) Vec16i32 { return nr_simd_cmpneq_Vec16f(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec16f(a: Vec16f, b: Vec16f) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt16f(a: Vec16f, b: Vec16f) Vec16i32 { return nr_simd_cmplt_Vec16f(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec16f(a: Vec16f, b: Vec16f) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt16f(a: Vec16f, b: Vec16f) Vec16i32 { return nr_simd_cmpgt_Vec16f(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec16f(mask: Vec16i32, a: Vec16f, b: Vec16f) Vec16f
[cfg("target_feature=avx512f")] [inline] pub fn Blend16f(mask: Vec16i32, a: Vec16f, b: Vec16f) Vec16f { return nr_simd_blend_Vec16f(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec8d(a: Vec8d, b: Vec8d) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq8d(a: Vec8d, b: Vec8d) Vec8i64 { return nr_simd_cmpeq_Vec8d(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec8d(a: Vec8d, b: Vec8d) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq8d(a: Vec8d, b: Vec8d) Vec8i64 { return nr_simd_cmpneq_Vec8d(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec8d(a: Vec8d, b: Vec8d) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt8d(a: Vec8d, b: Vec8d) Vec8i64 { return nr_simd_cmplt_Vec8d(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec8d(a: Vec8d, b: Vec8d) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt8d(a: Vec8d, b: Vec8d) Vec8i64 { return nr_simd_cmpgt_Vec8d(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec8d(mask: Vec8i64, a: Vec8d, b: Vec8d) Vec8d
[cfg("target_feature=avx512f")] [inline] pub fn Blend8d(mask: Vec8i64, a: Vec8d, b: Vec8d) Vec8d { return nr_simd_blend_Vec8d(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec16i32(a: Vec16i32, b: Vec16i32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq16i32(a: Vec16i32, b: Vec16i32) Vec16i32 { return nr_simd_cmpeq_Vec16i32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec16i32(a: Vec16i32, b: Vec16i32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq16i32(a: Vec16i32, b: Vec16i32) Vec16i32 { return nr_simd_cmpneq_Vec16i32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec16i32(a: Vec16i32, b: Vec16i32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt16i32(a: Vec16i32, b: Vec16i32) Vec16i32 { return nr_simd_cmplt_Vec16i32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec16i32(a: Vec16i32, b: Vec16i32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt16i32(a: Vec16i32, b: Vec16i32) Vec16i32 { return nr_simd_cmpgt_Vec16i32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec16i32(mask: Vec16i32, a: Vec16i32, b: Vec16i32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn Blend16i32(mask: Vec16i32, a: Vec16i32, b: Vec16i32) Vec16i32 { return nr_simd_blend_Vec16i32(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec16u32(a: Vec16u32, b: Vec16u32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq16u32(a: Vec16u32, b: Vec16u32) Vec16i32 { return nr_simd_cmpeq_Vec16u32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec16u32(a: Vec16u32, b: Vec16u32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq16u32(a: Vec16u32, b: Vec16u32) Vec16i32 { return nr_simd_cmpneq_Vec16u32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec16u32(a: Vec16u32, b: Vec16u32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt16u32(a: Vec16u32, b: Vec16u32) Vec16i32 { return nr_simd_cmplt_Vec16u32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec16u32(a: Vec16u32, b: Vec16u32) Vec16i32
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt16u32(a: Vec16u32, b: Vec16u32) Vec16i32 { return nr_simd_cmpgt_Vec16u32(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec16u32(mask: Vec16i32, a: Vec16u32, b: Vec16u32) Vec16u32
[cfg("target_feature=avx512f")] [inline] pub fn Blend16u32(mask: Vec16i32, a: Vec16u32, b: Vec16u32) Vec16u32 { return nr_simd_blend_Vec16u32(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec8i64(a: Vec8i64, b: Vec8i64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq8i64(a: Vec8i64, b: Vec8i64) Vec8i64 { return nr_simd_cmpeq_Vec8i64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec8i64(a: Vec8i64, b: Vec8i64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq8i64(a: Vec8i64, b: Vec8i64) Vec8i64 { return nr_simd_cmpneq_Vec8i64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec8i64(a: Vec8i64, b: Vec8i64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt8i64(a: Vec8i64, b: Vec8i64) Vec8i64 { return nr_simd_cmplt_Vec8i64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec8i64(a: Vec8i64, b: Vec8i64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt8i64(a: Vec8i64, b: Vec8i64) Vec8i64 { return nr_simd_cmpgt_Vec8i64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec8i64(mask: Vec8i64, a: Vec8i64, b: Vec8i64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn Blend8i64(mask: Vec8i64, a: Vec8i64, b: Vec8i64) Vec8i64 { return nr_simd_blend_Vec8i64(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec8u64(a: Vec8u64, b: Vec8u64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq8u64(a: Vec8u64, b: Vec8u64) Vec8i64 { return nr_simd_cmpeq_Vec8u64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec8u64(a: Vec8u64, b: Vec8u64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq8u64(a: Vec8u64, b: Vec8u64) Vec8i64 { return nr_simd_cmpneq_Vec8u64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec8u64(a: Vec8u64, b: Vec8u64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt8u64(a: Vec8u64, b: Vec8u64) Vec8i64 { return nr_simd_cmplt_Vec8u64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec8u64(a: Vec8u64, b: Vec8u64) Vec8i64
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt8u64(a: Vec8u64, b: Vec8u64) Vec8i64 { return nr_simd_cmpgt_Vec8u64(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec8u64(mask: Vec8i64, a: Vec8u64, b: Vec8u64) Vec8u64
[cfg("target_feature=avx512f")] [inline] pub fn Blend8u64(mask: Vec8i64, a: Vec8u64, b: Vec8u64) Vec8u64 { return nr_simd_blend_Vec8u64(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec64i8(a: Vec64i8, b: Vec64i8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq64i8(a: Vec64i8, b: Vec64i8) Vec64i8 { return nr_simd_cmpeq_Vec64i8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec64i8(a: Vec64i8, b: Vec64i8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq64i8(a: Vec64i8, b: Vec64i8) Vec64i8 { return nr_simd_cmpneq_Vec64i8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec64i8(a: Vec64i8, b: Vec64i8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt64i8(a: Vec64i8, b: Vec64i8) Vec64i8 { return nr_simd_cmplt_Vec64i8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec64i8(a: Vec64i8, b: Vec64i8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt64i8(a: Vec64i8, b: Vec64i8) Vec64i8 { return nr_simd_cmpgt_Vec64i8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec64i8(mask: Vec64i8, a: Vec64i8, b: Vec64i8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn Blend64i8(mask: Vec64i8, a: Vec64i8, b: Vec64i8) Vec64i8 { return nr_simd_blend_Vec64i8(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec64u8(a: Vec64u8, b: Vec64u8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq64u8(a: Vec64u8, b: Vec64u8) Vec64i8 { return nr_simd_cmpeq_Vec64u8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec64u8(a: Vec64u8, b: Vec64u8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq64u8(a: Vec64u8, b: Vec64u8) Vec64i8 { return nr_simd_cmpneq_Vec64u8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec64u8(a: Vec64u8, b: Vec64u8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt64u8(a: Vec64u8, b: Vec64u8) Vec64i8 { return nr_simd_cmplt_Vec64u8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec64u8(a: Vec64u8, b: Vec64u8) Vec64i8
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt64u8(a: Vec64u8, b: Vec64u8) Vec64i8 { return nr_simd_cmpgt_Vec64u8(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec64u8(mask: Vec64i8, a: Vec64u8, b: Vec64u8) Vec64u8
[cfg("target_feature=avx512f")] [inline] pub fn Blend64u8(mask: Vec64i8, a: Vec64u8, b: Vec64u8) Vec64u8 { return nr_simd_blend_Vec64u8(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec32i16(a: Vec32i16, b: Vec32i16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq32i16(a: Vec32i16, b: Vec32i16) Vec32i16 { return nr_simd_cmpeq_Vec32i16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec32i16(a: Vec32i16, b: Vec32i16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq32i16(a: Vec32i16, b: Vec32i16) Vec32i16 { return nr_simd_cmpneq_Vec32i16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec32i16(a: Vec32i16, b: Vec32i16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt32i16(a: Vec32i16, b: Vec32i16) Vec32i16 { return nr_simd_cmplt_Vec32i16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec32i16(a: Vec32i16, b: Vec32i16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt32i16(a: Vec32i16, b: Vec32i16) Vec32i16 { return nr_simd_cmpgt_Vec32i16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec32i16(mask: Vec32i16, a: Vec32i16, b: Vec32i16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn Blend32i16(mask: Vec32i16, a: Vec32i16, b: Vec32i16) Vec32i16 { return nr_simd_blend_Vec32i16(mask, a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpeq_Vec32u16(a: Vec32u16, b: Vec32u16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpEq32u16(a: Vec32u16, b: Vec32u16) Vec32i16 { return nr_simd_cmpeq_Vec32u16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpneq_Vec32u16(a: Vec32u16, b: Vec32u16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpNeq32u16(a: Vec32u16, b: Vec32u16) Vec32i16 { return nr_simd_cmpneq_Vec32u16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmplt_Vec32u16(a: Vec32u16, b: Vec32u16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpLt32u16(a: Vec32u16, b: Vec32u16) Vec32i16 { return nr_simd_cmplt_Vec32u16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_cmpgt_Vec32u16(a: Vec32u16, b: Vec32u16) Vec32i16
[cfg("target_feature=avx512f")] [inline] pub fn CmpGt32u16(a: Vec32u16, b: Vec32u16) Vec32i16 { return nr_simd_cmpgt_Vec32u16(a, b) }
[cfg("target_feature=avx512f")] extern fn nr_simd_blend_Vec32u16(mask: Vec32i16, a: Vec32u16, b: Vec32u16) Vec32u16
[cfg("target_feature=avx512f")] [inline] pub fn Blend32u16(mask: Vec32i16, a: Vec32u16, b: Vec32u16) Vec32u16 { return nr_simd_blend_Vec32u16(mask, a, b) }
extern fn nr_simd_cmpeq_Vec4d(a: Vec4d, b: Vec4d) Vec4i64
[inline] pub fn CmpEq4d(a: Vec4d, b: Vec4d) Vec4i64 { return nr_simd_cmpeq_Vec4d(a, b) }
extern fn nr_simd_cmpneq_Vec4d(a: Vec4d, b: Vec4d) Vec4i64
[inline] pub fn CmpNeq4d(a: Vec4d, b: Vec4d) Vec4i64 { return nr_simd_cmpneq_Vec4d(a, b) }
extern fn nr_simd_cmplt_Vec4d(a: Vec4d, b: Vec4d) Vec4i64
[inline] pub fn CmpLt4d(a: Vec4d, b: Vec4d) Vec4i64 { return nr_simd_cmplt_Vec4d(a, b) }
extern fn nr_simd_cmpgt_Vec4d(a: Vec4d, b: Vec4d) Vec4i64
[inline] pub fn CmpGt4d(a: Vec4d, b: Vec4d) Vec4i64 { return nr_simd_cmpgt_Vec4d(a, b) }
extern fn nr_simd_blend_Vec4d(mask: Vec4i64, a: Vec4d, b: Vec4d) Vec4d
[inline] pub fn Blend4d(mask: Vec4i64, a: Vec4d, b: Vec4d) Vec4d { return nr_simd_blend_Vec4d(mask, a, b) }
extern fn nr_simd_cmpeq_Vec8f(a: Vec8f, b: Vec8f) Vec8i
[inline] pub fn CmpEq8f(a: Vec8f, b: Vec8f) Vec8i { return nr_simd_cmpeq_Vec8f(a, b) }
extern fn nr_simd_cmpneq_Vec8f(a: Vec8f, b: Vec8f) Vec8i
[inline] pub fn CmpNeq8f(a: Vec8f, b: Vec8f) Vec8i { return nr_simd_cmpneq_Vec8f(a, b) }
extern fn nr_simd_cmplt_Vec8f(a: Vec8f, b: Vec8f) Vec8i
[inline] pub fn CmpLt8f(a: Vec8f, b: Vec8f) Vec8i { return nr_simd_cmplt_Vec8f(a, b) }
extern fn nr_simd_cmpgt_Vec8f(a: Vec8f, b: Vec8f) Vec8i
[inline] pub fn CmpGt8f(a: Vec8f, b: Vec8f) Vec8i { return nr_simd_cmpgt_Vec8f(a, b) }
extern fn nr_simd_blend_Vec8f(mask: Vec8i, a: Vec8f, b: Vec8f) Vec8f
[inline] pub fn Blend8f(mask: Vec8i, a: Vec8f, b: Vec8f) Vec8f { return nr_simd_blend_Vec8f(mask, a, b) }
extern fn nr_simd_cmpeq_Vec4f(a: Vec4f, b: Vec4f) Vec4i
[inline] pub fn CmpEq4f(a: Vec4f, b: Vec4f) Vec4i { return nr_simd_cmpeq_Vec4f(a, b) }
extern fn nr_simd_cmpneq_Vec4f(a: Vec4f, b: Vec4f) Vec4i
[inline] pub fn CmpNeq4f(a: Vec4f, b: Vec4f) Vec4i { return nr_simd_cmpneq_Vec4f(a, b) }
extern fn nr_simd_cmplt_Vec4f(a: Vec4f, b: Vec4f) Vec4i
[inline] pub fn CmpLt4f(a: Vec4f, b: Vec4f) Vec4i { return nr_simd_cmplt_Vec4f(a, b) }
extern fn nr_simd_cmpgt_Vec4f(a: Vec4f, b: Vec4f) Vec4i
[inline] pub fn CmpGt4f(a: Vec4f, b: Vec4f) Vec4i { return nr_simd_cmpgt_Vec4f(a, b) }
extern fn nr_simd_blend_Vec4f(mask: Vec4i, a: Vec4f, b: Vec4f) Vec4f
[inline] pub fn Blend4f(mask: Vec4i, a: Vec4f, b: Vec4f) Vec4f { return nr_simd_blend_Vec4f(mask, a, b) }
extern fn nr_simd_cmpeq_Vec8i(a: Vec8i, b: Vec8i) Vec8i
[inline] pub fn CmpEq8i(a: Vec8i, b: Vec8i) Vec8i { return nr_simd_cmpeq_Vec8i(a, b) }
extern fn nr_simd_cmpneq_Vec8i(a: Vec8i, b: Vec8i) Vec8i
[inline] pub fn CmpNeq8i(a: Vec8i, b: Vec8i) Vec8i { return nr_simd_cmpneq_Vec8i(a, b) }
extern fn nr_simd_cmplt_Vec8i(a: Vec8i, b: Vec8i) Vec8i
[inline] pub fn CmpLt8i(a: Vec8i, b: Vec8i) Vec8i { return nr_simd_cmplt_Vec8i(a, b) }
extern fn nr_simd_cmpgt_Vec8i(a: Vec8i, b: Vec8i) Vec8i
[inline] pub fn CmpGt8i(a: Vec8i, b: Vec8i) Vec8i { return nr_simd_cmpgt_Vec8i(a, b) }
extern fn nr_simd_blend_Vec8i(mask: Vec8i, a: Vec8i, b: Vec8i) Vec8i
[inline] pub fn Blend8i(mask: Vec8i, a: Vec8i, b: Vec8i) Vec8i { return nr_simd_blend_Vec8i(mask, a, b) }
extern fn nr_simd_cmpeq_Vec4i(a: Vec4i, b: Vec4i) Vec4i
[inline] pub fn CmpEq4i(a: Vec4i, b: Vec4i) Vec4i { return nr_simd_cmpeq_Vec4i(a, b) }
extern fn nr_simd_cmpneq_Vec4i(a: Vec4i, b: Vec4i) Vec4i
[inline] pub fn CmpNeq4i(a: Vec4i, b: Vec4i) Vec4i { return nr_simd_cmpneq_Vec4i(a, b) }
extern fn nr_simd_cmplt_Vec4i(a: Vec4i, b: Vec4i) Vec4i
[inline] pub fn CmpLt4i(a: Vec4i, b: Vec4i) Vec4i { return nr_simd_cmplt_Vec4i(a, b) }
extern fn nr_simd_cmpgt_Vec4i(a: Vec4i, b: Vec4i) Vec4i
[inline] pub fn CmpGt4i(a: Vec4i, b: Vec4i) Vec4i { return nr_simd_cmpgt_Vec4i(a, b) }
extern fn nr_simd_blend_Vec4i(mask: Vec4i, a: Vec4i, b: Vec4i) Vec4i
[inline] pub fn Blend4i(mask: Vec4i, a: Vec4i, b: Vec4i) Vec4i { return nr_simd_blend_Vec4i(mask, a, b) }

// --- PHASE 3: MASK TYPES ---
pub type Mask16i8 = Vec16i8
pub type Mask32i8 = Vec32i8
pub type Mask16u8 = Vec16i8
pub type Mask32u8 = Vec32i8
pub type Mask8i16 = Vec8i16
pub type Mask16i16 = Vec16i16
pub type Mask8u16 = Vec8i16
pub type Mask16u16 = Vec16i16
pub type Mask4u32 = Vec4i
pub type Mask8u32 = Vec8i
pub type Mask2i64 = Vec2i64
pub type Mask4i64 = Vec4i64
pub type Mask2u64 = Vec2i64
pub type Mask4u64 = Vec4i64
[cfg("target_feature=avx512f")] pub type Mask16f = Vec16i32
[cfg("target_feature=avx512f")] pub type Mask8d = Vec8i64
[cfg("target_feature=avx512f")] pub type Mask16i32 = Vec16i32
[cfg("target_feature=avx512f")] pub type Mask16u32 = Vec16i32
[cfg("target_feature=avx512f")] pub type Mask8i64 = Vec8i64
[cfg("target_feature=avx512f")] pub type Mask8u64 = Vec8i64
[cfg("target_feature=avx512f")] pub type Mask64i8 = Vec64i8
[cfg("target_feature=avx512f")] pub type Mask64u8 = Vec64i8
[cfg("target_feature=avx512f")] pub type Mask32i16 = Vec32i16
[cfg("target_feature=avx512f")] pub type Mask32u16 = Vec32i16
pub type Mask4d = Vec4i64
pub type Mask8f = Vec8i
pub type Mask4f = Vec4i
pub type Mask8i = Vec8i
pub type Mask4i = Vec4i

// --- PHASE 3: MASKED EXECUTION ---
[inline] pub fn VecAddMasked16i8(a: Vec16i8, b: Vec16i8, mask: Mask16i8) Vec16i8 { return Blend16i8(mask, VecAdd16i8(a, b), a) }
[inline] pub fn VecSubMasked16i8(a: Vec16i8, b: Vec16i8, mask: Mask16i8) Vec16i8 { return Blend16i8(mask, VecSub16i8(a, b), a) }
[inline] pub fn VecMulMasked16i8(a: Vec16i8, b: Vec16i8, mask: Mask16i8) Vec16i8 { return Blend16i8(mask, VecMul16i8(a, b), a) }
[inline] pub fn VecAddMasked32i8(a: Vec32i8, b: Vec32i8, mask: Mask32i8) Vec32i8 { return Blend32i8(mask, VecAdd32i8(a, b), a) }
[inline] pub fn VecSubMasked32i8(a: Vec32i8, b: Vec32i8, mask: Mask32i8) Vec32i8 { return Blend32i8(mask, VecSub32i8(a, b), a) }
[inline] pub fn VecMulMasked32i8(a: Vec32i8, b: Vec32i8, mask: Mask32i8) Vec32i8 { return Blend32i8(mask, VecMul32i8(a, b), a) }
[inline] pub fn VecAddMasked16u8(a: Vec16u8, b: Vec16u8, mask: Mask16u8) Vec16u8 { return Blend16u8(mask, VecAdd16u8(a, b), a) }
[inline] pub fn VecSubMasked16u8(a: Vec16u8, b: Vec16u8, mask: Mask16u8) Vec16u8 { return Blend16u8(mask, VecSub16u8(a, b), a) }
[inline] pub fn VecMulMasked16u8(a: Vec16u8, b: Vec16u8, mask: Mask16u8) Vec16u8 { return Blend16u8(mask, VecMul16u8(a, b), a) }
[inline] pub fn VecAddMasked32u8(a: Vec32u8, b: Vec32u8, mask: Mask32u8) Vec32u8 { return Blend32u8(mask, VecAdd32u8(a, b), a) }
[inline] pub fn VecSubMasked32u8(a: Vec32u8, b: Vec32u8, mask: Mask32u8) Vec32u8 { return Blend32u8(mask, VecSub32u8(a, b), a) }
[inline] pub fn VecMulMasked32u8(a: Vec32u8, b: Vec32u8, mask: Mask32u8) Vec32u8 { return Blend32u8(mask, VecMul32u8(a, b), a) }
[inline] pub fn VecAddMasked8i16(a: Vec8i16, b: Vec8i16, mask: Mask8i16) Vec8i16 { return Blend8i16(mask, VecAdd8i16(a, b), a) }
[inline] pub fn VecSubMasked8i16(a: Vec8i16, b: Vec8i16, mask: Mask8i16) Vec8i16 { return Blend8i16(mask, VecSub8i16(a, b), a) }
[inline] pub fn VecMulMasked8i16(a: Vec8i16, b: Vec8i16, mask: Mask8i16) Vec8i16 { return Blend8i16(mask, VecMul8i16(a, b), a) }
[inline] pub fn VecAddMasked16i16(a: Vec16i16, b: Vec16i16, mask: Mask16i16) Vec16i16 { return Blend16i16(mask, VecAdd16i16(a, b), a) }
[inline] pub fn VecSubMasked16i16(a: Vec16i16, b: Vec16i16, mask: Mask16i16) Vec16i16 { return Blend16i16(mask, VecSub16i16(a, b), a) }
[inline] pub fn VecMulMasked16i16(a: Vec16i16, b: Vec16i16, mask: Mask16i16) Vec16i16 { return Blend16i16(mask, VecMul16i16(a, b), a) }
[inline] pub fn VecAddMasked8u16(a: Vec8u16, b: Vec8u16, mask: Mask8u16) Vec8u16 { return Blend8u16(mask, VecAdd8u16(a, b), a) }
[inline] pub fn VecSubMasked8u16(a: Vec8u16, b: Vec8u16, mask: Mask8u16) Vec8u16 { return Blend8u16(mask, VecSub8u16(a, b), a) }
[inline] pub fn VecMulMasked8u16(a: Vec8u16, b: Vec8u16, mask: Mask8u16) Vec8u16 { return Blend8u16(mask, VecMul8u16(a, b), a) }
[inline] pub fn VecAddMasked16u16(a: Vec16u16, b: Vec16u16, mask: Mask16u16) Vec16u16 { return Blend16u16(mask, VecAdd16u16(a, b), a) }
[inline] pub fn VecSubMasked16u16(a: Vec16u16, b: Vec16u16, mask: Mask16u16) Vec16u16 { return Blend16u16(mask, VecSub16u16(a, b), a) }
[inline] pub fn VecMulMasked16u16(a: Vec16u16, b: Vec16u16, mask: Mask16u16) Vec16u16 { return Blend16u16(mask, VecMul16u16(a, b), a) }
[inline] pub fn VecAddMasked4u32(a: Vec4u32, b: Vec4u32, mask: Mask4u32) Vec4u32 { return Blend4u32(mask, VecAdd4u32(a, b), a) }
[inline] pub fn VecSubMasked4u32(a: Vec4u32, b: Vec4u32, mask: Mask4u32) Vec4u32 { return Blend4u32(mask, VecSub4u32(a, b), a) }
[inline] pub fn VecMulMasked4u32(a: Vec4u32, b: Vec4u32, mask: Mask4u32) Vec4u32 { return Blend4u32(mask, VecMul4u32(a, b), a) }
[inline] pub fn VecAddMasked8u32(a: Vec8u32, b: Vec8u32, mask: Mask8u32) Vec8u32 { return Blend8u32(mask, VecAdd8u32(a, b), a) }
[inline] pub fn VecSubMasked8u32(a: Vec8u32, b: Vec8u32, mask: Mask8u32) Vec8u32 { return Blend8u32(mask, VecSub8u32(a, b), a) }
[inline] pub fn VecMulMasked8u32(a: Vec8u32, b: Vec8u32, mask: Mask8u32) Vec8u32 { return Blend8u32(mask, VecMul8u32(a, b), a) }
[inline] pub fn VecAddMasked2i64(a: Vec2i64, b: Vec2i64, mask: Mask2i64) Vec2i64 { return Blend2i64(mask, VecAdd2i64(a, b), a) }
[inline] pub fn VecSubMasked2i64(a: Vec2i64, b: Vec2i64, mask: Mask2i64) Vec2i64 { return Blend2i64(mask, VecSub2i64(a, b), a) }
[inline] pub fn VecMulMasked2i64(a: Vec2i64, b: Vec2i64, mask: Mask2i64) Vec2i64 { return Blend2i64(mask, VecMul2i64(a, b), a) }
[inline] pub fn VecAddMasked4i64(a: Vec4i64, b: Vec4i64, mask: Mask4i64) Vec4i64 { return Blend4i64(mask, VecAdd4i64(a, b), a) }
[inline] pub fn VecSubMasked4i64(a: Vec4i64, b: Vec4i64, mask: Mask4i64) Vec4i64 { return Blend4i64(mask, VecSub4i64(a, b), a) }
[inline] pub fn VecMulMasked4i64(a: Vec4i64, b: Vec4i64, mask: Mask4i64) Vec4i64 { return Blend4i64(mask, VecMul4i64(a, b), a) }
[inline] pub fn VecAddMasked2u64(a: Vec2u64, b: Vec2u64, mask: Mask2u64) Vec2u64 { return Blend2u64(mask, VecAdd2u64(a, b), a) }
[inline] pub fn VecSubMasked2u64(a: Vec2u64, b: Vec2u64, mask: Mask2u64) Vec2u64 { return Blend2u64(mask, VecSub2u64(a, b), a) }
[inline] pub fn VecMulMasked2u64(a: Vec2u64, b: Vec2u64, mask: Mask2u64) Vec2u64 { return Blend2u64(mask, VecMul2u64(a, b), a) }
[inline] pub fn VecAddMasked4u64(a: Vec4u64, b: Vec4u64, mask: Mask4u64) Vec4u64 { return Blend4u64(mask, VecAdd4u64(a, b), a) }
[inline] pub fn VecSubMasked4u64(a: Vec4u64, b: Vec4u64, mask: Mask4u64) Vec4u64 { return Blend4u64(mask, VecSub4u64(a, b), a) }
[inline] pub fn VecMulMasked4u64(a: Vec4u64, b: Vec4u64, mask: Mask4u64) Vec4u64 { return Blend4u64(mask, VecMul4u64(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked16f(a: Vec16f, b: Vec16f, mask: Mask16f) Vec16f { return Blend16f(mask, VecAdd16f(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked16f(a: Vec16f, b: Vec16f, mask: Mask16f) Vec16f { return Blend16f(mask, VecSub16f(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked16f(a: Vec16f, b: Vec16f, mask: Mask16f) Vec16f { return Blend16f(mask, VecMul16f(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecDivMasked16f(a: Vec16f, b: Vec16f, mask: Mask16f) Vec16f { return Blend16f(mask, VecDiv16f(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked8d(a: Vec8d, b: Vec8d, mask: Mask8d) Vec8d { return Blend8d(mask, VecAdd8d(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked8d(a: Vec8d, b: Vec8d, mask: Mask8d) Vec8d { return Blend8d(mask, VecSub8d(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked8d(a: Vec8d, b: Vec8d, mask: Mask8d) Vec8d { return Blend8d(mask, VecMul8d(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecDivMasked8d(a: Vec8d, b: Vec8d, mask: Mask8d) Vec8d { return Blend8d(mask, VecDiv8d(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked16i32(a: Vec16i32, b: Vec16i32, mask: Mask16i32) Vec16i32 { return Blend16i32(mask, VecAdd16i32(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked16i32(a: Vec16i32, b: Vec16i32, mask: Mask16i32) Vec16i32 { return Blend16i32(mask, VecSub16i32(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked16i32(a: Vec16i32, b: Vec16i32, mask: Mask16i32) Vec16i32 { return Blend16i32(mask, VecMul16i32(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked16u32(a: Vec16u32, b: Vec16u32, mask: Mask16u32) Vec16u32 { return Blend16u32(mask, VecAdd16u32(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked16u32(a: Vec16u32, b: Vec16u32, mask: Mask16u32) Vec16u32 { return Blend16u32(mask, VecSub16u32(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked16u32(a: Vec16u32, b: Vec16u32, mask: Mask16u32) Vec16u32 { return Blend16u32(mask, VecMul16u32(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked8i64(a: Vec8i64, b: Vec8i64, mask: Mask8i64) Vec8i64 { return Blend8i64(mask, VecAdd8i64(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked8i64(a: Vec8i64, b: Vec8i64, mask: Mask8i64) Vec8i64 { return Blend8i64(mask, VecSub8i64(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked8i64(a: Vec8i64, b: Vec8i64, mask: Mask8i64) Vec8i64 { return Blend8i64(mask, VecMul8i64(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked8u64(a: Vec8u64, b: Vec8u64, mask: Mask8u64) Vec8u64 { return Blend8u64(mask, VecAdd8u64(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked8u64(a: Vec8u64, b: Vec8u64, mask: Mask8u64) Vec8u64 { return Blend8u64(mask, VecSub8u64(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked8u64(a: Vec8u64, b: Vec8u64, mask: Mask8u64) Vec8u64 { return Blend8u64(mask, VecMul8u64(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked64i8(a: Vec64i8, b: Vec64i8, mask: Mask64i8) Vec64i8 { return Blend64i8(mask, VecAdd64i8(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked64i8(a: Vec64i8, b: Vec64i8, mask: Mask64i8) Vec64i8 { return Blend64i8(mask, VecSub64i8(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked64i8(a: Vec64i8, b: Vec64i8, mask: Mask64i8) Vec64i8 { return Blend64i8(mask, VecMul64i8(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked64u8(a: Vec64u8, b: Vec64u8, mask: Mask64u8) Vec64u8 { return Blend64u8(mask, VecAdd64u8(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked64u8(a: Vec64u8, b: Vec64u8, mask: Mask64u8) Vec64u8 { return Blend64u8(mask, VecSub64u8(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked64u8(a: Vec64u8, b: Vec64u8, mask: Mask64u8) Vec64u8 { return Blend64u8(mask, VecMul64u8(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked32i16(a: Vec32i16, b: Vec32i16, mask: Mask32i16) Vec32i16 { return Blend32i16(mask, VecAdd32i16(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked32i16(a: Vec32i16, b: Vec32i16, mask: Mask32i16) Vec32i16 { return Blend32i16(mask, VecSub32i16(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked32i16(a: Vec32i16, b: Vec32i16, mask: Mask32i16) Vec32i16 { return Blend32i16(mask, VecMul32i16(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecAddMasked32u16(a: Vec32u16, b: Vec32u16, mask: Mask32u16) Vec32u16 { return Blend32u16(mask, VecAdd32u16(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecSubMasked32u16(a: Vec32u16, b: Vec32u16, mask: Mask32u16) Vec32u16 { return Blend32u16(mask, VecSub32u16(a, b), a) }
[cfg("target_feature=avx512f")] [inline] pub fn VecMulMasked32u16(a: Vec32u16, b: Vec32u16, mask: Mask32u16) Vec32u16 { return Blend32u16(mask, VecMul32u16(a, b), a) }
[inline] pub fn VecAddMasked8f(a: Vec8f, b: Vec8f, mask: Mask8f) Vec8f { return Blend8f(mask, VecAdd8f(a, b), a) }
[inline] pub fn VecSubMasked8f(a: Vec8f, b: Vec8f, mask: Mask8f) Vec8f { return Blend8f(mask, VecSub8f(a, b), a) }
[inline] pub fn VecMulMasked8f(a: Vec8f, b: Vec8f, mask: Mask8f) Vec8f { return Blend8f(mask, VecMul8f(a, b), a) }
[inline] pub fn VecDivMasked8f(a: Vec8f, b: Vec8f, mask: Mask8f) Vec8f { return Blend8f(mask, VecDiv8f(a, b), a) }
[inline] pub fn VecAddMasked4f(a: Vec4f, b: Vec4f, mask: Mask4f) Vec4f { return Blend4f(mask, VecAdd4f(a, b), a) }
[inline] pub fn VecSubMasked4f(a: Vec4f, b: Vec4f, mask: Mask4f) Vec4f { return Blend4f(mask, VecSub4f(a, b), a) }
[inline] pub fn VecMulMasked4f(a: Vec4f, b: Vec4f, mask: Mask4f) Vec4f { return Blend4f(mask, VecMul4f(a, b), a) }
[inline] pub fn VecDivMasked4f(a: Vec4f, b: Vec4f, mask: Mask4f) Vec4f { return Blend4f(mask, VecDiv4f(a, b), a) }
[inline] pub fn VecAddMasked8i(a: Vec8i, b: Vec8i, mask: Mask8i) Vec8i { return Blend8i(mask, VecAdd8i(a, b), a) }
[inline] pub fn VecSubMasked8i(a: Vec8i, b: Vec8i, mask: Mask8i) Vec8i { return Blend8i(mask, VecSub8i(a, b), a) }
[inline] pub fn VecMulMasked8i(a: Vec8i, b: Vec8i, mask: Mask8i) Vec8i { return Blend8i(mask, VecMul8i(a, b), a) }
[inline] pub fn VecAddMasked4i(a: Vec4i, b: Vec4i, mask: Mask4i) Vec4i { return Blend4i(mask, VecAdd4i(a, b), a) }
[inline] pub fn VecSubMasked4i(a: Vec4i, b: Vec4i, mask: Mask4i) Vec4i { return Blend4i(mask, VecSub4i(a, b), a) }
[inline] pub fn VecMulMasked4i(a: Vec4i, b: Vec4i, mask: Mask4i) Vec4i { return Blend4i(mask, VecMul4i(a, b), a) }
extern fn nr_simd_load_aligned_16i8(src: ptr) Vec16i8
extern fn nr_simd_store_aligned_16i8(dst: ptr, v: Vec16i8)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_16i8(src: ptr, indices: Vec16i32) Vec16i8
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_16i8(dst: ptr, indices: Vec16i32, v: Vec16i8)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_32i8(src: ptr) Vec32i8
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_32i8(dst: ptr, v: Vec32i8)
extern fn nr_simd_load_aligned_16u8(src: ptr) Vec16u8
extern fn nr_simd_store_aligned_16u8(dst: ptr, v: Vec16u8)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_16u8(src: ptr, indices: Vec16i32) Vec16u8
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_16u8(dst: ptr, indices: Vec16i32, v: Vec16u8)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_32u8(src: ptr) Vec32u8
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_32u8(dst: ptr, v: Vec32u8)
extern fn nr_simd_load_aligned_8i16(src: ptr) Vec8i16
extern fn nr_simd_store_aligned_8i16(dst: ptr, v: Vec8i16)
extern fn nr_simd_gather_8i16(src: ptr, indices: Vec8i) Vec8i16
extern fn nr_simd_scatter_8i16(dst: ptr, indices: Vec8i, v: Vec8i16)
extern fn nr_simd_load_aligned_16i16(src: ptr) Vec16i16
extern fn nr_simd_store_aligned_16i16(dst: ptr, v: Vec16i16)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_16i16(src: ptr, indices: Vec16i32) Vec16i16
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_16i16(dst: ptr, indices: Vec16i32, v: Vec16i16)
extern fn nr_simd_load_aligned_8u16(src: ptr) Vec8u16
extern fn nr_simd_store_aligned_8u16(dst: ptr, v: Vec8u16)
extern fn nr_simd_gather_8u16(src: ptr, indices: Vec8i) Vec8u16
extern fn nr_simd_scatter_8u16(dst: ptr, indices: Vec8i, v: Vec8u16)
extern fn nr_simd_load_aligned_16u16(src: ptr) Vec16u16
extern fn nr_simd_store_aligned_16u16(dst: ptr, v: Vec16u16)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_16u16(src: ptr, indices: Vec16i32) Vec16u16
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_16u16(dst: ptr, indices: Vec16i32, v: Vec16u16)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_4u32(src: ptr) Vec4u32
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_4u32(dst: ptr, v: Vec4u32)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_4u32(src: ptr, indices: Vec4i) Vec4u32
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_4u32(dst: ptr, indices: Vec4i, v: Vec4u32)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_8u32(src: ptr) Vec8u32
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_8u32(dst: ptr, v: Vec8u32)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_8u32(src: ptr, indices: Vec8i) Vec8u32
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_8u32(dst: ptr, indices: Vec8i, v: Vec8u32)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_2i64(src: ptr) Vec2i64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_2i64(dst: ptr, v: Vec2i64)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_2i64(src: ptr, indices: Vec2i64) Vec2i64
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_2i64(dst: ptr, indices: Vec2i64, v: Vec2i64)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_4i64(src: ptr) Vec4i64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_4i64(dst: ptr, v: Vec4i64)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_4i64(src: ptr, indices: Vec4i64) Vec4i64
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_4i64(dst: ptr, indices: Vec4i64, v: Vec4i64)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_2u64(src: ptr) Vec2u64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_2u64(dst: ptr, v: Vec2u64)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_2u64(src: ptr, indices: Vec2i64) Vec2u64
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_2u64(dst: ptr, indices: Vec2i64, v: Vec2u64)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_4u64(src: ptr) Vec4u64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_4u64(dst: ptr, v: Vec4u64)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_4u64(src: ptr, indices: Vec4i64) Vec4u64
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_4u64(dst: ptr, indices: Vec4i64, v: Vec4u64)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_16f(src: ptr) Vec16f
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_16f(dst: ptr, v: Vec16f)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_16f(src: ptr, indices: Vec16i32) Vec16f
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_16f(dst: ptr, indices: Vec16i32, v: Vec16f)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_8d(src: ptr) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_8d(dst: ptr, v: Vec8d)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_8d(src: ptr, indices: Vec8i64) Vec8d
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_8d(dst: ptr, indices: Vec8i64, v: Vec8d)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_16i32(src: ptr) Vec16i32
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_16i32(dst: ptr, v: Vec16i32)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_16i32(src: ptr, indices: Vec16i32) Vec16i32
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_16i32(dst: ptr, indices: Vec16i32, v: Vec16i32)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_16u32(src: ptr) Vec16u32
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_16u32(dst: ptr, v: Vec16u32)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_16u32(src: ptr, indices: Vec16i32) Vec16u32
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_16u32(dst: ptr, indices: Vec16i32, v: Vec16u32)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_8i64(src: ptr) Vec8i64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_8i64(dst: ptr, v: Vec8i64)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_8i64(src: ptr, indices: Vec8i64) Vec8i64
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_8i64(dst: ptr, indices: Vec8i64, v: Vec8i64)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_8u64(src: ptr) Vec8u64
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_8u64(dst: ptr, v: Vec8u64)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_8u64(src: ptr, indices: Vec8i64) Vec8u64
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_8u64(dst: ptr, indices: Vec8i64, v: Vec8u64)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_64i8(src: ptr) Vec64i8
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_64i8(dst: ptr, v: Vec64i8)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_64u8(src: ptr) Vec64u8
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_64u8(dst: ptr, v: Vec64u8)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_32i16(src: ptr) Vec32i16
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_32i16(dst: ptr, v: Vec32i16)
[cfg("target_feature=avx512f")] extern fn nr_simd_load_aligned_32u16(src: ptr) Vec32u16
[cfg("target_feature=avx512f")] extern fn nr_simd_store_aligned_32u16(dst: ptr, v: Vec32u16)
extern fn nr_simd_load_aligned_4d(src: ptr) Vec4d
extern fn nr_simd_store_aligned_4d(dst: ptr, v: Vec4d)
[cfg("target_feature=avx512f")] extern fn nr_simd_gather_4d(src: ptr, indices: Vec4i64) Vec4d
[cfg("target_feature=avx512f")] extern fn nr_simd_scatter_4d(dst: ptr, indices: Vec4i64, v: Vec4d)
extern fn nr_simd_load_aligned_8f(src: ptr) Vec8f
extern fn nr_simd_store_aligned_8f(dst: ptr, v: Vec8f)
extern fn nr_simd_gather_8f(src: ptr, indices: Vec8i) Vec8f
extern fn nr_simd_scatter_8f(dst: ptr, indices: Vec8i, v: Vec8f)
extern fn nr_simd_load_aligned_4f(src: ptr) Vec4f
extern fn nr_simd_store_aligned_4f(dst: ptr, v: Vec4f)
extern fn nr_simd_gather_4f(src: ptr, indices: Vec4i) Vec4f
extern fn nr_simd_scatter_4f(dst: ptr, indices: Vec4i, v: Vec4f)
extern fn nr_simd_load_aligned_8i(src: ptr) Vec8i
extern fn nr_simd_store_aligned_8i(dst: ptr, v: Vec8i)
extern fn nr_simd_gather_8i(src: ptr, indices: Vec8i) Vec8i
extern fn nr_simd_scatter_8i(dst: ptr, indices: Vec8i, v: Vec8i)
extern fn nr_simd_load_aligned_4i(src: ptr) Vec4i
extern fn nr_simd_store_aligned_4i(dst: ptr, v: Vec4i)
extern fn nr_simd_gather_4i(src: ptr, indices: Vec4i) Vec4i
extern fn nr_simd_scatter_4i(dst: ptr, indices: Vec4i, v: Vec4i)


// Phase 4 Wrappers
[inline] pub fn LoadAligned16i8(data: ptr) Vec16i8 { return nr_simd_load_aligned_16i8(data) }
[inline] pub fn StoreAligned16i8(data: ptr, vec: Vec16i8) { nr_simd_store_aligned_16i8(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather16i8(data: ptr, indices: Vec16i32) Vec16i8 { return nr_simd_gather_16i8(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter16i8(data: ptr, indices: Vec16i32, vec: Vec16i8) { nr_simd_scatter_16i8(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned32i8(data: ptr) Vec32i8 { return nr_simd_load_aligned_32i8(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned32i8(data: ptr, vec: Vec32i8) { nr_simd_store_aligned_32i8(data, vec) }
[inline] pub fn LoadAligned16u8(data: ptr) Vec16u8 { return nr_simd_load_aligned_16u8(data) }
[inline] pub fn StoreAligned16u8(data: ptr, vec: Vec16u8) { nr_simd_store_aligned_16u8(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather16u8(data: ptr, indices: Vec16i32) Vec16u8 { return nr_simd_gather_16u8(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter16u8(data: ptr, indices: Vec16i32, vec: Vec16u8) { nr_simd_scatter_16u8(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned32u8(data: ptr) Vec32u8 { return nr_simd_load_aligned_32u8(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned32u8(data: ptr, vec: Vec32u8) { nr_simd_store_aligned_32u8(data, vec) }
[inline] pub fn LoadAligned8i16(data: ptr) Vec8i16 { return nr_simd_load_aligned_8i16(data) }
[inline] pub fn StoreAligned8i16(data: ptr, vec: Vec8i16) { nr_simd_store_aligned_8i16(data, vec) }
[inline] pub fn Gather8i16(data: ptr, indices: Vec8i) Vec8i16 { return nr_simd_gather_8i16(data, indices) }
[inline] pub fn Scatter8i16(data: ptr, indices: Vec8i, vec: Vec8i16) { nr_simd_scatter_8i16(data, indices, vec) }
[inline] pub fn LoadAligned16i16(data: ptr) Vec16i16 { return nr_simd_load_aligned_16i16(data) }
[inline] pub fn StoreAligned16i16(data: ptr, vec: Vec16i16) { nr_simd_store_aligned_16i16(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather16i16(data: ptr, indices: Vec16i32) Vec16i16 { return nr_simd_gather_16i16(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter16i16(data: ptr, indices: Vec16i32, vec: Vec16i16) { nr_simd_scatter_16i16(data, indices, vec) }
[inline] pub fn LoadAligned8u16(data: ptr) Vec8u16 { return nr_simd_load_aligned_8u16(data) }
[inline] pub fn StoreAligned8u16(data: ptr, vec: Vec8u16) { nr_simd_store_aligned_8u16(data, vec) }
[inline] pub fn Gather8u16(data: ptr, indices: Vec8i) Vec8u16 { return nr_simd_gather_8u16(data, indices) }
[inline] pub fn Scatter8u16(data: ptr, indices: Vec8i, vec: Vec8u16) { nr_simd_scatter_8u16(data, indices, vec) }
[inline] pub fn LoadAligned16u16(data: ptr) Vec16u16 { return nr_simd_load_aligned_16u16(data) }
[inline] pub fn StoreAligned16u16(data: ptr, vec: Vec16u16) { nr_simd_store_aligned_16u16(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather16u16(data: ptr, indices: Vec16i32) Vec16u16 { return nr_simd_gather_16u16(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter16u16(data: ptr, indices: Vec16i32, vec: Vec16u16) { nr_simd_scatter_16u16(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned4u32(data: ptr) Vec4u32 { return nr_simd_load_aligned_4u32(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned4u32(data: ptr, vec: Vec4u32) { nr_simd_store_aligned_4u32(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather4u32(data: ptr, indices: Vec4i) Vec4u32 { return nr_simd_gather_4u32(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter4u32(data: ptr, indices: Vec4i, vec: Vec4u32) { nr_simd_scatter_4u32(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned8u32(data: ptr) Vec8u32 { return nr_simd_load_aligned_8u32(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned8u32(data: ptr, vec: Vec8u32) { nr_simd_store_aligned_8u32(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather8u32(data: ptr, indices: Vec8i) Vec8u32 { return nr_simd_gather_8u32(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter8u32(data: ptr, indices: Vec8i, vec: Vec8u32) { nr_simd_scatter_8u32(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned2i64(data: ptr) Vec2i64 { return nr_simd_load_aligned_2i64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned2i64(data: ptr, vec: Vec2i64) { nr_simd_store_aligned_2i64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather2i64(data: ptr, indices: Vec2i64) Vec2i64 { return nr_simd_gather_2i64(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter2i64(data: ptr, indices: Vec2i64, vec: Vec2i64) { nr_simd_scatter_2i64(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned4i64(data: ptr) Vec4i64 { return nr_simd_load_aligned_4i64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned4i64(data: ptr, vec: Vec4i64) { nr_simd_store_aligned_4i64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather4i64(data: ptr, indices: Vec4i64) Vec4i64 { return nr_simd_gather_4i64(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter4i64(data: ptr, indices: Vec4i64, vec: Vec4i64) { nr_simd_scatter_4i64(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned2u64(data: ptr) Vec2u64 { return nr_simd_load_aligned_2u64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned2u64(data: ptr, vec: Vec2u64) { nr_simd_store_aligned_2u64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather2u64(data: ptr, indices: Vec2i64) Vec2u64 { return nr_simd_gather_2u64(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter2u64(data: ptr, indices: Vec2i64, vec: Vec2u64) { nr_simd_scatter_2u64(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned4u64(data: ptr) Vec4u64 { return nr_simd_load_aligned_4u64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned4u64(data: ptr, vec: Vec4u64) { nr_simd_store_aligned_4u64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather4u64(data: ptr, indices: Vec4i64) Vec4u64 { return nr_simd_gather_4u64(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter4u64(data: ptr, indices: Vec4i64, vec: Vec4u64) { nr_simd_scatter_4u64(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned16f(data: ptr) Vec16f { return nr_simd_load_aligned_16f(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned16f(data: ptr, vec: Vec16f) { nr_simd_store_aligned_16f(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather16f(data: ptr, indices: Vec16i32) Vec16f { return nr_simd_gather_16f(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter16f(data: ptr, indices: Vec16i32, vec: Vec16f) { nr_simd_scatter_16f(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned8d(data: ptr) Vec8d { return nr_simd_load_aligned_8d(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned8d(data: ptr, vec: Vec8d) { nr_simd_store_aligned_8d(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather8d(data: ptr, indices: Vec8i64) Vec8d { return nr_simd_gather_8d(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter8d(data: ptr, indices: Vec8i64, vec: Vec8d) { nr_simd_scatter_8d(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned16i32(data: ptr) Vec16i32 { return nr_simd_load_aligned_16i32(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned16i32(data: ptr, vec: Vec16i32) { nr_simd_store_aligned_16i32(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather16i32(data: ptr, indices: Vec16i32) Vec16i32 { return nr_simd_gather_16i32(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter16i32(data: ptr, indices: Vec16i32, vec: Vec16i32) { nr_simd_scatter_16i32(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned16u32(data: ptr) Vec16u32 { return nr_simd_load_aligned_16u32(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned16u32(data: ptr, vec: Vec16u32) { nr_simd_store_aligned_16u32(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather16u32(data: ptr, indices: Vec16i32) Vec16u32 { return nr_simd_gather_16u32(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter16u32(data: ptr, indices: Vec16i32, vec: Vec16u32) { nr_simd_scatter_16u32(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned8i64(data: ptr) Vec8i64 { return nr_simd_load_aligned_8i64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned8i64(data: ptr, vec: Vec8i64) { nr_simd_store_aligned_8i64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather8i64(data: ptr, indices: Vec8i64) Vec8i64 { return nr_simd_gather_8i64(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter8i64(data: ptr, indices: Vec8i64, vec: Vec8i64) { nr_simd_scatter_8i64(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned8u64(data: ptr) Vec8u64 { return nr_simd_load_aligned_8u64(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned8u64(data: ptr, vec: Vec8u64) { nr_simd_store_aligned_8u64(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather8u64(data: ptr, indices: Vec8i64) Vec8u64 { return nr_simd_gather_8u64(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter8u64(data: ptr, indices: Vec8i64, vec: Vec8u64) { nr_simd_scatter_8u64(data, indices, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned64i8(data: ptr) Vec64i8 { return nr_simd_load_aligned_64i8(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned64i8(data: ptr, vec: Vec64i8) { nr_simd_store_aligned_64i8(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned64u8(data: ptr) Vec64u8 { return nr_simd_load_aligned_64u8(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned64u8(data: ptr, vec: Vec64u8) { nr_simd_store_aligned_64u8(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned32i16(data: ptr) Vec32i16 { return nr_simd_load_aligned_32i16(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned32i16(data: ptr, vec: Vec32i16) { nr_simd_store_aligned_32i16(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn LoadAligned32u16(data: ptr) Vec32u16 { return nr_simd_load_aligned_32u16(data) }
[cfg("target_feature=avx512f")] [inline] pub fn StoreAligned32u16(data: ptr, vec: Vec32u16) { nr_simd_store_aligned_32u16(data, vec) }
[inline] pub fn LoadAligned4d(data: ptr) Vec4d { return nr_simd_load_aligned_4d(data) }
[inline] pub fn StoreAligned4d(data: ptr, vec: Vec4d) { nr_simd_store_aligned_4d(data, vec) }
[cfg("target_feature=avx512f")] [inline] pub fn Gather4d(data: ptr, indices: Vec4i64) Vec4d { return nr_simd_gather_4d(data, indices) }
[cfg("target_feature=avx512f")] [inline] pub fn Scatter4d(data: ptr, indices: Vec4i64, vec: Vec4d) { nr_simd_scatter_4d(data, indices, vec) }
[inline] pub fn LoadAligned8f(data: ptr) Vec8f { return nr_simd_load_aligned_8f(data) }
[inline] pub fn StoreAligned8f(data: ptr, vec: Vec8f) { nr_simd_store_aligned_8f(data, vec) }
[inline] pub fn Gather8f(data: ptr, indices: Vec8i) Vec8f { return nr_simd_gather_8f(data, indices) }
[inline] pub fn Scatter8f(data: ptr, indices: Vec8i, vec: Vec8f) { nr_simd_scatter_8f(data, indices, vec) }
[inline] pub fn LoadAligned4f(data: ptr) Vec4f { return nr_simd_load_aligned_4f(data) }
[inline] pub fn StoreAligned4f(data: ptr, vec: Vec4f) { nr_simd_store_aligned_4f(data, vec) }
[inline] pub fn Gather4f(data: ptr, indices: Vec4i) Vec4f { return nr_simd_gather_4f(data, indices) }
[inline] pub fn Scatter4f(data: ptr, indices: Vec4i, vec: Vec4f) { nr_simd_scatter_4f(data, indices, vec) }
[inline] pub fn LoadAligned8i(data: ptr) Vec8i { return nr_simd_load_aligned_8i(data) }
[inline] pub fn StoreAligned8i(data: ptr, vec: Vec8i) { nr_simd_store_aligned_8i(data, vec) }
[inline] pub fn Gather8i(data: ptr, indices: Vec8i) Vec8i { return nr_simd_gather_8i(data, indices) }
[inline] pub fn Scatter8i(data: ptr, indices: Vec8i, vec: Vec8i) { nr_simd_scatter_8i(data, indices, vec) }
[inline] pub fn LoadAligned4i(data: ptr) Vec4i { return nr_simd_load_aligned_4i(data) }
[inline] pub fn StoreAligned4i(data: ptr, vec: Vec4i) { nr_simd_store_aligned_4i(data, vec) }
[inline] pub fn Gather4i(data: ptr, indices: Vec4i) Vec4i { return nr_simd_gather_4i(data, indices) }
[inline] pub fn Scatter4i(data: ptr, indices: Vec4i, vec: Vec4i) { nr_simd_scatter_4i(data, indices, vec) }



extern fn nr_simd_set8_f32(a: f32, b: f32, c: f32, d: f32, e: f32, f: f32, g: f32, h: f32) Vec8f
extern fn nr_simd_set8_i32(a: i32, b: i32, c: i32, d: i32, e: i32, f: i32, g: i32, h: i32) Vec8i

[inline] pub fn VecSet8f(a: f32, b: f32, c: f32, d: f32, e: f32, f: f32, g: f32, h: f32) Vec8f {
    return nr_simd_set8_f32(a, b, c, d, e, f, g, h)
}

[inline] pub fn VecSet8i(a: i32, b: i32, c: i32, d: i32, e: i32, f: i32, g: i32, h: i32) Vec8i {
    return nr_simd_set8_i32(a, b, c, d, e, f, g, h)
}

extern fn nr_simd_storeu_f32(dst: ptr, v: Vec8f)
[inline] pub fn StoreUnaligned8f(dst: ptr, v: Vec8f) { nr_simd_storeu_f32(dst, v) }

// Phase 5: Shuffles, Swizzles, and Permutations
[cfg("target_feature=avx2")] extern fn nr_simd_permute_8f(a: Vec8f, idx: Vec8i) Vec8f
[cfg("target_feature=avx2")] extern fn nr_simd_permute_8i(a: Vec8i, idx: Vec8i) Vec8i
[cfg("target_feature=avx")] extern fn nr_simd_permute_4d(a: Vec4d, idx: Vec4i64) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_swizzle_8f(a: Vec8f, idx: Vec8i) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_unpacklo_8f(a: Vec8f, b: Vec8f) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_unpackhi_8f(a: Vec8f, b: Vec8f) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_unpacklo_4d(a: Vec4d, b: Vec4d) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_unpackhi_4d(a: Vec4d, b: Vec4d) Vec4d

[cfg("target_feature=avx2")] [inline] pub fn Permute8f(a: Vec8f, idx: Vec8i) Vec8f { return nr_simd_permute_8f(a, idx) }
[cfg("target_feature=avx2")] [inline] pub fn Permute8i(a: Vec8i, idx: Vec8i) Vec8i { return nr_simd_permute_8i(a, idx) }
[cfg("target_feature=avx")] [inline] pub fn Permute4d(a: Vec4d, idx: Vec4i64) Vec4d { return nr_simd_permute_4d(a, idx) }
[cfg("target_feature=avx")] [inline] pub fn Swizzle8f(a: Vec8f, idx: Vec8i) Vec8f { return nr_simd_swizzle_8f(a, idx) }
[cfg("target_feature=avx")] [inline] pub fn UnpackLo8f(a: Vec8f, b: Vec8f) Vec8f { return nr_simd_unpacklo_8f(a, b) }
[cfg("target_feature=avx")] [inline] pub fn UnpackHi8f(a: Vec8f, b: Vec8f) Vec8f { return nr_simd_unpackhi_8f(a, b) }
[cfg("target_feature=avx")] [inline] pub fn UnpackLo4d(a: Vec4d, b: Vec4d) Vec4d { return nr_simd_unpacklo_4d(a, b) }
[cfg("target_feature=avx")] [inline] pub fn UnpackHi4d(a: Vec4d, b: Vec4d) Vec4d { return nr_simd_unpackhi_4d(a, b) }

// Phase 6: Conversions & Horizontal Reductions
[cfg("target_feature=avx")] extern fn nr_simd_cvt_8f_to_8i(a: Vec8f) Vec8i
[cfg("target_feature=avx")] extern fn nr_simd_cvt_8i_to_8f(a: Vec8i) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_cvt_4f_to_4d(a: Vec4f) Vec4d
[cfg("target_feature=avx")] extern fn nr_simd_cvt_4d_to_4f(a: Vec4d) Vec4f
[cfg("target_feature=avx")] extern fn nr_simd_reduce_add_8f(a: Vec8f) f32
[cfg("target_feature=avx")] extern fn nr_simd_reduce_add_4d(a: Vec4d) f64
[cfg("target_feature=avx")] extern fn nr_simd_reduce_max_8f(a: Vec8f) f32

[cfg("target_feature=avx")] [inline] pub fn Cvt8fTo8i(a: Vec8f) Vec8i { return nr_simd_cvt_8f_to_8i(a) }
[cfg("target_feature=avx")] [inline] pub fn Cvt8iTo8f(a: Vec8i) Vec8f { return nr_simd_cvt_8i_to_8f(a) }
[cfg("target_feature=avx")] [inline] pub fn Cvt4fTo4d(a: Vec4f) Vec4d { return nr_simd_cvt_4f_to_4d(a) }
[cfg("target_feature=avx")] [inline] pub fn Cvt4dTo4f(a: Vec4d) Vec4f { return nr_simd_cvt_4d_to_4f(a) }
[cfg("target_feature=avx")] [inline] pub fn ReduceAdd8f(a: Vec8f) f32 { return nr_simd_reduce_add_8f(a) }
[cfg("target_feature=avx")] [inline] pub fn ReduceAdd4d(a: Vec4d) f64 { return nr_simd_reduce_add_4d(a) }
[cfg("target_feature=avx")] [inline] pub fn ReduceMax8f(a: Vec8f) f32 { return nr_simd_reduce_max_8f(a) }

// Phase 7: Transcendental Math Approximations
[cfg("target_feature=avx")] extern fn nr_simd_exp_8f(a: Vec8f) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_sin_8f(a: Vec8f) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_cos_8f(a: Vec8f) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_log2_8f(a: Vec8f) Vec8f
[cfg("target_feature=avx")] extern fn nr_simd_log_8f(a: Vec8f) Vec8f

[cfg("target_feature=avx")] [inline] pub fn Exp8f(a: Vec8f) Vec8f { return nr_simd_exp_8f(a) }
[cfg("target_feature=avx")] [inline] pub fn Sin8f(a: Vec8f) Vec8f { return nr_simd_sin_8f(a) }
[cfg("target_feature=avx")] [inline] pub fn Cos8f(a: Vec8f) Vec8f { return nr_simd_cos_8f(a) }
[cfg("target_feature=avx")] [inline] pub fn Log28f(a: Vec8f) Vec8f { return nr_simd_log2_8f(a) }
[cfg("target_feature=avx")] [inline] pub fn Log8f(a: Vec8f) Vec8f { return nr_simd_log_8f(a) }
