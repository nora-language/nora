package main

import "collections"

pub type ComponentMeta = protocol {
    pub fn CreateArray() any
}

pub type ComponentMetaImpl[T] = struct { }

pub fn (_m: &ComponentMetaImpl[T]) CreateArray[T]() any {
    var vec = collections.NewVector[T](10)
    var res: any = @vec
    return res
}

pub fn main() {
    var meta_impl = alloc ComponentMetaImpl[i32] {}
    var meta: ComponentMeta = meta_impl
    var arr_any = meta.CreateArray()

    // Downcast any to &Vector[T]
    var vec_ref: &collections.Vector[i32] = collections.Vector[i32](&arr_any)
    
    // Perform push, which will crash or corrupt if vec_ref isn't correctly resolved
    vec_ref.Push(42)
    vec_ref.Push(100)

    if (vec_ref.size != 2) {
        panic("Size mismatch!")
    }
    
    if (#vec_ref.data[0] != 42) {
        panic("Value 0 mismatch!")
    }

    if (#vec_ref.data[1] != 100) {
        panic("Value 1 mismatch!")
    }
}
