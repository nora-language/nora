# The Nora Programming Language

Nora is a strictly typed, high-performance systems programming language designed to serve as a modern alternative to C++, Rust, and Go. It brings together the speed and predictability of low-level systems programming with a radically simplified approach to memory safety and concurrency.

Nora achieves **compile-time memory safety without a garbage collector** and **data-race-free concurrency without lock overhead** by leveraging a novel static analysis engine: the **Topological Lease Solver**.

## Key Features

### 🧠 Topological Lease Solver
Nora does not use a borrow checker that fights you. Instead, it uses a **Topological Lease Solver** that tracks variable dependencies, lifecycles, and automatically inserts RAII resource release calls (`Drop`, `PreDrop`). It understands partial moves, re-assignments, and handles memory safely behind the scenes.

### 🛡️ Explicit Ownership and Borrowing
Nora uses clear, minimal syntax markers to denote memory semantics, ensuring that data flow is always predictable:
- `@` **Owned Move**: Explicitly transfers ownership (consumption) of a value.
- `#` **Read-Only Borrow**: A safe, immutable lease of data.
- `&` **Mutable Borrow**: An exclusive, read-write lease.

### ⚡ Cooperative Fiber Runtime
Say goodbye to heavy OS threads and locking overhead. Nora features a built-in M:N cooperative fiber scheduler. It maps thousands of lightweight, stackless fibers onto OS threads. Combined with **Zero-Copy Channels**, concurrent data moves are compiled down to 8-byte pointer transfers, maximizing throughput.

### 🏗️ Type-Erased Shared Monomorphization
Unlike C++ or Rust where heavy generics can cause massive binary bloat and slow compile times, Nora uses a shared type-erasure model for generic structures. If generics only differ by pointers, Nora automatically merges them into a single, shared implementation—keeping your binaries tiny and compilation blazing fast.

### 🎯 C11 Backend Target
Nora compiles directly to highly optimized C11 code, allowing it to seamlessly tap into the vast ecosystem of C libraries, compile on any platform using `gcc` or `clang`, and easily integrate into existing codebases.

---

## A Quick Look at Nora

```nora
import "std/io"
import "std/collections"

pub type Vector[T] = struct {
    data: @T[]
    len: i32
    cap: i32
}

// Generics, explicit allocations, and mutable leases
pub fn NewVector[T](cap: i32) @Vector[T] {
    return alloc Vector[T] {
        data: alloc T[cap],
        len: 0,
        cap: cap,
    }
}

pub fn (v: &Vector[T]) Push[T](val: @T) {
    // ... vector push logic ...
}

pub fn main() {
    let my_vec = NewVector[str](10)
    my_vec.Push("Hello, Nora!")
    io.Println("Vector initialized successfully.")
}
```

---

## Getting Started

The Nora compiler comes batteries-included with a modern tooling suite.

**Initialize a new project:**
```bash
nora init my_app
cd my_app
```

**Run your project:**
```bash
nora run --release src/main.nr
```

### The Project Manifest (`nora.yaml`)
Nora uses a simple YAML-based manifest to handle dependencies, native C integrations, and compiler options cleanly:

```yaml
name: my_app
version: 1.0.0
entry: src/main.nr
output: my_app_bin
dependencies:
  my_lib:
    path: ../my_lib/src
    version: 1.0.0
```

---

## Project Philosophy

Our core directive:
> **Correctness, language consistency, and compiler maintainability are far more important than implementation speed.**

1. **Predictability:** Code must do exactly what it says. No hidden allocations, no silent memory copies, no invisible concurrency overhead.
2. **User Simplicity:** Syntactic constructs must remain clean and minimal.
3. **Documentation First:** No complex features or major architectural changes are implemented without a written specification and an Architecture Decision Record (ADR).

## Architecture & Pipeline

Nora's compiler is built in Go and operates in a multi-pass pipeline:
`Lexer -> Parser -> AST -> Symbol Scope Resolution -> Semantic Inference -> Topological Lease Solver -> HIR Lowering -> C11 Codegen -> Target Binary`

---

## Project Structure

A brief overview of the repository layout:

- **`pkg/`**: The core Go-based compiler source code. Contains the lexer, parser, semantic analyzer, topological lease solver, HIR lowerer, and the C11 code generator.
- **`core/`**: Foundational, non-allocating core primitives (like `Option`, `Result`, and core iterators) intrinsically linked to the compiler.
- **`std/`**: The Nora Standard Library. Contains advanced utilities like collections (`Vector`, `HashMap`), file I/O, networking, and the cooperative fiber runtime.
- **`examples/`**: Real-world sample projects written in Nora. Includes basic demos as well as complex architectural ports like `port_gecs` (a fully functioning ECS game engine).
- **`docs/`**: Our source of truth. Contains all language specifications, architecture decision records (ADRs), implementation plans, and technical walkthroughs.
- **`vscode-nora/`**: The official Visual Studio Code extension providing syntax highlighting and integration with our Language Server Protocol (LSP).
- **`pkg/cmd/test/`**: The integration testing suite. Contains extensive positive and negative (`fail_`) `.nr` files to rigorously test the compiler pipeline.

---

## Contributing
Please refer to the `docs/` folder for architectural decisions (`docs/adr/`), implementation plans (`docs/plans/`), and the official language specifications (`docs/specifications/`). All contributions must adhere strictly to the established compiler rules and ensure zero regressions in the integration test suite.

## Support and Donation
If you find the Nora Programming Language useful and would like to support its continued development, please consider making a donation. Your support helps us maintain the project, build new features, and expand the ecosystem.

## License
This project is licensed under the MIT License.
