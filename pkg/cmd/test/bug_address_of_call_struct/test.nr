pub type Vec3 = struct { x: f32, y: f32, z: f32 }

pub fn NewVec(x: f32) Vec3 {
    return Vec3 { x: x, y: x, z: x }
}

pub fn ProcessVec(v: #Vec3) f32 {
    return v.x
}

pub fn main() {
    var result = ProcessVec(@NewVec(f32(1.0)))
}
