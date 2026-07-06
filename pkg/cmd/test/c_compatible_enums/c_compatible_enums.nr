package c_compatible_enums

[repr("i32")]
type WGPUTextureFormat = enum {
    Undefined = 0,
    R8Unorm = 1,
    R8Snorm = 2,
    R8Uint = 3,
    R8Sint = 4,
    R16Uint = 13,
    R16Sint = 14,
    R16Float = 15,
    RG8Unorm = 16,
    RGBA8Unorm = 1 << 4 | 5,
}

pub fn (format: WGPUTextureFormat) is_valid() bool {
    return match format {
        WGPUTextureFormat.Undefined => false,
        _ => true,
    }
}

pub fn main() i32 {
    var fmt = WGPUTextureFormat.R16Float;
    if fmt.is_valid() {
        return 0;
    }
    return 1;
}
