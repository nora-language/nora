# Walkthrough: Platform-Conditional Manifest Configuration

## Overview
This feature introduces the ability to conditionally load dependencies and apply native C configurations based on the target operating system (e.g., `windows`, `linux`). This enables building cross-platform Nora applications where certain dependencies or C flags are specific to an OS.

## Changes Made

1. **Manifest Parsing (`pkg/cmd/nora/main.go`)**
   - Added `Platform` field to the `Dependency` struct to allow filtering dependencies based on the current OS.
   - Introduced `NativePlatformConfig` and updated `NativeConfig` to include an inline map `Platforms map[string]NativePlatformConfig`. This allows users to nest platform-specific configs under keys like `windows` or `linux`.
   - Added `resolveNativeConfig` helper to merge global native configurations with platform-specific ones during project initialization.
   - Updated `FileLoader.loadManifest` and `LoadProjectConfig` to filter out dependencies whose `platform` value does not match `runtime.GOOS`.
   - Updated the `nora init` generated template to include commented-out examples of platform-conditional blocks.

2. **LSP Integration (`pkg/lsp/handler.go`)**
   - Mirrored struct changes (`ProjectConfig`, `NativeConfig`, `Dependency`).
   - Updated `loadManifest` and `Load` logic to skip platform-mismatched dependencies, ensuring accurate syntax highlighting and autocomplete for the target OS.

3. **Compiler Testing (`pkg/cmd/nora/compiler_test.go`)**
   - Updated `TestCompilerWithTestFolder` to automatically load the `nora.yaml` manifest found in the specific test case folder. This was essential to enable proper end-to-end integration testing of project manifests.
   - Created the `pkg/cmd/test/manifest_platforms/` integration test folder containing:
     - `nora.yaml` specifying `win_dep` for Windows and `lin_dep` for Linux.
     - `main_windows.nr` and `main_linux.nr` to test positive platform dependency inclusion.
   - Created the `pkg/cmd/test/fail_platform_deps/` negative test folder containing files that expect compiler failures when attempting to import a dependency excluded by the current platform constraint.

4. **Documentation (`docs/specifications/project_manifest.md`)**
   - Documented the new `platform` field on dependencies.
   - Documented the structure and nested fields available for platform-specific native configurations.

## Validation Results

- **Unit and Integration Tests**: Ran `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder`. Positive integration tests correctly skip non-target files while successfully compiling valid imports. Negative integration tests correctly trigger "package not found" semantic errors when attempting to access platform-restricted modules.

This completes the implementation of platform-conditional project configuration.
