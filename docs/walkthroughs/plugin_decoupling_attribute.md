# Walkthrough: Decoupling Plugins from Core Compiler via [plugin_override]

## Summary of Changes

The Nora compiler has been successfully decoupled from the WebAssembly JSON serialization plugin. Previously, the core compiler files (nalyzer.go and hir_codegen.go) contained hardcoded references to 
r_serialize_json and 
r_deserialize_json. To make the plugin system scalable, we introduced a general-purpose attribute hook, [plugin_override], to handle these use cases dynamically for *any* plugin-generated function.

### 1. pkg/semantic/analyzer.go`n- **Removed Hardcodes**: The explicit string matching for nr_serialize_json was removed from esolveGenericCall.
- **General Plugin Hook**: Introduced getMangledPluginFuncName, which checks if the generic function has the [plugin_override] attribute. If it does, the generic call is mangled based on standard plugin naming rules: [TargetPkg]_[BaseFuncName]_[TargetType].

### 2. pkg/codegen/hir_codegen.go`n- **Removed Hardcodes**: The explicit string matches in Call, Spawn, and AddressOf instruction handlers were removed.
- **Dynamic C isExtern Check**: Since plugin overrides are marked extern fn in Nora code to satisfy the frontend parser, but are actually implemented fully in Nora by the WASM plugin generator, they must not be treated as C externs during codegen. The new logic checks if nStmt.IsExtern is true but also has [plugin_override]. If so, it disables isExtern, ensuring the standard _env_ptr and Nora function environment is passed correctly.

### 3. Standard Library (std/serialize/serialize.nr)
Tagged the builtin JSON intrinsics with the new attribute:
``nora
[plugin_override]
extern fn nr_serialize_json[T](val: #T) str

[plugin_override]
extern fn nr_deserialize_json[T](data: str) T
``n
## Validation Results
- The integration test suite (serialization_test.nr) was successfully run. The JSON serializer behaved exactly as intended, confirming the [plugin_override] system completely replaces the legacy hardcodes without breaking serialization.
