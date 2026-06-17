import "std/io"

type Node = struct {
    val: i32,
    next: @Node
}

type Tree = enum {
    Empty,
    Leaf(val: i32),
    Branch(node: @Node)
}

fn test_match_sum_move() {
    var n1 = alloc Node {
        val: 10,
        next: none
    }
    var t = Branch(@n1)

    match @t {
        Branch(n) => {
            var nv = n.val
            io.PrintLn("Matched branch with node value ${nv}")
            // n is @Node, should be dropped at end of block
        }
        Leaf(v) => {
            io.PrintLn("Leaf ${v}")
        }
        Empty => {
            io.PrintLn("Empty")
        }
    }
    // t should be dead
}

fn main() {
    test_match_sum_move()
    io.PrintLn("Match sum move test passed!")
}
