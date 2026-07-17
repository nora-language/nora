#ifndef NORA_SIMD_H
#define NORA_SIMD_H
#pragma GCC target("avx,avx2,fma")
#ifdef __clang__
#pragma clang attribute push (__attribute__((target("avx,avx2,fma"))), apply_to=function)
#endif
#include <x86intrin.h>
#include <stdint.h>

typedef __m256d simd_Vec4d;
typedef __m256 simd_Vec8f;
typedef __m128 simd_Vec4f;
typedef int simd_Vec8i __attribute__((__vector_size__(32)));
typedef int simd_Vec4i __attribute__((__vector_size__(16)));


// --- GENERATED TYPEDEFS ---
typedef char simd_Vec16i8 __attribute__((__vector_size__(16)));
typedef char simd_Vec32i8 __attribute__((__vector_size__(32)));
typedef unsigned char simd_Vec16u8 __attribute__((__vector_size__(16)));
typedef unsigned char simd_Vec32u8 __attribute__((__vector_size__(32)));
typedef short simd_Vec8i16 __attribute__((__vector_size__(16)));
typedef short simd_Vec16i16 __attribute__((__vector_size__(32)));
typedef unsigned short simd_Vec8u16 __attribute__((__vector_size__(16)));
typedef unsigned short simd_Vec16u16 __attribute__((__vector_size__(32)));
typedef unsigned int simd_Vec4u32 __attribute__((__vector_size__(16)));
typedef unsigned int simd_Vec8u32 __attribute__((__vector_size__(32)));
typedef long long simd_Vec2i64 __attribute__((__vector_size__(16)));
typedef long long simd_Vec4i64 __attribute__((__vector_size__(32)));
typedef unsigned long long simd_Vec2u64 __attribute__((__vector_size__(16)));
typedef unsigned long long simd_Vec4u64 __attribute__((__vector_size__(32)));

#if defined(__AVX512F__) || defined(__clang__)
typedef float simd_Vec16f __attribute__((__vector_size__(64)));
typedef double simd_Vec8d __attribute__((__vector_size__(64)));
typedef int simd_Vec16i32 __attribute__((__vector_size__(64)));
typedef unsigned int simd_Vec16u32 __attribute__((__vector_size__(64)));
typedef long long simd_Vec8i64 __attribute__((__vector_size__(64)));
typedef unsigned long long simd_Vec8u64 __attribute__((__vector_size__(64)));
typedef char simd_Vec64i8 __attribute__((__vector_size__(64)));
typedef unsigned char simd_Vec64u8 __attribute__((__vector_size__(64)));
typedef short simd_Vec32i16 __attribute__((__vector_size__(64)));
typedef unsigned short simd_Vec32u16 __attribute__((__vector_size__(64)));
#endif

static inline void* nr_simd_slice_ptr(double* slice, int index) {
    return &slice[index];
}

static inline simd_Vec4d nr_simd_load(void* src) {
    return _mm256_loadu_pd((double*)src);
}

static inline void nr_simd_store(void* dst, simd_Vec4d src) {
    _mm256_storeu_pd((double*)dst, src);
}

static inline simd_Vec4d nr_simd_set1(double val) {
    return _mm256_set1_pd(val);
}

static inline simd_Vec4d nr_simd_set4(double v0, double v1, double v2, double v3) {
    return _mm256_setr_pd(v0, v1, v2, v3);
}

static inline simd_Vec4d nr_simd_add(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_add_pd(a, b);
}

static inline simd_Vec4d nr_simd_sub(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_sub_pd(a, b);
}

static inline simd_Vec4d nr_simd_mul(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_mul_pd(a, b);
}

static inline simd_Vec4d nr_simd_div(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_div_pd(a, b);
}

static inline simd_Vec4d nr_simd_hadd(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_hadd_pd(a, b);
}

static inline simd_Vec4d nr_simd_blend_1100(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_blend_pd(a, b, 0b1100);
}

static inline simd_Vec4d nr_simd_blend_1010(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_blend_pd(a, b, 0b1010);
}

static inline simd_Vec4d nr_simd_blend_0011(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_blend_pd(a, b, 0b0011);
}

static inline simd_Vec4d nr_simd_permute_0101(simd_Vec4d a) {
    return _mm256_permute_pd(a, 0b0101);
}

static inline simd_Vec4d nr_simd_permute2f128_01(simd_Vec4d a) {
    return _mm256_permute2f128_pd(a, a, 0x01);
}

static inline simd_Vec4d nr_simd_permute2f128_1(simd_Vec4d a) {
    return _mm256_permute2f128_pd(a, a, 1);
}

static inline simd_Vec4d nr_simd_permute2f128_21(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_permute2f128_pd(a, b, 0x21);
}

static inline simd_Vec4d nr_simd_unpacklo(simd_Vec4d a) {
    return _mm256_unpacklo_pd(a, a);
}

static inline simd_Vec4d nr_simd_unpackhi(simd_Vec4d a) {
    return _mm256_unpackhi_pd(a, a);
}

static inline simd_Vec4d nr_simd_sqrt(simd_Vec4d a) {
    return _mm256_sqrt_pd(a);
}

static inline simd_Vec4d nr_simd_min(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_min_pd(a, b);
}

static inline simd_Vec4d nr_simd_max(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_max_pd(a, b);
}

static inline simd_Vec4d nr_simd_and(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_and_pd(a, b);
}

static inline simd_Vec4d nr_simd_or(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_or_pd(a, b);
}

static inline simd_Vec4d nr_simd_xor(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_xor_pd(a, b);
}

static inline simd_Vec4d nr_simd_andnot(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_andnot_pd(a, b);
}

// Approximation block from spectralnorm
static inline simd_Vec4d nr_simd_approx_recip(simd_Vec4d z) {
    __m128 q = _mm256_cvtpd_ps(z);
    q = _mm_rcp_ps(q);
    __m256d x = _mm256_cvtps_pd(q);
    __m256d w = _mm256_mul_pd(x, z);
    __m256d y = _mm256_set1_pd(3.0);
    z = _mm256_mul_pd(w, x);
    w = _mm256_sub_pd(w, y);
    x = _mm256_mul_pd(x, y);
    z = _mm256_mul_pd(z, w);
    z = _mm256_add_pd(z, x);
    return z;
}

// Vec8f (32-bit floats x 8)
static inline void* nr_simd_slice_ptr_f32(float* slice, int index) { return &slice[index]; }
static inline simd_Vec8f nr_simd_load_f32x8(void* src) { return _mm256_loadu_ps((float*)src); }
static inline void nr_simd_store_f32x8(void* dst, simd_Vec8f src) { _mm256_storeu_ps((float*)dst, src); }
static inline simd_Vec8f nr_simd_set1_f32x8(float val) { return _mm256_set1_ps(val); }
static inline simd_Vec8f nr_simd_add_f32x8(simd_Vec8f a, simd_Vec8f b) { return _mm256_add_ps(a, b); }
static inline simd_Vec8f nr_simd_sub_f32x8(simd_Vec8f a, simd_Vec8f b) { return _mm256_sub_ps(a, b); }
static inline simd_Vec8f nr_simd_mul_f32x8(simd_Vec8f a, simd_Vec8f b) { return _mm256_mul_ps(a, b); }
static inline simd_Vec8f nr_simd_div_f32x8(simd_Vec8f a, simd_Vec8f b) { return _mm256_div_ps(a, b); }

// Vec4f (32-bit floats x 4)
static inline simd_Vec4f nr_simd_load_f32x4(void* src) { return _mm_loadu_ps((float*)src); }
static inline void nr_simd_store_f32x4(void* dst, simd_Vec4f src) { _mm_storeu_ps((float*)dst, src); }
static inline simd_Vec4f nr_simd_set1_f32x4(float val) { return _mm_set1_ps(val); }
static inline simd_Vec4f nr_simd_add_f32x4(simd_Vec4f a, simd_Vec4f b) { return _mm_add_ps(a, b); }
static inline simd_Vec4f nr_simd_sub_f32x4(simd_Vec4f a, simd_Vec4f b) { return _mm_sub_ps(a, b); }
static inline simd_Vec4f nr_simd_mul_f32x4(simd_Vec4f a, simd_Vec4f b) { return _mm_mul_ps(a, b); }
static inline simd_Vec4f nr_simd_div_f32x4(simd_Vec4f a, simd_Vec4f b) { return _mm_div_ps(a, b); }

// Vec8i (32-bit integers x 8)
static inline void* nr_simd_slice_ptr_i32(int* slice, int index) { return &slice[index]; }
static inline simd_Vec8i nr_simd_load_i32x8(void* src) { return _mm256_loadu_si256((__m256i*)src); }
static inline void nr_simd_store_i32x8(void* dst, simd_Vec8i src) { _mm256_storeu_si256((__m256i*)dst, src); }
static inline simd_Vec8i nr_simd_set1_i32x8(int val) { return _mm256_set1_epi32(val); }
static inline simd_Vec8i nr_simd_add_i32x8(simd_Vec8i a, simd_Vec8i b) { return _mm256_add_epi32(a, b); }
static inline simd_Vec8i nr_simd_sub_i32x8(simd_Vec8i a, simd_Vec8i b) { return _mm256_sub_epi32(a, b); }
static inline simd_Vec8i nr_simd_mul_i32x8(simd_Vec8i a, simd_Vec8i b) { return _mm256_mullo_epi32(a, b); }

// Vec4i (32-bit integers x 4)
static inline simd_Vec4i nr_simd_load_i32x4(void* src) { return _mm_loadu_si128((__m128i*)src); }
static inline void nr_simd_store_i32x4(void* dst, simd_Vec4i src) { _mm_storeu_si128((__m128i*)dst, src); }
static inline simd_Vec4i nr_simd_set1_i32x4(int val) { return _mm_set1_epi32(val); }
static inline simd_Vec4i nr_simd_add_i32x4(simd_Vec4i a, simd_Vec4i b) { return _mm_add_epi32(a, b); }
static inline simd_Vec4i nr_simd_sub_i32x4(simd_Vec4i a, simd_Vec4i b) { return _mm_sub_epi32(a, b); }
static inline simd_Vec4i nr_simd_mul_i32x4(simd_Vec4i a, simd_Vec4i b) { return _mm_mullo_epi32(a, b); }

#ifdef __clang__
#pragma clang attribute pop
#endif


// --- GENERATED BINDINGS ---

static inline simd_Vec16i8 nr_simd_load_Vec16i8(void* src) {
    simd_Vec16i8 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec16i8));
    return res;
}
static inline void nr_simd_store_Vec16i8(void* dst, simd_Vec16i8 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec16i8));
}
static inline simd_Vec16i8 nr_simd_set1_Vec16i8(char val) {
    simd_Vec16i8 res;
    for (int i = 0; i < sizeof(simd_Vec16i8)/sizeof(char); i++) ((char*)&res)[i] = val;
    return res;
}

static inline simd_Vec32i8 nr_simd_load_Vec32i8(void* src) {
    simd_Vec32i8 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec32i8));
    return res;
}
static inline void nr_simd_store_Vec32i8(void* dst, simd_Vec32i8 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec32i8));
}
static inline simd_Vec32i8 nr_simd_set1_Vec32i8(char val) {
    simd_Vec32i8 res;
    for (int i = 0; i < sizeof(simd_Vec32i8)/sizeof(char); i++) ((char*)&res)[i] = val;
    return res;
}

static inline simd_Vec16u8 nr_simd_load_Vec16u8(void* src) {
    simd_Vec16u8 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec16u8));
    return res;
}
static inline void nr_simd_store_Vec16u8(void* dst, simd_Vec16u8 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec16u8));
}
static inline simd_Vec16u8 nr_simd_set1_Vec16u8(unsigned char val) {
    simd_Vec16u8 res;
    for (int i = 0; i < sizeof(simd_Vec16u8)/sizeof(unsigned char); i++) ((unsigned char*)&res)[i] = val;
    return res;
}

static inline simd_Vec32u8 nr_simd_load_Vec32u8(void* src) {
    simd_Vec32u8 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec32u8));
    return res;
}
static inline void nr_simd_store_Vec32u8(void* dst, simd_Vec32u8 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec32u8));
}
static inline simd_Vec32u8 nr_simd_set1_Vec32u8(unsigned char val) {
    simd_Vec32u8 res;
    for (int i = 0; i < sizeof(simd_Vec32u8)/sizeof(unsigned char); i++) ((unsigned char*)&res)[i] = val;
    return res;
}

static inline simd_Vec8i16 nr_simd_load_Vec8i16(void* src) {
    simd_Vec8i16 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec8i16));
    return res;
}
static inline void nr_simd_store_Vec8i16(void* dst, simd_Vec8i16 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec8i16));
}
static inline simd_Vec8i16 nr_simd_set1_Vec8i16(short val) {
    simd_Vec8i16 res;
    for (int i = 0; i < sizeof(simd_Vec8i16)/sizeof(short); i++) ((short*)&res)[i] = val;
    return res;
}

static inline simd_Vec16i16 nr_simd_load_Vec16i16(void* src) {
    simd_Vec16i16 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec16i16));
    return res;
}
static inline void nr_simd_store_Vec16i16(void* dst, simd_Vec16i16 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec16i16));
}
static inline simd_Vec16i16 nr_simd_set1_Vec16i16(short val) {
    simd_Vec16i16 res;
    for (int i = 0; i < sizeof(simd_Vec16i16)/sizeof(short); i++) ((short*)&res)[i] = val;
    return res;
}

static inline simd_Vec8u16 nr_simd_load_Vec8u16(void* src) {
    simd_Vec8u16 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec8u16));
    return res;
}
static inline void nr_simd_store_Vec8u16(void* dst, simd_Vec8u16 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec8u16));
}
static inline simd_Vec8u16 nr_simd_set1_Vec8u16(unsigned short val) {
    simd_Vec8u16 res;
    for (int i = 0; i < sizeof(simd_Vec8u16)/sizeof(unsigned short); i++) ((unsigned short*)&res)[i] = val;
    return res;
}

static inline simd_Vec16u16 nr_simd_load_Vec16u16(void* src) {
    simd_Vec16u16 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec16u16));
    return res;
}
static inline void nr_simd_store_Vec16u16(void* dst, simd_Vec16u16 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec16u16));
}
static inline simd_Vec16u16 nr_simd_set1_Vec16u16(unsigned short val) {
    simd_Vec16u16 res;
    for (int i = 0; i < sizeof(simd_Vec16u16)/sizeof(unsigned short); i++) ((unsigned short*)&res)[i] = val;
    return res;
}

static inline simd_Vec4u32 nr_simd_load_Vec4u32(void* src) {
    simd_Vec4u32 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec4u32));
    return res;
}
static inline void nr_simd_store_Vec4u32(void* dst, simd_Vec4u32 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec4u32));
}
static inline simd_Vec4u32 nr_simd_set1_Vec4u32(unsigned int val) {
    simd_Vec4u32 res;
    for (int i = 0; i < sizeof(simd_Vec4u32)/sizeof(unsigned int); i++) ((unsigned int*)&res)[i] = val;
    return res;
}

static inline simd_Vec8u32 nr_simd_load_Vec8u32(void* src) {
    simd_Vec8u32 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec8u32));
    return res;
}
static inline void nr_simd_store_Vec8u32(void* dst, simd_Vec8u32 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec8u32));
}
static inline simd_Vec8u32 nr_simd_set1_Vec8u32(unsigned int val) {
    simd_Vec8u32 res;
    for (int i = 0; i < sizeof(simd_Vec8u32)/sizeof(unsigned int); i++) ((unsigned int*)&res)[i] = val;
    return res;
}

static inline simd_Vec2i64 nr_simd_load_Vec2i64(void* src) {
    simd_Vec2i64 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec2i64));
    return res;
}
static inline void nr_simd_store_Vec2i64(void* dst, simd_Vec2i64 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec2i64));
}
static inline simd_Vec2i64 nr_simd_set1_Vec2i64(long long val) {
    simd_Vec2i64 res;
    for (int i = 0; i < sizeof(simd_Vec2i64)/sizeof(long long); i++) ((long long*)&res)[i] = val;
    return res;
}

static inline simd_Vec4i64 nr_simd_load_Vec4i64(void* src) {
    simd_Vec4i64 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec4i64));
    return res;
}
static inline void nr_simd_store_Vec4i64(void* dst, simd_Vec4i64 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec4i64));
}
static inline simd_Vec4i64 nr_simd_set1_Vec4i64(long long val) {
    simd_Vec4i64 res;
    for (int i = 0; i < sizeof(simd_Vec4i64)/sizeof(long long); i++) ((long long*)&res)[i] = val;
    return res;
}

static inline simd_Vec2u64 nr_simd_load_Vec2u64(void* src) {
    simd_Vec2u64 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec2u64));
    return res;
}
static inline void nr_simd_store_Vec2u64(void* dst, simd_Vec2u64 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec2u64));
}
static inline simd_Vec2u64 nr_simd_set1_Vec2u64(unsigned long long val) {
    simd_Vec2u64 res;
    for (int i = 0; i < sizeof(simd_Vec2u64)/sizeof(unsigned long long); i++) ((unsigned long long*)&res)[i] = val;
    return res;
}

static inline simd_Vec4u64 nr_simd_load_Vec4u64(void* src) {
    simd_Vec4u64 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec4u64));
    return res;
}
static inline void nr_simd_store_Vec4u64(void* dst, simd_Vec4u64 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec4u64));
}
static inline simd_Vec4u64 nr_simd_set1_Vec4u64(unsigned long long val) {
    simd_Vec4u64 res;
    for (int i = 0; i < sizeof(simd_Vec4u64)/sizeof(unsigned long long); i++) ((unsigned long long*)&res)[i] = val;
    return res;
}

#if defined(__AVX512F__) || defined(__clang__)
#ifdef __clang__
#pragma clang attribute push (__attribute__((target("avx512f"))), apply_to=function)
#endif

static inline simd_Vec16f nr_simd_load_Vec16f(void* src) {
    simd_Vec16f res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec16f));
    return res;
}
static inline void nr_simd_store_Vec16f(void* dst, simd_Vec16f src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec16f));
}
static inline simd_Vec16f nr_simd_set1_Vec16f(float val) {
    simd_Vec16f res;
    for (int i = 0; i < sizeof(simd_Vec16f)/sizeof(float); i++) ((float*)&res)[i] = val;
    return res;
}

static inline simd_Vec8d nr_simd_load_Vec8d(void* src) {
    simd_Vec8d res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec8d));
    return res;
}
static inline void nr_simd_store_Vec8d(void* dst, simd_Vec8d src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec8d));
}
static inline simd_Vec8d nr_simd_set1_Vec8d(double val) {
    simd_Vec8d res;
    for (int i = 0; i < sizeof(simd_Vec8d)/sizeof(double); i++) ((double*)&res)[i] = val;
    return res;
}

static inline simd_Vec16i32 nr_simd_load_Vec16i32(void* src) {
    simd_Vec16i32 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec16i32));
    return res;
}
static inline void nr_simd_store_Vec16i32(void* dst, simd_Vec16i32 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec16i32));
}
static inline simd_Vec16i32 nr_simd_set1_Vec16i32(int val) {
    simd_Vec16i32 res;
    for (int i = 0; i < sizeof(simd_Vec16i32)/sizeof(int); i++) ((int*)&res)[i] = val;
    return res;
}

static inline simd_Vec16u32 nr_simd_load_Vec16u32(void* src) {
    simd_Vec16u32 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec16u32));
    return res;
}
static inline void nr_simd_store_Vec16u32(void* dst, simd_Vec16u32 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec16u32));
}
static inline simd_Vec16u32 nr_simd_set1_Vec16u32(unsigned int val) {
    simd_Vec16u32 res;
    for (int i = 0; i < sizeof(simd_Vec16u32)/sizeof(unsigned int); i++) ((unsigned int*)&res)[i] = val;
    return res;
}

static inline simd_Vec8i64 nr_simd_load_Vec8i64(void* src) {
    simd_Vec8i64 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec8i64));
    return res;
}
static inline void nr_simd_store_Vec8i64(void* dst, simd_Vec8i64 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec8i64));
}
static inline simd_Vec8i64 nr_simd_set1_Vec8i64(long long val) {
    simd_Vec8i64 res;
    for (int i = 0; i < sizeof(simd_Vec8i64)/sizeof(long long); i++) ((long long*)&res)[i] = val;
    return res;
}

static inline simd_Vec8u64 nr_simd_load_Vec8u64(void* src) {
    simd_Vec8u64 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec8u64));
    return res;
}
static inline void nr_simd_store_Vec8u64(void* dst, simd_Vec8u64 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec8u64));
}
static inline simd_Vec8u64 nr_simd_set1_Vec8u64(unsigned long long val) {
    simd_Vec8u64 res;
    for (int i = 0; i < sizeof(simd_Vec8u64)/sizeof(unsigned long long); i++) ((unsigned long long*)&res)[i] = val;
    return res;
}

static inline simd_Vec64i8 nr_simd_load_Vec64i8(void* src) {
    simd_Vec64i8 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec64i8));
    return res;
}
static inline void nr_simd_store_Vec64i8(void* dst, simd_Vec64i8 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec64i8));
}
static inline simd_Vec64i8 nr_simd_set1_Vec64i8(char val) {
    simd_Vec64i8 res;
    for (int i = 0; i < sizeof(simd_Vec64i8)/sizeof(char); i++) ((char*)&res)[i] = val;
    return res;
}

static inline simd_Vec64u8 nr_simd_load_Vec64u8(void* src) {
    simd_Vec64u8 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec64u8));
    return res;
}
static inline void nr_simd_store_Vec64u8(void* dst, simd_Vec64u8 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec64u8));
}
static inline simd_Vec64u8 nr_simd_set1_Vec64u8(unsigned char val) {
    simd_Vec64u8 res;
    for (int i = 0; i < sizeof(simd_Vec64u8)/sizeof(unsigned char); i++) ((unsigned char*)&res)[i] = val;
    return res;
}

static inline simd_Vec32i16 nr_simd_load_Vec32i16(void* src) {
    simd_Vec32i16 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec32i16));
    return res;
}
static inline void nr_simd_store_Vec32i16(void* dst, simd_Vec32i16 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec32i16));
}
static inline simd_Vec32i16 nr_simd_set1_Vec32i16(short val) {
    simd_Vec32i16 res;
    for (int i = 0; i < sizeof(simd_Vec32i16)/sizeof(short); i++) ((short*)&res)[i] = val;
    return res;
}

static inline simd_Vec32u16 nr_simd_load_Vec32u16(void* src) {
    simd_Vec32u16 res;
    __builtin_memcpy(&res, src, sizeof(simd_Vec32u16));
    return res;
}
static inline void nr_simd_store_Vec32u16(void* dst, simd_Vec32u16 src) {
    __builtin_memcpy(dst, &src, sizeof(simd_Vec32u16));
}
static inline simd_Vec32u16 nr_simd_set1_Vec32u16(unsigned short val) {
    simd_Vec32u16 res;
    for (int i = 0; i < sizeof(simd_Vec32u16)/sizeof(unsigned short); i++) ((unsigned short*)&res)[i] = val;
    return res;
}
#ifdef __clang__
#pragma clang attribute pop
#endif
#endif


// --- PHASE 3: COMPARISONS & BLEND ---
static inline simd_Vec16i8 nr_simd_cmpeq_Vec16i8(simd_Vec16i8 a, simd_Vec16i8 b) { return (simd_Vec16i8)(a == b); }
static inline simd_Vec16i8 nr_simd_cmpneq_Vec16i8(simd_Vec16i8 a, simd_Vec16i8 b) { return (simd_Vec16i8)(a != b); }
static inline simd_Vec16i8 nr_simd_cmplt_Vec16i8(simd_Vec16i8 a, simd_Vec16i8 b) { return (simd_Vec16i8)(a < b); }
static inline simd_Vec16i8 nr_simd_cmpgt_Vec16i8(simd_Vec16i8 a, simd_Vec16i8 b) { return (simd_Vec16i8)(a > b); }
static inline simd_Vec16i8 nr_simd_blend_Vec16i8(simd_Vec16i8 mask, simd_Vec16i8 a, simd_Vec16i8 b) { return (simd_Vec16i8)(((simd_Vec16i8)a & mask) | ((simd_Vec16i8)b & ~mask)); }
static inline simd_Vec32i8 nr_simd_cmpeq_Vec32i8(simd_Vec32i8 a, simd_Vec32i8 b) { return (simd_Vec32i8)(a == b); }
static inline simd_Vec32i8 nr_simd_cmpneq_Vec32i8(simd_Vec32i8 a, simd_Vec32i8 b) { return (simd_Vec32i8)(a != b); }
static inline simd_Vec32i8 nr_simd_cmplt_Vec32i8(simd_Vec32i8 a, simd_Vec32i8 b) { return (simd_Vec32i8)(a < b); }
static inline simd_Vec32i8 nr_simd_cmpgt_Vec32i8(simd_Vec32i8 a, simd_Vec32i8 b) { return (simd_Vec32i8)(a > b); }
static inline simd_Vec32i8 nr_simd_blend_Vec32i8(simd_Vec32i8 mask, simd_Vec32i8 a, simd_Vec32i8 b) { return (simd_Vec32i8)(((simd_Vec32i8)a & mask) | ((simd_Vec32i8)b & ~mask)); }
static inline simd_Vec16i8 nr_simd_cmpeq_Vec16u8(simd_Vec16u8 a, simd_Vec16u8 b) { return (simd_Vec16i8)(a == b); }
static inline simd_Vec16i8 nr_simd_cmpneq_Vec16u8(simd_Vec16u8 a, simd_Vec16u8 b) { return (simd_Vec16i8)(a != b); }
static inline simd_Vec16i8 nr_simd_cmplt_Vec16u8(simd_Vec16u8 a, simd_Vec16u8 b) { return (simd_Vec16i8)(a < b); }
static inline simd_Vec16i8 nr_simd_cmpgt_Vec16u8(simd_Vec16u8 a, simd_Vec16u8 b) { return (simd_Vec16i8)(a > b); }
static inline simd_Vec16u8 nr_simd_blend_Vec16u8(simd_Vec16i8 mask, simd_Vec16u8 a, simd_Vec16u8 b) { return (simd_Vec16u8)(((simd_Vec16i8)a & mask) | ((simd_Vec16i8)b & ~mask)); }
static inline simd_Vec32i8 nr_simd_cmpeq_Vec32u8(simd_Vec32u8 a, simd_Vec32u8 b) { return (simd_Vec32i8)(a == b); }
static inline simd_Vec32i8 nr_simd_cmpneq_Vec32u8(simd_Vec32u8 a, simd_Vec32u8 b) { return (simd_Vec32i8)(a != b); }
static inline simd_Vec32i8 nr_simd_cmplt_Vec32u8(simd_Vec32u8 a, simd_Vec32u8 b) { return (simd_Vec32i8)(a < b); }
static inline simd_Vec32i8 nr_simd_cmpgt_Vec32u8(simd_Vec32u8 a, simd_Vec32u8 b) { return (simd_Vec32i8)(a > b); }
static inline simd_Vec32u8 nr_simd_blend_Vec32u8(simd_Vec32i8 mask, simd_Vec32u8 a, simd_Vec32u8 b) { return (simd_Vec32u8)(((simd_Vec32i8)a & mask) | ((simd_Vec32i8)b & ~mask)); }
static inline simd_Vec8i16 nr_simd_cmpeq_Vec8i16(simd_Vec8i16 a, simd_Vec8i16 b) { return (simd_Vec8i16)(a == b); }
static inline simd_Vec8i16 nr_simd_cmpneq_Vec8i16(simd_Vec8i16 a, simd_Vec8i16 b) { return (simd_Vec8i16)(a != b); }
static inline simd_Vec8i16 nr_simd_cmplt_Vec8i16(simd_Vec8i16 a, simd_Vec8i16 b) { return (simd_Vec8i16)(a < b); }
static inline simd_Vec8i16 nr_simd_cmpgt_Vec8i16(simd_Vec8i16 a, simd_Vec8i16 b) { return (simd_Vec8i16)(a > b); }
static inline simd_Vec8i16 nr_simd_blend_Vec8i16(simd_Vec8i16 mask, simd_Vec8i16 a, simd_Vec8i16 b) { return (simd_Vec8i16)(((simd_Vec8i16)a & mask) | ((simd_Vec8i16)b & ~mask)); }
static inline simd_Vec16i16 nr_simd_cmpeq_Vec16i16(simd_Vec16i16 a, simd_Vec16i16 b) { return (simd_Vec16i16)(a == b); }
static inline simd_Vec16i16 nr_simd_cmpneq_Vec16i16(simd_Vec16i16 a, simd_Vec16i16 b) { return (simd_Vec16i16)(a != b); }
static inline simd_Vec16i16 nr_simd_cmplt_Vec16i16(simd_Vec16i16 a, simd_Vec16i16 b) { return (simd_Vec16i16)(a < b); }
static inline simd_Vec16i16 nr_simd_cmpgt_Vec16i16(simd_Vec16i16 a, simd_Vec16i16 b) { return (simd_Vec16i16)(a > b); }
static inline simd_Vec16i16 nr_simd_blend_Vec16i16(simd_Vec16i16 mask, simd_Vec16i16 a, simd_Vec16i16 b) { return (simd_Vec16i16)(((simd_Vec16i16)a & mask) | ((simd_Vec16i16)b & ~mask)); }
static inline simd_Vec8i16 nr_simd_cmpeq_Vec8u16(simd_Vec8u16 a, simd_Vec8u16 b) { return (simd_Vec8i16)(a == b); }
static inline simd_Vec8i16 nr_simd_cmpneq_Vec8u16(simd_Vec8u16 a, simd_Vec8u16 b) { return (simd_Vec8i16)(a != b); }
static inline simd_Vec8i16 nr_simd_cmplt_Vec8u16(simd_Vec8u16 a, simd_Vec8u16 b) { return (simd_Vec8i16)(a < b); }
static inline simd_Vec8i16 nr_simd_cmpgt_Vec8u16(simd_Vec8u16 a, simd_Vec8u16 b) { return (simd_Vec8i16)(a > b); }
static inline simd_Vec8u16 nr_simd_blend_Vec8u16(simd_Vec8i16 mask, simd_Vec8u16 a, simd_Vec8u16 b) { return (simd_Vec8u16)(((simd_Vec8i16)a & mask) | ((simd_Vec8i16)b & ~mask)); }
static inline simd_Vec16i16 nr_simd_cmpeq_Vec16u16(simd_Vec16u16 a, simd_Vec16u16 b) { return (simd_Vec16i16)(a == b); }
static inline simd_Vec16i16 nr_simd_cmpneq_Vec16u16(simd_Vec16u16 a, simd_Vec16u16 b) { return (simd_Vec16i16)(a != b); }
static inline simd_Vec16i16 nr_simd_cmplt_Vec16u16(simd_Vec16u16 a, simd_Vec16u16 b) { return (simd_Vec16i16)(a < b); }
static inline simd_Vec16i16 nr_simd_cmpgt_Vec16u16(simd_Vec16u16 a, simd_Vec16u16 b) { return (simd_Vec16i16)(a > b); }
static inline simd_Vec16u16 nr_simd_blend_Vec16u16(simd_Vec16i16 mask, simd_Vec16u16 a, simd_Vec16u16 b) { return (simd_Vec16u16)(((simd_Vec16i16)a & mask) | ((simd_Vec16i16)b & ~mask)); }
static inline simd_Vec4i nr_simd_cmpeq_Vec4u32(simd_Vec4u32 a, simd_Vec4u32 b) { return (simd_Vec4i)(a == b); }
static inline simd_Vec4i nr_simd_cmpneq_Vec4u32(simd_Vec4u32 a, simd_Vec4u32 b) { return (simd_Vec4i)(a != b); }
static inline simd_Vec4i nr_simd_cmplt_Vec4u32(simd_Vec4u32 a, simd_Vec4u32 b) { return (simd_Vec4i)(a < b); }
static inline simd_Vec4i nr_simd_cmpgt_Vec4u32(simd_Vec4u32 a, simd_Vec4u32 b) { return (simd_Vec4i)(a > b); }
static inline simd_Vec4u32 nr_simd_blend_Vec4u32(simd_Vec4i mask, simd_Vec4u32 a, simd_Vec4u32 b) { return (simd_Vec4u32)(((simd_Vec4i)a & mask) | ((simd_Vec4i)b & ~mask)); }
static inline simd_Vec8i nr_simd_cmpeq_Vec8u32(simd_Vec8u32 a, simd_Vec8u32 b) { return (simd_Vec8i)(a == b); }
static inline simd_Vec8i nr_simd_cmpneq_Vec8u32(simd_Vec8u32 a, simd_Vec8u32 b) { return (simd_Vec8i)(a != b); }
static inline simd_Vec8i nr_simd_cmplt_Vec8u32(simd_Vec8u32 a, simd_Vec8u32 b) { return (simd_Vec8i)(a < b); }
static inline simd_Vec8i nr_simd_cmpgt_Vec8u32(simd_Vec8u32 a, simd_Vec8u32 b) { return (simd_Vec8i)(a > b); }
static inline simd_Vec8u32 nr_simd_blend_Vec8u32(simd_Vec8i mask, simd_Vec8u32 a, simd_Vec8u32 b) { return (simd_Vec8u32)(((simd_Vec8i)a & mask) | ((simd_Vec8i)b & ~mask)); }
static inline simd_Vec2i64 nr_simd_cmpeq_Vec2i64(simd_Vec2i64 a, simd_Vec2i64 b) { return (simd_Vec2i64)(a == b); }
static inline simd_Vec2i64 nr_simd_cmpneq_Vec2i64(simd_Vec2i64 a, simd_Vec2i64 b) { return (simd_Vec2i64)(a != b); }
static inline simd_Vec2i64 nr_simd_cmplt_Vec2i64(simd_Vec2i64 a, simd_Vec2i64 b) { return (simd_Vec2i64)(a < b); }
static inline simd_Vec2i64 nr_simd_cmpgt_Vec2i64(simd_Vec2i64 a, simd_Vec2i64 b) { return (simd_Vec2i64)(a > b); }
static inline simd_Vec2i64 nr_simd_blend_Vec2i64(simd_Vec2i64 mask, simd_Vec2i64 a, simd_Vec2i64 b) { return (simd_Vec2i64)(((simd_Vec2i64)a & mask) | ((simd_Vec2i64)b & ~mask)); }
static inline simd_Vec4i64 nr_simd_cmpeq_Vec4i64(simd_Vec4i64 a, simd_Vec4i64 b) { return (simd_Vec4i64)(a == b); }
static inline simd_Vec4i64 nr_simd_cmpneq_Vec4i64(simd_Vec4i64 a, simd_Vec4i64 b) { return (simd_Vec4i64)(a != b); }
static inline simd_Vec4i64 nr_simd_cmplt_Vec4i64(simd_Vec4i64 a, simd_Vec4i64 b) { return (simd_Vec4i64)(a < b); }
static inline simd_Vec4i64 nr_simd_cmpgt_Vec4i64(simd_Vec4i64 a, simd_Vec4i64 b) { return (simd_Vec4i64)(a > b); }
static inline simd_Vec4i64 nr_simd_blend_Vec4i64(simd_Vec4i64 mask, simd_Vec4i64 a, simd_Vec4i64 b) { return (simd_Vec4i64)(((simd_Vec4i64)a & mask) | ((simd_Vec4i64)b & ~mask)); }
static inline simd_Vec2i64 nr_simd_cmpeq_Vec2u64(simd_Vec2u64 a, simd_Vec2u64 b) { return (simd_Vec2i64)(a == b); }
static inline simd_Vec2i64 nr_simd_cmpneq_Vec2u64(simd_Vec2u64 a, simd_Vec2u64 b) { return (simd_Vec2i64)(a != b); }
static inline simd_Vec2i64 nr_simd_cmplt_Vec2u64(simd_Vec2u64 a, simd_Vec2u64 b) { return (simd_Vec2i64)(a < b); }
static inline simd_Vec2i64 nr_simd_cmpgt_Vec2u64(simd_Vec2u64 a, simd_Vec2u64 b) { return (simd_Vec2i64)(a > b); }
static inline simd_Vec2u64 nr_simd_blend_Vec2u64(simd_Vec2i64 mask, simd_Vec2u64 a, simd_Vec2u64 b) { return (simd_Vec2u64)(((simd_Vec2i64)a & mask) | ((simd_Vec2i64)b & ~mask)); }
static inline simd_Vec4i64 nr_simd_cmpeq_Vec4u64(simd_Vec4u64 a, simd_Vec4u64 b) { return (simd_Vec4i64)(a == b); }
static inline simd_Vec4i64 nr_simd_cmpneq_Vec4u64(simd_Vec4u64 a, simd_Vec4u64 b) { return (simd_Vec4i64)(a != b); }
static inline simd_Vec4i64 nr_simd_cmplt_Vec4u64(simd_Vec4u64 a, simd_Vec4u64 b) { return (simd_Vec4i64)(a < b); }
static inline simd_Vec4i64 nr_simd_cmpgt_Vec4u64(simd_Vec4u64 a, simd_Vec4u64 b) { return (simd_Vec4i64)(a > b); }
static inline simd_Vec4u64 nr_simd_blend_Vec4u64(simd_Vec4i64 mask, simd_Vec4u64 a, simd_Vec4u64 b) { return (simd_Vec4u64)(((simd_Vec4i64)a & mask) | ((simd_Vec4i64)b & ~mask)); }
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpeq_Vec16f(simd_Vec16f a, simd_Vec16f b) { return (simd_Vec16i32)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpneq_Vec16f(simd_Vec16f a, simd_Vec16f b) { return (simd_Vec16i32)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmplt_Vec16f(simd_Vec16f a, simd_Vec16f b) { return (simd_Vec16i32)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpgt_Vec16f(simd_Vec16f a, simd_Vec16f b) { return (simd_Vec16i32)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16f nr_simd_blend_Vec16f(simd_Vec16i32 mask, simd_Vec16f a, simd_Vec16f b) { return (simd_Vec16f)(((simd_Vec16i32)a & mask) | ((simd_Vec16i32)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpeq_Vec8d(simd_Vec8d a, simd_Vec8d b) { return (simd_Vec8i64)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpneq_Vec8d(simd_Vec8d a, simd_Vec8d b) { return (simd_Vec8i64)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmplt_Vec8d(simd_Vec8d a, simd_Vec8d b) { return (simd_Vec8i64)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpgt_Vec8d(simd_Vec8d a, simd_Vec8d b) { return (simd_Vec8i64)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8d nr_simd_blend_Vec8d(simd_Vec8i64 mask, simd_Vec8d a, simd_Vec8d b) { return (simd_Vec8d)(((simd_Vec8i64)a & mask) | ((simd_Vec8i64)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpeq_Vec16i32(simd_Vec16i32 a, simd_Vec16i32 b) { return (simd_Vec16i32)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpneq_Vec16i32(simd_Vec16i32 a, simd_Vec16i32 b) { return (simd_Vec16i32)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmplt_Vec16i32(simd_Vec16i32 a, simd_Vec16i32 b) { return (simd_Vec16i32)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpgt_Vec16i32(simd_Vec16i32 a, simd_Vec16i32 b) { return (simd_Vec16i32)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_blend_Vec16i32(simd_Vec16i32 mask, simd_Vec16i32 a, simd_Vec16i32 b) { return (simd_Vec16i32)(((simd_Vec16i32)a & mask) | ((simd_Vec16i32)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpeq_Vec16u32(simd_Vec16u32 a, simd_Vec16u32 b) { return (simd_Vec16i32)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpneq_Vec16u32(simd_Vec16u32 a, simd_Vec16u32 b) { return (simd_Vec16i32)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmplt_Vec16u32(simd_Vec16u32 a, simd_Vec16u32 b) { return (simd_Vec16i32)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16i32 nr_simd_cmpgt_Vec16u32(simd_Vec16u32 a, simd_Vec16u32 b) { return (simd_Vec16i32)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec16u32 nr_simd_blend_Vec16u32(simd_Vec16i32 mask, simd_Vec16u32 a, simd_Vec16u32 b) { return (simd_Vec16u32)(((simd_Vec16i32)a & mask) | ((simd_Vec16i32)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpeq_Vec8i64(simd_Vec8i64 a, simd_Vec8i64 b) { return (simd_Vec8i64)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpneq_Vec8i64(simd_Vec8i64 a, simd_Vec8i64 b) { return (simd_Vec8i64)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmplt_Vec8i64(simd_Vec8i64 a, simd_Vec8i64 b) { return (simd_Vec8i64)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpgt_Vec8i64(simd_Vec8i64 a, simd_Vec8i64 b) { return (simd_Vec8i64)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_blend_Vec8i64(simd_Vec8i64 mask, simd_Vec8i64 a, simd_Vec8i64 b) { return (simd_Vec8i64)(((simd_Vec8i64)a & mask) | ((simd_Vec8i64)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpeq_Vec8u64(simd_Vec8u64 a, simd_Vec8u64 b) { return (simd_Vec8i64)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpneq_Vec8u64(simd_Vec8u64 a, simd_Vec8u64 b) { return (simd_Vec8i64)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmplt_Vec8u64(simd_Vec8u64 a, simd_Vec8u64 b) { return (simd_Vec8i64)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8i64 nr_simd_cmpgt_Vec8u64(simd_Vec8u64 a, simd_Vec8u64 b) { return (simd_Vec8i64)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec8u64 nr_simd_blend_Vec8u64(simd_Vec8i64 mask, simd_Vec8u64 a, simd_Vec8u64 b) { return (simd_Vec8u64)(((simd_Vec8i64)a & mask) | ((simd_Vec8i64)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmpeq_Vec64i8(simd_Vec64i8 a, simd_Vec64i8 b) { return (simd_Vec64i8)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmpneq_Vec64i8(simd_Vec64i8 a, simd_Vec64i8 b) { return (simd_Vec64i8)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmplt_Vec64i8(simd_Vec64i8 a, simd_Vec64i8 b) { return (simd_Vec64i8)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmpgt_Vec64i8(simd_Vec64i8 a, simd_Vec64i8 b) { return (simd_Vec64i8)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_blend_Vec64i8(simd_Vec64i8 mask, simd_Vec64i8 a, simd_Vec64i8 b) { return (simd_Vec64i8)(((simd_Vec64i8)a & mask) | ((simd_Vec64i8)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmpeq_Vec64u8(simd_Vec64u8 a, simd_Vec64u8 b) { return (simd_Vec64i8)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmpneq_Vec64u8(simd_Vec64u8 a, simd_Vec64u8 b) { return (simd_Vec64i8)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmplt_Vec64u8(simd_Vec64u8 a, simd_Vec64u8 b) { return (simd_Vec64i8)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64i8 nr_simd_cmpgt_Vec64u8(simd_Vec64u8 a, simd_Vec64u8 b) { return (simd_Vec64i8)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec64u8 nr_simd_blend_Vec64u8(simd_Vec64i8 mask, simd_Vec64u8 a, simd_Vec64u8 b) { return (simd_Vec64u8)(((simd_Vec64i8)a & mask) | ((simd_Vec64i8)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmpeq_Vec32i16(simd_Vec32i16 a, simd_Vec32i16 b) { return (simd_Vec32i16)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmpneq_Vec32i16(simd_Vec32i16 a, simd_Vec32i16 b) { return (simd_Vec32i16)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmplt_Vec32i16(simd_Vec32i16 a, simd_Vec32i16 b) { return (simd_Vec32i16)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmpgt_Vec32i16(simd_Vec32i16 a, simd_Vec32i16 b) { return (simd_Vec32i16)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_blend_Vec32i16(simd_Vec32i16 mask, simd_Vec32i16 a, simd_Vec32i16 b) { return (simd_Vec32i16)(((simd_Vec32i16)a & mask) | ((simd_Vec32i16)b & ~mask)); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmpeq_Vec32u16(simd_Vec32u16 a, simd_Vec32u16 b) { return (simd_Vec32i16)(a == b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmpneq_Vec32u16(simd_Vec32u16 a, simd_Vec32u16 b) { return (simd_Vec32i16)(a != b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmplt_Vec32u16(simd_Vec32u16 a, simd_Vec32u16 b) { return (simd_Vec32i16)(a < b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32i16 nr_simd_cmpgt_Vec32u16(simd_Vec32u16 a, simd_Vec32u16 b) { return (simd_Vec32i16)(a > b); }
#endif
#if defined(__AVX512F__) || defined(__clang__)
static inline simd_Vec32u16 nr_simd_blend_Vec32u16(simd_Vec32i16 mask, simd_Vec32u16 a, simd_Vec32u16 b) { return (simd_Vec32u16)(((simd_Vec32i16)a & mask) | ((simd_Vec32i16)b & ~mask)); }
#endif

#if defined(__AVX512F__)
static inline simd_Vec8d nr_simd_set8(double v0, double v1, double v2, double v3, double v4, double v5, double v6, double v7) {
    return _mm512_setr_pd(v0, v1, v2, v3, v4, v5, v6, v7);
}
static inline simd_Vec8d nr_simd_setzero8d() {
    return _mm512_setzero_pd();
}
static inline simd_Vec8d nr_simd_sqrt8d(simd_Vec8d a) {
    return _mm512_sqrt_pd(a);
}
static inline simd_Vec8d nr_simd_min8d(simd_Vec8d a, simd_Vec8d b) {
    return _mm512_min_pd(a, b);
}
static inline simd_Vec8d nr_simd_max8d(simd_Vec8d a, simd_Vec8d b) {
    return _mm512_max_pd(a, b);
}
static inline simd_Vec8d nr_simd_and8d(simd_Vec8d a, simd_Vec8d b) {
    return _mm512_and_pd(a, b);
}
static inline simd_Vec8d nr_simd_or8d(simd_Vec8d a, simd_Vec8d b) {
    return _mm512_or_pd(a, b);
}
static inline simd_Vec8d nr_simd_xor8d(simd_Vec8d a, simd_Vec8d b) {
    return _mm512_xor_pd(a, b);
}
static inline simd_Vec8d nr_simd_andnot8d(simd_Vec8d a, simd_Vec8d b) {
    return _mm512_andnot_pd(a, b);
}
static inline double nr_simd_reduce_add8d(simd_Vec8d a) {
    return _mm512_reduce_add_pd(a);
}
#endif

static inline simd_Vec4i64 nr_simd_cmpeq_Vec4d(simd_Vec4d a, simd_Vec4d b) { return (simd_Vec4i64)(a == b); }
static inline simd_Vec4i64 nr_simd_cmpneq_Vec4d(simd_Vec4d a, simd_Vec4d b) { return (simd_Vec4i64)(a != b); }
static inline simd_Vec4i64 nr_simd_cmplt_Vec4d(simd_Vec4d a, simd_Vec4d b) { return (simd_Vec4i64)(a < b); }
static inline simd_Vec4i64 nr_simd_cmpgt_Vec4d(simd_Vec4d a, simd_Vec4d b) { return (simd_Vec4i64)(a > b); }
static inline simd_Vec4d nr_simd_blend_Vec4d(simd_Vec4i64 mask, simd_Vec4d a, simd_Vec4d b) { return (simd_Vec4d)(((simd_Vec4i64)a & mask) | ((simd_Vec4i64)b & ~mask)); }
static inline simd_Vec8i nr_simd_cmpeq_Vec8f(simd_Vec8f a, simd_Vec8f b) { return (simd_Vec8i)(a == b); }
static inline simd_Vec8i nr_simd_cmpneq_Vec8f(simd_Vec8f a, simd_Vec8f b) { return (simd_Vec8i)(a != b); }
static inline simd_Vec8i nr_simd_cmplt_Vec8f(simd_Vec8f a, simd_Vec8f b) { return (simd_Vec8i)(a < b); }
static inline simd_Vec8i nr_simd_cmpgt_Vec8f(simd_Vec8f a, simd_Vec8f b) { return (simd_Vec8i)(a > b); }
static inline simd_Vec8f nr_simd_blend_Vec8f(simd_Vec8i mask, simd_Vec8f a, simd_Vec8f b) { return (simd_Vec8f)(((simd_Vec8i)a & mask) | ((simd_Vec8i)b & ~mask)); }
static inline simd_Vec4i nr_simd_cmpeq_Vec4f(simd_Vec4f a, simd_Vec4f b) { return (simd_Vec4i)(a == b); }
static inline simd_Vec4i nr_simd_cmpneq_Vec4f(simd_Vec4f a, simd_Vec4f b) { return (simd_Vec4i)(a != b); }
static inline simd_Vec4i nr_simd_cmplt_Vec4f(simd_Vec4f a, simd_Vec4f b) { return (simd_Vec4i)(a < b); }
static inline simd_Vec4i nr_simd_cmpgt_Vec4f(simd_Vec4f a, simd_Vec4f b) { return (simd_Vec4i)(a > b); }
static inline simd_Vec4f nr_simd_blend_Vec4f(simd_Vec4i mask, simd_Vec4f a, simd_Vec4f b) { return (simd_Vec4f)(((simd_Vec4i)a & mask) | ((simd_Vec4i)b & ~mask)); }
static inline simd_Vec8i nr_simd_cmpeq_Vec8i(simd_Vec8i a, simd_Vec8i b) { return (simd_Vec8i)(a == b); }
static inline simd_Vec8i nr_simd_cmpneq_Vec8i(simd_Vec8i a, simd_Vec8i b) { return (simd_Vec8i)(a != b); }
static inline simd_Vec8i nr_simd_cmplt_Vec8i(simd_Vec8i a, simd_Vec8i b) { return (simd_Vec8i)(a < b); }
static inline simd_Vec8i nr_simd_cmpgt_Vec8i(simd_Vec8i a, simd_Vec8i b) { return (simd_Vec8i)(a > b); }
static inline simd_Vec8i nr_simd_blend_Vec8i(simd_Vec8i mask, simd_Vec8i a, simd_Vec8i b) { return (simd_Vec8i)(((simd_Vec8i)a & mask) | ((simd_Vec8i)b & ~mask)); }
static inline simd_Vec4i nr_simd_cmpeq_Vec4i(simd_Vec4i a, simd_Vec4i b) { return (simd_Vec4i)(a == b); }
static inline simd_Vec4i nr_simd_cmpneq_Vec4i(simd_Vec4i a, simd_Vec4i b) { return (simd_Vec4i)(a != b); }
static inline simd_Vec4i nr_simd_cmplt_Vec4i(simd_Vec4i a, simd_Vec4i b) { return (simd_Vec4i)(a < b); }
static inline simd_Vec4i nr_simd_cmpgt_Vec4i(simd_Vec4i a, simd_Vec4i b) { return (simd_Vec4i)(a > b); }
static inline simd_Vec4i nr_simd_blend_Vec4i(simd_Vec4i mask, simd_Vec4i a, simd_Vec4i b) { return (simd_Vec4i)(((simd_Vec4i)a & mask) | ((simd_Vec4i)b & ~mask)); }


// --- Phase 4 ---
static inline simd_Vec16i8 nr_simd_load_aligned_16i8(void* src) { return *(simd_Vec16i8*)src; }
static inline void nr_simd_store_aligned_16i8(void* dst, simd_Vec16i8 v) { *(simd_Vec16i8*)dst = v; }
static inline simd_Vec16i8 nr_simd_gather_16i8(void* src, simd_Vec16i32 indices) {
    simd_Vec16i8 res;
    for(int i=0; i<16; ++i) res[i] = ((int8_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_16i8(void* dst, simd_Vec16i32 indices, simd_Vec16i8 v) {
    for(int i=0; i<16; ++i) ((int8_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec32i8 nr_simd_load_aligned_32i8(void* src) { return *(simd_Vec32i8*)src; }
static inline void nr_simd_store_aligned_32i8(void* dst, simd_Vec32i8 v) { *(simd_Vec32i8*)dst = v; }
static inline simd_Vec16u8 nr_simd_load_aligned_16u8(void* src) { return *(simd_Vec16u8*)src; }
static inline void nr_simd_store_aligned_16u8(void* dst, simd_Vec16u8 v) { *(simd_Vec16u8*)dst = v; }
static inline simd_Vec16u8 nr_simd_gather_16u8(void* src, simd_Vec16i32 indices) {
    simd_Vec16u8 res;
    for(int i=0; i<16; ++i) res[i] = ((uint8_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_16u8(void* dst, simd_Vec16i32 indices, simd_Vec16u8 v) {
    for(int i=0; i<16; ++i) ((uint8_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec32u8 nr_simd_load_aligned_32u8(void* src) { return *(simd_Vec32u8*)src; }
static inline void nr_simd_store_aligned_32u8(void* dst, simd_Vec32u8 v) { *(simd_Vec32u8*)dst = v; }
static inline simd_Vec8i16 nr_simd_load_aligned_8i16(void* src) { return *(simd_Vec8i16*)src; }
static inline void nr_simd_store_aligned_8i16(void* dst, simd_Vec8i16 v) { *(simd_Vec8i16*)dst = v; }
static inline simd_Vec8i16 nr_simd_gather_8i16(void* src, simd_Vec8i indices) {
    simd_Vec8i16 res;
    for(int i=0; i<8; ++i) res[i] = ((int16_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8i16(void* dst, simd_Vec8i indices, simd_Vec8i16 v) {
    for(int i=0; i<8; ++i) ((int16_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec16i16 nr_simd_load_aligned_16i16(void* src) { return *(simd_Vec16i16*)src; }
static inline void nr_simd_store_aligned_16i16(void* dst, simd_Vec16i16 v) { *(simd_Vec16i16*)dst = v; }
static inline simd_Vec16i16 nr_simd_gather_16i16(void* src, simd_Vec16i32 indices) {
    simd_Vec16i16 res;
    for(int i=0; i<16; ++i) res[i] = ((int16_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_16i16(void* dst, simd_Vec16i32 indices, simd_Vec16i16 v) {
    for(int i=0; i<16; ++i) ((int16_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec8u16 nr_simd_load_aligned_8u16(void* src) { return *(simd_Vec8u16*)src; }
static inline void nr_simd_store_aligned_8u16(void* dst, simd_Vec8u16 v) { *(simd_Vec8u16*)dst = v; }
static inline simd_Vec8u16 nr_simd_gather_8u16(void* src, simd_Vec8i indices) {
    simd_Vec8u16 res;
    for(int i=0; i<8; ++i) res[i] = ((uint16_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8u16(void* dst, simd_Vec8i indices, simd_Vec8u16 v) {
    for(int i=0; i<8; ++i) ((uint16_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec16u16 nr_simd_load_aligned_16u16(void* src) { return *(simd_Vec16u16*)src; }
static inline void nr_simd_store_aligned_16u16(void* dst, simd_Vec16u16 v) { *(simd_Vec16u16*)dst = v; }
static inline simd_Vec16u16 nr_simd_gather_16u16(void* src, simd_Vec16i32 indices) {
    simd_Vec16u16 res;
    for(int i=0; i<16; ++i) res[i] = ((uint16_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_16u16(void* dst, simd_Vec16i32 indices, simd_Vec16u16 v) {
    for(int i=0; i<16; ++i) ((uint16_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec4u32 nr_simd_load_aligned_4u32(void* src) { return *(simd_Vec4u32*)src; }
static inline void nr_simd_store_aligned_4u32(void* dst, simd_Vec4u32 v) { *(simd_Vec4u32*)dst = v; }
static inline simd_Vec4u32 nr_simd_gather_4u32(void* src, simd_Vec4i indices) {
    simd_Vec4u32 res;
    for(int i=0; i<4; ++i) res[i] = ((uint32_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_4u32(void* dst, simd_Vec4i indices, simd_Vec4u32 v) {
    for(int i=0; i<4; ++i) ((uint32_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec8u32 nr_simd_load_aligned_8u32(void* src) { return *(simd_Vec8u32*)src; }
static inline void nr_simd_store_aligned_8u32(void* dst, simd_Vec8u32 v) { *(simd_Vec8u32*)dst = v; }
static inline simd_Vec8u32 nr_simd_gather_8u32(void* src, simd_Vec8i indices) {
    simd_Vec8u32 res;
    for(int i=0; i<8; ++i) res[i] = ((uint32_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8u32(void* dst, simd_Vec8i indices, simd_Vec8u32 v) {
    for(int i=0; i<8; ++i) ((uint32_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec2i64 nr_simd_load_aligned_2i64(void* src) { return *(simd_Vec2i64*)src; }
static inline void nr_simd_store_aligned_2i64(void* dst, simd_Vec2i64 v) { *(simd_Vec2i64*)dst = v; }
static inline simd_Vec2i64 nr_simd_gather_2i64(void* src, simd_Vec2i64 indices) {
    simd_Vec2i64 res;
    for(int i=0; i<2; ++i) res[i] = ((int64_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_2i64(void* dst, simd_Vec2i64 indices, simd_Vec2i64 v) {
    for(int i=0; i<2; ++i) ((int64_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec4i64 nr_simd_load_aligned_4i64(void* src) { return *(simd_Vec4i64*)src; }
static inline void nr_simd_store_aligned_4i64(void* dst, simd_Vec4i64 v) { *(simd_Vec4i64*)dst = v; }
static inline simd_Vec4i64 nr_simd_gather_4i64(void* src, simd_Vec4i64 indices) {
    simd_Vec4i64 res;
    for(int i=0; i<4; ++i) res[i] = ((int64_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_4i64(void* dst, simd_Vec4i64 indices, simd_Vec4i64 v) {
    for(int i=0; i<4; ++i) ((int64_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec2u64 nr_simd_load_aligned_2u64(void* src) { return *(simd_Vec2u64*)src; }
static inline void nr_simd_store_aligned_2u64(void* dst, simd_Vec2u64 v) { *(simd_Vec2u64*)dst = v; }
static inline simd_Vec2u64 nr_simd_gather_2u64(void* src, simd_Vec2i64 indices) {
    simd_Vec2u64 res;
    for(int i=0; i<2; ++i) res[i] = ((uint64_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_2u64(void* dst, simd_Vec2i64 indices, simd_Vec2u64 v) {
    for(int i=0; i<2; ++i) ((uint64_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec4u64 nr_simd_load_aligned_4u64(void* src) { return *(simd_Vec4u64*)src; }
static inline void nr_simd_store_aligned_4u64(void* dst, simd_Vec4u64 v) { *(simd_Vec4u64*)dst = v; }
static inline simd_Vec4u64 nr_simd_gather_4u64(void* src, simd_Vec4i64 indices) {
    simd_Vec4u64 res;
    for(int i=0; i<4; ++i) res[i] = ((uint64_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_4u64(void* dst, simd_Vec4i64 indices, simd_Vec4u64 v) {
    for(int i=0; i<4; ++i) ((uint64_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec16f nr_simd_load_aligned_16f(void* src) { return *(simd_Vec16f*)src; }
static inline void nr_simd_store_aligned_16f(void* dst, simd_Vec16f v) { *(simd_Vec16f*)dst = v; }
static inline simd_Vec16f nr_simd_gather_16f(void* src, simd_Vec16i32 indices) {
    simd_Vec16f res;
    for(int i=0; i<16; ++i) res[i] = ((float*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_16f(void* dst, simd_Vec16i32 indices, simd_Vec16f v) {
    for(int i=0; i<16; ++i) ((float*)dst)[indices[i]] = v[i];
}
static inline simd_Vec8d nr_simd_load_aligned_8d(void* src) { return *(simd_Vec8d*)src; }
static inline void nr_simd_store_aligned_8d(void* dst, simd_Vec8d v) { *(simd_Vec8d*)dst = v; }
static inline simd_Vec8d nr_simd_gather_8d(void* src, simd_Vec8i64 indices) {
    simd_Vec8d res;
    for(int i=0; i<8; ++i) res[i] = ((double*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8d(void* dst, simd_Vec8i64 indices, simd_Vec8d v) {
    for(int i=0; i<8; ++i) ((double*)dst)[indices[i]] = v[i];
}
static inline simd_Vec16i32 nr_simd_load_aligned_16i32(void* src) { return *(simd_Vec16i32*)src; }
static inline void nr_simd_store_aligned_16i32(void* dst, simd_Vec16i32 v) { *(simd_Vec16i32*)dst = v; }
static inline simd_Vec16i32 nr_simd_gather_16i32(void* src, simd_Vec16i32 indices) {
    simd_Vec16i32 res;
    for(int i=0; i<16; ++i) res[i] = ((int32_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_16i32(void* dst, simd_Vec16i32 indices, simd_Vec16i32 v) {
    for(int i=0; i<16; ++i) ((int32_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec16u32 nr_simd_load_aligned_16u32(void* src) { return *(simd_Vec16u32*)src; }
static inline void nr_simd_store_aligned_16u32(void* dst, simd_Vec16u32 v) { *(simd_Vec16u32*)dst = v; }
static inline simd_Vec16u32 nr_simd_gather_16u32(void* src, simd_Vec16i32 indices) {
    simd_Vec16u32 res;
    for(int i=0; i<16; ++i) res[i] = ((uint32_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_16u32(void* dst, simd_Vec16i32 indices, simd_Vec16u32 v) {
    for(int i=0; i<16; ++i) ((uint32_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec8i64 nr_simd_load_aligned_8i64(void* src) { return *(simd_Vec8i64*)src; }
static inline void nr_simd_store_aligned_8i64(void* dst, simd_Vec8i64 v) { *(simd_Vec8i64*)dst = v; }
static inline simd_Vec8i64 nr_simd_gather_8i64(void* src, simd_Vec8i64 indices) {
    simd_Vec8i64 res;
    for(int i=0; i<8; ++i) res[i] = ((int64_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8i64(void* dst, simd_Vec8i64 indices, simd_Vec8i64 v) {
    for(int i=0; i<8; ++i) ((int64_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec8u64 nr_simd_load_aligned_8u64(void* src) { return *(simd_Vec8u64*)src; }
static inline void nr_simd_store_aligned_8u64(void* dst, simd_Vec8u64 v) { *(simd_Vec8u64*)dst = v; }
static inline simd_Vec8u64 nr_simd_gather_8u64(void* src, simd_Vec8i64 indices) {
    simd_Vec8u64 res;
    for(int i=0; i<8; ++i) res[i] = ((uint64_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8u64(void* dst, simd_Vec8i64 indices, simd_Vec8u64 v) {
    for(int i=0; i<8; ++i) ((uint64_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec64i8 nr_simd_load_aligned_64i8(void* src) { return *(simd_Vec64i8*)src; }
static inline void nr_simd_store_aligned_64i8(void* dst, simd_Vec64i8 v) { *(simd_Vec64i8*)dst = v; }
static inline simd_Vec64u8 nr_simd_load_aligned_64u8(void* src) { return *(simd_Vec64u8*)src; }
static inline void nr_simd_store_aligned_64u8(void* dst, simd_Vec64u8 v) { *(simd_Vec64u8*)dst = v; }
static inline simd_Vec32i16 nr_simd_load_aligned_32i16(void* src) { return *(simd_Vec32i16*)src; }
static inline void nr_simd_store_aligned_32i16(void* dst, simd_Vec32i16 v) { *(simd_Vec32i16*)dst = v; }
static inline simd_Vec32u16 nr_simd_load_aligned_32u16(void* src) { return *(simd_Vec32u16*)src; }
static inline void nr_simd_store_aligned_32u16(void* dst, simd_Vec32u16 v) { *(simd_Vec32u16*)dst = v; }
static inline simd_Vec4d nr_simd_load_aligned_4d(void* src) { return *(simd_Vec4d*)src; }
static inline void nr_simd_store_aligned_4d(void* dst, simd_Vec4d v) { *(simd_Vec4d*)dst = v; }
static inline simd_Vec4d nr_simd_gather_4d(void* src, simd_Vec4i64 indices) {
    simd_Vec4d res;
    for(int i=0; i<4; ++i) res[i] = ((double*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_4d(void* dst, simd_Vec4i64 indices, simd_Vec4d v) {
    for(int i=0; i<4; ++i) ((double*)dst)[indices[i]] = v[i];
}
static inline simd_Vec8f nr_simd_load_aligned_8f(void* src) { return *(simd_Vec8f*)src; }
static inline void nr_simd_store_aligned_8f(void* dst, simd_Vec8f v) { *(simd_Vec8f*)dst = v; }
static inline simd_Vec8f nr_simd_gather_8f(void* src, simd_Vec8i indices) {
    simd_Vec8f res;
    for(int i=0; i<8; ++i) res[i] = ((float*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8f(void* dst, simd_Vec8i indices, simd_Vec8f v) {
    for(int i=0; i<8; ++i) ((float*)dst)[indices[i]] = v[i];
}
static inline simd_Vec4f nr_simd_load_aligned_4f(void* src) { return *(simd_Vec4f*)src; }
static inline void nr_simd_store_aligned_4f(void* dst, simd_Vec4f v) { *(simd_Vec4f*)dst = v; }
static inline simd_Vec4f nr_simd_gather_4f(void* src, simd_Vec4i indices) {
    simd_Vec4f res;
    for(int i=0; i<4; ++i) res[i] = ((float*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_4f(void* dst, simd_Vec4i indices, simd_Vec4f v) {
    for(int i=0; i<4; ++i) ((float*)dst)[indices[i]] = v[i];
}
static inline simd_Vec8i nr_simd_load_aligned_8i(void* src) { return *(simd_Vec8i*)src; }
static inline void nr_simd_store_aligned_8i(void* dst, simd_Vec8i v) { *(simd_Vec8i*)dst = v; }
static inline simd_Vec8i nr_simd_gather_8i(void* src, simd_Vec8i indices) {
    simd_Vec8i res;
    for(int i=0; i<8; ++i) res[i] = ((int32_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_8i(void* dst, simd_Vec8i indices, simd_Vec8i v) {
    for(int i=0; i<8; ++i) ((int32_t*)dst)[indices[i]] = v[i];
}
static inline simd_Vec4i nr_simd_load_aligned_4i(void* src) { return *(simd_Vec4i*)src; }
static inline void nr_simd_store_aligned_4i(void* dst, simd_Vec4i v) { *(simd_Vec4i*)dst = v; }
static inline simd_Vec4i nr_simd_gather_4i(void* src, simd_Vec4i indices) {
    simd_Vec4i res;
    for(int i=0; i<4; ++i) res[i] = ((int32_t*)src)[indices[i]];
    return res;
}
static inline void nr_simd_scatter_4i(void* dst, simd_Vec4i indices, simd_Vec4i v) {
    for(int i=0; i<4; ++i) ((int32_t*)dst)[indices[i]] = v[i];
}


static inline void nr_simd_storeu_f32(void* dst, simd_Vec8f v) {
    _mm256_storeu_ps((float*)dst, v);
}

static inline simd_Vec8f nr_simd_set8_f32(float f0, float f1, float f2, float f3, float f4, float f5, float f6, float f7) {
    return _mm256_set_ps(f7, f6, f5, f4, f3, f2, f1, f0);
}
static inline simd_Vec8i nr_simd_set8_i32(int i0, int i1, int i2, int i3, int i4, int i5, int i6, int i7) {
    return _mm256_set_epi32(i7, i6, i5, i4, i3, i2, i1, i0);
}

static inline simd_Vec8f nr_simd_permute_8f(simd_Vec8f a, simd_Vec8i idx) {
    return _mm256_permutevar8x32_ps(a, idx);
}
static inline simd_Vec8i nr_simd_permute_8i(simd_Vec8i a, simd_Vec8i idx) {
    return _mm256_permutevar8x32_epi32(a, idx);
}
static inline simd_Vec4d nr_simd_permute_4d(simd_Vec4d a, simd_Vec4i64 idx) {
    double a_arr[4];
    int64_t idx_arr[4];
    _mm256_storeu_pd(a_arr, a);
    _mm256_storeu_si256((__m256i*)idx_arr, idx);
    
    double res_arr[4];
    for(int i=0; i<4; i++) {
        res_arr[i] = a_arr[idx_arr[i]];
    }
    return _mm256_loadu_pd(res_arr);
}
static inline simd_Vec8f nr_simd_swizzle_8f(simd_Vec8f a, simd_Vec8i idx) {
    return _mm256_permutevar_ps(a, idx);
}
static inline simd_Vec8f nr_simd_unpacklo_8f(simd_Vec8f a, simd_Vec8f b) {
    return _mm256_unpacklo_ps(a, b);
}
static inline simd_Vec8f nr_simd_unpackhi_8f(simd_Vec8f a, simd_Vec8f b) {
    return _mm256_unpackhi_ps(a, b);
}
static inline simd_Vec4d nr_simd_unpacklo_4d(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_unpacklo_pd(a, b);
}
static inline simd_Vec4d nr_simd_unpackhi_4d(simd_Vec4d a, simd_Vec4d b) {
    return _mm256_unpackhi_pd(a, b);
}

// Phase 6: Conversions & Horizontal Reductions
static inline simd_Vec8i nr_simd_cvt_8f_to_8i(simd_Vec8f a) {
    return _mm256_cvttps_epi32(a);
}
static inline simd_Vec8f nr_simd_cvt_8i_to_8f(simd_Vec8i a) {
    return _mm256_cvtepi32_ps(a);
}
static inline simd_Vec4d nr_simd_cvt_4f_to_4d(simd_Vec4f a) {
    return _mm256_cvtps_pd(a);
}
static inline simd_Vec4f nr_simd_cvt_4d_to_4f(simd_Vec4d a) {
    return _mm256_cvtpd_ps(a);
}
static inline float nr_simd_reduce_add_8f(simd_Vec8f a) {
    simd_Vec8f t1 = _mm256_hadd_ps(a, a);
    simd_Vec8f t2 = _mm256_hadd_ps(t1, t1);
    __m128 lo = _mm256_castps256_ps128(t2);
    __m128 hi = _mm256_extractf128_ps(t2, 1);
    __m128 res = _mm_add_ps(lo, hi);
    return _mm_cvtss_f32(res);
}
static inline float nr_simd_reduce_max_8f(simd_Vec8f a) {
    __m128 lo = _mm256_castps256_ps128(a);
    __m128 hi = _mm256_extractf128_ps(a, 1);
    __m128 max1 = _mm_max_ps(lo, hi); 
    __m128 shuf = _mm_shuffle_ps(max1, max1, _MM_SHUFFLE(2, 3, 0, 1));
    __m128 max2 = _mm_max_ps(max1, shuf); 
    __m128 shuf2 = _mm_shuffle_ps(max2, max2, _MM_SHUFFLE(1, 0, 3, 2)); 
    __m128 max3 = _mm_max_ps(max2, shuf2); 
    return _mm_cvtss_f32(max3);
}
static inline double nr_simd_reduce_add_4d(simd_Vec4d a) {
    simd_Vec4d t1 = _mm256_hadd_pd(a, a);
    __m128d lo = _mm256_castpd256_pd128(t1);
    __m128d hi = _mm256_extractf128_pd(t1, 1);
    __m128d res = _mm_add_pd(lo, hi);
    return _mm_cvtsd_f64(res);
}

// Phase 7: Transcendental Math Approximations
#ifdef __clang__
#pragma clang attribute push (__attribute__((target("avx,avx2,fma"))), apply_to=function)
#endif
static inline simd_Vec8f nr_simd_exp_8f(simd_Vec8f x) {
    simd_Vec8f min_val = _mm256_set1_ps(-87.3365f);
    simd_Vec8f max_val = _mm256_set1_ps(88.7228f);
    x = _mm256_max_ps(min_val, x);
    x = _mm256_min_ps(max_val, x);
    
    simd_Vec8f inv_ln2 = _mm256_set1_ps(1.44269504088896341f);
    simd_Vec8f half = _mm256_set1_ps(0.5f);
    simd_Vec8f nx = _mm256_add_ps(_mm256_mul_ps(x, inv_ln2), half);
    simd_Vec8f n = _mm256_floor_ps(nx); 
    
    simd_Vec8f ln2_hi = _mm256_set1_ps(0.693359375f);
    simd_Vec8f ln2_lo = _mm256_set1_ps(-2.12194440e-4f);
    x = _mm256_sub_ps(x, _mm256_mul_ps(n, ln2_hi));
    x = _mm256_sub_ps(x, _mm256_mul_ps(n, ln2_lo));
    
    simd_Vec8f c2 = _mm256_set1_ps(1.0f / 2.0f);
    simd_Vec8f c3 = _mm256_set1_ps(1.0f / 6.0f);
    simd_Vec8f c4 = _mm256_set1_ps(1.0f / 24.0f);
    simd_Vec8f c5 = _mm256_set1_ps(1.0f / 120.0f);
    
    simd_Vec8f p = _mm256_fmadd_ps(c5, x, c4);
    p = _mm256_fmadd_ps(p, x, c3);
    p = _mm256_fmadd_ps(p, x, c2);
    p = _mm256_fmadd_ps(p, x, _mm256_set1_ps(1.0f)); 
    p = _mm256_fmadd_ps(p, x, _mm256_set1_ps(1.0f)); 
    
    __m256i n_int = _mm256_cvtps_epi32(n);
    n_int = _mm256_add_epi32(n_int, _mm256_set1_epi32(127));
    n_int = _mm256_slli_epi32(n_int, 23);
    simd_Vec8f exp2n = _mm256_castsi256_ps(n_int);
    
    return _mm256_mul_ps(p, exp2n);
}

static inline simd_Vec8f nr_simd_sin_8f(simd_Vec8f x) {
    simd_Vec8f inv_pi = _mm256_set1_ps(0.31830988618f);
    simd_Vec8f pi = _mm256_set1_ps(3.14159265359f);
    simd_Vec8f half = _mm256_set1_ps(0.5f);
    
    simd_Vec8f q = _mm256_fmadd_ps(x, inv_pi, half);
    q = _mm256_floor_ps(q); 
    
    x = _mm256_fnmadd_ps(q, pi, x); 
    
    __m256i q_int = _mm256_cvtps_epi32(q);
    __m256i odd = _mm256_slli_epi32(_mm256_and_si256(q_int, _mm256_set1_epi32(1)), 31);
    
    simd_Vec8f x2 = _mm256_mul_ps(x, x);
    simd_Vec8f c7 = _mm256_set1_ps(-1.984126984126984e-4f);
    simd_Vec8f c5 = _mm256_set1_ps(8.333333333333333e-3f);
    simd_Vec8f c3 = _mm256_set1_ps(-1.666666666666666e-1f);
    
    simd_Vec8f p = _mm256_fmadd_ps(c7, x2, c5);
    p = _mm256_fmadd_ps(p, x2, c3);
    p = _mm256_mul_ps(p, x2);
    p = _mm256_fmadd_ps(p, x, x); 
    
    return _mm256_xor_ps(p, _mm256_castsi256_ps(odd));
}

static inline simd_Vec8f nr_simd_cos_8f(simd_Vec8f x) {
    simd_Vec8f half_pi = _mm256_set1_ps(1.57079632679f);
    return nr_simd_sin_8f(_mm256_add_ps(x, half_pi));
}

static inline simd_Vec8f nr_simd_log2_8f(simd_Vec8f x) {
    __m256i i = _mm256_castps_si256(x);
    __m256i exp = _mm256_srli_epi32(i, 23);
    exp = _mm256_sub_epi32(exp, _mm256_set1_epi32(127));
    
    __m256i mant = _mm256_and_si256(i, _mm256_set1_epi32(0x007FFFFF));
    mant = _mm256_or_si256(mant, _mm256_set1_epi32(0x3F800000)); 
    
    simd_Vec8f m = _mm256_castsi256_ps(mant);
    simd_Vec8f e = _mm256_cvtepi32_ps(exp);
    
    simd_Vec8f z = _mm256_sub_ps(m, _mm256_set1_ps(1.0f));
    
    simd_Vec8f c1 = _mm256_set1_ps(1.44269504088896f);
    simd_Vec8f c2 = _mm256_set1_ps(-0.72134752044448f);
    simd_Vec8f c3 = _mm256_set1_ps(0.48089834696298f);
    simd_Vec8f c4 = _mm256_set1_ps(-0.36067376022224f);
    simd_Vec8f c5 = _mm256_set1_ps(0.28853900817779f);
    
    simd_Vec8f p = _mm256_fmadd_ps(c5, z, c4);
    p = _mm256_fmadd_ps(p, z, c3);
    p = _mm256_fmadd_ps(p, z, c2);
    p = _mm256_fmadd_ps(p, z, c1);
    p = _mm256_mul_ps(p, z);
    
    return _mm256_add_ps(e, p);
}

static inline simd_Vec8f nr_simd_log_8f(simd_Vec8f x) {
    simd_Vec8f l2 = nr_simd_log2_8f(x);
    return _mm256_mul_ps(l2, _mm256_set1_ps(0.69314718056f)); 
}

#ifdef __clang__
#pragma clang attribute pop
#endif

#endif
