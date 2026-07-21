package pkg_local_user

// Mirrors: cone.nr / constraints which import "src/math/math" (also package math)
// The local math has a generic Sqrt[T] — previously collided with core/math.Sqrt(f64)
import "../local/math"
import "io"

// Uses local generic math.Sqrt[T](x: T) T
pub fn RunLocalTest() {
    var b = math.Sqrt[f64](9.0)
    io.PrintLn("local math.Sqrt[f64](9.0) = ${b}")
}
