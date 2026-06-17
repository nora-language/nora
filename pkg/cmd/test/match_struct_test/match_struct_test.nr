import "io"

type Point = struct {
    x: i32,
    y: i32
}

type Rect = struct {
    top_left: Point,
    bottom_right: Point
}

fn test_struct_match() {
    var p = Point {
        x: 10,
        y: 20
    }

    match p {
        Point {
            x: 10,
            y: val
        } => {
            io.PrintLn("x is 10, y is ${val}")
        }
        Point {
            x: x_val,
            y: y_val
        } => {
            io.PrintLn("Point at ${x_val}, ${y_val}")
        }
    }

    var p2 = Point {
        x: 0,
        y: 0
    }
    match p2 {
        Point {
            x: 0,
            y: 0
        } => {
            io.PrintLn("Origin")
        }
        Point {
            x: _,
            y: _
        } => {
            io.PrintLn("Not Origin")
        }
    }

    var p3 = Point {
        x: 5,
        y: 15
    }
    match p3 {
        Point {
            x: 10
        } => {
            io.PrintLn("x is 10")
        }
        Point {
            y: 15
        } => {
            io.PrintLn("y is 15")
        }
        _ => {
            io.PrintLn("unknown")
        }
    }
}

fn test_struct_return() {
    var p = Point {
        x: 5,
        y: 5
    }

    var result_rect = match p {
        Point {
            x: 0,
            y: 0
        } => {
            Rect {
                top_left: Point {
                    x: 0,
                    y: 0
                },
                bottom_right: Point {
                    x: 10,
                    y: 10
                }
            }
        }
        Point {
            x: val,
            y: _
        } => {
            Rect {
                top_left: Point {
                    x: val,
                    y: 0
                },
                bottom_right: Point {
                    x: val + 10,
                    y: 10
                }
            }
        }
    }

    var x1 = result_rect.top_left.x
    var x2 = result_rect.bottom_right.x
    io.PrintLn("Rect x bounds: ${x1} to ${x2}")
}

fn test_nested_struct_match() {
    var r = Rect {
        top_left: Point {
            x: 0,
            y: 0
        },
        bottom_right: Point {
            x: 10,
            y: 10
        }
    }

    match r {
        Rect {
            top_left: Point {
                x: 0,
                y: 0
            },
            bottom_right: br
        } => {
            var brx = br.x
            var bry = br.y
            io.PrintLn("Starts at origin, ends at ${brx}, ${bry}")
        }
        _ => {
            io.PrintLn("Other rect")
        }
    }
}

fn main() {
    test_struct_match()
    test_struct_return()
    test_nested_struct_match()
}
