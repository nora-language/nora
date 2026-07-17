package main

import "std/simd"
import "io"

[cfg("target_os=linux")]
pub type ConfiguredVector = struct {
    val: f64
    os_name: str
}

[cfg("target_os=linux")]
pub fn GetConfigured() ConfiguredVector {
    return ConfiguredVector { val: 42.0, os_name: "linux" }
}

[cfg("target_os=windows")]
pub type ConfiguredVector = struct {
    val: i32
    os_name: str
}

[cfg("target_os=windows")]
pub fn GetConfigured() ConfiguredVector {
    return ConfiguredVector { val: 100, os_name: "windows" }
}

fn main() i32 {
    // 1. Verify standard SIMD types can be imported and used without redefinition collisions
    var v0 = simd.VecSet4(1.0, 2.0, 3.0, 4.0)
    var v1 = simd.VecSet1(5.0)
    var v_add = v0 + v1
    var arr = alloc f64[4]
    simd.VecStore(simd.VecSlicePtr(arr, 0), v_add)

    // 2. Verify CFG filtered types and functions do not collide during symbol harvesting & C codegen
    var c = GetConfigured()
    io.PrintLn("Configured OS: ${c.os_name}")

    if (c.os_name != "linux" && c.os_name != "windows") {
        panic("CFG resolution failed!")
    }

    io.PrintLn("SIMD compiler collision verification passed!")
    return 0
}
