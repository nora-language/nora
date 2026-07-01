# Investigation Report: Silent Exit and Premature Termination in `phase8_determinism`

## Problem
When executing the `phase8_determinism` simulation, the physics engine was failing to complete its step and was silently exiting. Initially, it appeared to be a cooperative scheduling or fiber unwinding issue related to `ParMap`, as commenting out `ParMap` calls allowed the simulation to print `--- Test Finished ---` and return an exit code of `0`.

The primary symptom was that `nora run` reported `EXIT CODE: 0` despite the simulation looping silently crashing in the background without completing.

## Reproduction
To isolate the bug, a reproduction test was created in `pkg/cmd/test/repro_solver_div_by_zero`. This test reproduces a zero division with `fixed64` math which causes the `EXCEPTION_INT_DIVIDE_BY_ZERO` hardware exception.

## Root Cause
There were **two interconnected bugs** creating this deceptive failure state:

1. **Compiler CLI Bug (`pkg/cmd/nora/main.go`)**: 
   The `nora run` command was discarding the OS exit code of the executed binary. When the compiled physics binary crashed with a hardware exception, `nora.exe` caught the error from the `cmd.Run()` execution but discarded it and exited with `0`. This masked the actual hardware exception (-1073741676 / `0xC0000094`).

2. **Physics Solver Logic Bug (`nora_physics/src/dynamics/solver.nr`)**:
   Inside the `ParMap` lambda dispatch in `system.nr`, the narrow phase passed contact manifolds to `solver.SolveContact`. Inside `SolveContact`, there were multiple unprotected divisions by zero:
   - When attempting to generate a safe "1" value using `one_safe = massB1 / massB1`, the solver did not account for situations where `massB1` was 0 (such as static or kinematic bodies colliding with each other), triggering `0 / 0`.
   - When applying impulses, calculating a negative one multiplier `var neg_one = (j - j) - (j / j)` resulted in `0 / 0` when the calculated impulse `j` was `0`.
   
   Because `fixed64_Fixed64_div` translates to C integer division, `0 / 0` triggers an `EXCEPTION_INT_DIVIDE_BY_ZERO` (-1073741676) at the OS level, terminating the fiber and process instantly.

## Fix
1. **CLI Fix**: Updated `main.go` inside the compiler frontend to explicitly check the `exec.ExitError` returned from `runCmd.Run()` and propagate the child process's exit code via `os.Exit(exitError.ExitCode())`.
2. **Solver Fix**: 
   - Refactored `solver.nr` to eagerly return early if both bodies in contact are static (mass 0).
   - Replaced division-based approximations (`massA1 / massA1`) with division-free assignments where possible, initializing `one_safe` and `neg_one` without dividing by zero.

## Validation
After applying the fixes, rebuilding `nora.exe`, and running `phase8_determinism`, the simulation successfully executes all steps, processes constraints, and reaches `--- Test Finished ---` cleanly with `EXIT CODE: 0`.
