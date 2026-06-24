# Investigation: R-Value Closures, Nested Structs, and Codegen Type Mismatches

## Problem
1. **Memory Leaks**: When passing closures as inline arguments (e.g., DoMap(fn(...) { ... })), the dynamically allocated closure environment (
r_closure_t) was leaked because the Topological Lease Solver only scheduled drops for named variables, ignoring anonymous r-values.
2. **Nested Struct Compilation Error**: Lambdas defined inside complex AST structures, like struct literals, failed to compile because the C-codegen could not find their _env_t struct declarations.
3. **Pointer/Value Return Mismatch**: After fixing memory leaks with move semantics, returning value types from lambda blocks resulted in error: returning 'DataObj' from a function with incompatible result type 'DataObj *'.

## Root Cause
1. **R-Value Drops**: R-values generated within expressions were not tracked for lifecycle management. The Lease Solver was unaware of st.LambdaExpression nodes that were not bound to variables.
2. **Nested Structs**: The HIR lowerer bypassed recursive traversal for complex literals by wrapping them directly in an *ast.ASTExpr fallback container, so nested lambdas were never discovered or added to the active function environment list.
3. **Type Mismatch**: The Ret instruction check was using the explicit GetType() of the AST node, which was evaluated as a static pointer (@DataObj). However, because it was wrapping the move in a GCC statement expression to disable drop flags, the generated C string was a value type (DataObj). Because the static check saw a pointer, it skipped the branch designed to heap-allocate and move the consumed value via 
r_malloc.

## Fix
1. Updated pkg/topology/solver.go to explicitly detect and track temporary r-values, especially st.LambdaExpression, scheduling AnonymousDrops at the end of the statement.
2. Added collectHiddenLambdas AST inspection pass to the default: case of lowerExpression in pkg/hir/lower.go to aggressively find and register any hidden lambdas.
3. Replaced g.cType(i.Val.GetType()) with g.cTypeOfOperand(i.Val) inside the Ret switch block in pkg/codegen/hir_codegen.go to accurately identify the evaluated codegen type and trigger 
r_malloc allocation correctly.

## Validation
epro_multi_lambda_leak.nr passes without memory leaks and without C type mismatches. The full test suite confirms zero regressions.
