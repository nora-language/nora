# Walkthrough: Fix Moved Heap Pointer Deletion Leak / Crash

This walkthrough documents the fix for a critical compiler transpilation bug where the codegen incorrectly generated `nr_free` calls on moved heap pointers, resulting in invalid memory drops (memory leaks) or segmentation faults/crashes in programs using recursive structures and lambdas (such as `bst_test` and `repro_multi_lambda_leak`).

## 1. Problem Description

When a heap pointer was explicitly moved via `@` (e.g. `return @res` or passing `@bst` to a function), the transpiler generated C code that:
1. Copied the pointer to a temporary.
2. Called `nr_free(res)` to release the source variable.
3. Returned/assigned the temporary pointer, which now pointed to freed memory.

### Root Cause
In the previous commit, a utility was introduced to determine if a moved heap pointer needed to be freed during assignment or calls. However, this utility unwrapped the lease pointer type (`types.UnwrapLease(t)`) before checking if it was a pointer. For a type like `@Node[i32]`, unwrapping the lease yielded the underlying struct type `Node[i32]`, which was not recognized as a pointer. As a result, the code generator assumed the destination was a value type and incorrectly freed the heap pointer, leading to use-after-free bugs.

---

## 2. Implementation & Fixes

We replaced the manual conceptually-pointer checks with the compiler's robust `isPointerTypeInC(t)` function, which correctly detects if a type is represented as a pointer in C (including leased pointers).

### A. Prefix Expression Moves
In [expressions.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/expressions.go#L400-L404), we replaced the manual unwrap check with:
```diff
-						ut := types.UnwrapLease(t)
-						isConceptuallyPointer := false
-						if ut.GetKind() == types.KindPointer || ut.Name() == "str" || ut.Name() == "ptr" {
-							isConceptuallyPointer = true
-						} else if _, ok := ut.(*types.ListType); ok {
-							isConceptuallyPointer = true
-						} else if _, ok := ut.(*types.MapType); ok {
-							isConceptuallyPointer = true
-						} else if _, ok := ut.(*types.ChanType); ok {
-							isConceptuallyPointer = true
-						}
-						
-						if !isConceptuallyPointer {
+						if !g.isPointerTypeInC(t) {
 							g.buf.WriteString(fmt.Sprintf("nr_free(%s); %s = NULL; ", id.Value, id.Value))
 						} else {
 							g.buf.WriteString(fmt.Sprintf("%s = NULL; ", id.Value))
```

### B. Statement-Level Moved Heap Pointers
In [hir_codegen.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/hir_codegen.go#L387-L405) and [hir_codegen.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/hir_codegen.go#L1523) / [hir_codegen.go](file:///e:/Project/Project%20Chronos/second/pkg/codegen/hir_codegen.go#L1542), we replaced similar manual checks with `!g.isPointerTypeInC(destType)` and `!g.isPointerTypeInC(paramType)` respectively.

---

## 3. Verification

The entire integration test suite was rerun. Limiting parallelism on Windows allowed the compiler and C toolchain to execute cleanly without resource contention timeouts.

All 170+ compiler integration tests now pass successfully.
```powershell
go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder -parallel 4
```
Output:
```text
PASS
ok  	github.com/DwiYI/Project-Nora/pkg/cmd/nora	323.691s
```

---

## 4. WSL Test Panic: stdlib_io_test

**Issue**: The integration test `stdlib_io_test.nr` consistently failed on WSL/Linux with `Panic: line mismatch`, outputting `Parsed line: Parsed lin` instead of `next_line `.

**Root Cause Analysis**:
A severe memory lifecycle and aliasing bug was identified in `std/io/io.nr`:
1. `Scanner.Text()` returned an owned `str` primitive.
2. When the user assigned `var val_str = s.Text()`, it created a shallow copy of the `s.current_token` pointer but marked it as "owned" by the caller.
3. When `s.NextLine()` was called, it allocated a new buffer for `current_token`, and the compiler automatically dropped the *old* `current_token`.
4. This freed the memory that `val_str` was still pointing to.
5. Because memory was prematurely freed, the `malloc` inside `io.PrintLn` reused the freed block for string interpolation, physically overwriting the new `val_line` buffer or allowing old bytes to leak and causing it to become `"Parsed lin"`.
6. This memory corruption caused the `val_line != "next_line "` assertion to fail. The discrepancy between Windows and Linux arose simply due to differences in how the underlying libc `malloc` reuses recently freed chunks.

**Resolution**:
- Modified `Scanner.Text()` to return a read-only borrow (`#str`) instead of an owned `str`.
```nora
pub fn (self: &Scanner) Text() #str {
    return self.current_token
}
```
- This matches Nora's strict ownership semantics: the `Scanner` retains ownership of the dynamically allocated token buffer, and the caller merely borrows a read-only lease to it.
- Verified that all Integration tests, including WSL execution under valgrind and native, now pass cleanly.
