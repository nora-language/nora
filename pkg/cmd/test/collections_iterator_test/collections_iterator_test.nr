package main

import "collections"
import "io"

fn Sum[T: collections.Iterator[i32]](it: T) i32 {
    var sum = 0
    while (true) {
        match it.Next() {
            Some(val) => {
                sum = sum + val
            }
            None => {
                break;
            }
        }
    }
    return sum
}

fn main() i32 {
    io.PrintLn("--- Collections Iterator Test ---")

    // 1. Test Vector Iterator
    var vec = collections.NewVector[i32](4)
    vec.Push(10)
    vec.Push(20)
    vec.Push(30)

    io.PrintLn("Direct Vector Iteration:")
    var v_iter = vec.Iter()
    while (true) {
        match v_iter.Next() {
            Some(val) => {
                io.Print("  Value: ")
                io.PrintLn(val)
            }
            None => {
                break;
            }
        }
    }

    // 2. Test LinkedList Iterator
    var list = collections.NewLinkedList[i32]()
    list.PushBack(100)
    list.PushBack(200)
    list.PushBack(300)

    io.PrintLn("Direct LinkedList Iteration:")
    var l_iter = list.Iter()
    while (true) {
        match l_iter.Next() {
            Some(val) => {
                io.Print("  Value: ")
                io.PrintLn(val)
            }
            None => {
                break;
            }
        }
    }

    // 3. Test Generic Polymorphic Interface Constraint (Sum)
    var v_iter2 = vec.Iter()
    var v_sum = Sum(v_iter2)
    io.Print("Vector sum: ")
    io.PrintLn(v_sum)

    var l_iter2 = list.Iter()
    var l_sum = Sum(l_iter2)
    io.Print("LinkedList sum: ")
    io.PrintLn(l_sum)

    io.PrintLn("Iterator tests completed successfully!")
    return 0
}
