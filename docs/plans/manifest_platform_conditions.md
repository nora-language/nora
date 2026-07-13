# Platform-Conditional Manifest Configuration

**Status:** Planned  
**Author:** Nora Core Team  
**Date:** 2026-07-13

---

## Goal

Add support for **platform-conditional sections** in `nora.yaml` for the `native:` block and `dependencies:`. This allows cross-platform projects to declare different native headers, linker flags, dynamic libraries, and dependencies per target OS — directly in the manifest — without workarounds.

**Specification:** [`docs/specifications/manifest_platform_conditions.md`](../specifications/manifest_platform_conditions.md)

---

## Affected Compiler Components

| Component | File | Change Type |
|---|---|---|
| `ProjectConfig` struct | `pkg/cmd/nora/main.go` | Modify — add platform map fields |
| `NativeConfig` struct | `pkg/cmd/nora/main.go` | Modify — extend YAML tags |
| `Dependency` struct | `pkg/cmd/nora/main.go` | Modify — add `Platform` field |
| `LoadProjectConfig()` | `pkg/cmd/nora/main.go` | Modify — merge platform-specific config |
| `FileLoader` | `pkg/cmd/nora/main.go` | Modify — filter platform-excluded dependencies |
| LSP manifest reader | `pkg/lsp/handler.go` | Modify — reflect same manifest changes |

---

## Implementation Checklist

### Phase 1 — Data Structures

- [ ] **Extend `NativeConfig`** to hold a `Platforms map[string]NativePlatformConfig` field.
  - `NativePlatformConfig` contains the same list fields as `NativeConfig` (headers, cflags, dynamic_libs, etc.) but only the additive/override-able fields — not the platform map itself.
  - YAML key: `windows:`, `linux:`, `darwin:`, `wasm:`, `wasi:`

- [ ] **Extend `Dependency` struct** with `Platform string` field.
  - YAML key: `platform:`
  - Empty string means "all platforms".

```go
// Proposed NativePlatformConfig
type NativePlatformConfig struct {
    CompilerConfig `yaml:",inline"`
    DynamicLibs    []string `yaml:"dynamic_libs"`
    StaticLibs     []string `yaml:"static_libs"`
    IncludeDirs    []string `yaml:"include_dirs"`
    LibDirs        []string `yaml:"lib_dirs"`
    Headers        []string `yaml:"headers"`
    SourceFiles    []string `yaml:"source_files"`
}

// Updated NativeConfig
type NativeConfig struct {
    CompilerConfig `yaml:",inline"`
    DynamicLibs    []string                       `yaml:"dynamic_libs"`
    StaticLibs     []string                       `yaml:"static_libs"`
    IncludeDirs    []string                       `yaml:"include_dirs"`
    LibDirs        []string                       `yaml:"lib_dirs"`
    Headers        []string                       `yaml:"headers"`
    SourceFiles    []string                       `yaml:"source_files"`
    Platforms      map[string]NativePlatformConfig `yaml:",inline"`  // keys: "windows", "linux", etc.
}

// Updated Dependency
type Dependency struct {
    Path     string `yaml:"path"`
    Version  string `yaml:"version"`
    Platform string `yaml:"platform"` // optional, empty = all
}
```

### Phase 2 — Config Resolution

- [ ] **Add `resolveNativeConfig(base NativeConfig, targetOS string) NativeConfig`** helper.
  - Reads `base.Platforms[targetOS]` if it exists.
  - Concatenates all list fields: `global + platform_specific`.
  - Overrides scalar fields (like `Compiler`) if platform-specific value is non-empty.
  - Returns a flat merged `NativeConfig` with `Platforms` cleared.

- [ ] **Update `LoadProjectConfig()`** to call `resolveNativeConfig` after YAML unmarshalling, passing in the target OS.
  - The target OS must be passed into `LoadProjectConfig` or resolved before calling it.

- [ ] **Update `FileLoader` dependency loading** to skip dependencies where `dep.Platform != "" && dep.Platform != targetOS`.

### Phase 3 — LSP Integration

- [ ] Mirror the same struct changes in `pkg/lsp/handler.go`'s manifest parsing logic.
- [ ] LSP should pass host OS (`runtime.GOOS`) for dependency filtering (same as the CLI default).

### Phase 4 — `nora init` Template

- [ ] Update the generated `nora.yaml` template to include **commented-out examples** of the platform block for discoverability:

```yaml
native:
  # cflags: ["-Wall"]
  # windows:
  #   dynamic_libs: ["ws2_32"]
  #   headers: ["winsock2.h"]
  # linux:
  #   dynamic_libs: ["pthread"]
  #   headers: ["sys/socket.h"]
```

### Phase 5 — Diagnostics

- [ ] Emit a **warning** (not error) when an unknown key appears under `native:` that is neither a known field nor a known OS name.
- [ ] No diagnostic for missing platform key — it simply has no effect.

---

## Test Plan

### Unit Tests
- [ ] `resolveNativeConfig` with only global fields → passthrough unchanged.
- [ ] `resolveNativeConfig` with global + matching platform → lists concatenated, scalar overridden.
- [ ] `resolveNativeConfig` with global + non-matching platform → platform block ignored.
- [ ] `resolveNativeConfig` with no global, platform-only → platform fields used as-is.

### Integration Tests
- [ ] Positive test: Project with `native.linux:` compiles on Linux, `native.windows:` ignored.
- [ ] Positive test: Dependency with `platform: linux` is excluded from resolution on Windows (no linker errors, no symbol errors).
- [ ] Negative test: Code references a symbol from a platform-excluded dependency → `undefined identifier` error.
- [ ] `nora init` generates a `nora.yaml` with commented-out platform examples.

---

## Risks

| Risk | Mitigation |
|---|---|
| YAML `inline` map key conflicts with `NativeConfig` field names | Use explicit known OS set to separate platform keys from field keys at unmarshal time |
| `NativePlatformConfig` vs `NativeConfig` type duplication | Use embedding or a shared interface to avoid drift between the two structs |
| LSP manifest parsing diverges from CLI | Extract manifest loading into a shared `pkg/manifest` package used by both |

---

## Completion Criteria

- [ ] Specification exists at `docs/specifications/manifest_platform_conditions.md`.
- [ ] `ProjectConfig`, `NativeConfig`, and `Dependency` structs updated.
- [ ] `resolveNativeConfig()` helper implemented and unit-tested.
- [ ] Dependency filtering by `platform:` implemented.
- [ ] LSP updated to reflect struct changes.
- [ ] `nora init` template updated with commented-out platform examples.
- [ ] All new tests pass.
- [ ] `docs/specifications/project_manifest.md` updated to document the new fields.
- [ ] Walkthrough written under `docs/walkthroughs/`.
