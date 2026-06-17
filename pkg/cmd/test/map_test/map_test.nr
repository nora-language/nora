package main

import "io"

fn main() i32 {
    io.PrintLn("--- Map Test ---")

    // 1. Map Literal
    var m = {"apple": 1, "banana": 2, "cherry": 3}
    io.PrintLn("apple: ${m["apple"]}")
    io.PrintLn("banana: ${m["banana"]}")
    io.PrintLn("cherry: ${m["cherry"]}")

    // 2. Explicit Update
    m["apple"] = 10
    io.PrintLn("apple (updated): ${m["apple"]}")

    // 3. make(Map)
    var scores = make(Map[str, i32])
    scores["player1"] = 100
    scores["player2"] = 200
    io.PrintLn("player1: ${scores["player1"]}, player2: ${scores["player2"]}")

    // 4. Int keys
    var counts = {1: "one", 2: "two"}
    io.PrintLn("1: ${counts[1]}, 2: ${counts[2]}")

    // 5. make(List)
    var list = make(List[i32], 5)
    list[0] = 42
    io.PrintLn("list[0]: ${list[0]}")

    io.PrintLn("Map and List test completed.")
    return 0
}
