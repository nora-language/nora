# Implementation Plan: Parallel C Object Compilation via Worker Pool

- **Status**: Completed
- **Metadata**:
  - Author: Antigravity
  - Date: 2026-08-21
  - Target: Compiler CLI (`pkg/cmd/nora/main.go`)

## Goal
Accelerate Nora compilation times by parallelizing the external C compiler (`clang`, `gcc`, `cl.exe`) object file generation across CPU cores using a bounded Go worker pool (`runtime.NumCPU()`).

## Affected Compiler Components
- `pkg/cmd/nora/main.go`:
  - Package object compilation loop (lines 1964-2017)
  - Native C runtime / source compilation loop (lines 2075-2237)
  - Build cache catalog updates (`catalog.Packages`)

## Implementation Checklist
- [x] Define compilation worker task structure (`packageCompileTask` and `nativeCompileTask`) with input source, output object, and package names.
- [x] Implement a worker pool executor utilizing `runtime.NumCPU()` worker goroutines and `sync.WaitGroup`.
- [x] Add thread-safe `sync.Mutex` for updating `catalog.Packages` and collecting `allPackageObjects`.
- [x] Implement fail-fast error capture to abort compilation early if any package fails to compile.
- [x] Parallelize native C dependencies (`opts.Native.SourceFiles`) within a concurrent worker pool.
- [x] Maintain clean mutex-guarded console output in both default and `--verbose` modes.

## Test Plan
1. **Benchmark Cold Build**: Measure compilation time of `nora_physics/examples/basic.nr` (55 packages) before vs. after.
2. **Benchmark Incremental Rebuild**: Verify cache hits and incremental rebuild integrity.
3. **Compiler Error Handling**: Test that compiler diagnostic errors in any package are caught and reported cleanly without deadlocks or hangs.
4. **Integration Test Suite**: Run `go test ./pkg/...` and all integration tests.

## Risks & Mitigations
- **Subprocess saturation**: Spawning too many compiler processes could exhaust system memory.
  - *Mitigation*: Limit concurrency strictly to `runtime.NumCPU()`.
- **Interleaved console output**: Concurrent compilation outputting messages at the same time.
  - *Mitigation*: Mutex-guard console prints or buffer output per job.
- **Race conditions on Cache Catalog**: Multiple goroutines writing to `catalog.Packages`.
  - *Mitigation*: Protect the cache map with a `sync.Mutex`.

## Completion Criteria
- Cold build of multi-package projects achieves at least a 5x speedup on systems with 8+ CPU cores.
- Full test suite passes without regressions.
