package main

import "std/io"

type Person = struct {
    name: str,
    age: i32
}

fn CreatePerson(name: str, age: i32) @Person {
    return alloc Person {
        name: name,
        age: age
    }
}

fn main() i32 {
    var person = CreatePerson("Alice", 30)
    io.PrintLn("Name: ${person.name}")
    io.PrintLn("Age: ${person.age}")

    CreatePerson("Bob", 25)
    return 0
}
