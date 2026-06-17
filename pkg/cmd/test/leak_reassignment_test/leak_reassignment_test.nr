package main

import "io"

var dropped_first = false
var dropped_original = false
var dropped_arr_original = false

type Element = struct {
    name: str
}

pub fn (self: &Element) drop() {
    if (self.name == "first") {
        dropped_first = true
    } else if (self.name == "original") {
        dropped_original = true
    } else if (self.name == "arr_original") {
        dropped_arr_original = true
    }
}

type Container = struct {
    el: Element
}

fn main() {
    io.PrintLn("=== Starting Leak Reassignment Test ===")

    // 1. Local variable reassignment
    var el = Element {
        name: "first"
    }
    el = Element { // Should trigger drop on the first one!
        name: "second"
    }

    if (!dropped_first) {
        panic("Error: Expected 'first' to be dropped on local reassignment")
    }

    // 2. Struct field reassignment
    var container = Container {
        el: Element {
            name: "original"
        }
    }
    container.el = Element { // Should trigger drop on original!
        name: "new"
    }

    if (!dropped_original) {
        panic("Error: Expected 'original' to be dropped on field reassignment")
    }

    // 3. List element reassignment
    var arr = make(List[Element], 5)
    arr[0] = Element {
        name: "arr_original"
    }
    arr[0] = Element { // Should trigger drop on arr_original!
        name: "arr_new"
    }

    if (!dropped_arr_original) {
        panic("Error: Expected 'arr_original' to be dropped on array element reassignment")
    }

    io.PrintLn("=== All Leak Reassignment Tests Passed! ===")
}
