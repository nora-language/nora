package main

import "io"
import "collections"
import "hash"
fn print_option_str(opt: collections.Option[#str]) {
    match opt {
        Some(val) => {
            io.Print("  Some('")
            io.Print(val)
            io.PrintLn("')")
        }
        None => {
            io.PrintLn("  None")
        }
    }
}

fn print_option_i32(opt: collections.Option[#i32]) {
    match opt {
        Some(val) => {
            io.Print("  Some(")
            io.Print(val)
            io.PrintLn(")")
        }
        None => {
            io.PrintLn("  None")
        }
    }
}

fn main() i32 {
    io.PrintLn("--- Collections HashMap Test ---")

    // 1. Initialize HashMap[i32, str] with capacity 4 to test dynamic rehashing
    var m = collections.NewHashMap[i32, str](4)
    io.Print("Initial Len: ")
    io.PrintLn(m.Len())

    io.PrintLn("Inserting 1 -> 'one'...")
    m.Insert(1, "one")
    io.PrintLn("Inserting 2 -> 'two'...")
    m.Insert(2, "two")

    io.Print("Len after 2 inserts: ")
    io.PrintLn(m.Len())
    io.PrintLn("Get(1):")
    var g1 = m.Get(1)
    print_option_str(g1)
    io.PrintLn("Get(2):")
    var g2 = m.Get(2)
    print_option_str(g2)

    // 2. Trigger Rehash (3rd insertion exceeds 70% threshold of 4 capacity)
    io.PrintLn("Inserting 3 -> 'three' (triggering rehash)...")
    m.Insert(3, "three")
    io.Print("Len after rehash insert: ")
    io.PrintLn(m.Len())
    io.PrintLn("Get(1):")
    var g3 = m.Get(1)
    print_option_str(g3)
    io.PrintLn("Get(2):")
    var g4 = m.Get(2)
    print_option_str(g4)
    io.PrintLn("Get(3):")
    var g5 = m.Get(3)
    print_option_str(g5)

    // 3. Test Replacement
    io.PrintLn("Replacing 2 with 'double-two'...")
    m.Insert(2, "double-two")
    io.PrintLn("Get(2) after replacement:")
    var g6 = m.Get(2)
    print_option_str(g6)

    // 4. Test Contains
    io.Print("Contains(3): ")
    io.PrintLn(m.Contains(3))
    io.Print("Contains(4): ")
    io.PrintLn(m.Contains(4))

    // 5. Test Removal
    io.PrintLn("Removing 1...")
    var removed = m.Remove(1)
    io.PrintLn("Removed value:")
    match removed {
        Some(val) => {
            io.Print("  Some('")
            io.Print(val)
            io.PrintLn("')")
        }
        None => {
            io.PrintLn("  None")
        }
    }
    io.Print("Len after removal: ")
    io.PrintLn(m.Len())
    io.PrintLn("Get(1) after removal:")
    var g7 = m.Get(1)
    print_option_str(g7)

    io.PrintLn("")
    io.PrintLn("--- String HashMap Test ---")

    // 6. Initialize HashMap[str, i32]
    var sm = collections.NewHashMap[str, i32](8)
    io.PrintLn("Inserting 'hello' -> 100...")
    sm.Insert("hello", 100)
    io.PrintLn("Inserting 'nora' -> 200...")
    sm.Insert("nora", 200)

    io.Print("Len: ")
    io.PrintLn(sm.Len())
    io.PrintLn("Get('hello'):")
    var sg1 = sm.Get("hello")
    print_option_i32(sg1)
    io.PrintLn("Get('nora'):")
    var sg2 = sm.Get("nora")
    print_option_i32(sg2)
    io.PrintLn("Get('other'):")
    var sg3 = sm.Get("other")
    print_option_i32(sg3)

    io.PrintLn("Collections HashMap tests finished successfully!")
    return 0
}
