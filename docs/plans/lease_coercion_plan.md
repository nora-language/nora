# Implementation Plan: Primitive Lease Coercion

## Title
Implicit Primitive Lease Coercion

## Status
Completed

## Metadata
- Date: 2026-06-02
- Author: Antigravity

## Goal
Improve ergonomics by allowing implicit copy-by-value coercion of primitive leases to their base types, removing the need for value copy hacks like `+ 0`.

## Affected Compiler Components
- `pkg/types/types.go` (type compatibility & assignability checking)
- `pkg/semantic/analyzer.go` (argument validation and lease tracking)

## Implementation Checklist
- [x] Modify `types.IsAssignable` to permit assigning `#T` or `&T` to `T` if `T` is a copy-by-value type.
- [x] Update `verifyCallArguments` in `pkg/semantic/analyzer.go` to skip `"cannot move borrowed value"` check for non-owned types.
- [x] Add regression test folder `pkg/cmd/test/lease_coercion_test/`.
- [x] Clean up GECS `examples/port_gecs/gecs/src/archetype.nr`.

## Test Plan
- Run `lease_coercion_test.nr` through the compiler to verify it runs and produces expected value.
- Rebuild the compiler and run GECS basic example.

## Risks
- Incorrectly bypassing moves for owned types. (Mitigated by strictly checking `IsOwnedType`).

## Completion Criteria
- All tests pass, GECS compiles and executes without `+ 0` math hacks.
