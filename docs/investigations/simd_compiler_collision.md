# SIMD Name Collision Investigation

**Status:** Completed
**Date:** 2026-07-17

## Problem
During the implementation of Phase 3 (SIMD Masking & Blending), the compiler failed to compile the generated C code with errors like:
duplicate symbol _Vec4d_Mask
edefinition of 'simd_Vec4d'
and similar redefinition errors in the C compiler output.

## Reproduction
1. Add new SIMD wrapper functions in simd.nr.
2. Generate C headers 
ora_simd.h and append to them.
3. Compile a project importing simd.

## Root Cause
The Nora compiler's inlining pass and AST lowering process was generating colliding symbol names for generic instantiation and function wrappers. Specifically, the generated C code created duplicate function definitions for SIMD intrinsic wrappers due to a missing FilterCfg pass in the dependency solver.
Additionally, when dynamically generating simd.nr bindings (Phase 4), appending the generated Phase 4 wrappers resulted in duplicate declarations within the same scope because the append script did not check if the definitions already existed.

## Fix
1. **Compiler Fix**: In pkg/topology/solver.go, a call to FilterCfg(prog) was added inside the SolveDependencies function. This ensures that unused conditional compilation paths (like [cfg("target_feature=avx512f")] on non-AVX-512 machines) are stripped *before* symbol resolution, preventing the compiler from generating duplicate/conflicting symbols for different architectures.
2. **Generator Fix**: The Python generation script (gen_phase4_safe.py) was updated to use strict regex word boundaries (\b) when resolving vector types. This prevents false positives (e.g., matching Vec4i inside Vec4i64), which had caused duplicate AVX-512 cfg tags to be emitted on non-AVX-512 structs.
3. **Prefixes**: For the C-layer 
ora_simd.h, all manually added C wrappers were prefixed appropriately (
r_simd_...) to avoid polluting the global namespace.

## Validation
Integration tests in simd_test.nr were compiled on Windows x64. The generated C code was verified to be free of edefinition errors. Linker errors (LNK2019, LNK1120) have been fully resolved.
