package main

import "math"

pub fn main() {
	var f32_val = f32(1.0)
	var f32_bits = math.Float32bits(f32_val)
	if f32_bits != u32(0x3f800000) {
		panic("f32 bits mismatch")
	}

	var f32_restored = math.Uint32bitsToFloat(f32_bits)
	if f32_restored != f32_val {
		panic("f32 restored mismatch")
	}

	var f64_val = f64(1.0)
	var f64_bits = math.Float64bits(f64_val)
	if f64_bits != u64(0x3ff0000000000000) {
		panic("f64 bits mismatch")
	}

	var f64_restored = math.Uint64bitsToFloat(f64_bits)
	if f64_restored != f64_val {
		panic("f64 restored mismatch")
	}
}
