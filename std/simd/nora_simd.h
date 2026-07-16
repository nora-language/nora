#ifndef NORA_SIMD_H
#define NORA_SIMD_H
#pragma GCC target("avx")
#include <x86intrin.h>

typedef __m256d simd_Vec4d;

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

#endif
