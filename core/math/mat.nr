package math

pub type Mat4 = struct {
    m: f32[16]
}

pub fn NewMat4() Mat4 {
    return Mat4 { m: [f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0)] }
}

pub fn Mat4Identity(out: &Mat4) {
    var i = 0
    while i < 16 { 
        out.m[i] = f32(0.0)
        i++ 
    }
    out.m[0] = f32(1.0)
    out.m[5] = f32(1.0)
    out.m[10] = f32(1.0)
    out.m[15] = f32(1.0)
}

pub fn Mat4Perspective(out: &Mat4, fov: f32, aspect: f32, near_plane: f32, far_plane: f32) {
    var f = f32(1.0) / f32(Tan(f64(fov * f32(0.5))))
    
    var i = 0
    while i < 16 { 
        out.m[i] = f32(0.0)
        i++ 
    }
    
    out.m[0] = f / aspect
    out.m[5] = f
    out.m[10] = far_plane / (near_plane - far_plane)
    out.m[11] = -f32(1.0)
    out.m[14] = (near_plane * far_plane) / (near_plane - far_plane)
}

pub fn Mat4Translation(out: &Mat4, v: #Vec3) {
    Mat4Identity(out)
    out.m[12] = v.x
    out.m[13] = v.y
    out.m[14] = v.z
}

pub fn Mat4Mul(out: &Mat4, a: #Mat4, b: #Mat4) {
    var c = 0
    while c < 4 {
        var r = 0
        while r < 4 {
            out.m[c*4 + r] = 
                a.m[0*4 + r] * b.m[c*4 + 0] +
                a.m[1*4 + r] * b.m[c*4 + 1] +
                a.m[2*4 + r] * b.m[c*4 + 2] +
                a.m[3*4 + r] * b.m[c*4 + 3]
            r++
        }
        c++
    }
}

pub fn Mat4LookAt(out: &Mat4, eye: #Vec3, target: #Vec3, up: #Vec3) {
    var f = Vec3Normalize(#Vec3Sub(target, eye))
    var s = Vec3Normalize(#Vec3Cross(#f, up))
    var u = Vec3Cross(#s, #f)
    
    Mat4Identity(out)
    out.m[0] = s.x
    out.m[4] = s.y
    out.m[8] = s.z
    
    out.m[1] = u.x
    out.m[5] = u.y
    out.m[9] = u.z
    
    out.m[2] = -f.x
    out.m[6] = -f.y
    out.m[10] = -f.z
    
    out.m[12] = -Vec3Dot(#s, eye)
    out.m[13] = -Vec3Dot(#u, eye)
    out.m[14] = Vec3Dot(#f, eye)
}
