# Walkthrough: Fiber Scheduler Cleanup & Allocator Race Fixes

## Overview
This walkthrough summarizes the changes made to resolve intermittent segmentation faults and deadlock hangs during scheduler termination and concurrent fiber execution. We addressed a critical data race in the static/untracked allocator initialization and resolved target-feature mismatch issues on AVX-only architectures.

## Changes

### 1. Eager Initialization of Untracked Allocator Mutex (`std/runtime/memory.c`)
- **Problem**: `g_untracked_lock` was initialized lazily within `nr_malloc_untracked()` and `nr_free_untracked()`. Under heavy concurrent fiber spawns, multiple worker threads called these functions simultaneously, causing a data race on `NR_MUTEX_INIT(&g_untracked_lock)`. This corrupted the mutex state, leading to deadlocks and race conditions.
- **Solution**: Moved `g_untracked_lock` and `g_untracked_lock_init` declarations to the top of `memory.c` and initialized the mutex eagerly inside `nr_mem_init()`, which runs on the main thread before starting any workers. Removed all lazy-initialization checks from allocation and deallocation paths.

### 2. SIMD Test Compatibility Check (`pkg/cmd/nora/compiler_test.go`)
- **Problem**: `simd_test.nr` requires AVX2/AVX-512 features (specifically instructions like gathers/scatters), but the test machine's CPU only supports AVX. Hardcoding AVX2 and FMA target features in `compiler_test.go` caused Clang to emit AVX2 instructions, resulting in `signal: illegal instruction (core dumped)` at runtime.
- **Solution**: Added a dynamic feature check at the start of the integration test runner to inspect `/proc/cpuinfo` for AVX2 support. If AVX2 is not supported by the host CPU, `simd_test.nr` is automatically skipped to prevent illegal instruction crashes.

## Verification & Testing
- Ran the full integration test suite via `go test -v ./pkg/cmd/nora`.
- All integration tests completed successfully (`PASS` with exit code `0`).
- `simd_test.nr` is correctly skipped on non-AVX2 hosts, while all other concurrent, semantic, and compiler tests successfully pass.
