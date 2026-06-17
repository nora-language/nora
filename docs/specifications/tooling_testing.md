# Integration Testing Framework

## Overview

The Nora compiler enforces correctness through an extensive internal integration test suite located in `pkg/cmd/test/`. The suite is executed using the `Nora test` CLI command. Every language feature, edge case, and diagnostic error relies on these tests to prevent regressions.

## Positive vs. Negative Tests

The test runner explicitly classifies tests into two distinct categories based on their filenames and expected behavior.

### 1. Positive Tests
Standard `.nr` files (e.g., `primitive_cast_test.nr`, `sum_type_test.nr`) are evaluated as **Positive Tests**.
For a positive test to pass, it must:
1. Lex and parse without syntax errors.
2. Pass semantic type-checking and name resolution.
3. Pass the Topological Lease Solver without emitting lifetime or memory safety violations.
4. Successfully compile to C11, generate a binary via the host compiler (GCC/Clang), and execute with a `0` exit code.

### 2. Negative Tests
Any file starting with `fail_` or containing the word `violation` in its filename (e.g., `frozen_violation.nr`, `fail_type_mismatch.nr`) is evaluated as a **Negative Test**.
For a negative test to pass, it must **fail** compilation and emit at least one compiler diagnostic error (e.g., from `pkg/diag`). If a negative test successfully compiles, the test runner marks it as a failure, indicating a severe regression in the compiler's static analysis.

## Memory Leak Detection

When running tests, the compiled binaries are often generated using the `--debug-memory` flag, which activates the `nr_mem_report()` tracking inside the C11 runtime.

### Strict Zero-Leak Policy
Historically, the fiber scheduler had known infrastructure memory overhead. This has since been completely resolved; the `scheduler_cleanup()` routine now explicitly sweeps and frees all terminated and active fiber stacks. 

Because of this, the Nora test runner expects a perfectly clean memory state. **Zero leaks are tolerated.** If a test leaks any memory (even 1 byte), the `nr_mem_report()` tracking will flag it and the test runner will fail the test. This ensures that no memory regressions are ever introduced into the standard library or core code generation pipeline.
