pub type Vector3[T] = struct { x: T, y: T, z: T } 

pub fn (v: #Vector3[T]) Normalize[T]() @Vector3[T] { 
    return alloc Vector3[T] { x: v.x, y: v.y, z: v.z } 
} 
pub fn (v: #Vector3[T]) MulScalar[T](s: T) @Vector3[T] { 
    return alloc Vector3[T] { x: v.x, y: v.y, z: v.z } 
}

pub type Shape[T] = interface {
    fn Support(dir: #Vector3[T]) @Vector3[T]
}

pub type Sphere[T] = struct {
    radius: T
}

pub fn (s: #Sphere[T]) Support[T](dir: #Vector3[T]) @Vector3[T] {
    // Workaround for primitive moves using Sub operation
    var zero = s.radius - s.radius
    var cloned_radius = s.radius - zero
    
    var norm_dir = dir.Normalize[T]()
    return norm_dir.MulScalar[T](cloned_radius)
}

// Issue 3: Parser Corruption on inline function pointer with pointer return type
pub fn CallSupportInline[T](c: #Sphere[T], d: #Vector3[T], f: fn(#Sphere[T], #Vector3[T]) @Vector3[T]) @Vector3[T] {
    return f(c, d)
}

// Issue 4: Generic Type Aliases
pub type SupportFunc[T] = fn(#Sphere[T], #Vector3[T]) @Vector3[T]

// Issue 2: Strict Semantic Method Resolution
pub fn CallSupportGeneric[T, ShapeA](s: #ShapeA, d: #Vector3[T]) @Vector3[T] {
    // return s.Support[T](d) // Intentional semantic error
}

pub fn main() {
    var sphere = alloc Sphere[f64] { radius: 5.0 }
    var dir = alloc Vector3[f64] { x: 1.0, y: 0.0, z: 0.0 }
    
    // Issue 1: Generic Interface Codegen
    var shape: Shape[f64] = sphere
    var support_res1 = shape.Support(dir)
    
    // Issue 3
    // CallSupportInline(sphere)
}
