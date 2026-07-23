package generic_method_copy

pub type Vector3[T: Copy] = struct {
    x: T,
    y: T,
    z: T
}

pub fn (v: &Vector3[T]) CloneX(val: T) {
    // Before the compiler fix, T here would lose its 'Copy' constraint 
    // because it was implicitly extracted from the receiver without checking the struct definition.
    // That would cause 'val' to be treated as an owned type, resulting in a 'use of moved value' error below.
    var a = val
    var b = val
    v.x = a
    v.y = b
}

pub fn main() {
    var v = alloc Vector3[f64]{x: 0.0, y: 0.0, z: 0.0}
    v.CloneX(42.0)
}
