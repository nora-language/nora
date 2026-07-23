# Investigation: Nora Language Design Limitations & Ergonomic Boundaries

**Status:** Active Assessment (July 2026)  
**Author:** DeepMind Antigravity Team & Nora Language Core Contributors  
**Target Subsystems:** Frontend (`pkg/parser`), Semantic Analyzer (`pkg/semantic`), Lease Solver (`pkg/topology`), Concurrency Runtime (`pkg/hir`, `std/runtime`)

---

## Executive Overview

During the development and testing of non-trivial applications (including `nora_physics`, `nora_wgpu`, and advanced generic libraries), developers encounter specific language and compiler boundaries. While Nora's static **Topological Lease Solver** guarantees zero-GC memory safety and data-race-free execution, certain syntactic and architectural constraints limit developer ergonomics or restrict code patterns common in other systems languages.

This investigation formally documents the 5 primary unresolved language limitations, their root causes in the compiler frontend/backend, the underlying **design & safety rationale** behind each choice, their negative test reproductions under `pkg/cmd/test/`, and their expected future evolution in the Nora specification.

---

## Discovered Limitations & Architectural Rationales

### 1. No Structural Method Resolution (Duck-Typing) for Generics

* **Description:** Attempting to invoke a method (e.g. `shape.Support()`) directly on an unconstrained generic parameter `[T]` fails during semantic analysis.
* **Reproduction Code:**
  ```nora
  fn Intersect[T](shape: #T) {
      var vec = shape.Support() // Fails: T has no field or method 'Support'
  }
  ```
* **Root Cause:** Unlike C++ templates which lazily resolve method signatures during monomorphization, Nora's semantic analyzer (`pkg/semantic/analyzer.go`) strictly enforces type bounds upfront. Without an explicit interface constraint (`[T: Shape]`), method resolution on generic type parameters is forbidden.
* **Design & Safety Rationale:**
  * **Compilation Speed & Clean Error Diagnostics:** Lazy structural resolution (like C++ templates) leads to massive compilation slowdowns and obscure, deeply nested error cascades when instantiations fail.
  * **Upfront Contract Enforceability:** Mandating explicit bounds ensures function interfaces are self-documenting and type-checked in isolation *before* any monomorphization takes place.
* **Compiler Error Output:**
  ```text
  Error: type parameter 'T' has no field or method 'Support'
    --> pkg/cmd/test/fail_limitation_unconstrained_generic_method/fail_limitation_unconstrained_generic_method.nr:12:21
  ```
* **Reproduction Test Path:** [fail_limitation_unconstrained_generic_method.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_unconstrained_generic_method/fail_limitation_unconstrained_generic_method.nr)
* **Expected Language Evolution:** Introduce structural type constraints (e.g. `[T: { fn Support() Vector3 }]`) or mandate explicit interface trait definitions.

---

### 2. Discarding Owned Return Values & Missing Blank Identifier (`_ = expr`)

* **Description:** Functions or builder methods that return an owned heap lease (`@T`) cannot be called as standalone expression statements without binding the return value to a variable.
* **Reproduction Code:**
  ```nora
  pub fn NewBuilder() @Builder { ... }

  pub fn main() {
      NewBuilder() // Fails: cannot discard owned value
  }
  ```
* **Root Cause:** To prevent resource and memory leaks, the Topological Lease Solver (`pkg/topology/solver.go`) mandates that all tracked pointers (`@T`) must be assigned, moved, or passed to a function. However, Nora currently lacks a blank identifier statement (`_ = expr`) or explicit drop syntax to safely discard unneeded owned leases.
* **Design & Safety Rationale:**
  * **Zero-GC Memory Leak Prevention:** In a language without garbage collection, silently dropping an owned heap lease (`@T`) without explicit ownership accounting creates invisible memory leaks.
  * **Explicit RAII Lifetime Tracking:** The compiler forces developers to acknowledge resource management by assigning tracked values to variables so the Topological Lease Solver can deterministically insert `Drop()` calls at scope exit.
* **Compiler Error Output:**
  ```text
  Error: cannot discard owned value of type '@Builder'. it must be assigned, moved, or passed to a function to avoid memory leaks
    --> pkg/cmd/test/fail_limitation_builder_discard_owned/fail_limitation_builder_discard_owned.nr:20:5
  ```
* **Reproduction Test Path:** [fail_limitation_builder_discard_owned.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_builder_discard_owned/fail_limitation_builder_discard_owned.nr)
* **Expected Language Evolution:** Implement blank identifier assignment syntax (`_ = expression`) or automatic scope-end `Drop` insertion for unassigned RValue expressions.

---

### 3. Cross-Fiber Write Lease Prohibition (`spawn &x`)

* **Description:** Spawning a fiber (`spawn`) while passing a mutable write lease (`&T`) is strictly rejected by the semantic analyzer.
* **Reproduction Code:**
  ```nora
  fn worker(val: &i32) { val = val + 1 }

  pub fn main() {
      var x = 10
      scope {
          spawn worker(&x) // Fails: cannot pass mutable lease across fiber boundary
      }
  }
  ```
* **Root Cause:** The borrow checker cannot statically prove at compile-time that memory slots passed into fibers (e.g., array elements `&bodies[i]`) are disjoint across concurrent execution units. To eliminate data races, `spawn` forbids mutable references (`&T`).
* **Design & Safety Rationale:**
  * **Data-Race-Free Concurrency by Construction:** Nora eliminates multi-threading data races statically. Allowing mutable write leases (`&T`) across concurrent fibers would open the door to race conditions and memory aliasing bugs.
  * **Disjoint Memory Limits:** Proving that `&array[i]` and `&array[j]` do not alias in a dynamic `while` loop requires complex dependent path analysis that is impractical in a linear static analysis pass.
* **Compiler Error Output:**
  ```text
  Error: cannot pass mutable lease (write) across fiber boundary
    --> pkg/cmd/test/fail_limitation_spawn_mutable_lease/fail_limitation_spawn_mutable_lease.nr:19:23
  ```
* **Reproduction Test Path:** [fail_limitation_spawn_mutable_lease.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_spawn_mutable_lease/fail_limitation_spawn_mutable_lease.nr)
* **Expected Language Evolution:** Expand standard library disjoint-borrow primitives (e.g. `Vector.ParMap`) and introduce compiler-verified disjoint slice partitioning.

---

### 4. No Parameter Names in Function Type Aliases

* **Description:** Defining function signature type aliases containing named parameters triggers a parser error.
* **Reproduction Code:**
  ```nora
  pub type Callback = fn(val: i32, msg: str) // Fails
  ```
* **Root Cause:** The recursive descent parser (`pkg/parser/parser.go`) routes `fn(...)` inside type declarations to `parseFunctionType()`, which expects a comma-separated list of types rather than `name: type` pairs.
* **Design & Safety Rationale:**
  * **Grammar Simplicity & Token Unambiguity:** Omitting parameter names keeps function type expressions minimal and prevents parsing ambiguities between struct field definitions (`name: type`) and function parameter signatures in the parser AST.
  * **Structural Type Equivalence:** Function types in Nora are evaluated purely based on input/output parameter type layouts rather than parameter names during monomorphization and type checking.
* **Compiler Error Output:**
  ```text
  Error: function types cannot have parameter names
    --> pkg/cmd/test/fail_limitation_fn_type_named_params/fail_limitation_fn_type_named_params.nr:11:24
  ```
* **Reproduction Test Path:** [fail_limitation_fn_type_named_params.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_fn_type_named_params/fail_limitation_fn_type_named_params.nr)
* **Expected Language Evolution:** Enhance `parseFunctionType()` to optionally accept and discard parameter documentation names (e.g., `fn(val: i32, msg: str)`).

---

### 5. No Inner-Function Statement-Level `[cfg]` Conditional Compilation

* **Description:** Annotating individual statements inside function bodies with `[cfg(target = "...")]` causes a syntax error.
* **Reproduction Code:**
  ```nora
  pub fn main() {
      [cfg(target = "windows")] // Fails
      var x = 10
  }
  ```
* **Root Cause:** Conditional compilation filtering (`pkg/semantic/cfg.go`) is currently executed as a top-level AST pass on `ast.FunctionDecl` and `ast.TypeDecl` nodes before semantic analysis. Inner statement nodes are not evaluated for `[cfg]` attributes.
* **Design & Safety Rationale:**
  * **Compiler Pass Isolation & Performance:** Filtering `[cfg]` at the top-level AST pass allows the compiler to prune unused functions early before type resolution, keeping compilation extremely fast.
  * **AST Complexity Prevention:** Allowing statement-level `[cfg]` inside function blocks complicates High-Level Intermediate Representation (HIR) lowering, variable scope tracking, and Language Server Protocol (LSP) code navigation.
* **Compiler Error Output:**
  ```text
  Error: expected next token to be ), got IDENT instead
    --> pkg/cmd/test/fail_limitation_inner_cfg/fail_limitation_inner_cfg.nr:12:10
  ```
* **Reproduction Test Path:** [fail_limitation_inner_cfg.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_inner_cfg/fail_limitation_inner_cfg.nr)
* **Expected Language Evolution:** Extend `[cfg]` evaluation to `ast.BlockStmt` and statement-level AST nodes to allow inline platform-specific code branching.

---

## Validation & Test Suite Matrix

All 5 limitations are actively tracked and validated in the compiler integration test suite under `pkg/cmd/test/`:

| Limitation Area | Test Folder | Test File | Primary Design Rationale |
|---|---|---|---|
| **Generic Duck-Typing** | `fail_limitation_unconstrained_generic_method` | [fail_limitation_unconstrained_generic_method.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_unconstrained_generic_method/fail_limitation_unconstrained_generic_method.nr) | Fast compilation & upfront contract safety |
| **Owned Return Discard** | `fail_limitation_builder_discard_owned` | [fail_limitation_builder_discard_owned.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_builder_discard_owned/fail_limitation_builder_discard_owned.nr) | Zero-GC memory leak prevention & RAII tracking |
| **Fiber Mutable Lease** | `fail_limitation_spawn_mutable_lease` | [fail_limitation_spawn_mutable_lease.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_spawn_mutable_lease/fail_limitation_spawn_mutable_lease.nr) | Static data-race elimination across threads |
| **Function Type Names** | `fail_limitation_fn_type_named_params` | [fail_limitation_fn_type_named_params.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_fn_type_named_params/fail_limitation_fn_type_named_params.nr) | Parser grammar simplicity & type equivalence |
| **Inner Statement `[cfg]`** | `fail_limitation_inner_cfg` | [fail_limitation_inner_cfg.nr](file:///e:/Project/Project%20Nora/nora/pkg/cmd/test/fail_limitation_inner_cfg/fail_limitation_inner_cfg.nr) | Early AST pruning & incremental build speed |

Each test is compiled via `Nora build` to verify that the compiler emits the exact expected diagnostic error exit code (Status 1).
