package main

[repr("i32")]
pub type Color = enum {
    Undefined = 0,
    Red = 1,
    Green = 2,
    Blue = 3
}

pub type TextureDesc = struct {
    format: Color
}

pub type Pass = struct {
    depth_format: Color
}

pub fn (self: &Pass) Resize() {
    var _desc1 = alloc TextureDesc {
        format: self.depth_format
    }
    
    var _desc2 = alloc TextureDesc {
        format: self.depth_format
    }

    if self.depth_format == Color.Undefined {
        panic("BUG: depth_format was moved and zeroed!")
    }
}

pub fn main() {
    var pass = alloc Pass {
        depth_format: Color.Green
    }
    pass.Resize()
}
