import "std/io"

type Node[T] = struct {
    val: T,
    next: @Node[T]
}

fn test_match_generic() {
    var n1 = alloc Node[i32] {
        val: 42,
        next: none
    }
    var n2 = alloc Node[i32] {
        val: 100,
        next: @n1
    }

    match @n2 {
        Node {
            val: v,
            next: n
        } => {
            io.PrintLn("Matched generic node with value ${v}")
            // v is i32, n is @Node[i32]
            match @n {
                Node {
                    val: v2,
                    next: n2
                } => {
                    io.PrintLn("Matched inner node with value ${v2}")
                }
                _ => {
                    io.PrintLn("No inner node")
                }
            }
        }
    }
}

fn main() {
    test_match_generic()
    io.PrintLn("Match generic test passed!")
}
