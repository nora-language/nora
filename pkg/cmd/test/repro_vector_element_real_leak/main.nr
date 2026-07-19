package main

import "collections"
import "io"

pub type InnerBuffer = struct {
    id: i32
}

pub fn (_self: &InnerBuffer) drop() {
    io.PrintLn("InnerBuffer dropped")
}

pub type MeshElement = struct {
    vb: @InnerBuffer
}

pub type ContainerModel = struct {
    meshes: @collections.Vector[MeshElement]
}

pub fn (_self: &ContainerModel) drop() {
    io.PrintLn("ContainerModel custom drop ran!")
}

fn reproduce_vector_element_leak() {
    var vec = collections.NewVector[MeshElement](2)
    var buf = alloc InnerBuffer { id: 42 }
    var mesh = alloc MeshElement { vb: @buf }
    vec.Push(@mesh)
    
    var _model = alloc ContainerModel { meshes: @vec }
    // When `_model` goes out of scope, `_model.meshes` (@collections.Vector[MeshElement]) is dropped by AutoDropMethod.
    // However, because Vector[T] uses shared monomorphization (`_ptr`), freeing vec.data (@MeshElement[]) only releases the array memory,
    // skipping element-level recursive RAII drop of `vb: @InnerBuffer`.
}

fn main() i32 {
    reproduce_vector_element_leak()
    return 0
}
