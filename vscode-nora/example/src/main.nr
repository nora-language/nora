package main

import "collections"
import "io"
import "iter"

type Person = struct {
    name: str,
    age: int
}

fn CreatePerson(name: str, age: int) @Person {
    return alloc Person {
        name: name,
        age: age
    }
}

fn main() i32 {
    var person = CreatePerson("Alice", 30)
    var lala = CreatePerson("Bob", 25)
    var lili = CreatePerson("Charlie", 35)
    var lolo = CreatePerson("David", 40)
    var lulu = CreatePerson("Eve", 28)

    // Using std/vector for a Vector of Person
    var people = collections.NewVector[Person](2)

    // We move (@) ownership of the structs into the Vector
    people.Push(@person)
    people.Push(@lala)
    people.Push(@lili)
    people.Push(@lolo)
    people.Push(@lulu)

    io.PrintLn("--- Vector Elements ---")

    // Iterate using a simple while loop
    var i = 0
    while i < people.Len() {
        var p = people.Get(i)
        io.PrintLn("Name: ${p.name}, Age: ${p.age}")
        i++
    }

    // You can also use the ForEach method!
    io.PrintLn("--- Using ForEach ---")
    people.ForEach(fn(p: #Person) {
            io.PrintLn("Name: ${p.name}, Age: ${p.age}")
        })
    io.PrintLn("--- Using for-in ---")

    for p in people.Iter() {
        io.PrintLn("Name: ${p.name}, Age: ${p.age}")
    }

    return 0
}
