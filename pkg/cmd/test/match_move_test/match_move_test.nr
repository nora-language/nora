import "std/io"

type Node = struct {
    val: i32,
    next: @Node
}

fn test_match_move() {
    var n1 = alloc Node {
        val: 1,
        next: none
    }
    var n2 = alloc Node {
        val: 2,
        next: @n1
    }

    match @n2 {
        Node {
            val: v,
            next: n
        } => {
            io.PrintLn("Matched node with value ${v}")
            // n is @Node, should be dropped here or moved
            if (n != none) {
                var next_val = n.val
                io.PrintLn("Next node value: ${next_val}")
            }
        }
    }
    // n2 should be dead here
}

fn main() {
    test_match_move()
    io.PrintLn("Match move test passed!")
}
