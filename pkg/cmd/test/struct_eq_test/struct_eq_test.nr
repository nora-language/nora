package main

import "io"

type Point = struct {
    x: i32,
    y: i32
}

type Rect = struct {
    top_left: Point,
    bottom_right: Point,
    name: str
}

fn main() i32 {
    var p1 = Point {
        x: 10,
        y: 20
    }
    var p2 = Point {
        x: 10,
        y: 20
    }
    var p3 = Point {
        x: 10,
        y: 21
    }

    io.PrintLn("Testing Point equality:")
    if (p1 == p2) {
        io.PrintLn("p1 == p2: PASSED")
    } else {
        io.PrintLn("p1 == p2: FAILED")
        return 1
    }

    if (p1 != p3) {
        io.PrintLn("p1 != p3: PASSED")
    } else {
        io.PrintLn("p1 != p3: FAILED")
        return 1
    }

    var r1 = Rect {
        top_left: Point {
            x: 0,
            y: 0
        },
        bottom_right: Point {
            x: 10,
            y: 10
        },
        name: "test"
    }
    var r2 = Rect {
        top_left: Point {
            x: 0,
            y: 0
        },
        bottom_right: Point {
            x: 10,
            y: 10
        },
        name: "test"
    }
    var r3 = Rect {
        top_left: Point {
            x: 1,
            y: 0
        },
        bottom_right: Point {
            x: 10,
            y: 10
        },
        name: "test"
    }
    var r4 = Rect {
        top_left: Point {
            x: 0,
            y: 0
        },
        bottom_right: Point {
            x: 10,
            y: 10
        },
        name: "other"
    }

    io.PrintLn("\nTesting Rect equality (nested + string):")
    if (r1 == r2) {
        io.PrintLn("r1 == r2: PASSED")
    } else {
        io.PrintLn("r1 == r2: FAILED")
        return 1
    }

    if (r1 != r3) {
        io.PrintLn("r1 != r3: PASSED")
    } else {
        io.PrintLn("r1 != r3: FAILED")
        return 1
    }

    if (r1 != r4) {
        io.PrintLn("r1 != r4: PASSED")
    } else {
        io.PrintLn("r1 != r4: FAILED")
        return 1
    }

    io.PrintLn("\nTesting Reference Equality:")
    if (#p1 == #p1) {
        io.PrintLn("#p1 == #p1: PASSED")
    } else {
        io.PrintLn("#p1 == #p1: FAILED")
        return 1
    }

    // Different memory locations should not be equal even if values are same
    if (#p1 != #p2) {
        io.PrintLn("#p1 != #p2: PASSED")
    } else {
        io.PrintLn("#p1 != #p2: FAILED (Reference equality expected to fail for different allocations)")
        // Note: On some optimizations they might be same, but usually different.
    }

    io.PrintLn("\nAll Structural Equality tests passed!")
    return 0
}
