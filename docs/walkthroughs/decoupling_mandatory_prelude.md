# Walkthrough: Decoupling Mandatory Prelude Injection

We have successfully decoupled the Nora compiler and Language Server from inherently requiring the `std/prelude.nr` file! This fulfills Step 2 of the `compiler_stdlib_coupling.md` report and opens the door for fully independent `#![no_std]` environments (like bare-metal kernels).

## What Was Changed

### 1. `no_stdlib` Project Configuration
We updated the core `ProjectConfig` and `LSPFileLoader` structures to recognize a new configuration flag. A developer can now specify `no_stdlib: true` inside their `nora.yaml` project manifest to prevent the automatic injection of the standard library prelude.

### 2. CLI `--no-stdlib` Flag
We added a new CLI flag to the `Nora build`, `Nora run`, and `Nora test` commands. Passing `--no-stdlib` overrides the project manifest and guarantees that the compiler will strictly skip auto-loading `std/prelude.nr`.

### 3. Language Server (LSP) Support
The Language Server (`pkg/lsp/handler.go`) was refactored so that `loader.loadManifest` runs *before* the prelude is evaluated. If a user is working on a `#![no_std]` project, their IDE will accurately respect the `nora.yaml` flag, preventing false "duplicate definition" or missing syntax errors that would otherwise occur if the LSP forcibly injected the prelude into a bare-metal environment.

### 4. Decoupling Compiler Primitive Errors
Previously, if a user typed `0..10` and the `Range` struct was not found, the compiler would crash with a hardcoded `compiler error: Range type not found in prelude`. 

We have refactored `pkg/semantic/analyzer.go` to emit a generalized error: `"compiler error: Range type not found in scope (required for range syntax)"`. This elegantly signals to the user that they can supply their own `Range` struct definition if they are operating without the standard library.

## Validation Results

The full Nora integration test suite (`go test ./pkg/cmd/nora -v`) completed in ~327 seconds with exactly **0 failures**. 

> [!TIP]
> The compiler will now flawlessly fall back on user-defined primitives if the `--no-stdlib` flag is provided, paving the way for embedded systems development in Nora!
