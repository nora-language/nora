# Custom Drop & Vector Element RAII Cleanup Fix Plan

## Status
Proposed / Ready for Implementation

## Metadata
- **Author**: Antigravity & User
- **Affected Compiler Components**: `pkg/codegen/generator.go`, `pkg/topology/solver.go`
- **Related Investigation**: `docs/investigations/custom_drop_suppresses_auto_drop_and_vector_element_leaks.md`

## Goal
Eliminate memory leaks caused when a struct defines an explicit `drop()` method (`pub fn (self: &Model) drop()`) by ensuring that the compiler's generated `AutoDropMethod` (`nr_drop_Model`) automatically invokes both the user-defined `drop()` method and recursively drops all owned `@` fields (`model.meshes`, `model.materials`).

## Architectural Solution

### Priority Alignment (`GEMINI.md` Priority Hierarchy)
1. **Language Consistency**: In modern systems languages with deterministic RAII (like Rust), implementing custom `Drop` on a struct allows executing custom code *right before* fields are destroyed. Custom `drop()` must never suppress automatic recursive teardown of `@` owned struct fields.
2. **Predictability**: Defining `drop()` on `Model` to destroy WGPU device handles must not silently cause its collection fields (`@collections.Vector[Mesh]`) to leak.
3. **Compiler Maintainability**: We cleanly decouple looking up explicit user methods (`getUserDropMethod`) from requesting unified struct/sum auto-drop wrappers (`getDropMethod`).

### Part 1: Decouple User `drop` Lookup (`getUserDropMethod`) from Auto-Drop Request (`getDropMethod`)
Currently across `pkg/codegen/generator.go`:
```go
dropMethod := g.getDropMethod(t)
if dropMethod == "" && g.hasOwnedFields(st) {
    dropMethod = g.requestAutoDrop(st)
}
```
If `getDropMethod(t)` locates a user-defined `drop()` method (`nr_pkg_gfx_Model_drop`), `requestAutoDrop(st)` is skipped and `@` fields (`self->meshes`) are never freed.

We will refactor `getDropMethod` in `pkg/codegen/generator.go`:
1. Rename current `getDropMethod(t types.NRType)` logic (which inspects `g.SemanticInfo.MethodSymbols`) to `getUserDropMethod(t types.NRType) string`.
2. Update `getDropMethod(t types.NRType) string` to:
   ```go
   func (g *Generator) getDropMethod(t types.NRType) string {
       if t == nil {
           return ""
       }
       base := t
       if pt, ok := t.(*types.PointerType); ok {
           base = pt.Base
       }
       // If the struct has owned linear fields (@), ALWAYS return the unified AutoDropMethod wrapper
       if st, ok := base.(*types.StructType); ok && g.hasOwnedFields(st) {
           return g.requestAutoDrop(st)
       }
       if sum, ok := base.(*types.SumType); ok && g.hasOwnedSumFields(sum) {
           return g.requestAutoDrop(sum)
       }
       // Otherwise (e.g. struct with raw handles like Buffer), return user method directly
       return g.getUserDropMethod(t)
   }
   ```

### Part 2: Invoke User `drop()` Inside `AutoDropMethod` (`emitAutoDropMethods`)
In `pkg/codegen/generator.go` (`emitAutoDropMethods`):
When emitting the C body for `void nr_drop_Model(Model* self)`:
```go
if st, ok := t.(*types.StructType); ok {
    // 1. First execute user-defined custom drop method if present
    if userDrop := g.getUserDropMethod(t); userDrop != "" {
        g.emit("    %s(NULL, self);", userDrop)
    }
    // 2. Then automatically drop all owned linear (@) fields
    for _, fName := range st.FieldNames {
        fType := st.Fields[fName]
        if types.IsOwnedType(fType) {
            fExpr := fmt.Sprintf("self->%s", fName)
            g.emitDrop(fExpr, fType, g.isPointerTypeInC(fType))
        }
    }
}
```

## Implementation Checklist
- [ ] Rename `getDropMethod(t)` symbol lookup in `pkg/codegen/generator.go` to `getUserDropMethod(t)`.
- [ ] Implement new `getDropMethod(t)` inside `pkg/codegen/generator.go` routing structs/sum-types with `@` fields through `requestAutoDrop`.
- [ ] Update `emitAutoDropMethods()` in `pkg/codegen/generator.go` to invoke `userDrop := g.getUserDropMethod(t)` prior to field teardown loops.
- [ ] Validate `go test -v ./pkg/cmd/nora` across all existing tests.
- [ ] Verify `repro_vector_element_real_leak` passes when `ContainerModel.drop()` is present (`isExpectedLeak` vs `0` leaks verification).
- [ ] Create positive regression test confirming custom `drop()` runs in lockstep with auto-dropping `@` struct fields.

## Test Plan & Verification
1. Run `go test ./pkg/cmd/nora -run TestCompilerWithTestFolder` to ensure zero regressions across all integration tests.
2. Verify that defining `pub fn (self: &Model) drop()` on structs with `@` collection fields outputs both custom drop logs and clean 0-byte memory leak reports under `-debug-memory`.
