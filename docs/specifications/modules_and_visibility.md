# Modules & Visibility

## Overview

Nora uses a straightforward package-based module system to organize code, manage namespaces, and enforce encapsulation boundaries. The core keywords for modularity are `package`, `import`, and `pub`.

## Syntax

### 1. Defining a Package (`package`)

Every Nora file must begin with a `package` declaration. All files within the same directory must declare the same package name and are considered part of the same logical module.

```nora
package math_utils
```

*Note:* Executables must have an entry point within `package main` (typically `src/main.nr`).

### 2. Importing Packages (`import`)

To use code from another package, you use the `import` keyword followed by the module path as a string literal.

```nora
import "io"               // Standard library
import "net/http"         // Standard library nested
import "my_lib"           // External dependency defined in nora.yaml
import "std/crypto/sha"   // Explicit std path resolution
```

Once imported, the package name acts as a namespace (e.g., `io.PrintLn()`).

### 3. Visibility & Access Control (`pub`)

By default, all types, functions, and struct fields in Nora are **private** (accessible only within the same package). 

To expose an API to external packages, you must explicitly use the `pub` keyword.

#### Public Functions
```nora
pub fn calculate_sum(a: i32, b: i32) i32 {
    return a + b
}

// Private helper
fn internal_check() bool {
    return true
}
```

#### Public Types and Fields
You must make the type `pub`, and you must also explicitly make individual fields `pub` if you want external packages to read or write to them.

```nora
pub type User = struct {
    pub id: i32,    // External packages can access
    name: str       // Private to this package
}
```

## Resolution Semantics

When the compiler resolves an `import "x"`, it checks in this order:
1.  **Dependencies:** Checks `nora.yaml` to see if `"x"` is explicitly mapped to a local or remote dependency.
2.  **Standard Library:** Checks the `std/` fallback path for core libraries (like `io`, `net/http`).
3.  **Local Directory:** Checks if it's a valid relative path within the workspace.

## Errors & Diagnostics

*   **Package Not Found:** Triggers when an `import` string cannot be resolved via `nora.yaml` or the standard library.
*   **Unused Import:** Importing a module but never referencing its exported symbols generates a compile-time warning (`unused_import_warning_test.nr`).
*   **Visibility Violation:** Attempting to access a non-`pub` function or field from outside its origin package yields an error.
