package main

import "collections"

pub type Serializer = interface {
    fn Serialize() i32
}

pub type SerializerImpl[T] = struct {
    val: T
}

pub fn (s: #SerializerImpl[T]) Serialize() i32 {
    return 42
}

fn main() i32 {
    var m = collections.NewHashMap[i32, Serializer](10)
    
    var impl = SerializerImpl[i32] { val: 1 }
    var s: Serializer = impl

    m.Insert(1, s)
    return 0
}
