# Project Manifest (`nora.yaml`)

## Overview

Every Nora workspace—whether an executable application or a reusable library—is defined by a project manifest file named `nora.yaml` located at the root of the project directory. It replaces traditional `.mod` or `.toml` files used in other languages.

## Motivation

The manifest acts as the central source of truth for the Nora compiler (`Nora build`), the Language Server Protocol (LSP), and the package manager. It explicitly defines how the project should be built, linked against native C code, and which dependencies it requires, ensuring reproducible builds across different environments.

## File Structure & Fields

A complete `nora.yaml` file looks like this:

```yaml
name: my_project
version: 1.0.0
language: 0.1.0
entry: src/main.nr
output: my_project_bin
no_stdlib: false
allow_unsafe: false
plugins: []
dependencies:
  my_lib:
    path: ../my_lib/src
    version: 1.0.0
native:
  compiler: clang
  opt_release: "-O3"
  opt_debug: "-O0"
  debug_symbols: "-g"
  out_flag: "-o"
  inc_flag: "-I"
  define_flag: "-D"
  lib_dir_flag: "-L"
  cflags: ["-pthread"]
  dynamic_libs: ["m", "dl"]
  static_libs: ["../my_lib/libcustom.a"]
  include_dirs: ["include/"]
  lib_dirs: ["lib/"]
  headers: ["my_header.h"]
  source_files: ["my_c_implementation.c"]
```

### Core Fields

*   **`name`** (string): The name of your project or library. This is the name other projects will use when importing your library.
*   **`version`** (string): The current version of your project (Semantic Versioning is recommended).
*   **`language`** (string): The minimum required version of the Nora compiler to build this project.
*   **`entry`** (string): The relative path to the main entry point file (usually `src/main.nr` for executables or `src/lib.nr` for libraries).
*   **`output`** (string): The desired name of the compiled binary executable.
*   **`no_stdlib`** (boolean): Optional flag. If `true`, the compiler and Language Server will not automatically inject the standard library prelude (`std/prelude.nr`). Used for bare-metal and `#![no_std]` equivalent environments.
*   **`allow_unsafe`** (boolean): Optional flag. If `true`, the compiler automatically grants the `[unsafe]` capability to all files in this project, eliminating the need for users to pass `--allow-unsafe` globally on the CLI.
*   **`plugins`** (list of strings): Any compiler plugins required during the build phase.

### Dependencies

The `dependencies` section maps import module names to their physical locations and version constraints.

*   **`[module_name]`**: The key is the name you use in your `import` statements (e.g., `import "my_lib"`).
    *   **`path`**: The local relative or absolute path to the dependency's source code (must point to a directory containing its own `nora.yaml`).
    *   **`version`**: The version constraint for the dependency.

### Native Integration (`native`)

Because Nora compiles to C11, the manifest provides native build configurations to link directly against C code.

*   **`compiler`** (string): Overrides the default native compiler (e.g., `clang`, `gcc`, `cl`).
*   **`opt_release`** (string): The optimization flag used when compiling in release mode (e.g., `-O3` or `/O2`).
*   **`opt_debug`** (string): The optimization flag used when compiling in debug mode (e.g., `-O0` or `/Od`).
*   **`debug_symbols`** (string): The flag used to generate debug symbols (e.g., `-g` or `/Zi`).
*   **`out_flag`** (string): The flag used to specify the output binary name (e.g., `-o` or `/Fe:`).
*   **`inc_flag`** (string): The flag used to specify include directories (e.g., `-I` or `/I`).
*   **`define_flag`** (string): The flag used to define preprocessor macros (e.g., `-D` or `/D`).
*   **`lib_dir_flag`** (string): The flag used to specify library search directories (e.g., `-L` or `/link /LIBPATH:`).
*   **`cflags`** (list of strings): Custom compiler flags passed directly to the native C compiler during the final build step.
*   **`dynamic_libs`** (list of strings): Names of dynamic libraries to link against (e.g., `m` for libm.so).
*   **`static_libs`** (list of strings): Paths to static library files to link into the final binary.
*   **`include_dirs`** (list of strings): Paths to directories containing C headers, appended using `inc_flag`.
*   **`lib_dirs`** (list of strings): Paths to directories containing pre-compiled C libraries, appended using `lib_dir_flag`.
*   **`headers`** (list of strings): Local C header files to include in the generated C output.
*   **`source_files`** (list of strings): Local `.c` implementation files to compile and link alongside the Nora-generated C code.

## Usage

When you run `Nora init <project-name>`, a default `nora.yaml` is automatically generated for you. 

When you run `Nora build`, the compiler:
1. Parses `nora.yaml`.
2. Resolves all `dependencies` recursively.
3. Compiles the Nora code to a unified C output.
4. Invokes the `native.compiler` with the specified `cflags`, linking the generated C code with your provided `native.source_files`.

## Errors & Diagnostics

*   **Missing Manifest:** If `Nora build` or the LSP cannot find `nora.yaml` in the current or parent directories, compilation will fail.
*   **Unresolved Dependency:** If an imported package is not listed in `dependencies` and does not exist in the standard library (`std/`), a compile-time "package not found" error is thrown.
