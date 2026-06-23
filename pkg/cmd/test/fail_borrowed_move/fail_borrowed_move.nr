pub type Vector[T] = struct {
    val: T
}

pub fn (v: #Vector[T]) MulScalar[T](s: T) @Vector[T] {
    return alloc Vector[T] { val: s }
}

pub type Sphere[T] = struct {
    radius: T
}

pub fn (s: #Sphere[T]) Support[T](dir: #Vector[T]) @Vector[T] {
    // s.radius is accessed via a borrowed struct (`#Sphere[T]`).
    // s.radius has generic type `T`, which is strictly linear.
    // Calling MulScalar[T](s: T) demands ownership of T.
    // The compiler correctly catches this and throws a topology error.
    return dir.MulScalar[T](s.radius) // ERROR: cannot move out of borrowed context
}

pub fn main() {
    var sphere = alloc Sphere[f64] { radius: 5.0 }
    var dir = alloc Vector[f64] { val: 1.0 }
    var res = sphere.Support[f64](dir)
}
