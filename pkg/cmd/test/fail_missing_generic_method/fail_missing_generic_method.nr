package main

type Vector3[T] = struct {
    x: T,
    y: T,
    z: T
}

pub fn TriggerCrash[T](val: T) {
    var v1 = Vector3[T]{x: val, y: val, z: val}
    var v2 = v1.Clone[T]()
}

fn main() i32 {
    TriggerCrash[f64](0.0)
    return 0
}
