package math

[intrinsic("bitcast")]
pub fn Float32bits(f: f32) u32 {
	return u32(f)
}

[intrinsic("bitcast")]
pub fn Float64bits(f: f64) u64 {
	return u64(f)
}

[intrinsic("bitcast")]
pub fn Uint32bitsToFloat(u: u32) f32 {
	return f32(u)
}

[intrinsic("bitcast")]
pub fn Uint64bitsToFloat(u: u64) f64 {
	return f64(u)
}
