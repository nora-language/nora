# Investigation: Topological Lease Solver Move Errors on Value Re-use

**Status:** Completed
**Date:** June 27, 2026
**Component:** `pkg/topology` (Topological Lease Solver)

## Problem
In the Nora language, when a static constant or local variable holding a primitive (like floating point zero) is reused across multiple mathematical expressions, the Topological Lease Solver intercepts the re-use and throws a "use of moved value" compilation error.

For example, using `var zero = dt - dt` and then later reusing `zero` as `(vel + zero)` multiple times inside a loop results in errors such as:
`Error: use of moved value 'zero'`
`= note: value moved here at ...`

## Reproduction
1. Define a float value locally: `var z = 0.0`.
2. Use this value in multiple assignments or operations: 
   ```nora
   var a = b + z
   var c = d + z
   ```
3. Run the Nora compiler. The lease solver graph will build a dependency where `z` is consumed (moved) at the first addition `var a = b + z`.
4. The second expression `var c = d + z` will trigger the compiler error.

## Root Cause
Nora is structurally linear. The Topological Lease Solver (which runs between `pkg/semantic` and `pkg/hir`) tracks assignments and field accesses to trace variable lifecycle boundaries to automatically inject RAII `Drops` and prevent data-race leaks.

Unlike some languages where primitives intrinsically implement an implicit "Copy" trait, Nora strictly manages dependencies via moves or leases (`#` and `&`). When `z` is used by value in an operation that isn't explicitly passing a read-only lease `#z`, the lease solver assumes the ownership is transferred (moved) to the evaluating expression, consuming the variable entirely.

## Fix (Workaround Applied)
Since basic arithmetic operators (`+`, `-`, `*`, `/`) natively accept values, and taking a read-only lease `#` for every single scalar parameter is syntactically heavy or unsupported by the generic math operations, the best workaround is to construct primitive values dynamically at the call site. 

Instead of creating a local variable `var zero = dt - dt` and reusing it, the zero value is calculated dynamically when needed via `(dt - dt)` in-place.
```nora
// Before (Fails move check)
var zero = dt - dt
wh.suspension_relative_velocity = proj_vel + zero
var f_damp = proj_vel + zero

// After (Passes topological lease checks)
wh.suspension_relative_velocity = proj_vel + (dt - dt)
var f_damp = proj_vel + (dt - dt)
```
This correctly avoids linear consumption because each `(dt - dt)` constructs a new ephemeral scalar that is safely consumed exactly once by the arithmetic expression.

## Validation
Applying the dynamic evaluation pattern across the `raycast_vehicle` implementation allowed the solver to complete the dependency topology without any moved value errors, seamlessly moving down the pipeline to AST-to-HIR compilation.
