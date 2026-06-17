# Walkthrough: Primitive Lease Coercion

This document walks through the completed implementation of primitive lease coercion.

## Changes Made

### 1. Type Assignability
In [types.go](file:///e:/Project/Project%20Chronos/second/pkg/types/types.go), we added a rule to `IsAssignable`:
```go
	if pt, ok := from.(*PointerType); ok && pt.Leased {
		if Equals(underlyingStructOrBase(pt.Base), underlyingStructOrBase(to)) {
			if to.Name() == "str" || to.Name() == "ptr" || pt.Kind == LeaseMove || !IsOwnedType(to) {
				return true
			}
		}
	}
```
If the destination type `to` is a copy-by-value/non-owned type (i.e. `!IsOwnedType(to)`), assigning any lease of that type (e.g. `#i32` or `&i32`) to `to` is now permitted.

### 2. Semantic Analysis
In [analyzer.go](file:///e:/Project/Project%20Chronos/second/pkg/semantic/analyzer.go), we updated the argument move check:
```go
				if pt, ok := argSym.Type.(*types.PointerType); ok && pt.Leased && pt.Kind != types.LeaseMove {
					if types.IsOwnedType(types.UnwrapLease(argSym.Type)) {
						sa.AddError(arg.Pos(), "cannot move borrowed value '%s'", argSym.Name)
						return
					}
				}
```
This skips the "cannot move borrowed value" semantic error for primitive types (e.g., when passing `#i32` to a generic parameter expecting `LeaseMove`).

### 3. GECS Library Cleanups
In GECS [archetype.nr](file:///e:/Project/Project%20Chronos/second/examples/port_gecs/gecs/src/archetype.nr), we successfully removed all instances of the `+ 0` math hack:
```diff
 pub fn HashI32(k: #i32) i32 {
-    return k + 0
+    return k
 }
```

## Validation & Testing
We verified the implementation using two methods:
1. Running the new integration test case [lease_coercion_test.nr](file:///e:/Project/Project%20Chronos/second/pkg/cmd/test/lease_coercion_test/lease_coercion_test.nr), which prints `Coercion result: 42`.
2. Verifying the complete GECS library compiles and executes successfully:
   ```bash
   nora.exe run --example basic
   ```
   **Output:**
   ```text
   GECS Full Port Test Successful!
   ```
