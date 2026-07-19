# Investigation: Builder Pattern Assignment Requirement Bug

**Status**: Completed

## Problem
When utilizing the builder pattern in Nora (such as with `gfx.PipelineBuilder` or `gfx.MaterialBuilder`), developers are forced to assign the result of every method call to a dummy variable (e.g., `var _1 = builder.WithWGSLShader(...)`). If the method call is left as a standalone expression statement without assignment, the compiler either emits an error (such as "value evaluated but not used") or fails to process the AST node properly.

## Reproduction
To reproduce the bug, attempt to use a builder pattern where methods return a pointer to `self` (`&Builder`), but do not assign the result to a variable:

```nora
var pb = gfx.NewPipelineBuilder(#device)
// The following line causes an error because the compiler believes dropping the returned `&PipelineBuilder` might cause a memory leak:
pb.WithWGSLShader(wgsl_code, "vs_main", "fs_main").WithDepthTest(48)
```

Because of this bug, developers either have to chain all methods and assign the final result to a dummy variable, or assign every step sequentially:
```nora
var pb = gfx.NewPipelineBuilder(#device)
var _1 = pb.WithWGSLShader(wgsl_code, "vs_main", "fs_main")
           .WithDepthTest(48)
           .Build()
// Or sequentially:
// var _1 = pb.WithWGSLShader(...)
// var _2 = pb.WithDepthTest(...)
```

## Root Cause
Nora's compiler front-end and Topological Lease Solver correctly identify that a function is returning a tracked pointer type (like `&PipelineBuilder` or `@PipelineBuilder`). To prevent memory leaks or dangling pointers, the solver strictly enforces that returned tracked values must be consumed (assigned, moved, or passed to a function). The exact error emitted is:
`Error: cannot discard owned value of type '&PipelineBuilder'. it must be assigned, moved, or passed to a function to avoid memory leaks`

While method chaining *does* work in Nora (e.g., `builder.A().B().C()`), the strict memory safety checks forbid discarding the *final* returned value of the chain. Because Nora currently lacks a blank identifier syntax (like `_ = expr`) to explicitly tell the solver "I am intentionally discarding this lease safely", developers are forced to bind the returned value to a dummy variable to appease the compiler.

## Fix
In the short term, the workaround is to assign every builder method return value to a sequential dummy variable (e.g., `var _1 = ...`, `var _2 = ...`).

*Future Consideration*: The Nora compiler should be updated to natively support expression statements that discard their return values. The topological lease solver should be instructed to immediately `Drop` or safely ignore the returned pointer if it isn't bound to a variable. Alternatively, introducing a blank identifier syntax (`_ = expression`) similar to Go would allow developers to explicitly discard returns cleanly.

## Validation
By assigning the result of all pipeline and material builder methods to variables (`var _4`, `var _5`, etc.) in `examples/phong/main.nr` and `examples/pbr/main.nr`, the compilation completes successfully, bypassing the bug.
