package pkg_generic

import "pkg_math"

pub type RackPinion[T] = struct {
    val: pkg_math.Vector3[T]
}

pub fn NewRackPinion[T](v: pkg_math.Vector3[T]) @RackPinion[T] {
    return alloc RackPinion[T] {
        val: v
    }
}

pub fn (self: #RackPinion[T]) Dummy[T]() bool {
    return true
}

pub type Constraint[T] = enum {
    RPVariant(rp: @RackPinion[T])
}
