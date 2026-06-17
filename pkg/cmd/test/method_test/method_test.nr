package main

import "io"

type Point = struct {
    x: i32,
    y: i32
}

fn (p: Point) print() {
    io.PrintLn("Point(${p.x}, ${p.y})")
}

fn (p: &Point) move(dx: i32, dy: i32) {
    p.x = p.x + dx
    p.y = p.y + dy
}

fn main() i32 {
    var p = Point {
        x: 10,
        y: 20
    }
    p.print()

    p.move(5, -5)
    p.print()

    return 0
}
