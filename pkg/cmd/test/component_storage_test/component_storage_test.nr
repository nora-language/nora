package main

import "io"

// -------------------------------------------------------------
// Definitions for both Option A and Option B
// -------------------------------------------------------------
type Component = interface {
    fn Value() i32
    fn Name() str
}

type Position = struct {
    x: i32,
    y: i32
}

fn (self: #Position) Value() i32 {
    return self.x + self.y
}

fn (self: #Position) Name() str {
    return "Position"
}

type Velocity = struct {
    dx: i32,
    dy: i32
}

fn (self: #Velocity) Value() i32 {
    return self.dx * self.dy
}

fn (self: #Velocity) Name() str {
    return "Velocity"
}

fn main() i32 {
    io.PrintLn("--- Starting Safe Heterogeneous Component Storage Test ---")

    var pos = Position {
        x: 15,
        y: 25
    }
    var vel = Velocity {
        dx: 3,
        dy: 4
    }

    // =============================================================
    // TEST OPTION A: Downcasting/Type Assertions on "any"
    // =============================================================
    io.PrintLn("[Test Option A] Testing casting from any to concrete type...")

    var pos_any: any = #pos
    var vel_any: any = #vel

    // Retrieve concrete structs back from "any" using type cast syntax
    var pos_back: #Position = Position(pos_any)
    var vel_back: #Velocity = Velocity(vel_any)

    io.PrintLn("Retrieved Position back: x=${pos_back.x}, y=${pos_back.y}")
    io.PrintLn("Retrieved Velocity back: dx=${vel_back.dx}, dy=${vel_back.dy}")

    if (pos_back.x != 15 || pos_back.y != 25) {
        panic("Option A: Position retrieval verification failed!")
    }
    if (vel_back.dx != 3 || vel_back.dy != 4) {
        panic("Option A: Velocity retrieval verification failed!")
    }
    io.PrintLn("[Test Option A] Passed successfully!")

    // =============================================================
    // TEST OPTION B: Component-Specific Protocols (Dynamic Dispatch)
    // =============================================================
    io.PrintLn("[Test Option B] Testing storage using Component protocol...")

    var comp_list = make(List[Component])
    comp_list = append(comp_list, #pos)
    comp_list = append(comp_list, #vel)

    io.PrintLn("Heterogeneous Component List length: ${len(comp_list)}")

    var i = 0
    while i < len(comp_list) {
        var comp = comp_list[i]
        io.PrintLn("Component ${i}: Name=${comp.Name()}, Value=${comp.Value()}")
        i = i + 1
    }

    // Verify first element is Position and has correct value (15 + 25 = 40)
    var first = comp_list[0]
    if (first.Value() != 40) {
        panic("Option B: Position value verification failed!")
    }

    // Verify second element is Velocity and has correct value (3 * 4 = 12)
    var second = comp_list[1]
    if (second.Value() != 12) {
        panic("Option B: Velocity value verification failed!")
    }

    io.PrintLn("[Test Option B] Passed successfully!")
    io.PrintLn("All Safe Component Storage Tests passed successfully!")
    return 0
}
