package pkg_core_user

// Mirrors: gfx_viewer/viewer.nr which imports core "math"
import "math"
import "io"

// Uses core/math.Sqrt(f64) f64
pub fn RunCoreTest() {
    var a = math.Sqrt(4.0)
    io.PrintLn("core math.Sqrt(4.0) = ${a}")
}
