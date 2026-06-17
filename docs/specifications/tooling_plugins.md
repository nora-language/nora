# WebAssembly Plugin System

## Overview

Nora features a highly robust, sandboxed plugin system implemented via `pkg/plugin`. This system enables developers to extend the compiler with custom AST macros, code generators, and linters without compromising the host machine's security or stability. 

## The Wazero Sandbox

Nora integrates the `github.com/tetratelabs/wazero` WebAssembly runtime. When a plugin is loaded, it executes completely isolated inside a WASM sandbox. This ensures that third-party compiler plugins cannot maliciously read the file system, execute shell commands, or access the network without explicit, scoped permission.

## Memory Exchange Protocol

Because the WASM plugin runs in an isolated memory space, the Nora compiler and the plugin communicate via a strict JSON memory-exchange protocol:

1.  **Allocation:** The compiler calls an exported `plugin_alloc` function inside the WASM module to allocate contiguous memory.
2.  **Request:** The compiler serializes the AST node or contextual data into a JSON string and writes it directly into the allocated WASM memory.
3.  **Execution:** The compiler invokes the specific macro function (e.g., `macro_derive_json`).
4.  **Response:** The WASM module processes the JSON, allocates memory for the result, and returns the pointer. The compiler reads the null-terminated JSON response.
5.  **Reset:** The compiler calls `plugin_reset` to clean the plugin's heap for the next execution.

## Native Go Macros

To provide fallback functionality and zero-overhead internal macros, the `PluginManager` also supports registering Native Go Macros.

```go
func (m *PluginManager) RegisterNativeMacro(pluginName string, macroName string, handler NativeMacro)
```

If the compiler attempts to invoke a macro and the corresponding `.wasm` file is not found or fails to load, it will check the native registry. If a match is found, the Go function is executed directly, bypassing the WebAssembly sandbox.

## Adding Plugins

Plugins are registered in the project manifest (`nora.yaml`) under the `plugins` array. During the `Nora build` initialization phase, the compiler reads the manifest, resolves the paths to the pre-compiled `.wasm` files, and instantiates them via the `PluginManager`.
