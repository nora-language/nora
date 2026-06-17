# Compiler CLI Toolchain

## Overview

The Nora compiler comes with an integrated command-line interface (CLI) to handle project initialization, building, execution, formatting, and Language Server Protocol (LSP) functionality. All actions are routed through the central `Nora` binary.

## Core Commands

### 1. Initialization (`Nora init`)
Initializes a new Nora workspace in the current directory, generating a standard `nora.yaml` manifest and standard directory layout.

*   `Nora init <project-name>`: Scaffolds a standard executable project (with `src/main.nr`).
*   `Nora init --lib <project-name>` (or `-l`): Scaffolds a library project. This automatically generates a `src/lib.nr` file and creates a runnable example under `examples/basic.nr` for a premium Developer Experience (DX).

### 2. Build (`Nora build`)
Compiles the Nora project to an executable or library. It reads the `nora.yaml` file to resolve dependencies and compile the source files.

*   `Nora build`: Compiles the entry point specified in the manifest.
*   `Nora build [filename]`: Compiles a specific `.nr` file, bypassing the manifest entry point.
*   `Nora build --example <name>`: Builds a specific project example located in the `examples/` directory.

### 3. Execution (`Nora run`)
Compiles the target and immediately executes the resulting binary. It accepts the same path arguments as `build`.

### 4. Integration Tests (`Nora test`)
Executes the internal compiler integration test suite or project-level tests.

### 5. Formatting (`Nora fmt`)
Invokes the built-in code formatter on the workspace to ensure strict adherence to the official Nora style guide.

### 6. Language Server (`Nora lsp`)
Starts the Language Server Protocol (LSP) process, communicating over standard I/O for IDE integration (e.g., VS Code).

## Compilation Flags

The `Nora build` and `Nora run` commands accept several flags to alter the compilation process:

*   **`-r`, `--release`**: Compiles the binary in release mode. Strips local diagnostics and applies aggressive C11 compiler optimizations (e.g., `-O3`).
*   **`-d`, `--debug`**: Compiles the binary in debug mode (default). Retains local diagnostic tracking and line numbers.
*   **`--no-stdlib`**: Disables the automatic injection of the standard library prelude (`std/prelude.nr`). This is used for compiling bare-metal environments or `#![no_std]` equivalent code.
*   **`--debug-memory`**: Activates deep memory leak checking. The generated C binary will automatically invoke `nr_mem_report()` upon termination to list all unfreed allocations.
*   **`-g`**: Emits source-mapping and C-level debug symbols for use with standard debuggers (like GDB or LLDB).
*   **`--target <triple>`**: Cross-compiles the output. Supported targets include native platforms (e.g., `windows-amd64`) and WebAssembly variants (`wasm`, `wasi`).
*   **`--verbose`**: Outputs highly detailed logs representing the internal compilation pipeline, AST lowering stages, lease solving metrics, and the final GCC/Clang invocation commands.
