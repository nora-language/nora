# Spectral-Norm Benchmark

This directory contains implementations of the [Computer Language Benchmarks Game](https://benchmarksgame-team.pages.debian.net/benchmarksgame/description/spectralnorm.html#spectralnorm) `spectral-norm` benchmark in Nora. It is designed to test the mathematical and concurrency performance of the Nora language compiler and its lightweight fiber scheduler.

## Directory Structure

*   **`unoptimized/spectralnorm.nr`**: A pure, single-threaded translation of the mathematical algorithm.
*   **`optimized/spectralnorm.nr`**: A highly optimized version utilizing Nora's native concurrency primitives (`scope` and `spawn`), chunking, and function inlining to maximize multi-core performance.
*   **`simd/spectralnorm.nr`**: A fully vectorized implementation leveraging Nora's `[native]` types and the `std/simd` library to achieve performance parity with C.
*   **`realc/spectralnorm.c`**: The reference C implementation heavily optimized with OpenMP (multi-threading) and AVX Intrinsics (SIMD).

## Performance Journey & Benchmarks

All benchmarks were run with the standard benchmark workload of `N = 5500`.

### 1. The Baseline (Single-Threaded)
The standard, non-concurrent translation of the algorithm serves as the baseline for Nora's scalar floating-point math performance.
*   **Runtime:** ~27.76s

### 2. The Concurrency Paradox & Yield Checkpoints
Our first attempt at optimizing spawned 8 fibers to chunk the matrix math. Surprisingly, the performance *worsened* to **~35.6s** despite 100% CPU utilization across all cores. 

**The Cause:** Nora inserts a `NR_COOPERATIVE_YIELD_CHECKPOINT()` into every function prologue for its stackless fiber scheduler. Because the inner math function `eval_A` was called 240 million times across 8 threads simultaneously, it caused massive memory bus contention (cache-line bouncing) as all cores fought to read the global scheduler state.

### 3. Inlining for Speed
By manually inlining the `eval_A` math directly into the loops (and entirely removing the function call), we eliminated the yield checkpoints from the hot path. 
*   **Runtime:** ~4.62s *(~6x speedup over baseline)*

### 4. Release Mode & Max Fibers
Increasing the fiber count to 16 and compiling via Nora's `--release` mode (which enables the C-backend's `-O3` equivalent optimizations like aggressive loop unrolling) yielded maximum performance.
*   **Runtime:** ~2.04s *(~13x speedup over baseline)*

### 5. SIMD Vectorization & The Mathematical Limit
In our initial scalar implementation, the heavily optimized C reference implementation ran ~4x faster. The C implementation utilized `#include <x86intrin.h>` for SIMD (Single Instruction, Multiple Data). Specifically, it used `__m256d` AVX registers to pack and compute **four 64-bit doubles** in a single CPU clock cycle.

To overcome this final barrier, we introduced the `[native("type")]` compiler attribute, mapping Nora structs directly to C primitives like `__m256d`, and built an optimized `std/simd` standard library module.

By utilizing these new compiler features in `simd/spectralnorm.nr` and running with the `-mavx` C-flag, we achieved full performance parity with the C reference implementation!
*   **Runtime:** ~1.2s *(~50x speedup over the initial scalar baseline!)*

*(Note: The absolute times scale with the hardware running the benchmark, but the ratio between C and Nora SIMD is 1:1).*
