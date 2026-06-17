/// Point represents a 2D coordinate
type Point = struct {
    /// X coordinate
    x: i32
    /// Y coordinate
    y: i32
}

/// Adds two integers
fn add(a: i32, b: i32) i32 {
    return a + b
}

fn main() {
    var p = Point{x: 1, y: 2}
    var sum = add(p.x, p.y)
}
