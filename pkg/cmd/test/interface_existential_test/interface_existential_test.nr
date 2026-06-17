package main

import "io"

type Greeter = interface {
    fn greet() void
}

type Describer = interface {
    fn describe() void
}

type User = struct {
    name: str,
    age: i32
}

fn (self: #User) greet() {
    io.PrintLn("Hello, I am User: ${self.name}, age ${self.age}")
}

fn (self: #User) describe() {
    io.PrintLn("User profile: Name=${self.name}")
}

type Robot = struct {
    model: str,
    id: i32
}

fn (self: #Robot) greet() {
    io.PrintLn("Beep Boop, Robot model ${self.model} (ID: ${self.id})")
}

type Device = struct {
    serial: str
}

fn (self: #Device) describe() {
    io.PrintLn("Device Serial: ${self.serial}")
}

fn printAny(x: #any) {
    // Test safe binding of borrowed existential
    var _y: #any = x
    io.PrintLn("Exiting printAny with borrowed any")
}

fn printAnyValue(x: any) {
    // Test safe move / consumption of existential
    var _y: any = x
    io.PrintLn("Exiting printAnyValue with owned any")
}

fn main() i32 {
    io.PrintLn("--- Starting Complete Safe Existential Test ---")

    var u = User {
        name: "Alice",
        age: 30
    }

    var r = Robot {
        model: "X-100",
        id: 42
    }

    var d = Device {
        serial: "SN-999"
    }

    // 1. Concrete Struct to any (existential type erasure)
    io.PrintLn("1. Casting concrete to any...")
    var a1: any = #u
    var a2: any = #r
    var a3: any = #d

    printAny(#a1)
    printAny(#a2)
    printAny(#a3)

    // 2. Interface to any (interface-to-interface casting)
    io.PrintLn("2. Casting interface to any...")
    var g1: Greeter = #u
    var g2: Greeter = #r
    var desc: Describer = #d

    var a4: any = g1
    var a5: any = g2
    var a6: any = desc

    printAny(#a4)
    printAny(#a5)
    printAny(#a6)

    // 3. Heterogeneous list storing 'any'
    io.PrintLn("3. Constructing Heterogeneous List[any]...")
    var list = make(List[any])
    list = append(list, #u)
    list = append(list, #r)
    list = append(list, #d)
    list = append(list, a1)
    list = append(list, a2)
    list = append(list, a3)
    list = append(list, a4)
    list = append(list, a5)
    list = append(list, a6)

    io.PrintLn("List size after appends: ${len(list)}")

    // 4. Access and borrow from List[any]
    io.PrintLn("4. Accessing elements from List[any]...")
    var i = 0
    while i < len(list) {
        io.PrintLn("Item index ${i}:")
        printAny(#list[i])
        i++
    }

    // 5. Passing existential value by-move
    io.PrintLn("5. Moving existential by-value...")
    var moved_existential: any = list[0]
    printAnyValue(moved_existential)

    // 6. Retrieve value back from existential via type cast
    io.PrintLn("6. Retrieving concrete value from existential...")
    var moved_existential2: any = list[3]
    var u_back: #User = User(moved_existential2)
    io.PrintLn("Retrieved User name: ${u_back.name}, age ${u_back.age}")

    io.PrintLn("All safe existential (any) test scenarios completed successfully!")
    return 0
}
