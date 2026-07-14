package math

pub type Vec2 = struct {
    x: f32,
    y: f32
}

pub fn NewVec2(x: f32, y: f32) Vec2 { return Vec2 { x: x, y: y } }

pub type Vec3 = struct {
    x: f32,
    y: f32,
    z: f32
}

pub fn NewVec3(x: f32, y: f32, z: f32) Vec3 { return Vec3 { x: x, y: y, z: z } }

pub type Vec4 = struct {
    x: f32,
    y: f32,
    z: f32,
    w: f32
}

pub fn NewVec4(x: f32, y: f32, z: f32, w: f32) Vec4 { return Vec4 { x: x, y: y, z: z, w: w } }

pub fn Vec3Sub(a: #Vec3, b: #Vec3) Vec3 {
    return NewVec3(a.x - b.x, a.y - b.y, a.z - b.z)
}

pub fn Vec3Dot(a: #Vec3, b: #Vec3) f32 {
    return a.x * b.x + a.y * b.y + a.z * b.z
}

pub fn Vec3Cross(a: #Vec3, b: #Vec3) Vec3 {
    return NewVec3(
        a.y * b.z - a.z * b.y,
        a.z * b.x - a.x * b.z,
        a.x * b.y - a.y * b.x
    )
}

pub fn Vec3Length(a: #Vec3) f32 {
    return f32(Sqrt(f64(Vec3Dot(a, a))))
}

pub fn Vec3Normalize(a: #Vec3) Vec3 {
    var len = Vec3Length(a)
    if len == f32(0.0) { return NewVec3(f32(0.0), f32(0.0), f32(0.0)) }
    return NewVec3(a.x / len, a.y / len, a.z / len)
}
