# Compiler Investigation: Monomorphizer Silent Panic on Missing Method

## Problem
When a generic struct is called with a method that does not exist (e.g., `v1.Clone[T]()` on a 3D vector where `Clone` is not defined), the Nora compiler fails to emit a standard semantic error ("undefined method"). Instead, it crashes silently during the compilation process.

## Reproduction
To reproduce the issue, attempt to call an undefined method on a generic type instantiation within a `.nr` file:

```nora
import "src/math/vector"

pub fn TriggerCrash[T]() {
    var v1 = vector.NewVector3[T](0.0, 0.0, 0.0)
    
    // The vector.Vector3[T] struct has no Clone[T]() method defined.
    // Instead of a semantic error, this causes the compiler to panic.
    var v2 = v1.Clone[T]() 
}
```
Running `nora run` on a file containing this will result in the compiler exiting abruptly with a panic instead of generating a proper diagnostic message.

## Root Cause
The root cause of the issue lies in the Nora compiler's `pkg/codegen` (or `pkg/semantic`) pipeline, specifically during the type-erased shared monomorphization phase. When the semantic analyzer processes a method call on a generic type, it defers certain method resolution checks to the monomorphizer. However, the monomorphizer assumes that any method it receives has already been validated. When it attempts to look up the undefined method's AST node to generate the specialized generic implementation, it encounters a nil pointer (or missing dictionary entry) and panics without recovering or emitting a proper error diagnostic (`pkg/diag`).

## Fix
*(This represents the required compiler-side fix)*
The semantic analyzer (`pkg/semantic`) must eagerly validate method existence on generic types *before* passing the AST down to the monomorphizer. If the method is not found, the semantic pass should immediately emit a diagnostic error (e.g., `Error: undefined method 'Clone' for type 'Vector3[T]'`) and halt the pipeline gracefully, preventing the monomorphizer from attempting to process invalid AST nodes. 

## Validation
* Create a negative integration test (`pkg/cmd/test/fail_missing_generic_method.nr`) containing the reproduction code.
* Verify that the old compiler binary crashes/panics when running the test.
* Verify that after the fix, the compiler successfully catches the error during the semantic pass and outputs a clear diagnostic message, passing the negative test runner.
