package rendering

pub type Vertex = struct {
    px: f32, py: f32, pz: f32,
    cx: f32, cy: f32, cz: f32,
    nx: f32, ny: f32, nz: f32
}

pub fn CreateCubeVertices() Vertex[] {
    var verts = make(Vertex[], 36)
    
    // Front face (Red)
    verts[0] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(1.0) }
    verts[1] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(1.0) }
    verts[2] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(1.0) }
    verts[3] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(1.0) }
    verts[4] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(1.0) }
    verts[5] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(1.0) }

    // Back face (Green)
    verts[6] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(-1.0) }
    verts[7] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(-1.0) }
    verts[8] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(-1.0) }
    verts[9] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(-1.0) }
    verts[10] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(-1.0) }
    verts[11] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(0.0), nx: f32(0.0), ny: f32(0.0), nz: f32(-1.0) }

    // Left face (Blue)
    verts[12] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(0.0), cz: f32(1.0), nx: f32(-1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[13] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(0.0), cy: f32(0.0), cz: f32(1.0), nx: f32(-1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[14] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(1.0), cx: f32(0.0), cy: f32(0.0), cz: f32(1.0), nx: f32(-1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[15] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(1.0), cx: f32(0.0), cy: f32(0.0), cz: f32(1.0), nx: f32(-1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[16] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(0.0), cz: f32(1.0), nx: f32(-1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[17] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(0.0), cz: f32(1.0), nx: f32(-1.0), ny: f32(0.0), nz: f32(0.0) }

    // Right face (Yellow)
    verts[18] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(1.0), cz: f32(0.0), nx: f32(1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[19] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(1.0), cy: f32(1.0), cz: f32(0.0), nx: f32(1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[20] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(1.0), cy: f32(1.0), cz: f32(0.0), nx: f32(1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[21] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(1.0), cy: f32(1.0), cz: f32(0.0), nx: f32(1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[22] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(1.0), cz: f32(0.0), nx: f32(1.0), ny: f32(0.0), nz: f32(0.0) }
    verts[23] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(1.0), cz: f32(0.0), nx: f32(1.0), ny: f32(0.0), nz: f32(0.0) }

    // Top face (Cyan)
    verts[24] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(1.0), nx: f32(0.0), ny: f32(1.0), nz: f32(0.0) }
    verts[25] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(1.0), nx: f32(0.0), ny: f32(1.0), nz: f32(0.0) }
    verts[26] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(1.0), nx: f32(0.0), ny: f32(1.0), nz: f32(0.0) }
    verts[27] = Vertex { px: f32(1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(1.0), nx: f32(0.0), ny: f32(1.0), nz: f32(0.0) }
    verts[28] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(-1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(1.0), nx: f32(0.0), ny: f32(1.0), nz: f32(0.0) }
    verts[29] = Vertex { px: f32(-1.0), py: f32(1.0), pz: f32(1.0), cx: f32(0.0), cy: f32(1.0), cz: f32(1.0), nx: f32(0.0), ny: f32(1.0), nz: f32(0.0) }

    // Bottom face (Magenta)
    verts[30] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(1.0), nx: f32(0.0), ny: f32(-1.0), nz: f32(0.0) }
    verts[31] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(1.0), nx: f32(0.0), ny: f32(-1.0), nz: f32(0.0) }
    verts[32] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(1.0), nx: f32(0.0), ny: f32(-1.0), nz: f32(0.0) }
    verts[33] = Vertex { px: f32(-1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(1.0), nx: f32(0.0), ny: f32(-1.0), nz: f32(0.0) }
    verts[34] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(-1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(1.0), nx: f32(0.0), ny: f32(-1.0), nz: f32(0.0) }
    verts[35] = Vertex { px: f32(1.0), py: f32(-1.0), pz: f32(1.0), cx: f32(1.0), cy: f32(0.0), cz: f32(1.0), nx: f32(0.0), ny: f32(-1.0), nz: f32(0.0) }

    return verts
}
