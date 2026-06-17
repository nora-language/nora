package core

pub type Mat4 = struct {
    m00: f32, m01: f32, m02: f32, m03: f32,
    m10: f32, m11: f32, m12: f32, m13: f32,
    m20: f32, m21: f32, m22: f32, m23: f32,
    m30: f32, m31: f32, m32: f32, m33: f32
}

pub fn Mat4Identity() Mat4 {
    return Mat4 {
        m00: f32(1.0), m01: f32(0.0), m02: f32(0.0), m03: f32(0.0),
        m10: f32(0.0), m11: f32(1.0), m12: f32(0.0), m13: f32(0.0),
        m20: f32(0.0), m21: f32(0.0), m22: f32(1.0), m23: f32(0.0),
        m30: f32(0.0), m31: f32(0.0), m32: f32(0.0), m33: f32(1.0)
    }
}

extern fn tanf(x: f32) f32
extern fn cosf(x: f32) f32
extern fn sinf(x: f32) f32
extern fn sqrtf(x: f32) f32

pub fn Mat4Perspective(fov_rad: f32, aspect: f32, z_near: f32, z_far: f32) Mat4 {
    var f = f32(1.0) / tanf(fov_rad * f32(0.5))
    
    return Mat4 {
        m00: f / aspect, m01: f32(0.0), m02: f32(0.0),                  m03: f32(0.0),
        m10: f32(0.0),   m11: f,        m12: f32(0.0),                  m13: f32(0.0),
        m20: f32(0.0),   m21: f32(0.0), m22: z_far / (z_near - z_far),        m23: f32(-1.0),
        m30: f32(0.0),   m31: f32(0.0), m32: (z_near * z_far) / (z_near - z_far), m33: f32(0.0)
    }
}

pub fn Mat4Translation(x: f32, y: f32, z: f32) Mat4 {
    return Mat4 {
        m00: f32(1.0), m01: f32(0.0), m02: f32(0.0), m03: f32(0.0),
        m10: f32(0.0), m11: f32(1.0), m12: f32(0.0), m13: f32(0.0),
        m20: f32(0.0), m21: f32(0.0), m22: f32(1.0), m23: f32(0.0),
        m30: x,        m31: y,        m32: z,        m33: f32(1.0)
    }
}

pub fn Mat4Scale(x: f32, y: f32, z: f32) Mat4 {
    return Mat4 {
        m00: x,        m01: f32(0.0), m02: f32(0.0), m03: f32(0.0),
        m10: f32(0.0), m11: y,        m12: f32(0.0), m13: f32(0.0),
        m20: f32(0.0), m21: f32(0.0), m22: z,        m23: f32(0.0),
        m30: f32(0.0), m31: f32(0.0), m32: f32(0.0), m33: f32(1.0)
    }
}

pub fn Mat4LookAt(eye_x: f32, eye_y: f32, eye_z: f32, center_x: f32, center_y: f32, center_z: f32, up_x: f32, up_y: f32, up_z: f32) Mat4 {
    var f_x = center_x - eye_x
    var f_y = center_y - eye_y
    var f_z = center_z - eye_z
    
    var f_len = sqrtf(f_x*f_x + f_y*f_y + f_z*f_z)
    f_x = f_x / f_len
    f_y = f_y / f_len
    f_z = f_z / f_len
    
    var s_x = f_y * up_z - f_z * up_y
    var s_y = f_z * up_x - f_x * up_z
    var s_z = f_x * up_y - f_y * up_x
    
    var s_len = sqrtf(s_x*s_x + s_y*s_y + s_z*s_z)
    s_x = s_x / s_len
    s_y = s_y / s_len
    s_z = s_z / s_len
    
    var u_x = s_y * f_z - s_z * f_y
    var u_y = s_z * f_x - s_x * f_z
    var u_z = s_x * f_y - s_y * f_x
    
    return Mat4 {
        m00: s_x, m01: u_x, m02: -f_x, m03: f32(0.0),
        m10: s_y, m11: u_y, m12: -f_y, m13: f32(0.0),
        m20: s_z, m21: u_z, m22: -f_z, m23: f32(0.0),
        m30: -(s_x*eye_x + s_y*eye_y + s_z*eye_z),
        m31: -(u_x*eye_x + u_y*eye_y + u_z*eye_z),
        m32: (f_x*eye_x + f_y*eye_y + f_z*eye_z),
        m33: f32(1.0)
    }
}

pub fn Mat4Multiply(a: Mat4, b: Mat4) Mat4 {
    return Mat4 {
        m00: a.m00*b.m00 + a.m01*b.m10 + a.m02*b.m20 + a.m03*b.m30,
        m01: a.m00*b.m01 + a.m01*b.m11 + a.m02*b.m21 + a.m03*b.m31,
        m02: a.m00*b.m02 + a.m01*b.m12 + a.m02*b.m22 + a.m03*b.m32,
        m03: a.m00*b.m03 + a.m01*b.m13 + a.m02*b.m23 + a.m03*b.m33,

        m10: a.m10*b.m00 + a.m11*b.m10 + a.m12*b.m20 + a.m13*b.m30,
        m11: a.m10*b.m01 + a.m11*b.m11 + a.m12*b.m21 + a.m13*b.m31,
        m12: a.m10*b.m02 + a.m11*b.m12 + a.m12*b.m22 + a.m13*b.m32,
        m13: a.m10*b.m03 + a.m11*b.m13 + a.m12*b.m23 + a.m13*b.m33,

        m20: a.m20*b.m00 + a.m21*b.m10 + a.m22*b.m20 + a.m23*b.m30,
        m21: a.m20*b.m01 + a.m21*b.m11 + a.m22*b.m21 + a.m23*b.m31,
        m22: a.m20*b.m02 + a.m21*b.m12 + a.m22*b.m22 + a.m23*b.m32,
        m23: a.m20*b.m03 + a.m21*b.m13 + a.m22*b.m23 + a.m23*b.m33,

        m30: a.m30*b.m00 + a.m31*b.m10 + a.m32*b.m20 + a.m33*b.m30,
        m31: a.m30*b.m01 + a.m31*b.m11 + a.m32*b.m21 + a.m33*b.m31,
        m32: a.m30*b.m02 + a.m31*b.m12 + a.m32*b.m22 + a.m33*b.m32,
        m33: a.m30*b.m03 + a.m31*b.m13 + a.m32*b.m23 + a.m33*b.m33
    }
}

