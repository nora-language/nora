import "std/io"

type Node[T] = struct {
    val: T
    next: @Node[T]
}

fn main() {
    var n1 = alloc Node[i32] { val: 42, next: none }
    var n2 = alloc Node[i32] { val: 100, next: @n1 }

    var curr = @n2
    while (curr != none) {
        match @curr {
            Node { val: v, next: n } => {
                io.PrintLn("Processing node $v")
                curr = @n
            }
        }
    }
    io.PrintLn("Done")
}
