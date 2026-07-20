package main

pub fn main() {
    var sb = NewSoftBody[f64](0.0)
    sb.Solve[f64]()
}

pub type SoftBody[T: Copy] = struct {
    data: T
}

pub fn NewSoftBody[T: Copy](val: T) @SoftBody[T] {
    return alloc SoftBody[T] { data: val }
}

pub type Vector3[T: Copy] = struct {
    x: T
}

pub fn ConsumeVec[T: Copy](v: @Vector3[T]) {}

pub fn (sb: &SoftBody[T]) Solve[T: Copy]() {
    var v_dummy = alloc Vector3[T] { x: sb.data }

    // First block: declares `g1` as a leased pointer to Vector3 (no drop flag).
    // The C type will be `Vector3* g1;`.
    var i = 0
    while i < 1 {
        var g1 = #v_dummy
        var _use_it = g1.x
        i = i + 1
    }

    // Second block: declares `g1` as an owned Vector3 allocation (has drop flag).
    // The C type will also be `Vector3* g1;`, so no pointer mismatch warnings are generated,
    // but without the compiler fix, `nr_df_g1` would fail to generate!
    var j = 0
    while j < 1 {
        var g1 = alloc Vector3[T] { x: sb.data }
        ConsumeVec[T](g1) 
        j = j + 1
    }
}
